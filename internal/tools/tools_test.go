package tools_test

import (
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
