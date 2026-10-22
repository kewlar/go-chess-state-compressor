package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
	"example.com/go-chess/internal/tools"
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

func TryDecodeEarlyPosition(moveNum int, data *bitfield.BitField) (*chess.State, bool) {
	var state *chess.State
	switch moveNum {
	case 0:
		state = chess.NewStartingState()
	case 1:
		num := int(data.ReadInt(5))
		state = chess.NewStartingState()
		move := chess.FirstWhiteMoves()[num]
		state.Board[move.To] = state.Board[move.From]
		state.Board[move.From] = chess.Empty
		if state.Board[move.To].Type() == chess.Pawn && tools.Abs(move.From.Rank()-move.To.Rank()) == 2 {
			state.EnPassant = chess.NewSquareFromRankAndFile((move.From.Rank()+move.To.Rank())/2, move.To.File())
		}
		if state.Board[move.To].Type() != chess.Pawn {
			state.Halfmove++
		}
		state.SideToMove = chess.Black
	case 2:
		num := int(data.ReadInt(9))
		whiteMove, blackMove := num/20, num%20

		state = chess.NewStartingState()
		// Apply move by White.
		move := chess.FirstWhiteMoves()[whiteMove]
		state.Board[move.To] = state.Board[move.From]
		state.Board[move.From] = chess.Empty
		if state.Board[move.To].Type() == chess.Pawn && tools.Abs(move.From.Rank()-move.To.Rank()) == 2 {
			state.EnPassant = chess.NewSquareFromRankAndFile((move.From.Rank()+move.To.Rank())/2, move.To.File())
		}
		if state.Board[move.To].Type() != chess.Pawn {
			state.Halfmove++
		}
		// Apply move by Black.
		move = chess.FirstBlackMoves()[blackMove]
		state.Board[move.To] = state.Board[move.From]
		state.Board[move.From] = chess.Empty
		if state.Board[move.To].Type() == chess.Pawn && tools.Abs(move.From.Rank()-move.To.Rank()) == 2 {
			state.EnPassant = chess.NewSquareFromRankAndFile((move.From.Rank()+move.To.Rank())/2, move.To.File())
		} else {
			state.EnPassant = chess.InvalidSquare
		}
		if state.Board[move.To].Type() != chess.Pawn {
			state.Halfmove++
		} else {
			state.Halfmove = 0
		}

		state.Fullmove = 2
		state.SideToMove = chess.White
	default:
		return nil, false
	}

	return state, true
}
