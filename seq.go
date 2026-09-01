// Package iter provides lazy, composable sequences over Go 1.27 iterators.
//
// Values flow through a pipeline of generic methods, so lookup cost is paid
// only on the final terminal method:
//
//	iter.Of(1, 2, 3, 4).Filter(isEven).Map(square).Collect()
//
// Range-over-func makes consumers the driver: breaking out of a `range` on a
// returned sequence stops upstream work immediately.
//
// Fallible operations (MapErr, FlatMapErr) run every element and pair the
// outcome as a Result[U], a sequence of (result, error): (result, nil) on
// success, (zero, err) on failure. The error travels as data, so it can be
// logged with Tap, dropped with IgnoreErr, collected with Errors, or surfaced
// with CollectErr.
package iter

import (
	"slices"

	stditer "iter"
)

// Seq is a lazy sequence of values of type T.
type Seq[T any] struct{ seq stditer.Seq[T] }

// Seq2 is a lazy sequence of key/value pairs.
type Seq2[K, V any] struct{ seq stditer.Seq2[K, V] }

// From wraps a standard iterator into a Seq.
func From[T any](seq stditer.Seq[T]) Seq[T] { return Seq[T]{seq: seq} }

// From2 wraps a standard pair iterator into a Seq2.
func From2[K, V any](seq stditer.Seq2[K, V]) Seq2[K, V] { return Seq2[K, V]{seq: seq} }

// Of builds a Seq from its arguments.
func Of[T any](vals ...T) Seq[T] { return From(slices.Values(vals)) }

// Seq exposes the underlying standard iterator, for ranging directly:
//
//	for v := range seq.Seq() { ... }
func (s Seq[T]) Seq() stditer.Seq[T] { return s.seq }

// Range yields from up to, but not including, to, stepping by 1.
func Range(from, to int) Seq[int] {
	return From(func(yield func(int) bool) {
		for i := from; i < to; i++ {
			if !yield(i) {
				return
			}
		}
	})
}

// Repeat yields n copies of v.
func Repeat[T any](n int, v T) Seq[T] {
	return From(func(yield func(T) bool) {
		for range n {
			if !yield(v) {
				return
			}
		}
	})
}

// RepeatBy yields the values produced by f(0), f(1), ..., f(n-1).
func RepeatBy[T any](n int, f func(int) T) Seq[T] {
	return From(func(yield func(T) bool) {
		for i := range n {
			if !yield(f(i)) {
				return
			}
		}
	})
}

// Concat concatenates the sequences in order.
func Concat[T any](seqs ...Seq[T]) Seq[T] {
	return From(func(yield func(T) bool) {
		for _, s := range seqs {
			for v := range s.seq {
				if !yield(v) {
					return
				}
			}
		}
	})
}

// Map applies f to each element.
func (s Seq[T]) Map[U any](f func(T) U) Seq[U] {
	return From(func(yield func(U) bool) {
		for v := range s.seq {
			if !yield(f(v)) {
				return
			}
		}
	})
}

// MapErr applies f to each element, pairing every outcome: (result, nil) on
// success, (zero, err) on failure. The error pairs never stop the pipeline;
// handling them is up to the consumer (IgnoreErr, Errors, CollectErr).
func (s Seq[T]) MapErr[U any](f func(T) (U, error)) Result[U] {
	return FromResult(func(yield func(U, error) bool) {
		for v := range s.seq {
			u, err := f(v)
			if err != nil {
				if !yield(*new(U), err) {
					return
				}
				continue
			}
			if !yield(u, nil) {
				return
			}
		}
	})
}

// Filter keeps elements for which f returns true.
func (s Seq[T]) Filter(f func(T) bool) Seq[T] {
	return From(func(yield func(T) bool) {
		for v := range s.seq {
			if f(v) && !yield(v) {
				return
			}
		}
	})
}

// SkipErr keeps elements for which check returns nil, skipping the failures
// and moving on to the next one.
func (s Seq[T]) SkipErr(check func(T) error) Seq[T] {
	return From(func(yield func(T) bool) {
		for v := range s.seq {
			if check(v) != nil {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// FlatMap concatenates the sequences produced by f.
func (s Seq[T]) FlatMap[U any](f func(T) Seq[U]) Seq[U] {
	return From(func(yield func(U) bool) {
		for v := range s.seq {
			for u := range f(v).seq {
				if !yield(u) {
					return
				}
			}
		}
	})
}

// FlatMapErr is FlatMap with an error source: each element produces either a
// run of (result, nil) pairs or a single (zero, err) pair.
func (s Seq[T]) FlatMapErr[U any](f func(T) (Seq[U], error)) Result[U] {
	return FromResult(func(yield func(U, error) bool) {
		for v := range s.seq {
			sub, err := f(v)
			if err != nil {
				if !yield(*new(U), err) {
					return
				}
				continue
			}
			for u := range sub.seq {
				if !yield(u, nil) {
					return
				}
			}
		}
	})
}

// Take keeps at most n elements.
func (s Seq[T]) Take(n int) Seq[T] {
	return From(func(yield func(T) bool) {
		for v := range s.seq {
			if n <= 0 {
				return
			}
			n--
			if !yield(v) {
				return
			}
		}
	})
}

// Drop discards the first n elements.
func (s Seq[T]) Drop(n int) Seq[T] {
	return From(func(yield func(T) bool) {
		for v := range s.seq {
			if n > 0 {
				n--
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// TakeWhile yields elements from the start while pred is true.
func (s Seq[T]) TakeWhile(pred func(T) bool) Seq[T] {
	return From(func(yield func(T) bool) {
		for v := range s.seq {
			if !pred(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	})
}

// DropWhile discards elements from the start while pred is true.
func (s Seq[T]) DropWhile(pred func(T) bool) Seq[T] {
	return From(func(yield func(T) bool) {
		drop := true
		for v := range s.seq {
			if drop && pred(v) {
				continue
			}
			drop = false
			if !yield(v) {
				return
			}
		}
	})
}

// Tap inspects every element as it flows through, then yields it unchanged.
// It is the Go counterpart of Elixir's |> tap/2: a hook for logging and
// side effects in the middle of a pipeline.
func (s Seq[T]) Tap(f func(T)) Seq[T] {
	return From(func(yield func(T) bool) {
		for v := range s.seq {
			f(v)
			if !yield(v) {
				return
			}
		}
	})
}

// Chunk groups the sequence into slices of at most n elements. It panics if n
// is not positive.
//
// Chunk is a package function, not a method: a generic method returning
// Seq[[]T] instantiates the receiver's own type with []T and trips the
// compiler's instantiation-cycle check.
func Chunk[T any](s Seq[T], n int) Seq[[]T] {
	if n <= 0 {
		panic("iter: non-positive chunk size")
	}
	return From(func(yield func([]T) bool) {
		var chunk []T
		for v := range s.seq {
			chunk = append(chunk, v)
			if len(chunk) == n {
				if !yield(chunk) {
					return
				}
				chunk = nil
			}
		}
		if len(chunk) > 0 && !yield(chunk) {
			return
		}
	})
}
