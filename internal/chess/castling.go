package chess

import "strings"

type Castling map[rune]bool

var castlingOptions = []rune{'K', 'Q', 'k', 'q'}

// NewCastling returns Castling populated with availability for the respective castling options.
func NewCastling(availability []bool) Castling {
	castling := Castling{}
	for idx, option := range castlingOptions {
		castling[option] = availability[idx]
	}
	return castling
}

// BoolSlice returns the boolean values for castling availability for the respective castling options.
func (c Castling) BoolSlice() []bool {
	s := []bool{}
	for _, option := range castlingOptions {
		s = append(s, c[option])
	}
	return s
}

// IsAvailable returns true if at least one castling option is available.
func (c Castling) IsAvailable() bool {
	for _, option := range castlingOptions {
		if c[option] {
			return true
		}
	}
	return false
}

// String returns castling availability as FEN-compatible string.
func (c Castling) String() string {
	if !c.IsAvailable() {
		return "-"
	}
	s := strings.Builder{}
	for _, option := range castlingOptions {
		if c[option] {
			s.WriteRune(option)
		}
	}
	return s.String()
}
