// Package tools contains miscellaneous utility functions.
package tools

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
