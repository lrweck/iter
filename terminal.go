package iter

import "slices"

// Collect eagerly evaluates the sequence into a slice.
func (s Seq[T]) Collect() []T { return slices.Collect(s.iter()) }

// First returns the first element, or ok=false if the sequence is empty.
func (s Seq[T]) First() (T, bool) {
	var zero T
	if s.seq == nil {
		return zero, false
	}
	for v := range s.iter() {
		return v, true
	}
	return zero, false
}

// Count returns the number of elements.
func (s Seq[T]) Count() int {
	n := 0
	for range s.iter() {
		n++
	}
	return n
}

// CountBy counts the elements for which pred is true.
func (s Seq[T]) CountBy(pred func(T) bool) int {
	n := 0
	for v := range s.iter() {
		if pred(v) {
			n++
		}
	}
	return n
}

// Some reports whether pred is true for at least one element, stopping early.
func (s Seq[T]) Some(pred func(T) bool) bool {
	for v := range s.iter() {
		if pred(v) {
			return true
		}
	}
	return false
}

// Every reports whether pred is true for every element, stopping early.
func (s Seq[T]) Every(pred func(T) bool) bool {
	for v := range s.iter() {
		if !pred(v) {
			return false
		}
	}
	return true
}

// None reports whether pred is false for every element.
func (s Seq[T]) None(pred func(T) bool) bool {
	return !s.Some(pred)
}

// Find returns the first element for which pred is true, or ok=false.
func (s Seq[T]) Find(pred func(T) bool) (T, bool) {
	for v := range s.iter() {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Contains reports whether the sequence contains v. Like Uniq, it is a
// package function because comparable applies to the element type; predicate
// checks are covered by Some.
func Contains[T comparable](s Seq[T], v T) bool {
	for x := range s.iter() {
		if x == v {
			return true
		}
	}
	return false
}

// Last returns the final element, or ok=false if the sequence is empty.
func (s Seq[T]) Last() (T, bool) {
	var last T
	ok := false
	for v := range s.iter() {
		last, ok = v, true
	}
	return last, ok
}

// Reduce folds the sequence into a single value starting from init.
func (s Seq[T]) Reduce[U any](init U, f func(U, T) U) U {
	acc := init
	for v := range s.iter() {
		acc = f(acc, v)
	}
	return acc
}

// ReduceErr is Reduce with a fallible step. It stops at the first error,
// returning the zero value of U and the error. The remaining elements are
// not processed.
func (s Seq[T]) ReduceErr[U any](init U, f func(U, T) (U, error)) (U, error) {
	acc := init
	for v := range s.iter() {
		next, err := f(acc, v)
		if err != nil {
			var zero U
			return zero, err
		}
		acc = next
	}
	return acc, nil
}

// ForEach drives the sequence, calling f for every element.
func (s Seq[T]) ForEach(f func(T)) {
	for v := range s.iter() {
		f(v)
	}
}
