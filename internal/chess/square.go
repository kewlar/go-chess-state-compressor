package chess

// Square is a chess square, indexed from 0 to 63.
type Square int

// InvalidSquare is a virtual scale used for cases where square is not existing, i.e. en passant square when there is no en passant.
const InvalidSquare Square = 64

// Rank returns the rank of the square, numbered 0 to 7.
func (s Square) Rank() int {
	return 7 - int(s)>>3
}

// File returns the file of the square, numbered 0 to 7.
func (s Square) File() int {
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
	return string('a'+s.File()) + string('1'+s.Rank())
}
