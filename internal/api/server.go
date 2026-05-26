package api

import (
	"net/http"

	"log/slog"

	"github.com/Royal17x/search-top/internal/stoplist"
	"github.com/Royal17x/search-top/internal/window"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /partial/top", s.handlePartialTop)
	mux.HandleFunc("GET /partial/stoplist", s.handlePartialStoplist)
	mux.HandleFunc("POST /partial/stoplist", s.handlePartialStoplistAdd)
	mux.HandleFunc("DELETE /partial/stoplist/{word}", s.handlePartialStoplistRemove)

	mux.HandleFunc("GET /api/v1/top", s.handleTop)
	mux.HandleFunc("GET /api/v1/stoplist", s.handleStoplistList)
	mux.HandleFunc("POST /api/v1/stoplist", s.handleStoplistAdd)
	mux.HandleFunc("DELETE /api/v1/stoplist/{word}", s.handleStoplistRemove)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.Handle("GET /metrics", promhttp.Handler())

	var h http.Handler = mux
	h = otelhttp.NewHandler(h, "search-top")
	h = recovery(h)
	h = logging(s.log)(h)
	return h
}
