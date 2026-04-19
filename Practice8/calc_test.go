package main

import "testing"

func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d; want %d", got, want)
	}
}

func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"both positive", 2, 3, 5},
		{"positive + zero", 5, 0, 5},
		{"negative + positive", -1, 4, 3},
		{"both negative", -2, -3, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"both positive", 10, 3, 7},
		{"positive minus zero", 5, 0, 5},
		{"negative minus positive", -4, 3, -7},
		{"both negative", -5, -2, -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Subtract(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		tests := []struct {
			name string
			a, b int
			want int
		}{
			{"positive divided by positive", 10, 2, 5},
			{"negative divided by positive", -9, 3, -3},
			{"zero divided by non-zero", 0, 5, 0},
			{"positive divided by negative", 10, -2, -5},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Divide(tt.a, tt.b)
				if err != nil {
					t.Fatalf("Divide(%d, %d) returned unexpected error: %v", tt.a, tt.b, err)
				}
				if got != tt.want {
					t.Errorf("Divide(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
				}
			})
		}
	})

	t.Run("division by zero", func(t *testing.T) {
		_, err := Divide(10, 0)
		if err == nil {
			t.Error("Divide(10, 0) expected an error, got nil")
		}
	})
}
