package chess

type Piece rune

const (
	Empty Piece = 0

	Pawn   Piece = 'p'
	King   Piece = 'k'
	Queen  Piece = 'q'
	Rook   Piece = 'r'
	Bishop Piece = 'b'
	Knight Piece = 'n'

	BlackPawn   Piece = 'p'
	BlackKing   Piece = 'k'
	BlackQueen  Piece = 'q'
	BlackRook   Piece = 'r'
	BlackBishop Piece = 'b'
	BlackKnight Piece = 'n'

	WhitePawn   Piece = 'P'
	WhiteKing   Piece = 'K'
	WhiteQueen  Piece = 'Q'
	WhiteRook   Piece = 'R'
	WhiteBishop Piece = 'B'
	WhiteKnight Piece = 'N'
)

// NewPiece returns a new piece, given pieceType and color.
func NewPiece(pieceType rune, color Color) Piece {
	p := Piece(pieceType)
	if color == White {
		p &= 0b1101_1111
	}
	return p
}

// IsEmpty returns true if piece is an empty square.
func (p Piece) IsEmpty() bool {
	return p == Empty
}

// IsOccupied returns true if the piece is not an empty square.
func (p Piece) IsOccupied() bool {
	return p != Empty
}

// IsValid returns true if piece is either a valid chess piece or an empty square.
func (p Piece) IsValid() bool {
	t := p.Type()
	return p.IsEmpty() || t == Pawn || t == King || t == Queen || t == Rook || t == Bishop || t == Knight
}

// Type returns the piece type, rejecting color information.
func (p Piece) Type() Piece {
	if p == Empty {
		return p
	}
	return p | 32
}

// Color returns the piece color.
func (p Piece) Color() Color {
	return Color(p&32 == 0)
}

// String returns piece name as a single character FEN/PGN compatible string, with "." standing for empty squares.
func (p Piece) String() string {
	if p == Empty {
		return "." // Useful for rendering chess board in ASCII.
	}

	return string(rune(p))
}
