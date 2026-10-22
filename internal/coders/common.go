package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
)

// TryEncodeEarlyPosition tries to encode one of the 421 positions possible in first 3 plies in a single magic number. Returns the bitfield and bool true if succeeded, nil and false otherwise.
func TryEncodeEarlyPosition(state *chess.State) (*bitfield.BitField, bool) {
	bf := &bitfield.BitField{}

	// Encode starting position.
	if state.Fullmove == 1 && state.SideToMove == chess.White && state.EnPassant == chess.InvalidSquare && state.Halfmove == 0 && state.Castling.String() == "KQkq" && len(startingPositionDiff(state.Board)) == 0 {
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

// startingPositionDiff returns the difference between the starting position and the given position.
func startingPositionDiff(board chess.Board) map[chess.Square]chess.Piece {
	startingPosition, _ := chess.ParseFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	diff := map[chess.Square]chess.Piece{}
	for i := range 64 {
		if board[i] != startingPosition.Board[i] {
			diff[chess.Square(i)] = board[i]
		}
	}
	return diff
}
