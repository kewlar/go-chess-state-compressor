package chess_test

import (
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

func TestNewColor(t *testing.T) {
	tests := []struct {
		name     string
		c        rune
		expected chess.Color
	}{
		{name: "White", c: 'w', expected: chess.White},
		{name: "Black", c: 'b', expected: chess.Black},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := chess.NewColor(tt.c)
			if actual != tt.expected {
				t.Errorf("NewColor('%s') = %s; expected %s", string(tt.c), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestColorOther(t *testing.T) {
	tests := []struct {
		name     string
		color    chess.Color
		expected chess.Color
	}{
		{name: "White flips to Black", color: chess.White, expected: chess.Black},
		{name: "Black flips to White", color: chess.Black, expected: chess.White},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.color.Other()
			if actual != tt.expected {
				t.Errorf("%s.Other() = %s; expected %s", tt.color.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		name     string
		color    chess.Color
		expected string
	}{
		{name: "White", color: chess.White, expected: "White"},
		{name: "Black", color: chess.Black, expected: "Black"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.color.String()
			if actual != tt.expected {
				t.Errorf("%s.String() = %q; expected %q", tt.color.String(), actual, tt.expected)
			}
		})
	}
}

func TestColorRune(t *testing.T) {
	tests := []struct {
		name     string
		color    chess.Color
		expected rune
	}{
		{name: "White", color: chess.White, expected: 'w'},
		{name: "Black", color: chess.Black, expected: 'b'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.color.Rune()
			if actual != tt.expected {
				t.Errorf("%s.Rune() = %q; expected %q", tt.color.String(), actual, tt.expected)
			}
		})
	}
}
