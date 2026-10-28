package squareskipper

import (
	"example.com/go-chess/internal/chess"
)

type (
	SquareSkipper map[chess.Square]bool // SquareSkipper allows shrinking a 64-square bitboard by removing certain cells whose value is known from other information, and re-inserting them later.
	Mode          int                   // Mode identifies the mode SquareSkipper can be configured to run in - Occupancy or Color.
)

const (
	Occupancy Mode = iota // In Occupancy mode, true values of the SquareSkipper map identify cells that are for certain occupied, and false - cells that are for certain empty.
	Color                 // In Color mode, true values of the SquareSkipper map identify cells that are for certain White (contain White pieces), and false - cells that are Black or empty.
)

// NewSquareSkipper returns a new SquareSkipper.
func NewSquareSkipper() SquareSkipper {
	return SquareSkipper{}
}

// SkipCastling configures the SquareSkipper to discard squares of castling-available pieces for given castling availability when deflating a bitboard.
func (s SquareSkipper) SkipCastling(castling chess.Castling, mode Mode) {
	for option, available := range castling {
		if available {
			s.SkipCastlingOption(option, mode)
		}
	}
}

// SkipCastlingOption configures the SquareSkipper to discard squares of castling-available pieces for given castling option when deflating a bitboard.
func (s SquareSkipper) SkipCastlingOption(castling rune, mode Mode) {
	isOccupancy := mode == Occupancy
	switch castling {
	case 'K':
		s[chess.E1] = true
		s[chess.H1] = true
	case 'Q':
		s[chess.E1] = true
		s[chess.A1] = true
	case 'k':
		s[chess.E8] = isOccupancy
		s[chess.H8] = isOccupancy
	case 'q':
		s[chess.E8] = isOccupancy
		s[chess.A8] = isOccupancy
	}
}

// SkipEmpty configures the SquareSkipper to discard the empty squares when deflating. Only meant to be used on a 64-square color bitboard.
func (s SquareSkipper) SkipEmpty(occupancy []bool) {
	for square, isOccupied := range occupancy {
		if !isOccupied {
			s[chess.Square(square)] = false
		}
	}
}

// SkipOriginalRanks configures the SquareSkipper to discard the occupied squares that are on original ranks when deflating. Only meant to be used on a 64-square color bitboard.
func (s SquareSkipper) SkipOriginalRanks(occupancy []bool) {
	// White
	for _, rank := range []int{0, 1} {
		for file := range 8 {
			square := chess.NewSquareFromRankAndFile(rank, file)
			if occupancy[square] {
				s[square] = true
			}
		}
	}
	// Black
	for _, rank := range []int{6, 7} {
		for file := range 8 {
			square := chess.NewSquareFromRankAndFile(rank, file)
			if occupancy[square] {
				s[square] = false
			}
		}
	}
}

// SkipEnPassant configures the SquareSkipper to discard the squares with en passant pawn, and the two empty squares behind it.
func (s SquareSkipper) SkipEnPassant(enPassant chess.Square, mode Mode) {
	enPassantPawn := chess.Square(int(enPassant) ^ 0b1000)
	enPassantPawnOrigin := chess.Square(int(enPassant) ^ 0b11000)
	s[enPassantPawn] = mode == Occupancy || enPassantPawn.Rank() == 3
	s[enPassant] = false
	s[enPassantPawnOrigin] = false
}

// Length returns the number of cells the SquareSkipper is configured to skip.
func (s SquareSkipper) Length() int {
	return len(s)
}

// Deflate returns a copy of l slice with configured squares removed.
func (s SquareSkipper) Deflate(l []bool) []bool {
	out := []bool{}
	for i, x := range l {
		if _, ok := s[chess.Square(i)]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// Inflate returns a copy of l square with configured squares reinserted.
func (s SquareSkipper) Inflate(l []bool) []bool {
	out := []bool{}
	idx := 0
	for i := range 64 {
		if value, ok := s[chess.Square(i)]; ok {
			out = append(out, value)
		} else {
			out = append(out, l[idx])
			idx++
		}
	}
	return out
}
