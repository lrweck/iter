package iter

import "testing"

func TestSumMaxMin(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"sum", func(t *testing.T) {
			if got := Sum(Of(1, 2, 3)); got != 6 {
				t.Fatalf("Sum got %d, want 6", got)
			}
		}},
		{"sum empty", func(t *testing.T) {
			if got := Sum(Of[int]()); got != 0 {
				t.Fatalf("Sum got %d, want 0", got)
			}
		}},
		{"max", func(t *testing.T) {
			if got, ok := Max(Of(3, 1, 4, 1)); got != 4 || !ok {
				t.Fatalf("Max got %d,%v want 4,true", got, ok)
			}
		}},
		{"max empty", func(t *testing.T) {
			if got, ok := Max(Of[int]()); got != 0 || ok {
				t.Fatalf("Max got %d,%v want 0,false", got, ok)
			}
		}},
		{"min", func(t *testing.T) {
			if got, ok := Min(Of(5, -1, 3)); got != -1 || !ok {
				t.Fatalf("Min got %d,%v want -1,true", got, ok)
			}
		}},
		{"min empty", func(t *testing.T) {
			if got, ok := Min(Of[int]()); got != 0 || ok {
				t.Fatalf("Min got %d,%v want 0,false", got, ok)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSumComplex(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"sum", func(t *testing.T) {
			if got := Sum(Of(1+2i, 3+4i)); got != 4+6i {
				t.Fatalf("Sum got %v, want 4+6i", got)
			}
		}},
		{"empty", func(t *testing.T) {
			if got := Sum(Of[complex128]()); got != 0 {
				t.Fatalf("Sum got %v, want 0", got)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSumByMaxByMinBy(t *testing.T) {
	identity := func(v int) int { return v }
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"sum", func(t *testing.T) {
			if got := Of(1, 2, 3).SumBy(identity); got != 6 {
				t.Fatalf("SumBy got %d, want 6", got)
			}
		}},
		{"sum empty", func(t *testing.T) {
			if got := Of[int]().SumBy(identity); got != 0 {
				t.Fatalf("SumBy got %d, want 0", got)
			}
		}},
		{"max", func(t *testing.T) {
			if got, ok := Of(3, 1, 4).MaxBy(identity); got != 4 || !ok {
				t.Fatalf("MaxBy got %d,%v want 4,true", got, ok)
			}
		}},
		{"max empty", func(t *testing.T) {
			if got, ok := Of[int]().MaxBy(identity); got != 0 || ok {
				t.Fatalf("MaxBy got %d,%v want 0,false", got, ok)
			}
		}},
		{"min", func(t *testing.T) {
			if got, ok := Of(5, -1, 3).MinBy(identity); got != -1 || !ok {
				t.Fatalf("MinBy got %d,%v want -1,true", got, ok)
			}
		}},
		{"min empty", func(t *testing.T) {
			if got, ok := Of[int]().MinBy(identity); got != 0 || ok {
				t.Fatalf("MinBy got %d,%v want 0,false", got, ok)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
