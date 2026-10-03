package arithmetic

func FrequenciesToIntervals(frequencies map[rune]int) map[rune]*Interval {
	tot := 0
	for _, freq := range frequencies {
		tot += freq
	}
	total := int64(tot)

	cum := 0
	intervals := map[rune]*Interval{}
	for r, freq := range frequencies {
		i0, i1 := int64(cum), int64(cum+freq)
		intervals[r] = &Interval{NewRational(i0, total), NewRational(i1, total)}
		cum += freq
	}

	return intervals
}
