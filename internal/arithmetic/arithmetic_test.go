package arithmetic_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/arithmetic"
)

func TestConcurrentArithmetic(t *testing.T) {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}
	intervals := arithmetic.FrequenciesToIntervals(frequencies)
	sample := []rune("rnbqkbnrpppppppppppppppprnbqkbnr")

	wg := sync.WaitGroup{}
	for range 16 {
		wg.Go(func() {
			for range 20 {
				encoded := arithmetic.Encode(sample, intervals)
				decoded := arithmetic.Decode(encoded, len(sample), intervals)
				if !slices.Equal(sample, decoded) {
					t.Errorf("Decode(Encode(%q)) = %q; want %q", string(sample), string(decoded), string(sample))
				}
			}
		})
	}
	wg.Wait()
}
