package chess

type Move struct {
	From, To Square
}

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
