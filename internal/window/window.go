package window

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Royal17x/search-top/internal/metrics"
)

const (
	bucketCount    = 30
	bucketInterval = 10 * time.Second
)

type bucket struct {
	mu     sync.Mutex
	counts map[string]int32
}

func newBucket() *bucket {
	return &bucket{counts: make(map[string]int32)}
}

func (b *bucket) inc(q string) {
	b.mu.Lock()
	b.counts[q]++
	b.mu.Unlock()
}

func (b *bucket) emptyAndGet() map[string]int32 {
	b.mu.Lock()
	m := b.counts
	b.counts = make(map[string]int32)
	b.mu.Unlock()
	return m
}

func (b *bucket) snapshot() map[string]int32 {
	b.mu.Lock()
	cp := make(map[string]int32, len(b.counts))
	for k, v := range b.counts {
		cp[k] = v
	}
	b.mu.Unlock()
	return cp
}

type TrendingWindow struct {
	buckets [bucketCount]*bucket
	current atomic.Int32
	stopCh  chan struct{}
}

func NewTrendingWindow() *TrendingWindow {
	w := &TrendingWindow{stopCh: make(chan struct{})}
	for i := range w.buckets {
		w.buckets[i] = newBucket()
	}
	go w.tick()
	return w
}

func (w *TrendingWindow) Record(query string) {
	w.buckets[w.current.Load()%bucketCount].inc(query)
}

func (w *TrendingWindow) tick() {
	t := time.NewTicker(bucketInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			next := (w.current.Load() + 1) % bucketCount
			w.buckets[next].emptyAndGet()
			w.current.Add(1)
		case <-w.stopCh:
			return
		}
	}
}

func (w *TrendingWindow) Aggregate(blocked map[string]struct{}) map[string]int32 {
	start := time.Now()
	defer func() {
		metrics.AggregationDuration.Observe(time.Since(start).Seconds())
	}()

	totals := make(map[string]int32)
	for _, b := range w.buckets {
		for q, c := range b.snapshot() {
			if _, ok := blocked[q]; ok {
				continue
			}
			totals[q] += c
		}
	}
	return totals
}

func (w *TrendingWindow) Close() { close(w.stopCh) }
