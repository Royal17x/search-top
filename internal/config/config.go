package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr          string
	KafkaBrokers      []string
	KafkaTopic        string
	KafkaGroupID      string
	TrendingTop       int
	CacheRefresh      time.Duration
	AnomalyWindow     time.Duration
	AnomalyMaxAllowed int
}

func Load() Config {
	return Config{
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
		KafkaBrokers:      strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:        env("KAFKA_TOPIC", "search-events"),
		KafkaGroupID:      env("KAFKA_GROUP_ID", "search-top-consumer"),
		TrendingTop:       envInt("TRENDING_TOP_N", 10),
		CacheRefresh:      envDuration("CACHE_REFRESH", time.Second),
		AnomalyWindow:     envDuration("ANOMALY_WINDOW", time.Minute),
		AnomalyMaxAllowed: envInt("ANOMALY_MAX_ALLOWED", 20),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
