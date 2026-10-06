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

func NewPiece(pieceType rune, color Color) Piece {
	p := Piece(pieceType)
	if color == White {
		p &= 0b1101_1111
	}
	return p
}

func (p Piece) IsEmpty() bool {
	return p == Empty
}

func (p Piece) IsOccupied() bool {
	return p != Empty
}

func (p Piece) IsValid() bool {
	t := p.Type()
	return p.IsEmpty() || t == Pawn || t == King || t == Queen || t == Rook || t == Bishop || t == Knight
}

func (p Piece) Type() Piece {
	return p | 32
}

func (p Piece) Color() Color {
	return Color(p&32 == 0)
}

func (p Piece) String() string {
	if p == Empty {
		return "."
	}

	return string(rune(p))
}
