package arithmetic_test

import (
	"math/big"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/arithmetic"
)

func TestRationalEquality(t *testing.T) {
	tests := []struct {
		name         string
		num1, denom1 int64
		num2, denom2 *big.Int
		expected     int
	}{
		{name: "equal", num1: 1, denom1: 2, num2: big.NewInt(11), denom2: big.NewInt(22), expected: 0},
		{name: "less", num1: 1, denom1: 3, num2: big.NewInt(11), denom2: big.NewInt(22), expected: -1},
		{name: "more", num1: 1, denom1: 2, num2: big.NewInt(11), denom2: big.NewInt(23), expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := arithmetic.NewRational(tt.num1, tt.denom1)
			r2 := arithmetic.NewBigRational(tt.num2, tt.denom2)
			actual := r1.Cmp(r2)
			if actual != tt.expected {
				t.Errorf("(%s).Cmp(%s) = %d; expected %d", r1.String(), r2.String(), actual, tt.expected)
			}
		})
	}
}

func TestRationalFromBits(t *testing.T) {
	tests := []struct {
		name       string
		bits       []bool
		num, denom int64
	}{
		{name: "Zero", bits: []bool{false}, num: 0, denom: 1},
		{name: "Zeros", bits: []bool{false, false, false}, num: 0, denom: 1},
		{name: "1/2", bits: []bool{true}, num: 1, denom: 2},
		{name: "3/4", bits: []bool{false, false, true, true}, num: 3, denom: 4},
		{name: "13/16", bits: []bool{true, false, true, true}, num: 13, denom: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := arithmetic.NewRationalFromBits(tt.bits)
			expected := arithmetic.NewRational(tt.num, tt.denom)
			if actual.Cmp(expected) != 0 {
				t.Errorf("NewRationalFromBits(%#v) = %s; expected %s", tt.bits, actual.String(), expected.String())
			}
		})
	}
}

func TestRationalAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Rational
		expected *arithmetic.Rational
	}{
		{name: "Zero", a: arithmetic.NewRational(0, 1), b: arithmetic.NewRational(0, 2), expected: arithmetic.NewRational(0, 1)},
		{name: "Single number", a: arithmetic.NewRational(0, 1), b: arithmetic.NewRational(4, 2), expected: arithmetic.NewRational(2, 1)},
		{name: "Full addition", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(5, 7), expected: arithmetic.NewRational(29, 21)},
		{name: "Fraction simplification", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(1, 6), expected: arithmetic.NewRational(5, 6)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := arithmetic.NewRational(0, 1)
			actual.Add(tt.a, tt.b)
			if actual.Cmp(tt.expected) != 0 {
				t.Errorf("Add(%s, %s) = %s; expected %s", tt.a.String(), tt.b.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestRationalSub(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Rational
		expected *arithmetic.Rational
	}{
		{name: "Zero", a: arithmetic.NewRational(0, 1), b: arithmetic.NewRational(0, 2), expected: arithmetic.NewRational(0, 1)},
		{name: "Single number", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(0, 2), expected: arithmetic.NewRational(2, 3)},
		{name: "Full subtraction", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(5, 7), expected: arithmetic.NewRational(-1, 21)},
		{name: "Fraction simplification", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(1, 6), expected: arithmetic.NewRational(1, 2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := arithmetic.NewRational(0, 1)
			actual.Sub(tt.a, tt.b)
			if actual.Cmp(tt.expected) != 0 {
				t.Errorf("Sub(%s, %s) = %s; expected %s", tt.a.String(), tt.b.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestRationalMul(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Rational
		expected *arithmetic.Rational
	}{
		{name: "Zero", a: arithmetic.NewRational(5, 1), b: arithmetic.NewRational(0, 2), expected: arithmetic.NewRational(0, 1)},
		{name: "Single number", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(2, 2), expected: arithmetic.NewRational(2, 3)},
		{name: "Full multiplication", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(5, 7), expected: arithmetic.NewRational(10, 21)},
		{name: "Fraction simplification", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(1, 6), expected: arithmetic.NewRational(1, 9)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := arithmetic.NewRational(0, 1)
			actual.Mul(tt.a, tt.b)
			if actual.Cmp(tt.expected) != 0 {
				t.Errorf("Mul(%s, %s) = %s; expected %s", tt.a.String(), tt.b.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestRationalDiv(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *arithmetic.Rational
		expected *arithmetic.Rational
	}{
		{name: "Zero", a: arithmetic.NewRational(0, 1), b: arithmetic.NewRational(7, 2), expected: arithmetic.NewRational(0, 1)},
		{name: "Single number", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(2, 2), expected: arithmetic.NewRational(2, 3)},
		{name: "Full division", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(5, 7), expected: arithmetic.NewRational(14, 15)},
		{name: "Fraction simplification", a: arithmetic.NewRational(2, 3), b: arithmetic.NewRational(1, 6), expected: arithmetic.NewRational(4, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := arithmetic.NewRational(0, 1)
			actual.Div(tt.a, tt.b)
			if actual.Cmp(tt.expected) != 0 {
				t.Errorf("Div(%s, %s) = %s; expected %s", tt.a.String(), tt.b.String(), actual.String(), tt.expected.String())
			}
		})
	}
}
