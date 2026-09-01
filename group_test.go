package iter

import (
	"maps"
	"slices"
	"testing"
)

type user struct {
	name string
	id   int
}

func TestUniq(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"dedupe", func(t *testing.T) {
			checkEqual(t, Uniq(Of(1, 2, 2, 1, 3)).Collect(), []int{1, 2, 3})
		}},
		{"empty", func(t *testing.T) {
			checkEqual(t, Uniq(Of[int]()).Collect(), nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestUniqByFunc(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"mod three", func(t *testing.T) {
			checkEqual(t, Of(0, 1, 2, 3, 4, 5).UniqByFunc(func(v int) int { return v % 3 }).Collect(), []int{0, 1, 2})
		}},
		{"non-comparable via key", func(t *testing.T) {
			checkEqual(t, Of(1).UniqByFunc(func(v int) int { return v }).Collect(), []int{1})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestGroupByFunc(t *testing.T) {
	eq := func(a, b []string) bool { return slices.Equal(a, b) }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"buckets", func(t *testing.T) {
			got := Of("a", "b", "a").GroupByFunc(func(s string) string { return s })
			if !maps.EqualFunc(got, map[string][]string{"a": {"a", "a"}, "b": {"b"}}, eq) {
				t.Fatalf("GroupByFunc got %v, want map[a:[a a] b:[b]]", got)
			}
		}},
		{"empty", func(t *testing.T) {
			got := Of[string]().GroupByFunc(func(s string) string { return s })
			if !maps.EqualFunc(got, map[string][]string{}, eq) {
				t.Fatalf("GroupByFunc got %v, want empty map", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestKeyByFunc(t *testing.T) {
	pivot := func(u user) string { return u.name }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"pivot", func(t *testing.T) {
			got := Of(user{"a", 1}, user{"b", 2}).KeyByFunc(pivot)
			if !maps.Equal(got, map[string]user{"a": {"a", 1}, "b": {"b", 2}}) {
				t.Fatalf("KeyByFunc got %v, want pivot map", got)
			}
		}},
		{"last wins", func(t *testing.T) {
			got := Of(user{"a", 1}, user{"a", 3}).KeyByFunc(pivot)
			if !maps.Equal(got, map[string]user{"a": {"a", 3}}) {
				t.Fatalf("KeyByFunc got %v, want last wins", got)
			}
		}},
		{"empty", func(t *testing.T) {
			got := Of[user]().KeyByFunc(pivot)
			if !maps.Equal(got, map[string]user{}) {
				t.Fatalf("KeyByFunc got %v, want empty map", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
