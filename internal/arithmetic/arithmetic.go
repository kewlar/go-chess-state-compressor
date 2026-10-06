package arithmetic

import (
	"fmt"
	"maps"
	"slices"

	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
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
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(i), n, occupancy, color, state, data[:idx])
			intervals := FrequenciesToIntervals(freqs2)
			r := data[idx]
			if freqs2[r] == 0 {
				panic(fmt.Sprintf("didn't expect '%s' at square %s", string(r), chess.Square(i).String()))
			}
			interval.Mul(intervals[r])
			idx++

			freqs = entropy.UpdateFrequencies(r, n-idx, freqs)
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
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(square), n, occupancy, color, state, data)
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

			freqs = entropy.UpdateFrequencies(piece, n-idx, freqs)
		}
	}

	return data
}
