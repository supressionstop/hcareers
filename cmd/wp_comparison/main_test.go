package main

import "testing"

var (
	n_10     []string
	n_100    []string
	n_1_000  []string
	n_10_000 []string
)

func init() {
	n_10 = readUrls(file_10, false)
	n_100 = readUrls(file_100, false)
	n_1_000 = readUrls(file_1_000, false)
	n_10_000 = readUrls(file_10_000, false)
}

func BenchmarkGoEachRequest10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goEachRequestStub(n_10, false)
	}
}

func BenchmarkGoEachRequest100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goEachRequestStub(n_100, false)
	}
}

func BenchmarkGoEachRequest1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goEachRequestStub(n_1_000, false)
	}
}

func BenchmarkGoEachRequest10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goEachRequestStub(n_10_000, false)
	}
}
