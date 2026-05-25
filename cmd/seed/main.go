package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	queries := []string{
		"кроссовки", "куртка", "платье", "телефон",
		"ноутбук", "сумка", "часы", "кроссовки",
		"кроссовки", "куртка", "платье", "телефон",
		"кроссовки", "ноутбук", "куртка",
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9094"},
		Topic:   "search-events",
	})
	defer w.Close()

	now := time.Now().UTC()
	msgs := make([]kafka.Message, len(queries))
	for i, q := range queries {
		body, _ := json.Marshal(map[string]any{
			"query":     q,
			"user_id":   fmt.Sprintf("u%d", rand.Intn(100)),
			"timestamp": now.Format(time.RFC3339),
		})
		msgs[i] = kafka.Message{Value: body}
	}

	if err := w.WriteMessages(context.Background(), msgs...); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("sent %d messages to search-events\n", len(msgs))
}
