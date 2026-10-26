package chess_test

import (
	"testing"

	"example.com/go-chess/internal/chess"
)

type squareTuple struct {
	square chess.Square
	piece  chess.Piece
}

func TestOccupiedSquares(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		expected []squareTuple
	}{
		{
			name: "Returns the occupied squares in order",
			fen:  "k7/P7/P7/P7/P7/P7/P7/R3K3 w Q - 1 35",
			expected: []squareTuple{
				{chess.A8, chess.BlackKing},
				{chess.A7, chess.WhitePawn},
				{chess.A6, chess.WhitePawn},
				{chess.A5, chess.WhitePawn},
				{chess.A4, chess.WhitePawn},
				{chess.A3, chess.WhitePawn},
				{chess.A2, chess.WhitePawn},
				{chess.A1, chess.WhiteRook},
				{chess.E1, chess.WhiteKing},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, _ := chess.ParseFen(tt.fen)
			actual := []squareTuple{}
			for square, piece := range state.Board.OccupiedSquares() {
				actual = append(actual, squareTuple{square, piece})
			}
			if !squaresEqual(actual, tt.expected) {
				t.Errorf("OccupiedSquares() for FEN %q = %#v; expected %#v", tt.fen, actual, tt.expected)
			}
		})
	}
}

func squaresEqual(a, b []squareTuple) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].square != b[i].square || a[i].piece != b[i].piece {
			return false
		}
	}
	return true
}
