package api

import (
	"encoding/json"
	"net/http"

	"github.com/Royal17x/search-top/internal/metrics"
)

type wordRequest struct {
	Word string `json:"word"`
}

func (s *Server) handleStoplistAdd(w http.ResponseWriter, r *http.Request) {
	var req wordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, `{"error":"word required"}`, http.StatusBadRequest)
		return
	}
	s.stoplist.Add(req.Word)
	metrics.StoplistSize.Set(float64(len(s.stoplist.AllWords())))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleStoplistRemove(w http.ResponseWriter, r *http.Request) {
	word := r.PathValue("word")
	s.stoplist.Remove(word)
	metrics.StoplistSize.Set(float64(len(s.stoplist.AllWords())))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStoplistList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"words": s.stoplist.AllWords()})
}
