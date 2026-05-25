package api

import (
	"net/http"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Royal17x/search-top/internal/stoplist"
	"github.com/Royal17x/search-top/internal/window"
)

type Server struct {
	cache    *window.TrendingCache
	stoplist *stoplist.StopList
	log      *slog.Logger
}

func NewServer(cache *window.TrendingCache, sl *stoplist.StopList, log *slog.Logger) *Server {
	return &Server{cache: cache, stoplist: sl, log: log}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/top", s.handleTop)
	mux.HandleFunc("GET /api/v1/stoplist", s.handleStoplistList)
	mux.HandleFunc("POST /api/v1/stoplist", s.handleStoplistAdd)
	mux.HandleFunc("DELETE /api/v1/stoplist/{word}", s.handleStoplistRemove)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.Handle("GET /metrics", promhttp.Handler())

	var h http.Handler = mux
	h = recovery(h)
	h = logging(s.log)(h)
	return h
}
