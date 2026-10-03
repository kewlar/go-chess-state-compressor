package arithmetic

import (
	"fmt"
	"math/big"
)

type Interval struct {
	a, b *Rational
}

func NewInterval(a, b *Rational) *Interval {
	return &Interval{a, b}
}

func (a *Interval) Length() *Rational {
	l := NewRational(1, 1)

	return l.Sub(a.b, a.a)
}

func (a *Interval) Mul(b *Interval) *Interval {
	l := a.Length()
	a0, ba, bb := NewRational(1, 1), NewRational(1, 1), NewRational(1, 1)
	a0.Set(a.a)
	ba.Set(b.a)
	bb.Set(b.b)
	a.a.Add(a0, ba.Mul(ba, l))
	a.b.Add(a0, bb.Mul(bb, l))

	return a
}

func (a *Interval) Cover(x *Rational) bool {
	return x.Cmp(a.a) != -1 && x.Cmp(a.b) == -1
}

func (a *Interval) String() string {
	return fmt.Sprintf("[%s-%s)", a.a.String(), a.b.String())
}

func (a *Interval) ToBits() []bool {
	x, y := big.NewInt(0), big.NewInt(1)
	x1, one := big.NewInt(0), big.NewInt(1)
	c := NewBigRational(x, y)
	for c.Cmp(a.a) == -1 || c.Cmp(a.b) != -1 {
		x.Lsh(x, 1)
		y.Lsh(y, 1)
		d := NewBigRational(x, y)
		x1 = x1.Or(x, one)
		e := NewBigRational(x1, y)
		if d.Cmp(a.a) == -1 && e.Cmp(a.b) == -1 {
			x.Set(x1)
		}
		c = NewBigRational(x, y)
	}

	out := []bool{}
	for i := range y.BitLen() {
		out = append(out, x.Bit(i) == 1)
	}

	return out
}
