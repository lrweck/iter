package iter

import (
	"errors"
	"maps"

	stditer "iter"
)

// Result is a fallible sequence: pairs of (result, error), where success is
// (result, nil) and failure is (zero, err). The errors travel as data, so
// they can be inspected or dropped without stopping the pipeline.
//
// It is a Seq2 with the second type argument pre-bound to error, which is
// what lets its methods express error semantics that generic Seq2 methods
// cannot (Go does not allow a method to constrain its receiver's type
// parameters).
type Result[V any] struct{ seq stditer.Seq2[V, error] }

// Setup: Seq2 and Result iterators safe on the zero value.

func (s Seq2[K, V]) iter() stditer.Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]().seq
	}
	return s.seq
}

func (s Result[V]) iter() stditer.Seq2[V, error] {
	if s.seq == nil {
		return emptyResult[V]().seq
	}
	return s.seq
}

// FromResult wraps a standard pair iterator as a fallible sequence.
func FromResult[V any](seq stditer.Seq2[V, error]) Result[V] { return Result[V]{seq: seq} }

// FromMap wraps map iteration as a key/value Seq2 (order not defined).
func FromMap[K comparable, V any](m map[K]V) Seq2[K, V] { return From2(maps.All(m)) }

// FromKeys yields the keys of a map.
func FromKeys[K comparable, V any](m map[K]V) Seq[K] { return From(maps.Keys(m)) }

// FromValues yields the values of a map.
func FromValues[K comparable, V any](m map[K]V) Seq[V] { return From(maps.Values(m)) }

// Enumerate pairs each element with its index.
func (s Seq[T]) Enumerate() Seq2[int, T] {
	if s.seq == nil {
		return emptySeq2[int, T]()
	}
	return From2(func(yield func(int, T) bool) {
		i := 0
		for v := range s.seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	})
}

// Zip pairs each element with the corresponding element of o, stopping at the
// shorter sequence.
func (s Seq[T]) Zip[U any](o Seq[U]) Seq2[T, U] {
	if s.seq == nil {
		return emptySeq2[T, U]()
	}
	return From2(func(yield func(T, U) bool) {
		next, stop := stditer.Pull(o.iter())
		defer stop()
		for v := range s.seq {
			u, ok := next()
			if !ok {
				return
			}
			if !yield(v, u) {
				return
			}
		}
	})
}

// Tap inspects every pair as it flows through, then yields it unchanged; see
// Seq.Tap.
func (s Seq2[K, V]) Tap(f func(K, V)) Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]()
	}
	return From2(func(yield func(K, V) bool) {
		for k, v := range s.seq {
			f(k, v)
			if !yield(k, v) {
				return
			}
		}
	})
}

// Keys yields the keys of a pair sequence.
func (s Seq2[K, V]) Keys() Seq[K] {
	if s.seq == nil {
		return emptySeq[K]()
	}
	return From(func(yield func(K) bool) {
		for k := range s.seq {
			if !yield(k) {
				return
			}
		}
	})
}

// Values yields the values of a pair sequence.
func (s Seq2[K, V]) Values() Seq[V] {
	if s.seq == nil {
		return emptySeq[V]()
	}
	return From(func(yield func(V) bool) {
		for _, v := range s.seq {
			if !yield(v) {
				return
			}
		}
	})
}

// Seq exposes the underlying standard pair iterator, for ranging directly:
//
//	for k, v := range pairs.Seq() { ... }
func (s Seq2[K, V]) Seq() stditer.Seq2[K, V] { return s.seq }

// Count returns the number of pairs.
func (s Seq2[K, V]) Count() int {
	n := 0
	for range s.iter() {
		n++
	}
	return n
}

// Filter keeps the pairs for which f returns true.
func (s Seq2[K, V]) Filter(f func(K, V) bool) Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]()
	}
	return From2(func(yield func(K, V) bool) {
		for k, v := range s.seq {
			if f(k, v) && !yield(k, v) {
				return
			}
		}
	})
}

// Take keeps at most n pairs.
func (s Seq2[K, V]) Take(n int) Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]()
	}
	return From2(func(yield func(K, V) bool) {
		for k, v := range s.seq {
			if n <= 0 {
				return
			}
			n--
			if !yield(k, v) {
				return
			}
		}
	})
}

// Drop discards the first n pairs.
func (s Seq2[K, V]) Drop(n int) Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]()
	}
	return From2(func(yield func(K, V) bool) {
		for k, v := range s.seq {
			if n > 0 {
				n--
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	})
}

// TakeWhile yields pairs from the start while pred is true.
func (s Seq2[K, V]) TakeWhile(pred func(K, V) bool) Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]()
	}
	return From2(func(yield func(K, V) bool) {
		for k, v := range s.seq {
			if !pred(k, v) {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	})
}

// DropWhile discards pairs from the start while pred is true.
func (s Seq2[K, V]) DropWhile(pred func(K, V) bool) Seq2[K, V] {
	if s.seq == nil {
		return emptySeq2[K, V]()
	}
	return From2(func(yield func(K, V) bool) {
		drop := true
		for k, v := range s.seq {
			if drop && pred(k, v) {
				continue
			}
			drop = false
			if !yield(k, v) {
				return
			}
		}
	})
}

// Collect eagerly evaluates the pair sequence into a slice of KeyValue pairs.
func (s Seq2[K, V]) Collect() []KeyValue[K, V] {
	if s.seq == nil {
		return nil
	}
	var out []KeyValue[K, V]
	for k, v := range s.seq {
		out = append(out, KeyValue[K, V]{K: k, V: v})
	}
	return out
}

// Tap inspects every outcome as it flows through, then yields it unchanged.
// This is where fallible pipelines log their errors.
func (s Result[V]) Tap(f func(V, error)) Result[V] {
	if s.seq == nil {
		return emptyResult[V]()
	}
	return FromResult(func(yield func(V, error) bool) {
		for v, err := range s.seq {
			f(v, err)
			if !yield(v, err) {
				return
			}
		}
	})
}

// Take keeps at most n outcome pairs, stopping the upstream pull that early.
func (s Result[V]) Take(n int) Result[V] {
	if s.seq == nil {
		return emptyResult[V]()
	}
	return FromResult(func(yield func(V, error) bool) {
		for v, err := range s.seq {
			if n <= 0 {
				return
			}
			n--
			if !yield(v, err) {
				return
			}
		}
	})
}

// Filter keeps the outcome pairs for which f returns true.
func (s Result[V]) Filter(f func(V, error) bool) Result[V] {
	if s.seq == nil {
		return emptyResult[V]()
	}
	return FromResult(func(yield func(V, error) bool) {
		for v, err := range s.seq {
			if f(v, err) && !yield(v, err) {
				return
			}
		}
	})
}

// IgnoreErr drops the error pairs and yields only the successful values as a
// plain sequence from here on: the errors are skipped and the value is kept.
func (s Result[V]) IgnoreErr() Seq[V] {
	if s.seq == nil {
		return emptySeq[V]()
	}
	return From(func(yield func(V) bool) {
		for v, err := range s.seq {
			if err == nil && !yield(v) {
				return
			}
		}
	})
}

// Errors yields the error of each failed pair, for logging or counting.
func (s Result[V]) Errors() Seq[error] {
	if s.seq == nil {
		return emptySeq[error]()
	}
	return From(func(yield func(error) bool) {
		for _, err := range s.seq {
			if err != nil && !yield(err) {
				return
			}
		}
	})
}

// CollectErr eagerly evaluates the fallible sequence, returning every
// successful value and all failures joined into one error with errors.Join,
// or nil if nothing failed.
func (s Result[V]) CollectErr() ([]V, error) {
	var vals []V
	var errs []error
	for v, err := range s.iter() {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		vals = append(vals, v)
	}
	return vals, errors.Join(errs...)
}

// Count returns the number of outcome pairs.
func (s Result[V]) Count() int {
	n := 0
	for range s.iter() {
		n++
	}
	return n
}

// Seq exposes the underlying standard pair iterator, for ranging directly:
//
//	for v, err := range res.Seq() { ... }
func (s Result[V]) Seq() stditer.Seq2[V, error] { return s.seq }

// ToMap eagerly collects a pair sequence into a map.
func ToMap[K comparable, V any](s Seq2[K, V]) map[K]V { return maps.Collect(s.iter()) }
