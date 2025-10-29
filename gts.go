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
	t1 := ts.fieldTimeExtractor(ts.slice[i])
	t2 := ts.fieldTimeExtractor(ts.slice[j])
	ts.mu.RUnlock()
	return t1.Before(t2)
}

// LessDesc compares the time field of slice[i] and slice[j] for descending order.
// Returns true if the time of slice[i] is after that of slice[j].
func (ts *TimeSlice[T]) LessDesc(i, j int) bool {
	ts.mu.RLock()
	t1 := ts.fieldTimeExtractor(ts.slice[i])
	t2 := ts.fieldTimeExtractor(ts.slice[j])
	ts.mu.RUnlock()
	return t1.After(t2)
}

// Swap exchanges the elements at indices i and j in the slice.
// Implements sort.Interface.
func (ts *TimeSlice[T]) Swap(i, j int) {
	ts.mu.Lock()
	ts.slice[i], ts.slice[j] = ts.slice[j], ts.slice[i]
	ts.mu.Unlock()
}

// SortAsc sorts the underlying slice in ascending order according to the extracted time field (thread-safe).
func (ts *TimeSlice[T]) SortAsc() {
	ts.mu.Lock()
	sort.SliceStable(ts.slice, func(i, j int) bool {
		return ts.fieldTimeExtractor(ts.slice[i]).Before(ts.fieldTimeExtractor(ts.slice[j]))
	})
	ts.mu.Unlock()
}

// SortDesc sorts the underlying slice in descending order according to the extracted time field (thread-safe).
func (ts *TimeSlice[T]) SortDesc() {
	ts.mu.Lock()
	sort.SliceStable(ts.slice, func(i, j int) bool {
		return ts.fieldTimeExtractor(ts.slice[i]).After(ts.fieldTimeExtractor(ts.slice[j]))
	})
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
