package anomaly

import (
	"testing"
	"time"
)

func TestDetector_allows_under_max_allowed(t *testing.T) {
	d := New(time.Minute, 3)

	for i := 0; i < 3; i++ {
		if d.IsSus("u1", "кроссовки") {
			t.Fatalf("must not be sus on attempt: %d (max_allowed=3)", i+1)
		}
	}
}

func TestDetector_blocks_after_max_allowed(t *testing.T) {
	d := New(time.Minute, 3)

	for i := 0; i < 3; i++ {
		d.IsSus("u1", "кроссовки")
	}
	if !d.IsSus("u1", "кроссовки") {
		t.Error("want: sus after max allowed exceeded")
	}
}

func TestDetector_different_users_independent(t *testing.T) {
	d := New(time.Minute, 2)

	d.IsSus("u1", "кроссовки")
	d.IsSus("u1", "кроссовки")
	d.IsSus("u1", "кроссовки")

	if d.IsSus("u2", "кроссовки") {
		t.Error("u2 must not be affected by u1 exceeding the limit")
	}
}

func TestDetector_same_user_different_queries_independent(t *testing.T) {
	d := New(time.Minute, 1)

	d.IsSus("u1", "кроссовки")
	d.IsSus("u1", "кроссовки")

	if d.IsSus("u1", "куртка") {
		t.Error("different query for same user must not be sus")
	}
}

func TestDetector_resets_after_window(t *testing.T) {
	d := New(50*time.Millisecond, 1)

	d.IsSus("u1", "кроссовки")
	if !d.IsSus("u1", "кроссовки") {
		t.Error("want: sus before window reset")
	}

	time.Sleep(70 * time.Millisecond)

	if d.IsSus("u1", "кроссовки") {
		t.Error("must not be sus after window reset")
	}
}
