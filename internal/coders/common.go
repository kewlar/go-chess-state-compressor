package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
)

func TryEncodeEarlyPosition(state *chess.State) (*bitfield.BitField, bool) {
	bf := &bitfield.BitField{}

	// Encode starting position.
	if state.Fullmove == 1 && state.SideToMove == chess.White {
		bf.WriteInt(0, 21) // moveNum
		return bf, true
	}

	// Encode the board after first move by White
	if state.Fullmove == 1 && state.SideToMove == chess.Black {
		bf.WriteInt(1, 21) // moveNum
		whiteMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			panic("move not found")
		}
		bf.WriteInt(uint64(whiteMove), 5)
		return bf, true
	}

	// Encode the board after first move by Black
	if state.Fullmove == 2 && state.SideToMove == chess.White {
		bf.WriteInt(2, 21) // moveNum
		whiteMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			panic("move not found")
		}
		blackMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstBlackMoves())
		if blackMove == -1 { // Not found.
			panic("move not found")
		}

		bf.WriteInt(uint64(whiteMove*20+blackMove), 9)
		return bf, true
	}

	return nil, false
}
