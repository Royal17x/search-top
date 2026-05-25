package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/Royal17x/search-top/internal/metrics"
	"github.com/Royal17x/search-top/internal/window"
)

const maxEventAge = 5 * time.Minute

type AnomalyDetector interface {
	isSus(userID, query string) bool
}

type Consumer struct {
	reader   *kafka.Reader
	window   *window.TrendingWindow
	detector AnomalyDetector
	log      *slog.Logger
}

func NewConsumer(brokers []string, topic, groupID string, w *window.TrendingWindow, d AnomalyDetector, log *slog.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       1 << 20,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
	return &Consumer{reader: r, window: w, detector: d, log: log}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.log.Error("kafka read failed", "err", err)
			continue
		}
		c.handle(msg.Value)
	}
}

func (c *Consumer) handle(data []byte) {
	metrics.EventsConsumed.Inc()

	var ev SearchEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		metrics.EventsDropped.WithLabelValues("invalid").Inc()
		return
	}

	query := strings.ToLower(strings.TrimSpace(ev.Query))
	if query == "" {
		metrics.EventsDropped.WithLabelValues("invalid").Inc()
		return
	}

	if time.Since(ev.Timestamp) > maxEventAge {
		metrics.EventsDropped.WithLabelValues("stale").Inc()
		return
	}

	if c.detector.isSus(ev.UserID, query) {
		metrics.EventsDropped.WithLabelValues("anomaly").Inc()
		return
	}

	c.window.Record(query)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
