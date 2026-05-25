package stoplist

import (
	"strings"
	"sync"
)

type StopList struct {
	mu    sync.RWMutex
	words map[string]struct{}
}

func NewStopList() *StopList {
	return &StopList{words: make(map[string]struct{})}
}

func (sl *StopList) Add(word string) {
	w := cleanWord(word)
	if w == "" {
		return
	}
	sl.mu.Lock()
	sl.words[w] = struct{}{}
	sl.mu.Unlock()
}

func (sl *StopList) Remove(word string) {
	sl.mu.Lock()
	delete(sl.words, cleanWord(word))
	sl.mu.Unlock()
}

func (sl *StopList) Snapshot() map[string]struct{} {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	cp := make(map[string]struct{}, len(sl.words))
	for w := range sl.words {
		cp[w] = struct{}{}
	}
	return cp
}

func (sl *StopList) AllWords() []string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	out := make([]string, 0, len(sl.words))
	for w := range sl.words {
		out = append(out, w)
	}
	return out
}

func cleanWord(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
