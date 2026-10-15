package chess_test

import (
	"slices"
	"testing"

	"example.com/go-chess/internal/chess"
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
		expectedMove := chess.Move{From: chess.NewSquare("e2"), To: chess.NewSquare("e4")}
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

	t.Run("Includes e2e4", func(t *testing.T) {
		expectedMove := chess.Move{From: chess.NewSquare("e7"), To: chess.NewSquare("e5")}
		if !slices.Contains(chess.FirstBlackMoves(), expectedMove) {
			t.Error("FirstBlackMoves() did not contain e7e5")
		}
	})
}
