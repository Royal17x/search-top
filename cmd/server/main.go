package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Royal17x/search-top/internal/anomaly"
	"github.com/Royal17x/search-top/internal/api"
	"github.com/Royal17x/search-top/internal/config"
	"github.com/Royal17x/search-top/internal/consumer"
	"github.com/Royal17x/search-top/internal/stoplist"
	"github.com/Royal17x/search-top/internal/window"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	win := window.NewTrendingWindow()
	defer win.Close()

	sl := stoplist.NewStopList()
	detector := anomaly.New(cfg.AnomalyWindow, cfg.AnomalyMaxAllowed)
	cache := window.NewTopCache(win, sl, cfg.TrendingTop, cfg.CacheRefresh)
	defer cache.Close()

	kafkaConsumer := consumer.NewConsumer(
		cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID,
		win, detector, log,
	)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      api.NewServer(cache, sl, log).Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("http listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server died", "err", err)
			cancel()
		}
	}()

	go func() {
		log.Info("kafka consumer started", "topic", cfg.KafkaTopic)
		if err := kafkaConsumer.Run(ctx); err != nil {
			log.Error("consumer exited", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("graceful shutdown failed", "err", err)
	}
	if err := kafkaConsumer.Close(); err != nil {
		log.Error("consumer close failed", "err", err)
	}
}
