package codecs

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
	"example.com/go-chess/internal/tools"
)

var startingPosition = chess.NewStartingState().Board

// TryEncodeEarlyPosition tries to encode one of the 421 positions possible in first 2 plies in a single magic number. Returns the bitfield and bool true if succeeded, nil and false otherwise.
func TryEncodeEarlyPosition(state *chess.State) (*bitfield.BitField, bool) {
	bf := &bitfield.BitField{}

	isFullCastling := state.Castling.String() == "KQkq"
	diff := startingPositionDiff(state.Board)

	// Encode starting position.
	if state.Fullmove == 1 && state.SideToMove == chess.White && state.EnPassant == chess.InvalidSquare && state.Halfmove == 0 && isFullCastling && len(diff) == 0 {
		bf.WriteInt(0, 21) // moveNum
		return bf, true
	}

	// Encode the board after first move by White
	if state.Fullmove == 1 && state.SideToMove == chess.Black && isFullCastling && len(diff) == 2 {
		whiteMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			return nil, false
		}

		bf.WriteInt(1, 21) // moveNum
		bf.WriteInt(uint64(whiteMove), 5)
		return bf, true
	}

	// Encode the board after first move by Black
	if state.Fullmove == 2 && state.SideToMove == chess.White && isFullCastling && len(diff) == 4 {
		whiteMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			return nil, false
		}
		blackMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstBlackMoves())
		if blackMove == -1 { // Not found.
			return nil, false
		}

		bf.WriteInt(2, 21) // moveNum
		bf.WriteInt(uint64(whiteMove*20+blackMove), 9)
		return bf, true
	}

	return nil, false
}

// startingPositionDiff returns the difference between the starting position and the given position.
func startingPositionDiff(board chess.Board) map[chess.Square]chess.Piece {
	diff := map[chess.Square]chess.Piece{}
	for i := range 64 {
		if board[i] != startingPosition[i] {
			diff[chess.Square(i)] = board[i]
		}
	}
	return diff
}

// TryDecodeEarlyPosition tries to decode the early position given moveNum and access to encoded data. Returns the chess state and bool true if succeeded.
func TryDecodeEarlyPosition(moveNum int, data *bitfield.BitField) (*chess.State, bool) {
	var state *chess.State
	switch moveNum {
	case 0:
		state = chess.NewStartingState()
	case 1:
		state = chess.NewStartingState()
		whiteMove := int(data.ReadInt(5))
		applyMove(state, chess.FirstWhiteMoves()[whiteMove])
	case 2:
		state = chess.NewStartingState()
		num := int(data.ReadInt(9))
		whiteMove, blackMove := num/20, num%20
		applyMove(state, chess.FirstWhiteMoves()[whiteMove])
		applyMove(state, chess.FirstBlackMoves()[blackMove])
	default:
		return nil, false
	}

	return state, true
}

// applyMove modifies the state by applying the given move.
func applyMove(state *chess.State, move chess.Move) {
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
	state.SideToMove = state.SideToMove.Other()
	if state.SideToMove == chess.White {
		state.Fullmove++
	}
}

// separateBoardInformation separates board into separate variables for occupancy bitboard, color, and piece list.
func separateBoardInformation(board chess.Board) (occupancy uint64, color uint32, pieces []chess.Piece) {
	for square, piece := range board.OccupiedSquares() {
		occupancy |= 1 << square
		if piece.Color() == chess.White {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, piece)
	}
	return
}

// placePieces places pieces on board given an occupancy bitboard, and lists of pieces and their colors.
func placePieces(board *chess.Board, occupancy []bool, pieces []rune, colors []bool) {
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			board[i] = chess.NewPiece(pieces[idx], chess.Color(colors[idx]))
			idx++
		}
	}
}

func piecesToRunes(pieces []chess.Piece) []rune {
	runes := make([]rune, len(pieces))
	for i, piece := range pieces {
		runes[i] = rune(piece.Type())
	}
	return runes
}

func maybeOriginalColorRanks(state *chess.State) bool {
	return state.Fullmove < 3 || state.Fullmove == 3 && state.SideToMove == chess.White
}

func reallyOriginalColorRanks(board chess.Board) bool {
	for i := range 16 {
		if board[i] != chess.Empty && board[i].Color() != chess.Black {
			return false
		}
		j := 48 + i
		if board[j] != chess.Empty && board[j].Color() != chess.White {
			return false
		}
	}
	return true
}
