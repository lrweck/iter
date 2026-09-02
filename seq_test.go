package iter

import (
	"slices"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"square", func(t *testing.T) {
			checkEqual(t, Of(1, 2, 3).Map(func(v int) int { return v * v }).Collect(), []int{1, 4, 9})
		}},
		{"empty", func(t *testing.T) {
			checkEqual(t, Of[int]().Map(func(v int) int { return v * v }).Collect(), nil)
		}},
		{"nil receiver", func(t *testing.T) {
			var s Seq[int]
			checkEqual(t, s.Map(func(v int) int { return v * v }).Collect(), nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"even", func(t *testing.T) {
			checkEqual(t, Of(1, 2, 3, 4).Filter(func(v int) bool { return v%2 == 0 }).Collect(), []int{2, 4})
		}},
		{"none kept", func(t *testing.T) {
			checkEqual(t, Of(1, 3).Filter(func(v int) bool { return v%2 == 0 }).Collect(), nil)
		}},
		{"nil receiver", func(t *testing.T) {
			var s Seq[int]
			checkEqual(t, s.Filter(func(v int) bool { return v%2 == 0 }).Collect(), nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSkipErr(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"skip x", func(t *testing.T) {
			skip := func(s string) error {
				if s == "x" {
					return errSentinel
				}
				return nil
			}
			checkEqual(t, Of("a", "x", "bb").SkipErr(skip).Collect(), []string{"a", "bb"})
		}},
		{"none fail", func(t *testing.T) {
			checkEqual(t, Of("a", "b").SkipErr(func(string) error { return nil }).Collect(), []string{"a", "b"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestFlatMap(t *testing.T) {
	flat := func(s string) Seq[rune] { return Of(rune(s[0]), rune(s[1])) }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"chars", func(t *testing.T) {
			checkEqual(t, Of("ab", "cd").FlatMap(flat).Collect(), []rune{'a', 'b', 'c', 'd'})
		}},
		{"last char twice", func(t *testing.T) {
			twice := func(s string) Seq[rune] {
				c := rune(s[len(s)-1])
				return Of(c, c)
			}
			checkEqual(t, Of("ab", "c").FlatMap(twice).Collect(), []rune{'b', 'b', 'c', 'c'})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestTakeDrop(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"drop 2 take 2", func(t *testing.T) {
			checkEqual(t, Of(1, 2, 3, 4, 5).Drop(2).Take(2).Collect(), []int{3, 4})
		}},
		{"take none", func(t *testing.T) {
			checkEqual(t, Of(1, 2).Drop(0).Take(0).Collect(), nil)
		}},
		{"drop all", func(t *testing.T) {
			checkEqual(t, Of(1, 2).Drop(3).Take(0).Collect(), nil)
		}},
		{"take beyond", func(t *testing.T) {
			checkEqual(t, Of(1, 2).Drop(0).Take(5).Collect(), []int{1, 2})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestTakeDropWhile(t *testing.T) {
	lt4 := func(v int) bool { return v < 4 }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"take while", func(t *testing.T) {
			checkEqual(t, Of(1, 2, 3, 4, 1).TakeWhile(lt4).Collect(), []int{1, 2, 3})
		}},
		{"drop while", func(t *testing.T) {
			checkEqual(t, Of(1, 2, 3, 4, 1).DropWhile(lt4).Collect(), []int{4, 1})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestChunk(t *testing.T) {
	eq := func(a, b []int) bool { return slices.Equal(a, b) }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"remainder", func(t *testing.T) {
			got := Chunk(Of(1, 2, 3, 4, 5), 3).Collect()
			if !slices.EqualFunc(got, [][]int{{1, 2, 3}, {4, 5}}, eq) {
				t.Fatalf("chunks got %v, want [[1 2 3] [4 5]]", got)
			}
		}},
		{"exact multiple", func(t *testing.T) {
			got := Chunk(Of(1, 2, 3, 4, 5, 6), 3).Collect()
			if !slices.EqualFunc(got, [][]int{{1, 2, 3}, {4, 5, 6}}, eq) {
				t.Fatalf("chunks got %v, want [[1 2 3] [4 5 6]]", got)
			}
		}},
		{"empty", func(t *testing.T) {
			if got := Chunk(Of[int](), 3).Collect(); len(got) != 0 {
				t.Fatalf("chunks got %v, want nil", got)
			}
		}},
		{"chunk of one", func(t *testing.T) {
			got := Chunk(Of(1, 2, 3), 1).Collect()
			if !slices.EqualFunc(got, [][]int{{1}, {2}, {3}}, eq) {
				t.Fatalf("chunks got %v, want [[1] [2] [3]]", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSeq(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"ranges directly", func(t *testing.T) {
			var got []int
			for v := range Of(1, 2, 3).Seq() {
				got = append(got, v)
			}
			checkEqual(t, got, []int{1, 2, 3})
		}},
		{"nil Seq is empty", func(t *testing.T) {
			var s Seq[int]
			checkEqual(t, s.Collect(), nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestChunkNegativePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Chunk with n<=0 did not panic")
		}
	}()
	Chunk(Of(1, 2), 0)
}

func TestTap(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"logs and passes through", func(t *testing.T) {
			logged := 0
			got := Of(1, 2, 3).Tap(func(v int) { logged += v }).Collect()
			if logged != 6 {
				t.Fatalf("Tap logged %d, want 6", logged)
			}
			checkEqual(t, got, []int{1, 2, 3})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestGenerators(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"range", func(t *testing.T) {
			checkEqual(t, Range(2, 5, 1).Collect(), []int{2, 3, 4})
		}},
		{"range empty", func(t *testing.T) {
			checkEqual(t, Range(3, 3, 1).Collect(), nil)
		}},
		{"range step 2", func(t *testing.T) {
			checkEqual(t, Range(0, 10, 2).Collect(), []int{0, 2, 4, 6, 8})
		}},
		{"range negative step", func(t *testing.T) {
			checkEqual(t, Range(5, 1, -1).Collect(), []int{5, 4, 3, 2})
		}},
		{"range negative step 2", func(t *testing.T) {
			checkEqual(t, Range(10, 0, -3).Collect(), []int{10, 7, 4, 1})
		}},
		{"range empty by step", func(t *testing.T) {
			checkEqual(t, Range(0, 10, -1).Collect(), nil)
		}},
		{"range panics on zero step", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("Range with step=0 did not panic")
				}
			}()
			Range(0, 10, 0)
		}},
		{"repeat", func(t *testing.T) {
			checkEqual(t, Repeat(3, 7).Collect(), []int{7, 7, 7})
		}},
		{"repeat zero", func(t *testing.T) {
			checkEqual(t, Repeat(0, 7).Collect(), nil)
		}},
		{"repeatBy", func(t *testing.T) {
			checkEqual(t, RepeatBy(4, func(i int) int { return i * i }).Collect(), []int{0, 1, 4, 9})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestConcat(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"three", func(t *testing.T) {
			checkEqual(t, Concat(Of(1, 2), Of(3), Of(4, 5)).Collect(), []int{1, 2, 3, 4, 5})
		}},
		{"empty sources", func(t *testing.T) {
			checkEqual(t, Concat(Of[int](), Of(1)).Collect(), []int{1})
		}},
		{"no sources", func(t *testing.T) {
			if got := Concat[int]().Collect(); len(got) != 0 {
				t.Fatalf("Concat got %v, want empty", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestStops(t *testing.T) {
	flat := func(s string) Seq[rune] { return Of(rune(s[0]), rune(s[1])) }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"tap", func(t *testing.T) { checkStops(t, Of(1, 2, 3).Tap(func(int) {}).Take(1).Count()) }},
		{"skipErr", func(t *testing.T) {
			checkStops(t, Of("a", "b").SkipErr(func(string) error { return nil }).Take(1).Count())
		}},
		{"nested take", func(t *testing.T) { checkStops(t, Of(1, 2, 3).Take(3).Take(1).Count()) }},
		{"chunk", func(t *testing.T) { checkStops(t, Chunk(Of(1, 2, 3, 4), 2).Take(1).Count()) }},
		{"flatMap", func(t *testing.T) { checkStops(t, Of("ab", "cd").FlatMap(flat).Take(1).Count()) }},
		{"concat", func(t *testing.T) { checkStops(t, Concat(Of(1, 2), Of(3, 4)).Take(1).Count()) }},
		{"mapErr ok", func(t *testing.T) { checkStops(t, Of("aa", "bb", "x").MapErr(parseLen).IgnoreErr().Take(1).Count()) }},
		{"mapErr fail", func(t *testing.T) { checkStops(t, Of("aa", "x").MapErr(parseLen).Errors().Take(1).Count()) }},
		{"flatMapErr ok", func(t *testing.T) {
			checkStops(t, Of("ab", "cd").FlatMapErr(func(s string) (Seq[rune], error) {
				return Of(rune(s[0]), rune(s[1])), nil
			}).IgnoreErr().Take(1).Count())
		}},
		{"flatMapErr fail", func(t *testing.T) {
			checkStops(t, Of("aa", "x").FlatMapErr(func(s string) (Seq[int], error) {
				if s == "x" {
					return Seq[int]{}, errSentinel
				}
				return Of(len(s)), nil
			}).Errors().Take(1).Count())
		}},
		{"enumerate+keys", func(t *testing.T) { checkStops(t, Of("a", "b").Enumerate().Keys().Take(1).Count()) }},
		{"values", func(t *testing.T) { checkStops(t, Of("a", "b").Enumerate().Values().Take(1).Count()) }},
		{"zip+keys", func(t *testing.T) { checkStops(t, Of(1, 2, 3).Zip(Of("a", "b", "c")).Keys().Take(1).Count()) }},
		{"seq2 tap break", func(t *testing.T) {
			n := 0
			for range Of(1, 2).Enumerate().Tap(func(int, int) {}).Seq() {
				n++
				break
			}
			checkStops(t, n)
		}},
		{"r tap break", func(t *testing.T) {
			n := 0
			for range Of(1, 2).MapErr(func(i int) (int, error) { return i, nil }).Tap(func(int, error) {}).Seq() {
				n++
				break
			}
			checkStops(t, n)
		}},
		{"seq2 filter", func(t *testing.T) {
			checkStops(t, Of(1, 2, 3).Enumerate().Filter(func(k, v int) bool { return v > 1 }).Take(1).Count())
		}},
		{"seq2 take", func(t *testing.T) {
			checkStops(t, Of(1, 2, 3).Enumerate().Take(1).Count())
		}},
		{"seq2 drop", func(t *testing.T) {
			checkStops(t, Of(1, 2, 3).Enumerate().Drop(2).Take(1).Count())
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func checkStops(t *testing.T, n int) {
	t.Helper()
	if n != 1 {
		t.Fatalf("observed %d elements, want 1", n)
	}
}

func TestLazy(t *testing.T) {
	calls := 0
	Of(1, 2, 3, 4, 5).Map(func(v int) int {
		calls++
		return v
	}).Take(2).Collect()
	if calls != 3 {
		t.Fatalf("Map ran %d times, want 3", calls)
	}

	fcalls := 0
	Of(1, 2, 3, 4).Filter(func(v int) bool {
		fcalls++
		return v%2 == 0
	}).Take(1).Collect()
	if fcalls != 4 {
		t.Fatalf("Filter ran %d times, want 4", fcalls)
	}
}

func TestNilReceiver(t *testing.T) {
	var s Seq[int]

	// All pipeline methods should return empty, not panic.
	checkEqual(t, s.Map(func(v int) int { return v * 2 }).Collect(), nil)
	checkEqual(t, s.Filter(func(v int) bool { return true }).Collect(), nil)
	checkEqual(t, s.FlatMap(func(v int) Seq[int] { return Of(v) }).Collect(), nil)
	checkEqual(t, s.Take(5).Collect(), nil)
	checkEqual(t, s.Drop(2).Collect(), nil)
	checkEqual(t, s.TakeWhile(func(v int) bool { return true }).Collect(), nil)
	checkEqual(t, s.DropWhile(func(v int) bool { return false }).Collect(), nil)
	checkEqual(t, s.SkipErr(func(int) error { return nil }).Collect(), nil)
	checkEqual(t, s.Tap(func(int) {}).Collect(), nil)

	// Enumerate, Zip, MapErr, FlatMapErr.
	checkEqual(t, s.Enumerate().Keys().Collect(), nil)
	checkEqual(t, s.Enumerate().Values().Collect(), nil)
	checkEqual(t, s.Zip(Of(1)).Keys().Collect(), nil)
	checkEqual(t, s.MapErr(func(v int) (int, error) { return v, nil }).IgnoreErr().Collect(), nil)
	checkEqual(t, s.FlatMapErr(func(v int) (Seq[int], error) { return Of(v), nil }).Errors().Collect(), []error{})
}
