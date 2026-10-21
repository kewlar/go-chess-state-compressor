package arithmetic

import (
	"math/big"
	"sync"
)

// Temporary variables for calculations, to avoid allocating big.Ints on every operation.
type scratch struct {
	a, b                *big.Int
	l, a0, ba, bb, c, d *Rational
}

var scratchPool = sync.Pool{
	New: func() any {
		s := &scratch{
			a:  big.NewInt(0),
			b:  big.NewInt(0),
			l:  NewRational(1, 1),
			a0: NewRational(1, 1),
			ba: NewRational(1, 1),
			bb: NewRational(1, 1),
			c:  NewRational(1, 1),
			d:  NewRational(1, 1),
		}
		return s
	},
}

func getScratch() *scratch {
	return scratchPool.Get().(*scratch)
}

func putScratch(s *scratch) {
	scratchPool.Put(s)
}
