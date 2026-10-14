package chess

// Square is a chess square, indexed from 0 to 63.
type Square int

// InvalidSquare is a virtual scale used for cases where square is not existing, i.e. en passant square when there is no en passant.
const InvalidSquare Square = 64

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

// IsOnHomeRank returns if the square is either the home rank or the back rank.
func (s Square) IsOnHomeRank() bool {
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
