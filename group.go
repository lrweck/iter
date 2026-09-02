package iter

// Uniq keeps only the first occurrence of each equal element, deduplicating
// the sequence. Like Max/Min/Sum, it is a package function because its
// comparable constraint applies to the element type, which a method cannot
// re-restrict; the key-based variant UniqBy handles non-comparable
// elements as a method.
func Uniq[T comparable](s Seq[T]) Seq[T] {
	return From(func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range s.iter() {
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

// UniqBy keeps only the first element for each distinct key produced by key.
func (s Seq[T]) UniqBy[K comparable](key func(T) K) Seq[T] {
	return From(func(yield func(T) bool) {
		seen := make(map[K]struct{})
		for v := range s.iter() {
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

// GroupBy buckets elements by the key produced by f, returning a map of
// the buckets.
func (s Seq[T]) GroupBy[K comparable](f func(T) K) map[K][]T {
	m := make(map[K][]T)
	for v := range s.iter() {
		k := f(v)
		m[k] = append(m[k], v)
	}
	return m
}

// KeyBy pivots a single element per key: the last element for each key
// wins.
func (s Seq[T]) KeyBy[K comparable](key func(T) K) map[K]T {
	m := make(map[K]T)
	for v := range s.iter() {
		m[key(v)] = v
	}
	return m
}
