package gts

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type testEvent struct {
	Name string
	Date time.Time
}

func eventTime(e testEvent) time.Time {
	return e.Date
}

// ...existing code...

func generateLargeEvents(n int) []testEvent {
	events := make([]testEvent, n)
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		events[i] = testEvent{
			Name: "Event" + string(rune(i%26+65)),
			Date: base.Add(time.Duration(i) * time.Hour),
		}
	}
	return events
}

func BenchmarkSortAsc_10(b *testing.B)    { benchmarkSortAsc(b, 10) }
func BenchmarkSortAsc_100(b *testing.B)   { benchmarkSortAsc(b, 100) }
func BenchmarkSortAsc_500(b *testing.B)   { benchmarkSortAsc(b, 500) }
func BenchmarkSortAsc_1000(b *testing.B)  { benchmarkSortAsc(b, 1000) }
func BenchmarkSortAsc_5000(b *testing.B)  { benchmarkSortAsc(b, 5000) }
func BenchmarkSortAsc_10000(b *testing.B) { benchmarkSortAsc(b, 10000) }

func benchmarkSortAsc(b *testing.B, n int) {
	events := generateLargeEvents(n)
	ts := New(events, eventTime)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.SortAsc()
	}
}

func BenchmarkSortDesc_10(b *testing.B)    { benchmarkSortDesc(b, 10) }
func BenchmarkSortDesc_100(b *testing.B)   { benchmarkSortDesc(b, 100) }
func BenchmarkSortDesc_500(b *testing.B)   { benchmarkSortDesc(b, 500) }
func BenchmarkSortDesc_1000(b *testing.B)  { benchmarkSortDesc(b, 1000) }
func BenchmarkSortDesc_5000(b *testing.B)  { benchmarkSortDesc(b, 5000) }
func BenchmarkSortDesc_10000(b *testing.B) { benchmarkSortDesc(b, 10000) }

func benchmarkSortDesc(b *testing.B, n int) {
	events := generateLargeEvents(n)
	ts := New(events, eventTime)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.SortDesc()
	}
}

func TestLenAndSwap(t *testing.T) {
	ts := New([]testEvent{
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}, eventTime)
	if ts.Len() != 2 {
		t.Errorf("Len() = %d, want 2", ts.Len())
	}
	ts.Swap(0, 1)
	got := ts.Items()
	want := []testEvent{
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Swap() failed, got %v, want %v", got, want)
	}
}

func TestLessAscAndLessDesc(t *testing.T) {
	events := []testEvent{
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	ts := New(events, eventTime)
	if !ts.LessAsc(0, 1) {
		t.Error("LessAsc(0,1) should be true")
	}
	if ts.LessAsc(1, 0) {
		t.Error("LessAsc(1,0) should be false")
	}
	if !ts.LessDesc(1, 0) {
		t.Error("LessDesc(1,0) should be true")
	}
	if ts.LessDesc(0, 1) {
		t.Error("LessDesc(0,1) should be false")
	}
}

func TestSortAscAndSortDesc(t *testing.T) {
	events := []testEvent{
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"C", time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	ts := New(events, eventTime)
	ts.SortAsc()
	got := ts.Items()
	want := []testEvent{
		{"C", time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortAsc() got %v, want %v", got, want)
	}
	ts.SortDesc()
	got = ts.Items()
	want = []testEvent{
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"C", time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortDesc() got %v, want %v", got, want)
	}
}

func TestClone(t *testing.T) {
	events := []testEvent{
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	ts := New(events, eventTime)
	clone := ts.Clone()
	if !reflect.DeepEqual(ts.Items(), clone.Items()) {
		t.Error("Clone() did not copy items correctly")
	}
	clone.Swap(0, 1)
	if reflect.DeepEqual(ts.Items(), clone.Items()) {
		t.Error("Clone() should not share state with original")
	}
}

func TestConcurrencySafety(t *testing.T) {
	events := []testEvent{
		{"A", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"B", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	ts := New(events, eventTime)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ts.Len()
			ts.Items()
			ts.SortAsc()
			ts.SortDesc()
			ts.Clone()
		}(i)
	}
	wg.Wait()
}
