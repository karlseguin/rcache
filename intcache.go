package rcache

import (
	"sync"
	"time"
)

type IntFetcher func(key int) interface{}

type IntCache struct {
	sync.RWMutex
	fetcher      IntFetcher
	ttl          time.Duration
	items        map[int]*Item
	fetchingLock sync.Mutex
	fetchings    map[int]time.Time
}

func NewInt(fetcher IntFetcher, ttl time.Duration) *IntCache {
	c := &IntCache{
		ttl:       ttl,
		fetcher:   fetcher,
		items:     make(map[int]*Item),
		fetchings: make(map[int]time.Time),
	}
	go c.reaper()
	return c
}

func (c *IntCache) Get(key int) interface{} {
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

func (c *IntCache) Replace(key int, value interface{}) {
	c.RLock()
	_, exists := c.items[key]
	c.RUnlock()
	if exists == false {
		return
	}
	c.Set(key, value)
}

func (c *IntCache) Delete(key int) {
	c.Lock()
	delete(c.items, key)
	c.Unlock()
}

func (c *IntCache) Clear() {
	c.Lock()
	c.items = make(map[int]*Item)
	c.Unlock()
}

func (c *IntCache) fetch(key int) interface{} {
	value := c.fetcher(key)
	if value == nil {
		return nil
	}
	c.Set(key, value)
	return value
}

func (c *IntCache) cfetch(key int) {
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

func (c *IntCache) reaper() {
	scratch := make([]int, REAPER_LIMIT)
	for {
		time.Sleep(REAPER_FREQUENCY)
		c.reap(scratch)
	}
}

func (c *IntCache) Set(key int, value interface{}) {
	item := &Item{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	c.Lock()
	c.items[key] = item
	c.Unlock()
}

func (c *IntCache) reap(scratch []int) {
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
