package iter

import "cmp"

// Number is a constraint for all builtin numeric types, including complex
// (which cannot be ordered, so Max/Min stay on cmp.Ordered). It cannot be a
// method receiver constraint, so summing lives here as a package function, as
// math/rand/v2 does with its intType.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~complex64 | ~complex128
}

// Sum adds all elements. The empty sequence sums to zero.
func Sum[T Number](s Seq[T]) T {
	var total T
	for v := range s.seq {
		total += v
	}
	return total
}

// SumByFunc sums the keys produced by key. The empty sequence sums to zero.
func (s Seq[T]) SumByFunc[K Number](key func(T) K) K {
	var total K
	for v := range s.seq {
		total += key(v)
	}
	return total
}

// Max returns the greatest element, or ok=false if the sequence is empty.
func Max[T cmp.Ordered](s Seq[T]) (T, bool) {
	var best T
	ok := false
	for v := range s.seq {
		if !ok || cmp.Less(best, v) {
			best, ok = v, true
		}
	}
	return best, ok
}

// MaxByFunc returns the element with the greatest key, or ok=false if the
// sequence is empty.
func (s Seq[T]) MaxByFunc[K cmp.Ordered](key func(T) K) (T, bool) {
	var best T
	var bestKey K
	ok := false
	for v := range s.seq {
		k := key(v)
		if !ok || cmp.Less(bestKey, k) {
			best, bestKey, ok = v, k, true
		}
	}
	return best, ok
}

// Min returns the least element, or ok=false if the sequence is empty.
func Min[T cmp.Ordered](s Seq[T]) (T, bool) {
	var best T
	ok := false
	for v := range s.seq {
		if !ok || cmp.Less(v, best) {
			best, ok = v, true
		}
	}
	return best, ok
}

// MinByFunc returns the element with the least key, or ok=false if the
// sequence is empty.
func (s Seq[T]) MinByFunc[K cmp.Ordered](key func(T) K) (T, bool) {
	var best T
	var bestKey K
	ok := false
	for v := range s.seq {
		k := key(v)
		if !ok || cmp.Less(k, bestKey) {
			best, bestKey, ok = v, k, true
		}
	}
	return best, ok
}
