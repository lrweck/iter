package iter

import (
	"testing"
)

// =============================================================================
// Baselines: raw slice operations (no iterator overhead)
// =============================================================================

func BenchmarkOf_Collect_10(b *testing.B) {
	for b.Loop() {
		Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Collect()
	}
}

func BenchmarkOf_Collect_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Collect()
	}
}

// BenchmarkRaw_Map vs iter.Map
func BenchmarkRaw_Map_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		out := make([]int, len(s))
		for i, v := range s {
			out[i] = v * 2
		}
		_ = out
	}
}

func BenchmarkIter_Map_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Map(func(v int) int { return v * 2 }).Collect()
	}
}

// BenchmarkRaw_Filter vs iter.Filter
func BenchmarkRaw_Filter_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		out := make([]int, 0, len(s))
		for _, v := range s {
			if v%2 == 0 {
				out = append(out, v)
			}
		}
		_ = out
	}
}

func BenchmarkIter_Filter_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Filter(func(v int) bool { return v%2 == 0 }).Collect()
	}
}

// =============================================================================
// Pipeline depth: how much overhead does each layer add?
// =============================================================================

func BenchmarkPipeline_Map_1(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Map(func(v int) int { return v * 2 }).Collect()
	}
}

func BenchmarkPipeline_Map_3(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).
			Map(func(v int) int { return v * 2 }).
			Map(func(v int) int { return v + 1 }).
			Map(func(v int) int { return v * 3 }).
			Collect()
	}
}

func BenchmarkPipeline_Map_5(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).
			Map(func(v int) int { return v * 2 }).
			Map(func(v int) int { return v + 1 }).
			Map(func(v int) int { return v * 3 }).
			Map(func(v int) int { return v - 1 }).
			Map(func(v int) int { return v / 2 }).
			Collect()
	}
}

// =============================================================================
// Common real-world pipelines
// =============================================================================

func BenchmarkPipeline_MapFilterTake(b *testing.B) {
	s := make([]int, 10000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).
			Filter(func(v int) bool { return v%2 == 0 }).
			Map(func(v int) int { return v * v }).
			Take(100).
			Collect()
	}
}

func BenchmarkPipeline_Raw_MapFilterTake(b *testing.B) {
	s := make([]int, 10000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		out := make([]int, 0, 100)
		n := 0
		for _, v := range s {
			if v%2 == 0 {
				out = append(out, v*v)
				n++
				if n == 100 {
					break
				}
			}
		}
		_ = out
	}
}

// =============================================================================
// Take / Drop early stop
// =============================================================================

func BenchmarkTake_10_of_10000(b *testing.B) {
	s := make([]int, 10000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Take(10).Collect()
	}
}

func BenchmarkDrop_9990_of_10000(b *testing.B) {
	s := make([]int, 10000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Drop(9990).Collect()
	}
}

// =============================================================================
// FlatMap
// =============================================================================

func BenchmarkFlatMap_10x100(b *testing.B) {
	for b.Loop() {
		Range(0, 10, 1).
			FlatMap(func(i int) Seq[int] {
				return Range(0, 100, 1)
			}).
			Collect()
	}
}

// =============================================================================
// Enumerate + Keys/Values (Seq2 pipeline)
// =============================================================================

func BenchmarkEnumerate_Keys_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Enumerate().Keys().Collect()
	}
}

func BenchmarkEnumerate_Collect_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Enumerate().Collect()
	}
}

// =============================================================================
// Result pipeline: MapErr + IgnoreErr + Collect
// =============================================================================

func BenchmarkMapErr_AllOK_1000(b *testing.B) {
	s := make([]string, 1000)
	for i := range s {
		s[i] = "a"
	}
	for b.Loop() {
		Of(s...).
			MapErr(func(v string) (int, error) { return len(v), nil }).
			IgnoreErr().
			Collect()
	}
}

func BenchmarkMapErr_AllOK_Raw_1000(b *testing.B) {
	s := make([]string, 1000)
	for i := range s {
		s[i] = "a"
	}
	for b.Loop() {
		out := make([]int, 0, len(s))
		for _, v := range s {
			out = append(out, len(v))
		}
	}
}

// =============================================================================
// Generators
// =============================================================================

func BenchmarkRange_Collect_1000(b *testing.B) {
	for b.Loop() {
		Range(0, 1000, 1).Collect()
	}
}

func BenchmarkRepeat_Collect_1000(b *testing.B) {
	for b.Loop() {
		Repeat(1000, 42).Collect()
	}
}

func BenchmarkConcat_4x250(b *testing.B) {
	a := make([]int, 250)
	for i := range a {
		a[i] = i
	}
	for b.Loop() {
		Concat(Of(a...), Of(a...), Of(a...), Of(a...)).Collect()
	}
}

// =============================================================================
// Memory allocation: operations that should not allocate
// =============================================================================

func BenchmarkFirst_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).First()
	}
}

func BenchmarkCount_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Count()
	}
}

func BenchmarkSome_Early(b *testing.B) {
	// First element matches — early stop.
	s := make([]int, 1000)
	s[0] = -1 // sentinel
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Of(s...).Some(func(v int) bool { return v < 0 })
	}
}

func BenchmarkSum_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Sum(Of(s...))
	}
}

func BenchmarkSum_Raw_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		var total int
		for _, v := range s {
			total += v
		}
	}
}

// =============================================================================
// Uniq — map allocation overhead
// =============================================================================

func BenchmarkUniq_AllDistinct_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	for b.Loop() {
		Uniq(Of(s...)).Collect()
	}
}

func BenchmarkUniq_AllSame_1000(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = 42
	}
	for b.Loop() {
		Uniq(Of(s...)).Collect()
	}
}
