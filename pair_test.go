package iter

import (
	"errors"
	"maps"
	"slices"
	"testing"
)

func TestEnumerate(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"two", func(t *testing.T) {
			en := Of("a", "b").Enumerate()
			var ks []int
			var vs []string
			for k, v := range en.Seq() {
				ks = append(ks, k)
				vs = append(vs, v)
			}
			checkEqual(t, ks, []int{0, 1})
			checkEqual(t, vs, []string{"a", "b"})
			checkEqual(t, en.Values().Collect(), []string{"a", "b"})
		}},
		{"empty", func(t *testing.T) {
			en := Of[string]().Enumerate()
			var ks []int
			var vs []string
			for k, v := range en.Seq() {
				ks = append(ks, k)
				vs = append(vs, v)
			}
			checkEqual(t, ks, nil)
			checkEqual(t, vs, nil)
			checkEqual(t, en.Values().Collect(), nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSeq2Inspection(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"tap count keys", func(t *testing.T) {
			seen := 0
			res := Of("a", "b").Enumerate().Tap(func(i int, s string) { seen++ })
			if n := res.Count(); n != 2 || seen != 2 {
				t.Fatalf("Seq2 Count=%d seen=%d, want 2/2", n, seen)
			}
			checkEqual(t, res.Keys().Collect(), []int{0, 1})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestZip(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"shorter right", func(t *testing.T) {
			checkEqual(t, zipOut(Of(1, 2, 3).Zip(Of("a", "b"))), []pair{{1, "a"}, {2, "b"}})
		}},
		{"equal", func(t *testing.T) {
			checkEqual(t, zipOut(Of(1, 2).Zip(Of("a", "b"))), []pair{{1, "a"}, {2, "b"}})
		}},
		{"shorter left", func(t *testing.T) {
			checkEqual(t, zipOut(Of(1).Zip(Of("a", "b", "c"))), []pair{{1, "a"}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func zipOut(z Seq2[int, string]) []pair {
	var out []pair
	for k, v := range z.Seq() {
		out = append(out, pair{k, v})
	}
	return out
}

func TestToMap(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"enumerate", func(t *testing.T) {
			got := ToMap(Of("a", "b").Enumerate())
			if !maps.Equal(got, map[int]string{0: "a", 1: "b"}) {
				t.Fatalf("ToMap got %v, want map[0:a 1:b]", got)
			}
		}},
		{"empty", func(t *testing.T) {
			got := ToMap(Of[string]().Enumerate())
			if !maps.Equal(got, map[int]string{}) {
				t.Fatalf("ToMap got %v, want empty map", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestCollectErr(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"all ok", func(t *testing.T) {
			vals, err := Of("a", "bb").MapErr(parseLen).CollectErr()
			checkEqual(t, vals, []int{1, 2})
			if err != nil {
				t.Fatalf("CollectErr got %v, want nil", err)
			}
		}},
		{"some fail", func(t *testing.T) {
			vals, err := Of("aa", "b", "x", "cc").MapErr(parseLen).CollectErr()
			checkEqual(t, vals, []int{2, 1, 2})
			if !errors.Is(err, errSentinel) {
				t.Fatalf("CollectErr got %v, want sentinel", err)
			}
		}},
		{"two join", func(t *testing.T) {
			vals, err := Of("x", "a", "x").MapErr(parseLen).CollectErr()
			checkEqual(t, vals, []int{1})
			if err == nil || err.Error() != "bad input\nbad input" {
				t.Fatalf("CollectErr got %v, want 2 joined", err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestFlatMapErr(t *testing.T) {
	flat := func(v string) (Seq[int], error) {
		if v == "x" {
			return Seq[int]{}, errSentinel
		}
		return Of(len(v), len(v)), nil
	}
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"ok and fail", func(t *testing.T) {
			vals, err := Of("aa", "x").FlatMapErr(flat).CollectErr()
			checkEqual(t, vals, []int{2, 2})
			if !errors.Is(err, errSentinel) {
				t.Fatalf("FlatMapErr got %v, want sentinel", err)
			}
		}},
		{"all ok", func(t *testing.T) {
			vals, err := Of("ab").FlatMapErr(flat).CollectErr()
			checkEqual(t, vals, []int{2, 2})
			if err != nil {
				t.Fatalf("FlatMapErr got %v, want nil", err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestIgnoreErr(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"chain continues", func(t *testing.T) {
			got := Of("aa", "b", "x", "cc").MapErr(parseLen).IgnoreErr().Filter(func(v int) bool { return v > 1 }).Collect()
			checkEqual(t, got, []int{2, 2})
		}},
		{"no filter", func(t *testing.T) {
			checkEqual(t, Of("a", "x", "bb").MapErr(parseLen).IgnoreErr().Collect(), []int{1, 2})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"one", func(t *testing.T) {
			if got := Of("aa", "x", "bb").MapErr(parseLen).Errors().Count(); got != 1 {
				t.Fatalf("Errors got %d, want 1", got)
			}
		}},
		{"none", func(t *testing.T) {
			if got := Of("aa", "bb").MapErr(parseLen).Errors().Count(); got != 0 {
				t.Fatalf("Errors got %d, want 0", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestFromMap(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"from map", func(t *testing.T) {
			m := map[int]string{1: "a", 2: "b"}
			if got := ToMap(FromMap(m)); !maps.Equal(got, m) {
				t.Fatalf("FromMap got %v, want %v", got, m)
			}
		}},
		{"from keys", func(t *testing.T) {
			m := map[int]string{1: "a", 2: "b"}
			ks := FromKeys(m).Collect()
			slices.Sort(ks)
			checkEqual(t, ks, []int{1, 2})
		}},
		{"from values", func(t *testing.T) {
			m := map[int]string{1: "a", 2: "b"}
			vs := FromValues(m).Collect()
			slices.Sort(vs)
			checkEqual(t, vs, []string{"a", "b"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestResultTake(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"stops the pull before the failure", func(t *testing.T) {
			vals, err := Of("aa", "b", "x", "cc").MapErr(parseLen).Take(2).CollectErr()
			checkEqual(t, vals, []int{2, 1})
			if err != nil {
				t.Fatalf("Take+CollectErr got %v, want nil", err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestResultFilter(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"keep only failures", func(t *testing.T) {
			keepErr := func(v int, err error) bool { return err != nil }
			if got := Of("a", "x", "bb").MapErr(parseLen).Filter(keepErr).Errors().Count(); got != 1 {
				t.Fatalf("Filter kept %d failures, want 1", got)
			}
		}},
		{"keep only successes", func(t *testing.T) {
			keepOK := func(v int, err error) bool { return err == nil }
			got := Of("a", "x", "bb").MapErr(parseLen).Filter(keepOK).IgnoreErr().Collect()
			checkEqual(t, got, []int{1, 2})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRTap(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"sees all and the ok count", func(t *testing.T) {
			seen, ok := 0, 0
			Of("aa", "b", "x", "cc").MapErr(parseLen).Tap(func(v int, err error) {
				seen++
				if err == nil {
					ok++
				}
			}).Count()
			if seen != 4 || ok != 3 {
				t.Fatalf("Result.Tap saw %d total, %d ok, want 4/3", seen, ok)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRSeq(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"ranges directly", func(t *testing.T) {
			var vals []int
			for v, err := range Of("aa", "x").MapErr(parseLen).Seq() {
				if err == nil {
					vals = append(vals, v)
				}
			}
			checkEqual(t, vals, []int{2})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
