package rcache

import (
	"time"
)

type state int

const (
	ok = iota
	stale
	expired
)

type Item struct {
	expires time.Time
	value   interface{}
}

func (i *Item) State() state {
	d := time.Now().Sub(i.expires)
	if d < 0 {
		return ok
	}
	if d > GRACE_LIMIT {
		return expired
	}
	return stale
}
