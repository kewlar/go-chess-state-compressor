package chess

import "iter"

type Board [64]Piece

// OccupiedSquares returns an iterator over occupied squares of the board, which yields the occupied squares in order.
func (board Board) OccupiedSquares() iter.Seq2[Square, Piece] {
	return func(yield func(Square, Piece) bool) {
		for square, piece := range board {
			if !piece.IsEmpty() {
				if !yield(Square(square), piece) {
					return
				}
			}
		}
	}
}

// Render renders a board into string for printing. Used for debugging.
func (board Board) Render() string {
	out := ""
	for i := range 64 {
		out += " " + board[i].String()
		if i&7 == 7 {
			out += "\n"
		}
	}
	return out
}
