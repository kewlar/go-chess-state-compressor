package chess

// Color is a color of either a chess piece, or chessboard square.
// Color can be either White (true), or Black (false).
type Color bool

const (
	White Color = true
	Black Color = false
)

// NewColor returns a new Color. The rune argument can be either 'w' or 'b'.
func NewColor(c rune) Color {
	switch c {
	case 'w':
		return White
	case 'b':
		return Black
	}
	panic("unexpected color")
}

// Other returns the opposite color.
func (c Color) Other() Color {
	return !c
}

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
