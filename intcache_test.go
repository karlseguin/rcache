package rcache

import (
	. "github.com/karlseguin/expect"
	"strconv"
	"testing"
	"time"
)

type IntCacheTest struct {
	fetchCount int
}

func Test_IntCache(t *testing.T) {
	Expectify(new(IntCacheTest), t)
}

func (ct IntCacheTest) FetchesAnUncachedItem() {
	c := NewInt(ct.DumbFetcher, time.Minute)
	Expect(c.Get(1)).To.Equal("1-fetch-1")
}

func (ct IntCacheTest) ReturnsACachedItem() {
	c := NewInt(ct.DumbFetcher, time.Minute)
	c.Get(1)
	Expect(c.Get(1)).To.Equal("1-fetch-1")
}

func (ct IntCacheTest) FetchesAnExpiredItem() {
	c := NewInt(ct.DumbFetcher, time.Minute)
	c.Get(2)
	c.items[2].expires = time.Now().Add(-GRACE_LIMIT) //sue me
	Expect(c.Get(2)).To.Equal("2-fetch-2")
}

//sue me * 2
func (ct IntCacheTest) GracesAnItem() {
	c := NewInt(ct.DumbFetcher, time.Minute)
	c.Get(3)
	c.items[3].expires = time.Now().Add(-GRACE_LIMIT).Add(time.Second)
	Expect(c.Get(3)).To.Equal("3-fetch-1")
	time.Sleep(time.Millisecond)
	Expect(c.items[3].value.(string)).To.Equal("3-fetch-2")
}

//sue me * 2
func (ct IntCacheTest) ReapsExpiredItem() {
	REAPER_FREQUENCY = time.Millisecond
	defer func() { REAPER_FREQUENCY = time.Minute }()

	c := NewInt(ct.DumbFetcher, time.Minute)
	c.Get(1)
	c.Get(2)
	c.Get(3)
	c.Get(4)
	c.items[1].expires = time.Now().Add(-GRACE_LIMIT)
	c.items[3].expires = time.Now().Add(-GRACE_LIMIT)
	time.Sleep(time.Millisecond * 5)
	Expect(len(c.items)).To.Equal(2)
	Expect(c.Get(1)).To.Equal("1-fetch-5")
	Expect(c.Get(2)).To.Equal("2-fetch-2")
	Expect(c.Get(3)).To.Equal("3-fetch-6")
	Expect(c.Get(4)).To.Equal("4-fetch-4")
}

func (ct *IntCacheTest) DumbFetcher(key int) interface{} {
	ct.fetchCount += 1
	return strconv.Itoa(key) + "-fetch-" + strconv.Itoa(ct.fetchCount)
}

func (ct *IntCacheTest) Each(test func()) {
	ct.fetchCount = 0
	test()
}