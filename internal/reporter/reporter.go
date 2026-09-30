package reporter

import (
	"maps"
	"slices"
)

type Reporter struct {
	Histogram map[int]int
	Count     int
	Total     uint64
	Min       int
	Max       int
	WorstCase string
}

func NewReporter() *Reporter {
	return &Reporter{
		Histogram: map[int]int{},
		Min:       -1,
		Max:       -1,
	}
}

func (r *Reporter) Log(data string, compressedSize int) {
	r.Histogram[compressedSize]++
	r.Total += uint64(compressedSize)
	r.Count++
	if r.Min == -1 || compressedSize < r.Min {
		r.Min = compressedSize
	}
	if r.Max < compressedSize {
		r.Max = compressedSize
		r.WorstCase = data
	}
}

func (r *Reporter) Mean() float64 {
	if r.Count == 0 {
		return -1
	}
	return float64(r.Total) / float64(r.Count)
}

func (r *Reporter) Median() int {
	values := slices.Collect(maps.Keys(r.Histogram))
	slices.Sort(values)
	half := r.Count / 2
	cum := 0
	for _, value := range values {
		cum += r.Histogram[value]
		if cum > half {
			return value
		}
	}
	return -1
}
