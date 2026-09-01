package iter

import (
	"errors"
	"testing"
)

func TestFirst(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"first", func(t *testing.T) {
			if got, ok := Of(9, 8).First(); got != 9 || !ok {
				t.Fatalf("First got %d,%v want 9,true", got, ok)
			}
		}},
		{"empty", func(t *testing.T) {
			if got, ok := Of[int]().First(); got != 0 || ok {
				t.Fatalf("First got %d,%v want 0,false", got, ok)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"three", func(t *testing.T) {
			if got := Of(1, 2, 3).Count(); got != 3 {
				t.Fatalf("Count got %d, want 3", got)
			}
		}},
		{"empty", func(t *testing.T) {
			if got := Of[int]().Count(); got != 0 {
				t.Fatalf("Count got %d, want 0", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"product", func(t *testing.T) {
			if got := Of(2, 3, 4).Reduce(1, func(a, b int) int { return a * b }); got != 24 {
				t.Fatalf("Reduce got %d, want 24", got)
			}
		}},
		{"empty keeps init", func(t *testing.T) {
			if got := Of[int]().Reduce(5, func(a, b int) int { return a * b }); got != 5 {
				t.Fatalf("Reduce got %d, want 5", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestEach(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"sum", func(t *testing.T) {
			sum := 0
			Of(1, 2, 3).Each(func(v int) { sum += v })
			if sum != 6 {
				t.Fatalf("Each summed to %d, want 6", sum)
			}
		}},
		{"empty", func(t *testing.T) {
			sum := 0
			Of[int]().Each(func(v int) { sum += v })
			if sum != 0 {
				t.Fatalf("Each summed to %d, want 0", sum)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestPredicates(t *testing.T) {
	even := func(v int) bool { return v%2 == 0 }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"some", func(t *testing.T) {
			if got := Of(1, 3, 4).Some(even); !got {
				t.Fatalf("Some got %v, want true", got)
			}
		}},
		{"some none", func(t *testing.T) {
			if got := Of(1, 3).Some(even); got {
				t.Fatalf("Some got %v, want false", got)
			}
		}},
		{"every", func(t *testing.T) {
			if got := Of(2, 4).Every(even); !got {
				t.Fatalf("Every got %v, want true", got)
			}
		}},
		{"every not", func(t *testing.T) {
			if got := Of(2, 3).Every(even); got {
				t.Fatalf("Every got %v, want false", got)
			}
		}},
		{"every empty", func(t *testing.T) {
			if got := Of[int]().Every(even); !got {
				t.Fatalf("Every got %v, want true", got)
			}
		}},
		{"none", func(t *testing.T) {
			if got := Of(1, 3).None(even); !got {
				t.Fatalf("None got %v, want true", got)
			}
		}},
		{"none not", func(t *testing.T) {
			if got := Of(1, 2).None(even); got {
				t.Fatalf("None got %v, want false", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestFindLastCountBy(t *testing.T) {
	even := func(v int) bool { return v%2 == 0 }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"countBy", func(t *testing.T) {
			if got := Of(1, 2, 3, 4).CountBy(even); got != 2 {
				t.Fatalf("CountBy got %d, want 2", got)
			}
		}},
		{"find", func(t *testing.T) {
			if got, ok := Of(1, 3, 4).Find(even); got != 4 || !ok {
				t.Fatalf("Find got %d,%v want 4,true", got, ok)
			}
		}},
		{"find missing", func(t *testing.T) {
			if got, ok := Of(1, 3).Find(even); got != 0 || ok {
				t.Fatalf("Find got %d,%v want 0,false", got, ok)
			}
		}},
		{"last", func(t *testing.T) {
			if got, ok := Of(1, 2, 3).Last(); got != 3 || !ok {
				t.Fatalf("Last got %d,%v want 3,true", got, ok)
			}
		}},
		{"last empty", func(t *testing.T) {
			if got, ok := Of[int]().Last(); got != 0 || ok {
				t.Fatalf("Last got %d,%v want 0,false", got, ok)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"found", func(t *testing.T) {
			if !Contains(Of(3, 1, 4), 4) {
				t.Fatal("Contains want true")
			}
		}},
		{"missing", func(t *testing.T) {
			if Contains(Of(3, 1), 4) {
				t.Fatal("Contains want false")
			}
		}},
		{"empty", func(t *testing.T) {
			if Contains(Of[int](), 0) {
				t.Fatal("Contains want false")
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestReduceErr(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"first error kept", func(t *testing.T) {
			acc, err := Of(1, 2, 3).ReduceErr(0, failOnTwo)
			if acc != 4 {
				t.Fatalf("ReduceErr acc=%d, want 4", acc)
			}
			if !errors.Is(err, errSentinel) {
				t.Fatalf("ReduceErr err=%v, want sentinel", err)
			}
		}},
		{"all ok", func(t *testing.T) {
			acc, err := Of(3).ReduceErr(0, failOnTwo)
			if acc != 3 {
				t.Fatalf("ReduceErr acc=%d, want 3", acc)
			}
			if err != nil {
				t.Fatalf("ReduceErr err=%v, want nil", err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
