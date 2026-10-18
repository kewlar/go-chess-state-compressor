// Package tools contains miscellaneous utility functions.
package tools

import (
	"slices"
)

// Abs returns the absolute value of an integer number x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Sign returns the sign (-1, 0, or 1) of an integer number x.
func Sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// EliasGamma returns Elias gamma code for x.
func EliasGamma(x int) []bool {
	if x < 1 {
		return []bool{} // Cannot calculate Elias Gamma value for zero or negative numbers.
	}
	x0 := x
	out := []bool{}
	for x0 > 0 {
		out = append(out, x0&1 == 1)
		x0 >>= 1
	}
	i := 2
	for i <= x {
		out = append(out, false)
		i <<= 1
	}
	slices.Reverse(out)
	return out
}

// WithoutSquares removes squares-indexed elements from slice s and returns the result.
func WithoutSquares(s []bool, squares []int) []bool {
	out := []bool{}
	for i, x := range s {
		if !slices.Contains(squares, i) {
			out = append(out, x)
		}
	}
	return out
}

// WithSquares fills in 64-square board s with missing information given in squares.
func WithSquares(s []bool, squares map[int]bool) []bool {
	out := []bool{}
	idx := 0
	for i := range 64 {
		if value, ok := squares[i]; ok {
			out = append(out, value)
		} else {
			out = append(out, s[idx])
			idx++
		}
	}
	return out
}
