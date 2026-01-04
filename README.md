<img align="left" width="250"  src="logo.png" alt="go-timesort Logo" vspace="20" />

[![Go Reference](https://pkg.go.dev/badge/github.com/azrod/go-timesort.svg)](https://pkg.go.dev/github.com/azrod/go-timesort)
[![Go Report Card](https://goreportcard.com/badge/github.com/azrod/go-timesort)](https://goreportcard.com/report/github.com/azrod/go-timesort)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)](https://golang.org/dl/)

**Overview**

**go-timesort** is a lightweight and generic Go library that helps you organize and sort slices of any type based on a time or date field.

It is thread-safe, flexible, and designed for easy integration in your Go projects

---

## Learn More: Sorting Slices by Date in Go

For a detailed explanation of how Go's `sort.Interface` works and how to sort slices by date, check out this Medium post:

👉 [How to Sort a Golang Slice by Date](https://medium.com/@azrod/how-to-sort-a-golang-slice-by-date-0edce447407c)

## Compatibility

- Requires Go 1.18 or newer (uses Go generics)
- Zero external dependencies (pure standard library)

## Features

- Sort any slice by a time or date field (ascending or descending)
- Stable sorting (preserve order of equal elements)
- Thread-safe (safe for concurrent access)
- Easy to use with Go generics
- Clone slices for safe manipulation

## Installation

```bash
go get github.com/azrod/go-timesort
```

## Usage

### Define your struct

```go
type Event struct {
    Name string
    Date time.Time
}
```

### Create a TimeSort

```go
import gts "github.com/azrod/go-timesort"

users := []User{
    {Name: "Alice", CreatedAt: time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)},
    {Name: "Bob", CreatedAt: time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC)},
}

usersSorter := gts.New(users, func(u User) time.Time { return u.CreatedAt })
```

### Sort ascending

```go
usersSorter.SortAsc()
```

### Sort descending

```go
usersSorter.SortDesc()
```

### Get sorted items

```go
sortedUsers := usersSorter.Items()
for _, u := range sortedUsers {
    fmt.Println(u.Name, u.CreatedAt)
}
```

### Clone your slice

```go
copyUsers := usersSorter.Clone()
```

## Thread Safety

All operations on the underlying slice are protected by a mutex, making this library safe for concurrent use.

## Performance

The library is designed for efficiency, with sorting algorithms optimized for performance. Benchmarks are provided to help you understand the performance characteristics with different data sizes.

```txt
goarch: arm64
pkg: github.com/azrod/go-timesort
cpu: Apple M1 Pro
BenchmarkSortAsc_10-10           2749162             438.8 ns/op           1736 B/op          8 allocs/op
BenchmarkSortAsc_100-10          397156              2790 ns/op           15912 B/op         8 allocs/op
BenchmarkSortAsc_500-10          102378             11999 ns/op           77992 B/op         8 allocs/op
BenchmarkSortAsc_1000-10         50924              23876 ns/op          155817 B/op         8 allocs/op
BenchmarkSortAsc_5000-10         7962              135405 ns/op          778408 B/op         8 allocs/op
BenchmarkSortAsc_10000-10        3693              376569 ns/op         1532074 B/op         8 allocs/op
BenchmarkSortDesc_10-10         1000000             1013 ns/op           1736 B/op          8 allocs/op
BenchmarkSortDesc_100-10         76129             15766 ns/op          15912 B/op         8 allocs/op
BenchmarkSortDesc_500-10         14449             83490 ns/op          77992 B/op         8 allocs/op
BenchmarkSortDesc_1000-10        7136             171374 ns/op         155816 B/op         8 allocs/op
BenchmarkSortDesc_5000-10        1284             957855 ns/op         778408 B/op         8 allocs/op
BenchmarkSortDesc_10000-10       612             1983840 ns/op        1532072 B/op         8 allocs/op
```

## License

MIT

## Contributing

Pull requests and suggestions are welcome! Please open an issue or PR on GitHub.
