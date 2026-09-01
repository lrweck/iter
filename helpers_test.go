package iter

import (
	"errors"
	"slices"
	"testing"
)

// checkEqual fails when got differs from want, on the helper's line.
func checkEqual[T comparable](t *testing.T, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

var errSentinel = errors.New("bad input")

func parseLen(s string) (int, error) {
	if s == "x" {
		return 0, errSentinel
	}
	return len(s), nil
}

func failOnTwo(a, b int) (int, error) {
	if b == 2 {
		return 0, errSentinel
	}
	return a + b, nil
}

type pair struct {
	k int
	v string
}
