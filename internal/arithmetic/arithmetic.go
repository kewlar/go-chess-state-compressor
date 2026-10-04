package arithmetic

import (
	"maps"
	"slices"
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

func EncodeAdaptive(data []rune, occupancy []bool, frequencies map[rune]int, color []bool, side rune, fullmove int) []bool {
	freqs := maps.Clone(frequencies)
	interval := NewInterval(NewRational(0, 1), NewRational(1, 1))
	n := len(data)
	idx := 0
	for i := range 64 {
		if occupancy[i] {
			freqs2 := adjustFrequencies(freqs, i, n, occupancy, color, side, fullmove, data[:idx])
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

func DecodeAdaptive(bits []bool, n int, occupancy []bool, frequencies map[rune]int, color []bool, side rune, fullmove int) []rune {
	freqs := maps.Clone(frequencies)
	data := make([]rune, 0, n)
	x := NewRationalFromBits(bits)
	idx := 0
	for square, occupied := range occupancy {
		if occupied {
			freqs2 := adjustFrequencies(freqs, square, n, occupancy, color, side, fullmove, data)
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

func adjustFrequencies(frequencies map[rune]int, square int, numPieces int, occupancy, color []bool, side rune, fullmove int, placedPieces []rune) map[rune]int {
	freqs := maps.Clone(frequencies)

	rank, file := square>>3, square&7
	squareColor := (rank+file)&1 == 1

	// Pawns cannot be on home ranks.
	if rank == 0 || rank == 7 {
		freqs['p'] = 0
	}

	pieceIdx := len(placedPieces) // Number of the piece to be placed.
	pieceColor := color[pieceIdx]

	placedCounts := map[rune]int{} // Counts of pieces (by type) placed by the side to be placed.
	bishops := map[bool]bool{}
	i, idx := 0, 0
	for idx < len(placedPieces) {
		for !occupancy[i] {
			i++
		}
		if color[idx] == pieceColor {
			piece := placedPieces[idx]
			placedCounts[piece]++
			if piece == 'b' {
				rank, file := i>>3, i&7
				bishopColor := (rank+file)&1 == 1
				bishops[bishopColor] = true
			}
		}
		i++
		idx++
	}

	promotionPossible := fullmove >= 5 && numPieces < 32

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
	if !promotionPossible && bishops[squareColor] {
		freqs['b'] = 0
	}

	// TODO: Side not to move cannot have their King in check.
	// TODO: Kings cannot be within 1 square of one another.
	// TODO: If castling is available, the only piece on the Rook square can be a Rook.

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
