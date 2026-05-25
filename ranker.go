package main

import "container/heap"

type Entry struct {
	Query string `json:"query"`
	Count int32  `json:"count"`
}

type minHeap []Entry

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].Count < h[j].Count }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) { *h = append(*h, x.(Entry)) }
func (h *minHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func SelectTopQueries(counts map[string]int32, n int) []Entry {
	if n <= 0 {
		return nil
	}
	h := new(minHeap)
	heap.Init(h)

	for q, c := range counts {
		switch {
		case h.Len() < n:
			heap.Push(h, Entry{q, c})
		case c > (*h)[0].Count:
			(*h)[0] = Entry{q, c}
			heap.Fix(h, 0)
		}
	}

	res := make([]Entry, h.Len())
	for i := len(res) - 1; i >= 0; i-- {
		res[i] = heap.Pop(h).(Entry)
	}
	return res
}
