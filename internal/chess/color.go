package chess

// Color is a color of either a chess piece, or chessboard square.
// Color can be either White (true), or Black (false).
type Color bool

const (
	White Color = true
	Black Color = false
)

// String returns color name as string ("White" or "Black").
func (c Color) String() string {
	if c {
		return "White"
	}
	return "Black"
}

// Rune returns color name as rune ('w' or 'b').
func (c Color) Rune() rune {
	if c {
		return 'w'
	}
	return 'b'
}
