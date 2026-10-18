package tools_test

import (
	"slices"
	"testing"

	"example.com/go-chess/internal/tools"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		expected int
	}{
		{name: "positive numbers", x: 2, expected: 2},
		{name: "zero", x: 0, expected: 0},
		{name: "negative numbers", x: -3, expected: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tools.Abs(tt.x)
			if actual != tt.expected {
				t.Errorf("Abs(%d) = %d; want %d", tt.x, actual, tt.expected)
			}
		})
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		expected int
	}{
		{name: "positive numbers", x: 2, expected: 1},
		{name: "zero", x: 0, expected: 0},
		{name: "negative numbers", x: -3, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tools.Sign(tt.x)
			if actual != tt.expected {
				t.Errorf("Sign(%d) = %d; want %d", tt.x, actual, tt.expected)
			}
		})
	}
}

func TestEliasGamma(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		expected []bool
	}{
		{name: "negative numbers", x: -1, expected: []bool{}},
		{name: "zero", x: 0, expected: []bool{}},
		{name: "one", x: 1, expected: []bool{true}},
		{name: "two", x: 2, expected: []bool{false, true, false}},
		{name: "three", x: 3, expected: []bool{false, true, true}},
		{name: "four", x: 4, expected: []bool{false, false, true, false, false}},
		{name: "five", x: 5, expected: []bool{false, false, true, false, true}},
		{name: "sixty four", x: 64, expected: []bool{false, false, false, false, false, false, true, false, false, false, false, false, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tools.EliasGamma(tt.x)
			if !slices.Equal(actual, tt.expected) {
				t.Errorf("EliasGamma(%d) = %#v; want %#v", tt.x, actual, tt.expected)
			}
		})
	}
}

func TestWithoutSquares(t *testing.T) {
	input := []bool{false, true, true, false, true, false, true, false}
	squares := []int{1, 3, 4, 5}
	expected := []bool{false, true, true, false}
	actual := tools.WithoutSquares(input, squares)
	if !slices.Equal(actual, expected) {
		t.Errorf("WithoutSquares(%#v, %#v) = %#v; expected %#v", input, squares, actual, expected)
	}
}

func TestWithSquares(t *testing.T) {
	input := make([]bool, 32)
	for i := range 32 {
		input[i] = i < 16
	}
	squares := map[int]bool{}
	for i := range 32 {
		j := i*2 + 1
		squares[j] = i >= 16
	}
	expectedString := "1010101010101010101010101010101001010101010101010101010101010101"
	expected := make([]bool, 64)
	for i := range 64 {
		expected[i] = expectedString[i] == '1'
	}
	actual := tools.WithSquares(input, squares)
	if !slices.Equal(actual, expected) {
		t.Errorf("WithSquares(%#v, %#v) = %#v; expected %#v", input, squares, actual, expected)
	}
}
