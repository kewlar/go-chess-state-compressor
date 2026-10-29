package chess_test

import (
	"slices"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

func TestFirstWhiteMoves(t *testing.T) {
	t.Run("Generates 20 unique moves", func(t *testing.T) {
		moveMap := map[chess.Move]bool{}
		for _, move := range chess.FirstWhiteMoves() {
			moveMap[move] = true
		}
		expected := 20
		actual := len(moveMap)
		if actual != expected {
			t.Errorf("FirstWhiteMoves() generated %d moves; expected %d", actual, expected)
		}
	})

	t.Run("Includes e2e4", func(t *testing.T) {
		expectedMove := chess.Move{From: chess.E2, To: chess.E4}
		if !slices.Contains(chess.FirstWhiteMoves(), expectedMove) {
			t.Error("FirstWhiteMoves() did not contain e2e4")
		}
	})
}

func TestFirstBlackMoves(t *testing.T) {
	t.Run("Generates 20 unique moves", func(t *testing.T) {
		moveMap := map[chess.Move]bool{}
		for _, move := range chess.FirstBlackMoves() {
			moveMap[move] = true
		}
		expected := 20
		actual := len(moveMap)
		if actual != expected {
			t.Errorf("FirstBlackMoves() generated %d moves; expected %d", actual, expected)
		}
	})

	t.Run("Includes e7e5", func(t *testing.T) {
		expectedMove := chess.Move{From: chess.E7, To: chess.E5}
		if !slices.Contains(chess.FirstBlackMoves(), expectedMove) {
			t.Error("FirstBlackMoves() did not contain e7e5")
		}
	})
}
