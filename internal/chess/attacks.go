package chess

import (
	"fmt"

	"example.com/go-chess/internal/tools"
)

// Attacks returns true if the piece in the 'from' square can attack the piece on the 'target' square'.
func Attacks(board Board, from, target Square) bool {
	piece := board[from]
	if piece == Empty {
		return false
	}

	dr := target.Rank() - from.Rank()
	df := target.File() - from.File()
	adr, adf := tools.Abs(dr), tools.Abs(df)
	switch piece.Type() {
	case King:
		if adr <= 1 && adf <= 1 {
			return true
		}
	case Queen:
		if (adr == adf || adr == 0 || adf == 0) && rayHits(board, from, target) {
			return true
		}
	case Rook:
		if (adr == 0 || adf == 0) && rayHits(board, from, target) {
			return true
		}
	case Bishop:
		if (adr == adf) && rayHits(board, from, target) {
			return true
		}
	case Knight:
		if adr == 2 && adf == 1 || adr == 1 && adf == 2 {
			return true
		}
	case Pawn:
		if adr == 1 && adf == 1 {
			c := piece.Color()
			if c == White && dr == 1 || c == Black && dr == -1 {
				return true
			}
		}
	default:
		panic(fmt.Sprintf("Unexpected piece: %q", piece))
	}
	return false
}

// rayHits returns true if a piece on 'from' square has a line-of-sight visibility to 'target' square without any obstacles.
// WARNING: This function doesn't check if it's actually possible to hit the target square, and may loop forever.
func rayHits(board Board, from, target Square) bool {
	dr := target.Rank() - from.Rank()
	df := target.File() - from.File()
	ur, uf := tools.Sign(dr), tools.Sign(df)
	r, f := from.Rank()+ur, from.File()+uf
	for !(r == target.Rank() && f == target.File()) {
		square := NewSquareFromRankAndFile(r, f)
		if board[square] != 0 {
			return false
		}
		r += ur
		f += uf
	}
	return true
}
