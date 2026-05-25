package consumer

import "time"

type SearchEvent struct {
	Query     string    `json:"query"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
