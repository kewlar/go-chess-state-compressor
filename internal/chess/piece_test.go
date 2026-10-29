package chess_test

import (
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

func TestNewPiece(t *testing.T) {
	tests := []struct {
		name      string
		pieceType rune
		color     chess.Color
		expected  chess.Piece
	}{
		{name: "White Pawn", pieceType: rune(chess.Pawn), color: chess.White, expected: chess.WhitePawn},
		{name: "Black Queen", pieceType: rune(chess.Queen), color: chess.Black, expected: chess.BlackQueen},
		{name: "Black Bishop", pieceType: rune(chess.WhiteBishop), color: chess.Black, expected: chess.BlackBishop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := chess.NewPiece(tt.pieceType, tt.color)
			if actual != tt.expected {
				t.Errorf("NewPiece('%s', %q) = %q; expected: %q", string(tt.pieceType), tt.color.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestPieceIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		piece    chess.Piece
		expected bool
	}{
		{name: "Any piece", piece: chess.BlackPawn, expected: false},
		{name: "Empty square", piece: chess.Empty, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.piece.IsEmpty()
			if actual != tt.expected {
				t.Errorf("Piece(%q).IsEmpty() = %v; expected %v", tt.piece.String(), actual, tt.expected)
			}
		})
	}
}

func TestPieceIsOccupied(t *testing.T) {
	tests := []struct {
		name     string
		piece    chess.Piece
		expected bool
	}{
		{name: "Any piece", piece: chess.BlackPawn, expected: true},
		{name: "Empty square", piece: chess.Empty, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.piece.IsOccupied()
			if actual != tt.expected {
				t.Errorf("Piece(%q).IsOccupied() = %v; expected %v", tt.piece.String(), actual, tt.expected)
			}
		})
	}
}

func TestPieceIsValid(t *testing.T) {
	tests := []struct {
		name     string
		piece    chess.Piece
		expected bool
	}{
		{name: "A chess piece is valid", piece: chess.BlackBishop, expected: true},
		{name: "An empty square is valid", piece: chess.Empty, expected: true},
		{name: "An unknown chess piece is not valid", piece: chess.Piece('z'), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.piece.IsValid()
			if actual != tt.expected {
				t.Errorf("Piece(%q).IsValid() = %v; expected %v", tt.piece.String(), actual, tt.expected)
			}
		})
	}
}

func TestPieceType(t *testing.T) {
	tests := []struct {
		name     string
		piece    chess.Piece
		expected chess.Piece
	}{
		{name: "A Black Bishop is a Bishop", piece: chess.BlackBishop, expected: chess.Bishop},
		{name: "A White King is a King", piece: chess.WhiteKing, expected: chess.King},
		{name: "An empty square is still empty", piece: chess.Empty, expected: chess.Empty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.piece.Type()
			if actual != tt.expected {
				t.Errorf("Piece(%q).Type() = %q; expected %q", tt.piece.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestPieceColor(t *testing.T) {
	tests := []struct {
		name     string
		piece    chess.Piece
		expected chess.Color
	}{
		{name: "A Black Rook is Black", piece: chess.BlackRook, expected: chess.Black},
		{name: "A White Queen is White", piece: chess.WhiteQueen, expected: chess.White},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.piece.Color()
			if actual != tt.expected {
				t.Errorf("Piece(%q).Color() = %q; expected %q", tt.piece.String(), actual.String(), tt.expected.String())
			}
		})
	}
}

func TestPieceString(t *testing.T) {
	tests := []struct {
		name     string
		piece    chess.Piece
		expected string
	}{
		{name: "Black Rook", piece: chess.BlackRook, expected: "r"},
		{name: "White Pawn", piece: chess.WhitePawn, expected: "P"},
		{name: "An empty square", piece: chess.Empty, expected: "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.piece.String()
			if actual != tt.expected {
				t.Errorf("Piece(%q).String() = %q; expected %q", tt.piece.String(), actual, tt.expected)
			}
		})
	}
}
