package square_skipper

import (
	"example.com/go-chess/internal/chess"
)

type (
	SquareSkipper map[chess.Square]bool
	Mode          int
)

const (
	Occupancy Mode = iota
	Color
)

func NewSquareSkipper() *SquareSkipper {
	return &SquareSkipper{}
}

func (s *SquareSkipper) SkipCastling(castling chess.Castling, mode Mode) {
	for option, available := range castling {
		if available {
			s.SkipCastlingOption(option, mode)
		}
	}
}

func (s *SquareSkipper) SkipCastlingOption(castling rune, mode Mode) {
	isOccupancy := mode == Occupancy
	switch castling {
	case 'K':
		(*s)[chess.E1] = true
		(*s)[chess.H1] = true
	case 'Q':
		(*s)[chess.E1] = true
		(*s)[chess.A1] = true
	case 'k':
		(*s)[chess.E8] = isOccupancy
		(*s)[chess.H8] = isOccupancy
	case 'q':
		(*s)[chess.E8] = isOccupancy
		(*s)[chess.A8] = isOccupancy
	}
}

func (s *SquareSkipper) SkipEmpty(occupancy []bool) {
	for square, isOccupied := range occupancy {
		if !isOccupied {
			(*s)[chess.Square(square)] = false
		}
	}
}

func (s *SquareSkipper) SkipOriginalRanks(mode Mode) {
	// White
	for _, rank := range []int{0, 1} {
		for file := range 8 {
			square := chess.NewSquareFromRankAndFile(rank, file)
			(*s)[square] = true
		}
	}
	// Black
	for _, rank := range []int{6, 7} {
		for file := range 8 {
			square := chess.NewSquareFromRankAndFile(rank, file)
			(*s)[square] = mode == Occupancy
		}
	}
}

func (s *SquareSkipper) SkipEnPassant(enPassant chess.Square, mode Mode) {
	enPassantPawn := chess.Square(int(enPassant) ^ 0b1000)
	enPassantHome := chess.Square(int(enPassant) ^ 0b11000)
	(*s)[enPassantPawn] = mode == Occupancy || enPassantPawn.Rank() == 3
	(*s)[enPassant] = false
	(*s)[enPassantHome] = false
}

func (s *SquareSkipper) Length() int {
	return len(*s)
}

func (s *SquareSkipper) All() map[chess.Square]bool {
	return map[chess.Square]bool(*s)
}

func (s *SquareSkipper) Deflate(l []bool) []bool {
	out := []bool{}
	for i, x := range l {
		if _, ok := (*s)[chess.Square(i)]; !ok {
			out = append(out, x)
		}
	}
	return out
}

func (s *SquareSkipper) Inflate(l []bool) []bool {
	out := []bool{}
	idx := 0
	for i := range 64 {
		if value, ok := (*s)[chess.Square(i)]; ok {
			out = append(out, value)
		} else {
			out = append(out, l[idx])
			idx++
		}
	}
	return out
}
