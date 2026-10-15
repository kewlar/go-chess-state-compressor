package chess_test

import (
	"testing"

	"example.com/go-chess/internal/chess"
)

type pieceMap map[chess.Square]chess.Piece

func TestAttacks(t *testing.T) {
	tests := []struct {
		name     string
		pieces   pieceMap
		from     chess.Square
		to       []chess.Square
		expected bool
	}{
		// King
		{
			name:     "King attacks nearby square",
			pieces:   pieceMap{chess.E1: chess.WhiteKing},
			from:     chess.E1,
			to:       []chess.Square{chess.D1, chess.D2, chess.E2, chess.F2, chess.F1},
			expected: true,
		},
		{
			name:     "King cannot attack faraway square",
			pieces:   pieceMap{chess.E1: chess.WhiteKing},
			from:     chess.E1,
			to:       []chess.Square{chess.C1, chess.C3, chess.E3, chess.G3, chess.G1},
			expected: false,
		},
		// Queen
		{
			name:     "Queen attacks diagonally",
			pieces:   pieceMap{chess.E5: chess.BlackQueen},
			from:     chess.E5,
			to:       []chess.Square{chess.A1, chess.B8, chess.H8, chess.G3},
			expected: true,
		},
		{
			name:     "Queen attacks horizontally and vertically",
			pieces:   pieceMap{chess.E5: chess.WhiteQueen},
			from:     chess.E5,
			to:       []chess.Square{chess.A5, chess.E7, chess.E1, chess.H5},
			expected: true,
		},
		{
			name:     "Queen cannot attack squares that aren't diagonal or orthogonal from it",
			pieces:   pieceMap{chess.E5: chess.BlackQueen},
			from:     chess.E5,
			to:       []chess.Square{chess.D3, chess.C4, chess.H1, chess.D8},
			expected: false,
		},
		{
			name:     "Queen cannot attack squares that hidden behind other pieces",
			pieces:   pieceMap{chess.E5: chess.BlackQueen, chess.D6: chess.WhitePawn, chess.E4: chess.BlackRook},
			from:     chess.E5,
			to:       []chess.Square{chess.E1, chess.C7},
			expected: false,
		},
		// Rook
		{
			name:     "Rook attacks horizontally and vertically",
			pieces:   pieceMap{chess.E5: chess.BlackRook},
			from:     chess.E5,
			to:       []chess.Square{chess.A5, chess.E7, chess.E1, chess.H5},
			expected: true,
		},
		{
			name:     "Rook cannot attack diagonally",
			pieces:   pieceMap{chess.E5: chess.BlackRook},
			from:     chess.E5,
			to:       []chess.Square{chess.A1, chess.B8, chess.H8, chess.G3},
			expected: false,
		},
		{
			name:     "Rook cannot attack squares that aren't orthogonal from it",
			pieces:   pieceMap{chess.E5: chess.WhiteRook},
			from:     chess.E5,
			to:       []chess.Square{chess.D3, chess.C4, chess.H1, chess.D8},
			expected: false,
		},
		{
			name:     "Rook cannot attack squares that hidden behind other pieces",
			pieces:   pieceMap{chess.E5: chess.BlackRook, chess.E6: chess.WhitePawn, chess.D5: chess.BlackRook, chess.F5: chess.BlackBishop, chess.E4: chess.WhitePawn},
			from:     chess.E5,
			to:       []chess.Square{chess.A5, chess.E7, chess.E1, chess.H5},
			expected: false,
		},
		// Bishop
		{
			name:     "Bishop attacks diagonally",
			pieces:   pieceMap{chess.E5: chess.BlackBishop},
			from:     chess.E5,
			to:       []chess.Square{chess.A1, chess.B8, chess.H8, chess.G3},
			expected: true,
		},
		{
			name:     "Bishop cannot attack horizontally and vertically",
			pieces:   pieceMap{chess.E5: chess.BlackBishop},
			from:     chess.E5,
			to:       []chess.Square{chess.A5, chess.E7, chess.E1, chess.H5},
			expected: false,
		},
		{
			name:     "Bishop cannot attack squares that aren't diagonal from it",
			pieces:   pieceMap{chess.E5: chess.WhiteBishop},
			from:     chess.E5,
			to:       []chess.Square{chess.D3, chess.C4, chess.H1, chess.D8},
			expected: false,
		},
		{
			name:     "Bishop cannot attack squares that hidden behind other pieces",
			pieces:   pieceMap{chess.E5: chess.BlackBishop, chess.D4: chess.WhitePawn, chess.D6: chess.BlackBishop, chess.F4: chess.BlackBishop, chess.F6: chess.WhitePawn},
			from:     chess.E5,
			to:       []chess.Square{chess.A1, chess.B8, chess.H8, chess.G3},
			expected: false,
		},
		// Knight
		{
			name:     "Knight attacks 1 square straight + 1 square diagonally away from it",
			pieces:   pieceMap{chess.E5: chess.BlackKnight},
			from:     chess.E5,
			to:       []chess.Square{chess.C4, chess.D3, chess.F3, chess.G4, chess.G6, chess.F7, chess.D7, chess.C6},
			expected: true,
		},
		{
			name:     "Knight cannot attack diagonally",
			pieces:   pieceMap{chess.E5: chess.BlackKnight},
			from:     chess.E5,
			to:       []chess.Square{chess.A1, chess.B8, chess.H8, chess.G3},
			expected: false,
		},
		{
			name:     "Knight cannot attack horizontally and vertically",
			pieces:   pieceMap{chess.E5: chess.BlackKnight},
			from:     chess.E5,
			to:       []chess.Square{chess.A5, chess.E7, chess.E1, chess.H5},
			expected: false,
		},
		{
			name:     "Knight cannot attack squares that aren't exactly in its attack squares",
			pieces:   pieceMap{chess.E5: chess.WhiteKnight},
			from:     chess.E5,
			to:       []chess.Square{chess.C1, chess.A3, chess.F2, chess.F8},
			expected: false,
		},
		// Pawn
		{
			name:     "White Pawn attacks exactly 1 square diagonally in front of it",
			pieces:   pieceMap{chess.E2: chess.WhitePawn},
			from:     chess.E2,
			to:       []chess.Square{chess.D3, chess.F3},
			expected: true,
		},
		{
			name:     "Black Pawn attacks exactly 1 square diagonally in front of it",
			pieces:   pieceMap{chess.D6: chess.BlackPawn},
			from:     chess.D6,
			to:       []chess.Square{chess.C5, chess.E5},
			expected: true,
		},
		{
			name:     "White Pawn cannot attack any other squares",
			pieces:   pieceMap{chess.E2: chess.WhitePawn},
			from:     chess.E2,
			to:       []chess.Square{chess.D1, chess.E1, chess.F1, chess.D2, chess.F2, chess.E3},
			expected: false,
		},
		{
			name:     "Black Pawn cannot attack any other squares",
			pieces:   pieceMap{chess.D6: chess.BlackPawn},
			from:     chess.D6,
			to:       []chess.Square{chess.C7, chess.D7, chess.E7, chess.C6, chess.E6, chess.D5},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := chess.Board{}
			for square, piece := range tt.pieces {
				board[square] = piece
			}
			for _, to := range tt.to {
				actual := chess.Attacks(board, tt.from, to)
				if actual != tt.expected {
					t.Errorf("Attacks(%#v, %s, %s) = %v; expected %v", tt.pieces, tt.from.String(), to.String(), actual, tt.expected)
				}
			}
		})
	}
}
