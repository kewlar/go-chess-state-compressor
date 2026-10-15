package chess_test

import (
	"testing"

	"example.com/go-chess/internal/chess"
)

func TestNewStartingState(t *testing.T) {
	t.Run("Starting game state", func(t *testing.T) {
		state := chess.NewStartingState()
		actual := state.ToFen()
		expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
		if actual != expected {
			t.Errorf("NewStartingState().ToFen() = %q; expected %q", actual, expected)
		}
	})
}

func TestStateRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{name: "Starting state", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{name: "Bare Kings endgame", fen: "8/4k3/8/8/8/8/5K2/8 b - - 0 17"},
		{name: "Last position in the longest possible game", fen: "8/8/8/5K2/6Qk/8/8/8 b - - 150 8849"},
		{name: "18 Queens", fen: "3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80"},
		{name: "6 Pawns in a file", fen: "k7/P7/P7/P7/P7/P7/P7/R3K3 w Q - 1 35"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := chess.ParseFen(tt.fen)
			if err != nil {
				t.Errorf("ParseFen(%q) returned an error %v; expected no error", tt.fen, err)
			}
			actual := state.ToFen()
			if actual != tt.fen {
				t.Errorf("ParseFen(%q).ToFen() = %q; expected it to stay the same", tt.fen, actual)
			}
		})
	}
}

func TestParseFen(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		// Some of the tests are too strict for parsing an otherwise valid FEN, but we want to consider only valid chess games in this benchmark experiment.
		{name: "Too short", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0"},
		{name: "Too few ranks", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq - 0 1"},
		{name: "Too many ranks", fen: "rnbqkbnr/pppppppp/8/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{name: "Too few files", fen: "rnbqkbnr/pppppppp/7/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{name: "Too many files", fen: "rnbqkbnr/ppppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{name: "Invalid piece", fen: "rnbqkbnr/zppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{name: "Too few pieces", fen: "8/8/8/8/8/8/8/K w KQkq - 0 1"},
		{name: "Too many pieces", fen: "rnbqkbnr/pppppppp/p7/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{name: "Invalid side to move", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR W KQkq - 0 1"},
		{name: "Invalid castling options", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w AHah - 0 1"},
		{name: "Invalid en passant square", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e4 0 1"},
		{name: "Invalid halfmove clock", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - X 1"},
		{name: "Halfmove clock too low", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - -1 1"},
		{name: "Halfmove clock too high", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 151 1"},
		{name: "Invalid fullmove number", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 X"},
		{name: "Fullmove number too low", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0"},
		{name: "Fullmove number too high", fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 8850"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := chess.ParseFen(tt.fen)
			if err == nil {
				t.Errorf("expected ParseFen(%q) to return an error; got none", tt.fen)
			}
		})
	}
}
