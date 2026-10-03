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

func EncodeAdaptive(data []rune, occupancy uint64, frequencies map[rune]int) []bool {
	freqs := maps.Clone(frequencies)
	interval := NewInterval(NewRational(0, 1), NewRational(1, 1))
	n := len(data)
	idx := 0
	for i := range 64 {
		if occupancy&(1<<i) != 0 {
			r := data[idx]
			freqs2 := maps.Clone(freqs)
			rank := i >> 3
			if rank == 0 || rank == 7 {
				freqs2['p'] = 0
			}
			intervals := FrequenciesToIntervals(freqs2)
			interval.Mul(intervals[r])
			idx++

			freqs = updateFrequencies(r, n-idx, freqs)
		}
	}

	return interval.ToBits()
}

func DecodeAdaptive(bits []bool, n int, occupancy []bool, frequencies map[rune]int) []rune {
	freqs := maps.Clone(frequencies)
	data := make([]rune, 0, n)
	x := NewRationalFromBits(bits)
	idx := 0
	for square, occupied := range occupancy {
		if occupied {
			rank := square >> 3
			piece := rune(0)
			freqs2 := maps.Clone(freqs)
			if rank == 0 || rank == 7 {
				freqs2['p'] = 0
			}
			intervals := FrequenciesToIntervals(freqs2)
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
