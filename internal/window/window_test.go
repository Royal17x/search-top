package window

import "testing"

func TestRecord_and_aggregate(t *testing.T) {
	w := NewTrendingWindow()
	defer w.Close()

	w.Record("кроссовки")
	w.Record("кроссовки")
	w.Record("футболка")

	totals := w.Aggregate(nil)

	if totals["кроссовки"] != 2 {
		t.Errorf("want: 2, got: %d", totals["кроссовки"])
	}
	if totals["футболка"] != 1 {
		t.Errorf("want: 1, got: %d", totals["футболка"])
	}
}

func TestAggregate_stopList(t *testing.T) {
	w := NewTrendingWindow()
	defer w.Close()

	w.Record("стоп-слово")
	w.Record("кроссовки")

	blocked := map[string]struct{}{"стоп-слово": {}}
	totals := w.Aggregate(blocked)

	if _, ok := totals["стоп-слово"]; ok {
		t.Fatal("stopList word not found in aggregation")
	}
	if totals["кроссовки"] != 1 {
		t.Errorf("want: 1, got: %d", totals["кроссовки"])
	}
}

func TestRotation_clears_old_data(t *testing.T) {
	w := NewTrendingWindow()
	defer w.Close()

	w.Record("старая память")

	for i := 0; i < bucketCount; i++ {
		next := (w.current.Load() + 1) % bucketCount
		w.buckets[next].emptyAndGet()
		w.current.Add(1)
	}

	totals := w.Aggregate(nil)
	if totals["старая память"] != 0 {
		t.Errorf("failed to reset old data after rotation, got: %d", totals["старая память"])
	}
}
