package stoplist

import "testing"

func TestAdd_clean_words_input(t *testing.T) {
	sl := NewStopList()
	sl.Add("Кроссовки")
	sl.Add("  куртка ")
	sl.Add("")

	snap := sl.Snapshot()
	if _, ok := snap["кроссовки"]; !ok {
		t.Error("want: кроссовки in snapshot after uppercase input")
	}
	if _, ok := snap["куртка"]; !ok {
		t.Error("want: куртка in snapshot after trimmed input")
	}
	if len(snap) != 2 {
		t.Errorf("want: 2 words, got: %d", len(snap))
	}
}

func TestRemove(t *testing.T) {
	sl := NewStopList()
	sl.Add("кроссовки")
	sl.Remove("кроссовки")

	if words := sl.AllWords(); len(words) != 0 {
		t.Errorf("want: empty after remove, got: %v", words)
	}
}

func TestSnapshot_is_copy(t *testing.T) {
	sl := NewStopList()
	sl.Add("кроссовки")

	snap := sl.Snapshot()
	snap["новое"] = struct{}{}

	if _, ok := sl.Snapshot()["новое"]; ok {
		t.Error("snapshot must not affect the original")
	}
}

func TestAdd_duplicate_ignored(t *testing.T) {
	sl := NewStopList()
	sl.Add("кроссовки")
	sl.Add("кроссовки")
	sl.Add("кроссовки")

	if n := len(sl.AllWords()); n != 1 {
		t.Errorf("want: 1 unique word, got: %d", n)
	}
}
