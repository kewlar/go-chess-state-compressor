package chess

import "fmt"

// Attacks returns true if the piece in the 'from' square can attack the piece on the 'target' square'.
func Attacks(board Board, from, target Square) bool {
	piece := board[from]
	if piece == 0 {
		return false
	}

	dr := target.Rank() - from.Rank()
	df := target.File() - from.File()
	adr, adf := abs(dr), abs(df)
	switch piece | 32 {
	case 'k':
		if adr <= 1 && adf <= 1 {
			return true
		}
	case 'q':
		if (adr == adf || adr == 0 || adf == 0) && rayHits(board, from, target) {
			return true
		}
	case 'r':
		if (adr == 0 || adf == 0) && rayHits(board, from, target) {
			return true
		}
	case 'b':
		if (adr == adf) && rayHits(board, from, target) {
			return true
		}
	case 'n':
		if adr == 2 && adf == 1 || adr == 1 && adf == 2 {
			return true
		}
	case 'p':
		if adr == 1 && adf == 1 {
			c := Color(piece&32 == 0)
			if c == White && dr == 1 || c == Black && dr == -1 {
				return true
			}
		}
	default:
		panic(fmt.Sprintf("Unexpected piece: '%s'", string(piece)))
	}
	return false
}

// rayHits returns true if a piece on 'from' square has a line-of-sight visibility to 'target' square without any obstacles.
// WARNING: This function doesn't check if it's actually possible to hit the target square, and may loop forever.
func rayHits(board Board, from, target Square) bool {
	dr := target.Rank() - from.Rank()
	df := target.File() - from.File()
	ur, uf := sign(dr), sign(df)
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

// abs returns the absolute value of an integer number x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// sign returns the sign (-1, 0, or 1) of an integer number x.
func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// boardToString renders a board into string for printing. Used for debugging.
func boardToString(board Board) string {
	out := ""
	for i := range 64 {
		out += " " + board[i].String()
		if i&7 == 7 {
			out += "\n"
		}
	}
	return out
}
