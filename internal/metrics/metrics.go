package metrics

import "github.com/prometheus/client_golang/prometheus/promauto"
import "github.com/prometheus/client_golang/prometheus"

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
		Help: "Total calls to the top-N",
	})

	AggregationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "search_aggregate_duration_seconds",
		Help:    "Time spent aggregating bucket window into totals map",
		Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10),
	})

	StoplistSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "search_stoplist_size",
		Help: "Current number of words in stop list",
	})
)
