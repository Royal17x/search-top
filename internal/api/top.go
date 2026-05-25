package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Royal17x/search-top/internal/metrics"
	"github.com/Royal17x/search-top/internal/window"
)

type topResponse struct {
	Items []window.Entry `json:"items"`
	Total int            `json:"total"`
}

func (s *Server) handleTop(w http.ResponseWriter, r *http.Request) {
	metrics.TopRequests.Inc()

	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil || n <= 0 {
		n = 10
	}
	if n > 100 {
		n = 100
	}

	items := s.cache.Get()
	if len(items) > n {
		items = items[:n]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topResponse{Items: items, Total: len(items)})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
