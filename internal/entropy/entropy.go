package entropy

import (
	"maps"

	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

func AdjustFrequencies(frequencies map[rune]int, square chess.Square, numPieces int, occupancy, color []bool, state *chess.State, placedPieces []rune) map[rune]int {
	freqs := maps.Clone(frequencies)

	// If it's en-passant Pawn square, the only piece possible is a Pawn.
	if state.EnPassant.IsValid() {
		pawnSquare := state.EnPassant ^ 0b1000
		if square == pawnSquare {
			freqs['r'] = 0
			freqs['n'] = 0
			freqs['b'] = 0
			freqs['p'] = 1
			freqs['k'] = 0
			freqs['q'] = 0
			return freqs
		}
	}

	// If castling is available, the only piece on the Rook square can be a Rook.
	whiteCanCastle := state.Castling['K'] || state.Castling['Q']
	blackCanCastle := state.Castling['k'] || state.Castling['q']
	if (square == chess.A1 && state.Castling['Q'] || square == chess.H1 && state.Castling['K']) ||
		(square == chess.A8 && state.Castling['q'] || square == chess.H8 && state.Castling['k']) {
		freqs['r'] = 1
		freqs['n'] = 0
		freqs['b'] = 0
		freqs['p'] = 0
		freqs['k'] = 0
		freqs['q'] = 0
		return freqs
	}
	if square == chess.E1 && whiteCanCastle || square == chess.E8 && blackCanCastle {
		// If castling is available, the only piece on the King square can be a King.
		freqs['r'] = 0
		freqs['n'] = 0
		freqs['b'] = 0
		freqs['p'] = 0
		freqs['k'] = 1
		freqs['q'] = 0
		return freqs
	}

	// Pawns cannot be on home ranks.
	if square.IsOnHomeRank() {
		freqs['p'] = 0
	}

	pieceIdx := len(placedPieces) // Number of the piece to be placed.
	pieceColor := chess.Color(color[pieceIdx])

	placedCounts := map[rune]int{} // Counts of pieces (by type) placed by the side to be placed.
	bishops := map[chess.Color]bool{}
	kings := map[chess.Color]chess.Square{
		chess.White: chess.InvalidSquare,
		chess.Black: chess.InvalidSquare,
	}
	sq, idx := chess.Square(0), 0
	for idx < len(placedPieces) {
		for !occupancy[sq] {
			sq++
		}
		piece := placedPieces[idx]
		if chess.Color(color[idx]) == pieceColor {
			placedCounts[piece]++
			if piece == 'b' {
				bishopColor := sq.Color()
				bishops[bishopColor] = true
			}
		}
		if piece == 'k' {
			kings[chess.Color(color[idx])] = sq
		}
		sq++
		idx++
	}
	kingSquare := kings[pieceColor]
	opposingKingSquare := kings[pieceColor.Other()]

	promotionPossible := state.Fullmove >= 5 && numPieces < 32

	// If promotion couldn't happen, there cannot be more pieces of a kind than there were originally on the board.
	if !promotionPossible {
		if placedCounts['q'] == 1 {
			freqs['q'] = 0
		}
		if placedCounts['b'] == 2 {
			freqs['b'] = 0
		}
		if placedCounts['n'] == 2 {
			freqs['n'] = 0
		}
		if placedCounts['r'] == 2 {
			freqs['r'] = 0
		}
	}

	// If promotion couldn't happen, there cannot be two same-color Bishops on the same side.
	if !promotionPossible && bishops[square.Color()] {
		freqs['b'] = 0
	}

	// There cannot be more than one King per side.
	if kingSquare.IsValid() {
		freqs['k'] = 0
	}

	// Kings cannot be within 1 square of one another.
	if freqs['k'] > 0 && opposingKingSquare.IsValid() &&
		tools.Abs(square.Rank()-opposingKingSquare.Rank()) <= 1 && tools.Abs(square.File()-opposingKingSquare.File()) <= 1 {
		freqs['k'] = 0
	}

	// If a side can castle, a King cannot be anywhere else than on the King's square.
	if pieceColor == chess.White && whiteCanCastle && square != chess.E1 || pieceColor == chess.Black && blackCanCastle && square != chess.E8 {
		freqs['k'] = 0
	}

	// Side not to move cannot have their King in check.
	sideNotToMove := state.SideToMove.Other()
	if freqs['k'] > 0 && pieceColor == sideNotToMove && !kings[pieceColor].IsValid() {
		// Is anything by SideToMove on the board attacking our target square?
		isAttacking := false
		board := buildBoard(occupancy, color, placedPieces)
		for sq, piece := range board {
			if piece.IsValid() && piece.Color() == state.SideToMove && chess.Attacks(board, chess.Square(sq), square) {
				isAttacking = true
				break
			}
		}

		if isAttacking { // square is attacked by state.SideToMove
			freqs['k'] = 0
		}
	} else if pieceColor == state.SideToMove && kings[sideNotToMove].IsValid() {
		// Side to move cannot be checking other side's King.
		board := buildBoard(occupancy, color, placedPieces)
		for _, pieceType := range []rune{'r', 'b', 'n', 'q', 'p'} {
			if freqs[pieceType] == 0 {
				continue
			}
			piece := chess.NewPiece(pieceType, pieceColor)
			board2 := boardWith(board, piece, square)
			if chess.Attacks(board2, square, kings[sideNotToMove]) {
				freqs[pieceType] = 0
			}
		}
	}

	return freqs
}

func UpdateFrequencies(piece rune, remaining int, frequencies map[rune]int) map[rune]int {
	freqs := maps.Clone(frequencies)

	freqs[piece] = max(freqs[piece]-100, 0)
	if piece != 'k' && piece != 'p' {
		// One of the unencountered pawns may be promoted to a piece.
		if freqs[piece] == 0 && freqs['p'] >= 100 {
			switch piece {
			case 'q':
				freqs[piece] = 100
			case 'r':
				freqs[piece] = 3
			case 'b':
				freqs[piece] = 1
			case 'n':
				freqs[piece] = 1
			}
		}
	}

	remaining = remaining*100 - freqs['k']
	for r, c := range freqs {
		if r != 'k' {
			freqs[r] = min(c, remaining)
		}
	}

	return freqs
}

func PrepareFrequencies(pieceCount int, frequencies map[rune]int) map[rune]int {
	freqs := maps.Clone(frequencies)

	pieceCount = pieceCount*100 - freqs['k']
	for r, c := range freqs {
		if r != 'k' {
			freqs[r] = min(c, pieceCount)
		}
	}

	return freqs
}

func buildBoard(occupancy, color []bool, placedPieces []rune) chess.Board {
	board := chess.Board{}
	sq := 0
	idx := 0
	for idx < len(placedPieces) {
		for !occupancy[sq] {
			sq++
		}
		board[sq] = chess.NewPiece(placedPieces[idx], chess.Color(color[idx]))
		sq++
		idx++
	}

	return board
}

func boardWith(board chess.Board, piece chess.Piece, square chess.Square) chess.Board {
	board2 := chess.Board{}
	for i := range 64 {
		board2[i] = board[i]
	}
	board2[square] = piece
	return board2
}

func FindMatchingMoveIndex(board chess.Board, moves []chess.Move) int {
	for i, m := range moves {
		if board[m.From].IsEmpty() && board[m.To].IsOccupied() {
			return i
		}
	}
	return -1
}
