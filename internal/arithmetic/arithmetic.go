package arithmetic

import "slices"

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
