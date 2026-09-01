package iter

// Uniq keeps only the first occurrence of each equal element, deduplicating
// the sequence. Like Max/Min/Sum, it is a package function because its
// comparable constraint applies to the element type, which a method cannot
// re-restrict; the key-based variant UniqByFunc handles non-comparable
// elements as a method.
func Uniq[T comparable](s Seq[T]) Seq[T] {
	return From(func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range s.seq {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	})
}

// UniqByFunc keeps only the first element of each run of equal keys,
// deduplicating by the comparable key produced by key.
func (s Seq[T]) UniqByFunc[K comparable](key func(T) K) Seq[T] {
	return From(func(yield func(T) bool) {
		seen := make(map[K]struct{})
		for v := range s.seq {
			k := key(v)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	})
}

// GroupByFunc buckets elements by the key produced by f, returning a map of
// the buckets.
func (s Seq[T]) GroupByFunc[K comparable](f func(T) K) map[K][]T {
	m := make(map[K][]T)
	for v := range s.seq {
		k := f(v)
		m[k] = append(m[k], v)
	}
	return m
}

// KeyByFunc pivots a single element per key: the last element for each key
// wins.
func (s Seq[T]) KeyByFunc[K comparable](key func(T) K) map[K]T {
	m := make(map[K]T)
	for v := range s.seq {
		m[key(v)] = v
	}
	return m
}
