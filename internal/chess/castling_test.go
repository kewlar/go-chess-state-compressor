package chess_test

import (
	"slices"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

func TestCastling(t *testing.T) {
	tests := []struct {
		name              string
		castling          []bool
		expectedAvailable bool
		expectedString    string
	}{
		{name: "Full castling", castling: []bool{true, true, true, true}, expectedAvailable: true, expectedString: "KQkq"},
		{name: "Partial castling", castling: []bool{false, true, true, false}, expectedAvailable: true, expectedString: "Qk"},
		{name: "No castling", castling: []bool{false, false, false, false}, expectedAvailable: false, expectedString: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := chess.NewCastling(tt.castling)
			actual := test.BoolSlice()
			if !slices.Equal(actual, tt.castling) {
				t.Errorf("NewCastling(%#v).BoolSlice() = %#v; expected %#v", tt.castling, actual, tt.castling)
			}
			actualAvailable := test.IsAvailable()
			if actualAvailable != tt.expectedAvailable {
				t.Errorf("NewCastling(%#v).IsAvailable() = %v; expected %v", tt.castling, actualAvailable, tt.expectedAvailable)
			}
			actualString := test.String()
			if actualString != tt.expectedString {
				t.Errorf("NewCastling(%#v).String() = %v; expected %v", tt.castling, actualString, tt.expectedString)
			}
		})
	}
}

func TestCastlingFromString(t *testing.T) {
	tests := []string{
		"",
		"-",
		"q",
		"k",
		"kq",
		"Q",
		"Qq",
		"Qk",
		"Qkq",
		"K",
		"Kq",
		"Kk",
		"Kkq",
		"KQ",
		"KQq",
		"KQk",
		"KQkq",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			availability := tt
			if availability == "-" {
				availability = ""
			}
			expected := tt
			if expected == "" {
				expected = "-"
			}
			castling := chess.NewCastlingFromString(availability)
			actual := castling.String()
			if actual != expected {
				t.Errorf("NewCastlingFromString(%q).String() = %q; expected %q", availability, actual, expected)
			}
		})
	}
}
