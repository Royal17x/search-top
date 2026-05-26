package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

var queries = []string{
	"кроссовки", "куртка", "платье", "телефон", "ноутбук",
	"сумка", "часы", "наушники", "кофе", "книга",
	"кроссовки nike", "куртка зимняя", "платье вечернее",
	"iphone", "samsung", "airpods", "macbook", "xbox",
}

func main() {
	const (
		kafkaEvents   = 5000
		apiRequests   = 500
		goroutines    = 20
		anomalyUserID = "bot-001"
	)

	fmt.Println("search-top load test")
	fmt.Printf("kafka events: %d | api requests: %d | goroutines: %d\n\n",
		kafkaEvents, apiRequests, goroutines)

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9094"},
		Topic:   "search-events",
	})
	defer w.Close()

	msgs := make([]kafka.Message, 0, kafkaEvents)
	now := time.Now().UTC()

	for i := 0; i < kafkaEvents; i++ {
		query := queries[rand.Intn(len(queries))]
		userID := fmt.Sprintf("u%d", rand.Intn(200))

		if i%50 == 0 {
			userID = anomalyUserID
			query = "кроссовки"
		}

		body, _ := json.Marshal(map[string]any{
			"query":     query,
			"user_id":   userID,
			"timestamp": now.Format(time.RFC3339),
		})
		msgs = append(msgs, kafka.Message{Value: body})
	}

	start := time.Now()
	if err := w.WriteMessages(context.Background(), msgs...); err != nil {
		fmt.Println("kafka error:", err)
		return
	}
	fmt.Printf("kafka: %d events sent in %s\n", kafkaEvents, time.Since(start).Round(time.Millisecond))

	fmt.Println("  waiting 3s for consumer...")
	time.Sleep(3 * time.Second)

	var (
		success  atomic.Int64
		failed   atomic.Int64
		totalDur atomic.Int64
	)

	var wg sync.WaitGroup
	sem := make(chan struct{}, goroutines)
	apiStart := time.Now()

	for i := 0; i < apiRequests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			t := time.Now()
			resp, err := http.Get("http://localhost:8080/api/v1/top?n=10")
			dur := time.Since(t)

			if err != nil || resp.StatusCode != 200 {
				failed.Add(1)
				return
			}
			resp.Body.Close()
			success.Add(1)
			totalDur.Add(dur.Microseconds())
		}()
	}
	wg.Wait()

	total := success.Load() + failed.Load()
	avgUs := int64(0)
	if success.Load() > 0 {
		avgUs = totalDur.Load() / success.Load()
	}

	fmt.Printf("api:   %d/%d requests OK in %s | avg latency: %dµs\n",
		success.Load(), total,
		time.Since(apiStart).Round(time.Millisecond),
		avgUs,
	)

	fmt.Println("\n top-5 after load")
	resp, err := http.Get("http://localhost:8080/api/v1/top?n=5")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Query string `json:"query"`
			Count int32  `json:"count"`
		} `json:"items"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	for i, item := range result.Items {
		fmt.Printf("  #%d  %-20s %d\n", i+1, item.Query, item.Count)
	}

	fmt.Println("\n metrics")
	metricsResp, _ := http.Get("http://localhost:8080/metrics")
	if metricsResp != nil {
		defer metricsResp.Body.Close()
		buf := make([]byte, 4096)
		n, _ := metricsResp.Body.Read(buf)
		for _, line := range splitLines(string(buf[:n])) {
			if len(line) > 0 && line[0] != '#' &&
				(contains(line, "search_events") || contains(line, "search_top_requests")) {
				fmt.Println(" ", line)
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	return lines
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
