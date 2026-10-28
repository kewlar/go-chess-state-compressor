package chess

// Square is a chess square, indexed from 0 to 63.
type Square int

// InvalidSquare is a virtual scale used for cases where square is not existing, i.e. en passant square when there is no en passant.
const (
	InvalidSquare Square = 64

	A8 Square = 0
	B8 Square = 1
	C8 Square = 2
	D8 Square = 3
	E8 Square = 4
	F8 Square = 5
	G8 Square = 6
	H8 Square = 7

	A7 Square = 8
	B7 Square = 9
	C7 Square = 10
	D7 Square = 11
	E7 Square = 12
	F7 Square = 13
	G7 Square = 14
	H7 Square = 15

	A6 Square = 16
	B6 Square = 17
	C6 Square = 18
	D6 Square = 19
	E6 Square = 20
	F6 Square = 21
	G6 Square = 22
	H6 Square = 23

	A5 Square = 24
	B5 Square = 25
	C5 Square = 26
	D5 Square = 27
	E5 Square = 28
	F5 Square = 29
	G5 Square = 30
	H5 Square = 31

	A4 Square = 32
	B4 Square = 33
	C4 Square = 34
	D4 Square = 35
	E4 Square = 36
	F4 Square = 37
	G4 Square = 38
	H4 Square = 39

	A3 Square = 40
	B3 Square = 41
	C3 Square = 42
	D3 Square = 43
	E3 Square = 44
	F3 Square = 45
	G3 Square = 46
	H3 Square = 47

	A2 Square = 48
	B2 Square = 49
	C2 Square = 50
	D2 Square = 51
	E2 Square = 52
	F2 Square = 53
	G2 Square = 54
	H2 Square = 55

	A1 Square = 56
	B1 Square = 57
	C1 Square = 58
	D1 Square = 59
	E1 Square = 60
	F1 Square = 61
	G1 Square = 62
	H1 Square = 63
)

// NewSquare returns a new Square from coordinates given as algebraic string.
func NewSquare(coordinates string) Square {
	if len(coordinates) != 2 {
		return InvalidSquare
	}
	return NewSquareFromRankAndFile(int(coordinates[1]-'1'), int(coordinates[0]-'a'))
}

// NewSquareFromRankAndFile returns a Square at given rank and file.
func NewSquareFromRankAndFile(rank, file int) Square {
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return InvalidSquare
	}
	return Square((7-rank)<<3 + file&7)
}

// Rank returns the rank of the square, numbered 0 to 7.
func (s Square) Rank() int {
	if !s.IsValid() {
		return -1
	}
	return 7 - int(s)>>3
}

// File returns the file of the square, numbered 0 to 7.
func (s Square) File() int {
	if !s.IsValid() {
		return -1
	}
	return int(s) & 7
}

// Color returns the color of the square, either White or Black.
func (s Square) Color() Color {
	return Color((s.Rank()+s.File())&1 == 1)
}

// IsOnBackRank returns if the square is either the first rank or the last rank.
func (s Square) IsOnBackRank() bool {
	rank := s.Rank()
	return rank == 0 || rank == 7
}

// IsValid returns true if the square is valid.
func (s Square) IsValid() bool {
	return 0 <= s && s <= 63
}

// String returns square name in algebraic notation, from "a1" to "h8".
func (s Square) String() string {
	if !s.IsValid() {
		return "-"
	}
	return string([]rune{'a' + rune(s.File()), '1' + rune(s.Rank())})
}
