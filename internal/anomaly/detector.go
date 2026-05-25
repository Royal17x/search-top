package anomaly

import (
	"sync"
	"time"
)

type Detector struct {
	mu             sync.Mutex
	counts         map[string]int
	expiresAt      time.Time
	windowDuration time.Duration
	maxAllowed     int
}

func New(windowDuration time.Duration, maxAllowed int) *Detector {
	return &Detector{
		counts:         make(map[string]int),
		expiresAt:      time.Now().Add(windowDuration),
		windowDuration: windowDuration,
		maxAllowed:     maxAllowed,
	}
}

func (d *Detector) IsSus(userID, query string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if time.Now().After(d.expiresAt) {
		d.counts = make(map[string]int)
		d.expiresAt = time.Now().Add(d.windowDuration)
	}

	key := userID + "\x00" + query
	d.counts[key]++
	return d.counts[key] > d.maxAllowed
}
