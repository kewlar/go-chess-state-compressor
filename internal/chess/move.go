package chess

type Move struct {
	From, To Square
}

// FirstWhiteMoves returns the 20 first moves available for White.
func FirstWhiteMoves() []Move {
	moves := []Move{}
	for file := range 8 {
		// Possible first Pawn moves.
		moves = append(moves,
			Move{NewSquareFromRankAndFile(1, file), NewSquareFromRankAndFile(3, file)},
			Move{NewSquareFromRankAndFile(1, file), NewSquareFromRankAndFile(2, file)},
		)
	}
	// Possible first Knight moves.
	moves = append(moves,
		Move{NewSquareFromRankAndFile(0, 1), NewSquareFromRankAndFile(2, 0)},
		Move{NewSquareFromRankAndFile(0, 1), NewSquareFromRankAndFile(2, 2)},
		Move{NewSquareFromRankAndFile(0, 6), NewSquareFromRankAndFile(2, 5)},
		Move{NewSquareFromRankAndFile(0, 6), NewSquareFromRankAndFile(2, 7)},
	)

	return moves
}

// FirstBlackMoves returns the 20 first moves available for Black.
func FirstBlackMoves() []Move {
	moves := []Move{}
	for file := range 8 {
		// Possible first Pawn moves.
		moves = append(moves,
			Move{NewSquareFromRankAndFile(6, file), NewSquareFromRankAndFile(4, file)},
			Move{NewSquareFromRankAndFile(6, file), NewSquareFromRankAndFile(5, file)},
		)
	}
	// Possible first Knight moves.
	moves = append(moves,
		Move{NewSquareFromRankAndFile(7, 1), NewSquareFromRankAndFile(5, 0)},
		Move{NewSquareFromRankAndFile(7, 1), NewSquareFromRankAndFile(5, 2)},
		Move{NewSquareFromRankAndFile(7, 6), NewSquareFromRankAndFile(5, 5)},
		Move{NewSquareFromRankAndFile(7, 6), NewSquareFromRankAndFile(5, 7)},
	)

	return moves
}
