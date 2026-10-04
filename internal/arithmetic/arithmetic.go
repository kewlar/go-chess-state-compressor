package arithmetic

import (
	"maps"
	"slices"

	"example.com/go-chess/internal/chess"
)

func FrequenciesToIntervals(frequencies map[rune]int) map[rune]*Interval {
	pieces := []rune{}
	tot := 0
	for r, freq := range frequencies {
		tot += freq
		pieces = append(pieces, r)
	}
	total := int64(tot)

	slices.SortFunc(pieces, func(a, b rune) int {
		if frequencies[a] > frequencies[b] {
			return -1
		}
		if frequencies[a] < frequencies[b] {
			return 1
		}
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})

	cum := 0
	intervals := map[rune]*Interval{}
	for _, r := range pieces {
		freq := frequencies[r]
		i0, i1 := int64(cum), int64(cum+freq)
		intervals[r] = &Interval{NewRational(i0, total), NewRational(i1, total)}
		cum += freq
	}

	return intervals
}

func Encode(data []rune, intervals map[rune]*Interval) []bool {
	interval := NewInterval(NewRational(0, 1), NewRational(1, 1))
	for _, r := range data {
		interval.Mul(intervals[r])
	}

	return interval.ToBits()
}

func Decode(bits []bool, n int, intervals map[rune]*Interval) []rune {
	data := make([]rune, 0, n)
	x := NewRationalFromBits(bits)
	for range n {
		piece := rune(0)
		for r, i := range intervals {
			if i.Cover(x) {
				piece = r
				break
			}
		}
		if piece == 0 {
			panic("covering interval not found")
		}
		data = append(data, piece)
		x.ScaleWithInterval(intervals[piece])
	}

	return data
}

func EncodeAdaptive(data []rune, occupancy []bool, frequencies map[rune]int, color []bool, state *chess.State) []bool {
	freqs := maps.Clone(frequencies)
	interval := NewInterval(NewRational(0, 1), NewRational(1, 1))
	n := len(data)
	idx := 0
	for i := range 64 {
		if occupancy[i] {
			freqs2 := adjustFrequencies(freqs, chess.Square(i), n, occupancy, color, state, data[:idx])
			intervals := FrequenciesToIntervals(freqs2)
			r := data[idx]
			if freqs2[r] == 0 {
				panic("didn't expect this piece at this position")
			}
			interval.Mul(intervals[r])
			idx++

			freqs = updateFrequencies(r, n-idx, freqs)
		}
	}

	return interval.ToBits()
}

func DecodeAdaptive(bits []bool, n int, occupancy []bool, frequencies map[rune]int, color []bool, state *chess.State) []rune {
	freqs := maps.Clone(frequencies)
	data := make([]rune, 0, n)
	x := NewRationalFromBits(bits)
	idx := 0
	for square, occupied := range occupancy {
		if occupied {
			freqs2 := adjustFrequencies(freqs, chess.Square(square), n, occupancy, color, state, data)
			intervals := FrequenciesToIntervals(freqs2)
			piece := rune(0)
			for r, i := range intervals {
				if i.Cover(x) {
					piece = r
					break
				}
			}
			if piece == 0 {
				panic("covering interval not found")
			}
			data = append(data, piece)
			idx++
			x.ScaleWithInterval(intervals[piece])

			freqs = updateFrequencies(piece, n-idx, freqs)
		}
	}

	return data
}

func adjustFrequencies(frequencies map[rune]int, square chess.Square, numPieces int, occupancy, color []bool, state *chess.State, placedPieces []rune) map[rune]int {
	freqs := maps.Clone(frequencies)

	// Pawns cannot be on home ranks.
	if square.IsOnHomeRank() {
		freqs['p'] = 0
	}

	pieceIdx := len(placedPieces) // Number of the piece to be placed.
	pieceColor := chess.Color(color[pieceIdx])

	placedCounts := map[rune]int{} // Counts of pieces (by type) placed by the side to be placed.
	bishops := map[chess.Color]bool{}
	hasKing := false
	kingSquare := chess.InvalidSquare
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
			} else if piece == 'k' {
				hasKing = true
			}
		}
		if piece == 'k' {
			kingSquare = sq
		}
		sq++
		idx++
	}

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
	if hasKing {
		freqs['k'] = 0
	}

	// Kings cannot be within 1 square of one another.
	if kingSquare.IsValid() {
		deltaRank := square.Rank() - kingSquare.Rank()
		deltaFile := square.File() - kingSquare.File()
		if deltaRank < 0 {
			deltaRank *= -1
		}
		if deltaFile < 0 {
			deltaFile *= -1
		}
		if deltaRank <= 1 && deltaFile <= 1 {
			freqs['k'] = 0
		}
	}

	// If castling is available, the only piece on the Rook square can be a Rook.
	sqs := square.String()
	if pieceColor == chess.White && (sqs == "a1" && state.Castling['Q'] || sqs == "h1" && state.Castling['K']) ||
		pieceColor == chess.Black && (sqs == "a8" && state.Castling['q'] || sqs == "h8" && state.Castling['k']) {
		freqs['r'] = 1
		freqs['n'] = 0
		freqs['b'] = 0
		freqs['p'] = 0
		freqs['k'] = 0
		freqs['q'] = 0
	}

	// TODO: Side not to move cannot have their King in check.

	return freqs
}

func updateFrequencies(piece rune, remaining int, frequencies map[rune]int) map[rune]int {
	freqs := maps.Clone(frequencies)

	switch piece {
	case 'k':
		freqs[piece]--
	case 'p':
		freqs[piece]--
	default:
		freqs[piece]--
		// One of the unencountered pawns may be promoted to a piece.
		if freqs[piece] == 0 && freqs['p'] > 1 {
			freqs[piece] = 1
		}
	}
	remaining -= freqs['k']
	for r, c := range freqs {
		if r != 'k' {
			freqs[r] = min(c, remaining)
		}
	}

	return freqs
}
