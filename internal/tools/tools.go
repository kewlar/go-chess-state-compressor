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

// DecodeEliasGamma decodes variable length Elias gamma code into a positive integer.
func DecodeEliasGamma(readBit func() bool) int {
	n := 0
	for !readBit() {
		n++
	}
	// just read a 1-bit
	x := 0
	for range n {
		x <<= 1
		if readBit() {
			x |= 1
		}
	}
	return (1 << n) + x
}

// CountTrue returns a number of members in bool slice s that are true.
func CountTrue(s []bool) int {
	count := 0
	for _, x := range s {
		if x {
			count++
		}
	}
	return count
}
