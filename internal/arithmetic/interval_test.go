package arithmetic_test

import (
	"slices"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/arithmetic"
)

func TestIntervalLength(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Rational
		expected *arithmetic.Rational
	}{
		{"Interval length 1", arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1), arithmetic.NewRational(1, 1)},
		{"Interval length 5/32", arithmetic.NewRational(1, 32), arithmetic.NewRational(3, 16), arithmetic.NewRational(5, 32)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := arithmetic.NewInterval(tt.a, tt.b)
			actual := interval.Length()
			if actual.Cmp(tt.expected) != 0 {
				t.Errorf("%s.Length() = %s; expected %s", interval.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestIntervalMul(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Interval
		expected string // Interval rendered as string
	}{
		{
			name:     "Idempotent operation",
			a:        arithmetic.NewInterval(arithmetic.NewRational(0, 2), arithmetic.NewRational(2, 2)),
			b:        arithmetic.NewInterval(arithmetic.NewRational(0, 3), arithmetic.NewRational(1, 1)),
			expected: "[0/1-1/1)",
		},
		{
			name:     "Third quarter of a second third",
			a:        arithmetic.NewInterval(arithmetic.NewRational(1, 3), arithmetic.NewRational(2, 3)),
			b:        arithmetic.NewInterval(arithmetic.NewRational(1, 2), arithmetic.NewRational(3, 4)),
			expected: "[1/2-7/12)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := arithmetic.NewInterval(arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1))
			a.Mul(tt.a) // Effectivelly, a = Copy(tt.a)
			a.Mul(tt.b)
			actual := a.String()
			if actual != tt.expected {
				t.Errorf("%s.Mul(%s) = %s; expected %s", tt.a.String(), tt.b.String(), actual, tt.expected)
			}
		})
	}
}

func TestIntervalCover(t *testing.T) {
	tests := []struct {
		name     string
		a        *arithmetic.Interval
		z        *arithmetic.Rational
		expected bool
	}{
		{
			name:     "Middle of the interval is covered",
			a:        arithmetic.NewInterval(arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1)),
			z:        arithmetic.NewRational(1, 2),
			expected: true,
		},
		{
			name:     "Start of the interval is covered",
			a:        arithmetic.NewInterval(arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1)),
			z:        arithmetic.NewRational(0, 1),
			expected: true,
		},
		{
			name:     "End of the interval is not covered",
			a:        arithmetic.NewInterval(arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1)),
			z:        arithmetic.NewRational(1, 1),
			expected: false,
		},
		{
			name:     "Before the start of the interval is not covered",
			a:        arithmetic.NewInterval(arithmetic.NewRational(1, 2), arithmetic.NewRational(3, 4)),
			z:        arithmetic.NewRational(1, 4),
			expected: false,
		},
		{
			name:     "After the end of the interval is not covered",
			a:        arithmetic.NewInterval(arithmetic.NewRational(1, 2), arithmetic.NewRational(3, 4)),
			z:        arithmetic.NewRational(7, 8),
			expected: false,
		},
		{
			name:     "Arbitrary point inside interval is covered",
			a:        arithmetic.NewInterval(arithmetic.NewRational(1, 16), arithmetic.NewRational(7, 8)),
			z:        arithmetic.NewRational(7, 16),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.a.Cover(tt.z)
			if actual != tt.expected {
				t.Errorf("%s.Cover(%s) = %v; expected %v", tt.a.String(), tt.z.String(), actual, tt.expected)
			}
		})
	}
}

func TestIntervalToBits(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Rational
		expected []bool
	}{
		{
			name:     "Interval starts at zero",
			a:        arithmetic.NewRational(0, 1),
			b:        arithmetic.NewRational(767, 769),
			expected: []bool{},
		},
		{
			name:     "Interval starts just before 1/2",
			a:        arithmetic.NewRational(63, 128),
			b:        arithmetic.NewRational(767, 768),
			expected: []bool{true}, // 0b0.1 / 0b1.0
		},
		{
			name:     "Interval ends just after 1/4",
			a:        arithmetic.NewRational(1, 789),
			b:        arithmetic.NewRational(65, 256),
			expected: []bool{true, false}, // 0b0.01 / 0b1.00
		},
		{
			name:     "Interval conveniently covers 3/4",
			a:        arithmetic.NewRational(129, 256),
			b:        arithmetic.NewRational(1, 1),
			expected: []bool{true, true}, // 0b0.11 / 0b1.00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := arithmetic.NewInterval(tt.a, tt.b)
			actual := interval.ToBits()
			if !slices.Equal(actual, tt.expected) {
				t.Errorf("%s.ToBits() = %#v; expected %#v", interval.String(), actual, tt.expected)
			}
		})
	}
}
