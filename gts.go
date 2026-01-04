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

var (
	timeSlicePool sync.Pool // stores []time.Time
	intSlicePool  sync.Pool // stores []int
)

// SortAsc sorts the underlying slice in ascending order according to the extracted time field (thread-safe).
// This implementation uses the decorate-sort-undecorate pattern with lower allocations:
// - precompute a `[]time.Time` of keys
// - sort a slice of indices using the keys
// - reorder the slice once under a write lock
// Temporary slices for times and indices are reused via `sync.Pool` to reduce allocations.
func (ts *TimeSlice[T]) SortAsc() {
	// copy current slice under read lock
	ts.mu.RLock()
	orig := make([]T, len(ts.slice))
	copy(orig, ts.slice)
	ts.mu.RUnlock()

	n := len(orig)
	if n == 0 {
		return
	}
	// get or allocate times slice
	var times []time.Time
	if v := timeSlicePool.Get(); v != nil {
		times = v.([]time.Time)
		if cap(times) < n {
			times = make([]time.Time, n)
		} else {
			times = times[:n]
		}
	} else {
		times = make([]time.Time, n)
	}
	// fill times
	for i := 0; i < n; i++ {
		times[i] = ts.fieldTimeExtractor(orig[i])
	}
	// get or allocate indices slice
	var idxs []int
	if v := intSlicePool.Get(); v != nil {
		idxs = v.([]int)
		if cap(idxs) < n {
			idxs = make([]int, n)
		} else {
			idxs = idxs[:n]
		}
	} else {
		idxs = make([]int, n)
	}
	for i := 0; i < n; i++ {
		idxs[i] = i
	}
	// sort indices by times
	sort.SliceStable(idxs, func(i, j int) bool { return times[idxs[i]].Before(times[idxs[j]]) })
	// build new ordered slice
	newSlice := make([]T, n)
	for i := 0; i < n; i++ {
		newSlice[i] = orig[idxs[i]]
	}
	// replace under write lock
	ts.mu.Lock()
	ts.slice = newSlice
	ts.mu.Unlock()
	// put buffers back to pools (reset length to 0 but keep capacity)
	timeSlicePool.Put(times[:0])
	intSlicePool.Put(idxs[:0])
}

// SortDesc sorts the underlying slice in descending order according to the extracted time field (thread-safe).
// Uses the decorate-sort-undecorate pattern similar to SortAsc with index sorting reversed.
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
	// get or allocate times slice
	var times []time.Time
	if v := timeSlicePool.Get(); v != nil {
		times = v.([]time.Time)
		if cap(times) < n {
			times = make([]time.Time, n)
		} else {
			times = times[:n]
		}
	} else {
		times = make([]time.Time, n)
	}
	// fill times
	for i := 0; i < n; i++ {
		times[i] = ts.fieldTimeExtractor(orig[i])
	}
	// get or allocate indices slice
	var idxs []int
	if v := intSlicePool.Get(); v != nil {
		idxs = v.([]int)
		if cap(idxs) < n {
			idxs = make([]int, n)
		} else {
			idxs = idxs[:n]
		}
	} else {
		idxs = make([]int, n)
	}
	for i := 0; i < n; i++ {
		idxs[i] = i
	}
	// sort indices by times descending
	sort.SliceStable(idxs, func(i, j int) bool { return times[idxs[i]].After(times[idxs[j]]) })
	// build new ordered slice
	newSlice := make([]T, n)
	for i := 0; i < n; i++ {
		newSlice[i] = orig[idxs[i]]
	}
	// replace under write lock
	ts.mu.Lock()
	ts.slice = newSlice
	ts.mu.Unlock()
	// put buffers back to pools
	timeSlicePool.Put(times[:0])
	intSlicePool.Put(idxs[:0])
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
