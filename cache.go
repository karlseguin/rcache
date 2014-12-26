package rcache

import (
	"sync"
	"time"
)

var (
	GRACE_LIMIT      = time.Second * 20
	FETCH_TIME_LIMIT = time.Second * 5
	REAPER_FREQUENCY = time.Minute * 5
	REAPER_LIMIT     = 1000
)

type Fetcher func(key string) interface{}

type Cache struct {
	sync.RWMutex
	fetcher      Fetcher
	ttl          time.Duration
	items        map[string]*Item
	fetchingLock sync.Mutex
	fetchings    map[string]time.Time
}

func New(fetcher Fetcher, ttl time.Duration) *Cache {
	c := &Cache{
		ttl:       ttl,
		fetcher:   fetcher,
		items:     make(map[string]*Item),
		fetchings: make(map[string]time.Time),
	}
	go c.reaper()
	return c
}

func (c *Cache) Get(key string) interface{} {
	c.RLock()
	item, exists := c.items[key]
	c.RUnlock()
	if exists == false {
		return c.fetch(key)
	}
	state := item.State()
	if state == expired {
		return c.fetch(key)
	}
	if state == stale {
		go c.cfetch(key)
	}
	return item.value
}

func (c *Cache) Replace(key string, value interface{}) {
	c.RLock()
	_, exists := c.items[key]
	c.RUnlock()
	if exists == false {
		return
	}
	c.Set(key, value)
}

func (c *Cache) Delete(key string) {
	c.Lock()
	delete(c.items, key)
	c.Unlock()
}

func (c *Cache) Clear() {
	c.Lock()
	c.items = make(map[string]*Item)
	c.Unlock()
}

func (c *Cache) fetch(key string) interface{} {
	value := c.fetcher(key)
	if value == nil {
		return nil
	}
	c.Set(key, value)
	return value
}

func (c *Cache) cfetch(key string) {
	now := time.Now()
	c.fetchingLock.Lock()
	start, exists := c.fetchings[key]
	if exists && start.Add(FETCH_TIME_LIMIT).Before(now) {
		c.fetchingLock.Unlock()
		return
	}
	c.fetchings[key] = now
	c.fetchingLock.Unlock()

	value := c.fetcher(key)
	c.fetchingLock.Lock()
	delete(c.fetchings, key)
	c.fetchingLock.Unlock()
	if value != nil {
		c.Set(key, value)
	}
}

func (c *Cache) Set(key string, value interface{}) {
	item := &Item{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	c.Lock()
	c.items[key] = item
	c.Unlock()
}

func (c *Cache) reaper() {
	scratch := make([]string, REAPER_LIMIT)
	for {
		time.Sleep(REAPER_FREQUENCY)
		c.reap(scratch)
	}
}

func (c *Cache) reap(scratch []string) {
	count, victims := 0, 0
	c.RLock()
	for key, item := range c.items {
		if item.State() == expired {
			scratch[victims] = key
			victims++
		}
		count++
		if count == REAPER_LIMIT {
			break
		}
	}
	c.RUnlock()
	if victims == 0 {
		return
	}
	c.Lock()
	defer c.Unlock()
	for i := 0; i < victims; i++ {
		delete(c.items, scratch[i])
	}
}
