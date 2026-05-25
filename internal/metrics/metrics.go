package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsConsumed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "search_events_consumed_total",
		Help: "Total search events consumed from Kafka",
	})

	EventsDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "search_events_dropped_total",
		Help: "Events dropped before recording by reason",
	}, []string{"reason"})

	TopRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "search_top_requests_total",
		Help: "Total HTTP calls to GET /api/v1/top",
	})

	AggregationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "search_aggregate_duration_seconds",
		Help:    "Time spent aggregate all 30 buckets into totals map",
		Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10),
	})

	StoplistSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "search_stoplist_size",
		Help: "Current number of words in stop list",
	})
)
