package chess_test

import (
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

func TestSquareCasting(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected chess.Square
	}{
		{name: "a1", index: 56, expected: chess.A1},
		{name: "h8", index: 7, expected: chess.H8},
		{name: "Invalid square", index: 64, expected: chess.InvalidSquare},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := chess.Square(tt.index)
			if actual != tt.expected {
				t.Errorf("Square(%d) = %s; want %s", tt.index, actual.String(), tt.expected.String())
			}
		})
	}
}

func TestNewSquare(t *testing.T) {
	tests := []struct {
		name        string
		coordinates string
		expected    chess.Square
	}{
		{name: "a8", coordinates: "a8", expected: chess.A8},
		{name: "h1", coordinates: "h1", expected: chess.H1},
		{name: "Invalid square (empty)", coordinates: "", expected: chess.InvalidSquare},
		{name: "Invalid square (zzz)", coordinates: "zzz", expected: chess.InvalidSquare},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := chess.NewSquare(tt.coordinates)
			if actual != tt.expected {
				t.Errorf("NewSquare(%q) = Square(%d); want Square(%d)", tt.coordinates, int(actual), int(tt.expected))
			}
		})
	}
}

func TestNewSquareFromRankAndFile(t *testing.T) {
	tests := []struct {
		name       string
		rank, file int
		expected   chess.Square
	}{
		{name: "a1", rank: 0, file: 0, expected: chess.A1},
		{name: "h8", rank: 7, file: 7, expected: chess.H8},
		{name: "Invalid square (b0)", rank: 1, file: -1, expected: chess.InvalidSquare},
		{name: "Invalid square (c9)", rank: 2, file: 8, expected: chess.InvalidSquare},
		{name: "Invalid square (i0)", rank: 8, file: 0, expected: chess.InvalidSquare},
		{name: "Invalid square (-1, 0)", rank: -1, file: 0, expected: chess.InvalidSquare},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := chess.NewSquareFromRankAndFile(tt.rank, tt.file)
			if actual != tt.expected {
				t.Errorf("NewSquareFromRankAndFile(%d, %d) = %s; expected %s", tt.rank, tt.file, actual.String(), tt.expected.String())
			}
		})
	}
}

func TestSquareRank(t *testing.T) {
	tests := []struct {
		name     string
		square   chess.Square
		expected int
	}{
		{name: "a1", square: chess.A1, expected: 0},
		{name: "a8", square: chess.A8, expected: 7},
		{name: "h1", square: chess.H1, expected: 0},
		{name: "h8", square: chess.H8, expected: 7},
		{name: "Invalid square", square: chess.InvalidSquare, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.square.Rank()
			if actual != tt.expected {
				t.Errorf("Square(%q).Rank() = %d; expected %d", tt.square.String(), actual, tt.expected)
			}
		})
	}
}

func TestSquareFile(t *testing.T) {
	tests := []struct {
		name     string
		square   chess.Square
		expected int
	}{
		{name: "a1", square: chess.A1, expected: 0},
		{name: "a8", square: chess.A8, expected: 0},
		{name: "h1", square: chess.H1, expected: 7},
		{name: "h8", square: chess.H8, expected: 7},
		{name: "Invalid square", square: chess.InvalidSquare, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.square.File()
			if actual != tt.expected {
				t.Errorf("Square(%q).File() = %d; expected %d", tt.square.String(), actual, tt.expected)
			}
		})
	}
}

func TestSquareColor(t *testing.T) {
	tests := []struct {
		name     string
		square   chess.Square
		expected chess.Color
	}{
		{name: "a1", square: chess.A1, expected: chess.Black},
		{name: "a8", square: chess.A8, expected: chess.White},
		{name: "h1", square: chess.H1, expected: chess.White},
		{name: "h8", square: chess.H8, expected: chess.Black},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.square.Color()
			if actual != tt.expected {
				t.Errorf("Square(%q).Color() = %s; expected %s", tt.square.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestSquareIsOnBackRank(t *testing.T) {
	tests := []struct {
		name     string
		square   chess.Square
		expected bool
	}{
		{name: "a1", square: chess.A1, expected: true},
		{name: "b1", square: chess.B1, expected: true},
		{name: "g8", square: chess.G8, expected: true},
		{name: "h8", square: chess.H8, expected: true},
		{name: "a2", square: chess.A2, expected: false},
		{name: "b4", square: chess.B4, expected: false},
		{name: "g7", square: chess.G7, expected: false},
		{name: "h6", square: chess.H6, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.square.IsOnBackRank()
			if actual != tt.expected {
				t.Errorf("Square(%q).IsOnBackRank() = %v; expected %v", tt.square.String(), actual, tt.expected)
			}
		})
	}
}

func TestSquareIsValid(t *testing.T) {
	tests := []struct {
		name     string
		square   chess.Square
		expected bool
	}{
		{name: "a1", square: chess.A1, expected: true},
		{name: "b1", square: chess.B1, expected: true},
		{name: "g8", square: chess.G8, expected: true},
		{name: "h8", square: chess.H8, expected: true},
		{name: "Invalid square", square: chess.InvalidSquare, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.square.IsValid()
			if actual != tt.expected {
				t.Errorf("Square(%q).IsValid() = %v; expected %v", tt.square.String(), actual, tt.expected)
			}
		})
	}
}

func TestSquareString(t *testing.T) {
	tests := []struct {
		name     string
		square   chess.Square
		expected string
	}{
		{name: "a1", square: chess.A1, expected: "a1"},
		{name: "b2", square: chess.B2, expected: "b2"},
		{name: "g5", square: chess.G5, expected: "g5"},
		{name: "h8", square: chess.H8, expected: "h8"},
		{name: "Invalid square", square: chess.InvalidSquare, expected: "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.square.String()
			if actual != tt.expected {
				t.Errorf("Square(%q).String() = %q; expected %q", tt.square.String(), actual, tt.expected)
			}
		})
	}
}
