package gts

import (
	"sort"
	"sync"
	"time"
)

// TimeSlice is a generic slice wrapper that supports sorting elements by a time field extracted via fieldTimeExtractor.
// It is safe for concurrent use.
type TimeSlice[T any] struct {
	fieldTimeExtractor func(T) time.Time // extracts time from T
	slice              []T               // underlying slice of elements
	mu                 sync.RWMutex      // mutex for concurrency
}

// New creates and returns a new TimeSlice instance for the given fieldTimeExtractor function.
func New[T any](values []T, fieldTimeExtractor func(T) time.Time) *TimeSlice[T] {
	if fieldTimeExtractor == nil {
		panic("gts: fieldTimeExtractor cannot be nil")
	}
	return &TimeSlice[T]{
		fieldTimeExtractor: fieldTimeExtractor,
		slice:              values,
	}
}

// Len returns the length of the underlying slice.
// Implements sort.Interface.
func (ts *TimeSlice[T]) Len() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.slice)
}

// LessAsc compares the time field of slice[i] and slice[j] for ascending order.
// Returns true if the time of slice[i] is before that of slice[j].
func (ts *TimeSlice[T]) LessAsc(i, j int) bool {
	ts.mu.RLock()
	v1 := ts.slice[i]
	v2 := ts.slice[j]
	ts.mu.RUnlock()
	t1 := ts.fieldTimeExtractor(v1)
	t2 := ts.fieldTimeExtractor(v2)
	return t1.Before(t2)
}

// LessDesc compares the time field of slice[i] and slice[j] for descending order.
// Returns true if the time of slice[i] is after that of slice[j].
func (ts *TimeSlice[T]) LessDesc(i, j int) bool {
	ts.mu.RLock()
	v1 := ts.slice[i]
	v2 := ts.slice[j]
	ts.mu.RUnlock()
	t1 := ts.fieldTimeExtractor(v1)
	t2 := ts.fieldTimeExtractor(v2)
	return t1.After(t2)
}

// Swap exchanges the elements at indices i and j in the slice.
// Implements sort.Interface.
func (ts *TimeSlice[T]) Swap(i, j int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.slice[i], ts.slice[j] = ts.slice[j], ts.slice[i]
}

// SortAsc sorts the underlying slice in ascending order according to the extracted time field (thread-safe).
// This implementation uses the decorate-sort-undecorate pattern: it copies the
// slice, precomputes the time keys, sorts indices by key, and then reorders the
// original slice once under a write lock. This reduces calls to the extractor
// and avoids holding locks during comparison.
func (ts *TimeSlice[T]) SortAsc() {
	// copy current slice under read lock
	ts.mu.RLock()
	orig := make([]T, len(ts.slice))
	copy(orig, ts.slice)
	ts.mu.RUnlock()

	n := len(orig)
	type pair struct {
		idx int
		t   time.Time
	}
	if n == 0 {
		return
	}
	pairs := make([]pair, n)
	for i := 0; i < n; i++ {
		pairs[i] = pair{idx: i, t: ts.fieldTimeExtractor(orig[i])}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].t.Before(pairs[j].t)
	})
	// build new ordered slice
	newSlice := make([]T, n)
	for i := 0; i < n; i++ {
		newSlice[i] = orig[pairs[i].idx]
	}
	// replace under write lock
	ts.mu.Lock()
	ts.slice = newSlice
	ts.mu.Unlock()
}

// SortDesc sorts the underlying slice in descending order according to the extracted time field (thread-safe).
// Uses the decorate-sort-undecorate pattern similar to SortAsc.
func (ts *TimeSlice[T]) SortDesc() {
	// copy current slice under read lock
	ts.mu.RLock()
	orig := make([]T, len(ts.slice))
	copy(orig, ts.slice)
	ts.mu.RUnlock()

	n := len(orig)
	if n == 0 {
		return
	}
	type pair struct {
		idx int
		t   time.Time
	}
	pairs := make([]pair, n)
	for i := 0; i < n; i++ {
		pairs[i] = pair{idx: i, t: ts.fieldTimeExtractor(orig[i])}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].t.After(pairs[j].t)
	})
	// build new ordered slice
	newSlice := make([]T, n)
	for i := 0; i < n; i++ {
		newSlice[i] = orig[pairs[i].idx]
	}
	// replace under write lock
	ts.mu.Lock()
	ts.slice = newSlice
	ts.mu.Unlock()
}

// Items returns a copy of the underlying slice of items (thread-safe).
func (ts *TimeSlice[T]) Items() []T {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	copied := make([]T, len(ts.slice))
	copy(copied, ts.slice)
	return copied
}

// Clone returns a new TimeSlice with a copy of the underlying slice and the same fieldTimeExtractor.
// The clone does not share state with the original.
func (ts *TimeSlice[T]) Clone() *TimeSlice[T] {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	newSlice := make([]T, len(ts.slice))
	copy(newSlice, ts.slice)
	return &TimeSlice[T]{fieldTimeExtractor: ts.fieldTimeExtractor, slice: newSlice}
}
