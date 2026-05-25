package main

import (
	"sync/atomic"
	"time"
)

type BlocklistProvider interface {
	Snapshot() map[string]struct{}
}

type TrendingCache struct {
	ptr       atomic.Pointer[[]Entry]
	window    *TrendingWindow
	blocklist BlocklistProvider
	n         int
	stop      chan struct{}
}

func NewTopCache(w *TrendingWindow, bl BlocklistProvider, n int, refresh time.Duration) *TrendingCache {
	c := &TrendingCache{
		window:    w,
		blocklist: bl,
		n:         n,
		stop:      make(chan struct{}),
	}
	c.updateTop()
	go c.loop(refresh)
	return c
}

func (c *TrendingCache) Get() []Entry {
	if p := c.ptr.Load(); p != nil {
		return *p
	}
	return nil
}

func (c *TrendingCache) loop(d time.Duration) {
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.updateTop()
		case <-c.stop:
			return
		}
	}
}

func (c *TrendingCache) updateTop() {
	blocked := c.blocklist.Snapshot()
	counts := c.window.Aggregate(blocked)
	result := SelectTopQueries(counts, c.n)
	c.ptr.Store(&result)
}

func (c *TrendingCache) Close() { close(c.stop) }
