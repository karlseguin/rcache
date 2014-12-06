# rcache

rcache is a simple cache that periodically erases stale items and which can help mitigate the thundering herd problem. rcache's growth is unbound. It is meant to cache a small number of items.


## Usage:

```go
// thread-safe
apps := rcache.New(fetcher, time.Minute * 2)

func fetcher(key string) interface{} {
  // HIT YOUR DB, GET THE APP
  // ...
  return theApp
}

// in code:
app := apps.Get("spice").(*Application)
```

There's a short 20 second grace window in which an expired item will be returned. In other words, the real TTL of items placed in the above cache is 120-140seconds. Even if multiple goroutines concurrently GET an item which is in this grace window, only 1 call to fetcher will be executed.