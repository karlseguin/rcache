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

## Methods

- `Get(key string)` - Returns the value or nil. Either gets it from the cache or fetches it via the fetcher
- `Set(key string, value interface{})` - Sets the value
- `Replace(key string, value interface{})` - Replaces an existing value, extending it's TTL. Noop if key isn't already in the cache
- `Delete(key string)` - Deletes an item
- `Clear()` - removes all items from the cache

## Integer Keys

By default, the cache key is a string. You can create a cache that uses an integer key by using `NewInt` (instead of `New`). Everything works the same, except keys are `int`.

## LRU Cache Alternative

For a more powerful LRU cache, checkout [ccache](https://github.com/karlseguin/ccache)
