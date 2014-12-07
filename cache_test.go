package rcache

import (
	. "github.com/karlseguin/expect"
	"strconv"
	"testing"
	"time"
)

type CacheTest struct {
	fetchCount int
}

func Test_Cache(t *testing.T) {
	Expectify(new(CacheTest), t)
}

func (ct CacheTest) FetchesAnUncachedItem() {
	c := New(ct.DumbFetcher, time.Minute)
	Expect(c.Get("spice")).To.Equal("spice-fetch-1")
}

func (ct CacheTest) ReturnsACachedItem() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Get("spice")
	Expect(c.Get("spice")).To.Equal("spice-fetch-1")
}

func (ct CacheTest) DeletesAnItem() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Get("spice")
	c.Delete("spice")
	Expect(c.Get("spice")).To.Equal("spice-fetch-2")
}

func (ct CacheTest) ClearsTheCache() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Get("spice1")
	c.Get("spice2")
	c.Clear()
	Expect(c.Get("spice1")).To.Equal("spice1-fetch-3")
	Expect(c.Get("spice2")).To.Equal("spice2-fetch-4")
}

func (ct CacheTest) SetsAnItem() {
	c := New(nil, time.Minute)
	c.Set("power", 9001)
	c.Get("power")
	Expect(c.Get("power")).To.Equal(9001)
}

func (ct CacheTest) FetchesAnExpiredItem() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Get("power9000!")
	c.items["power9000!"].expires = time.Now().Add(-GRACE_LIMIT) //sue me
	Expect(c.Get("power9000!")).To.Equal("power9000!-fetch-2")
}

//sue me * 2
func (ct CacheTest) GracesAnItem() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Get("power9000!")
	c.items["power9000!"].expires = time.Now().Add(-GRACE_LIMIT).Add(time.Second)
	Expect(c.Get("power9000!")).To.Equal("power9000!-fetch-1")
	time.Sleep(time.Millisecond)
	Expect(c.items["power9000!"].value.(string)).To.Equal("power9000!-fetch-2")
}

//sue me * 2
func (ct CacheTest) ReapsExpiredItem() {
	REAPER_FREQUENCY = time.Millisecond
	defer func() { REAPER_FREQUENCY = time.Minute }()

	c := New(ct.DumbFetcher, time.Minute)
	c.Get("a")
	c.Get("b")
	c.Get("c")
	c.Get("d")
	c.items["a"].expires = time.Now().Add(-GRACE_LIMIT)
	c.items["c"].expires = time.Now().Add(-GRACE_LIMIT)
	time.Sleep(time.Millisecond * 5)
	Expect(len(c.items)).To.Equal(2)
	Expect(c.Get("a")).To.Equal("a-fetch-5")
	Expect(c.Get("b")).To.Equal("b-fetch-2")
	Expect(c.Get("c")).To.Equal("c-fetch-6")
	Expect(c.Get("d")).To.Equal("d-fetch-4")
}

func (ct *CacheTest) ReplaceNoopsOnNonExist() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Replace("a", "value")
	Expect(len(c.items)).To.Equal(0)
}

func (ct *CacheTest) ReplacesAValue() {
	c := New(ct.DumbFetcher, time.Minute)
	c.Get("a")
	c.Replace("a", "b")
	Expect(c.Get("a")).To.Equal("b")
}

func (ct *CacheTest) DumbFetcher(key string) interface{} {
	ct.fetchCount += 1
	return key + "-fetch-" + strconv.Itoa(ct.fetchCount)
}

func (ct *CacheTest) Each(test func()) {
	ct.fetchCount = 0
	test()
}
