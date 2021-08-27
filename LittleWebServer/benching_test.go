package main

import (
	"math/rand"
	"testing"
)
const (
	size = 1_000_000
	low = 100
	mid = 500000
	high = 999999
)

func BenchmarkFindingTests(b *testing.B) {
	b.Run("Random", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindRandom("fish")
		}
	})

	b.Run("Low", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindLow("fish")
		}
	})

	b.Run("Mid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindMid("fish")
		}
	})

	b.Run("High", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindHigh("fish")
		}
	})
}

var haystack = make([]string, size)

func FindRandom(needle string) int {
	//haystack := make([]string, size)
	haystack[rand.Intn(size)] = needle
	for i, e := range haystack {
		if needle == e {
			return i
		}
	}
	return -1
}
func FindLow(needle string) int {
	//haystack := make([]string, size)
	haystack[low] = needle
	for i, e := range haystack {
		if needle == e {
			return i
		}
	}
	return -1
}
func FindMid(needle string) int {
	//haystack := make([]string, size)
	haystack[mid] = needle
	for i, e := range haystack {
		if needle == e {
			return i
		}
	}
	return -1
}
func FindHigh(needle string) int {
	//haystack := make([]string, size)
	haystack[high] = needle

	for i, e := range haystack {
		if needle == e {
			return i
		}
	}
	return -1
}