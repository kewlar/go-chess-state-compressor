package arithmetic

import (
	"fmt"
	"math/big"
)

// Interval represents an interval between two Rationals [a..b).
type Interval struct {
	a, b *Rational
}

// NewInterval creates a new interval [a..b).
func NewInterval(a, b *Rational) *Interval {
	return &Interval{a, b}
}

// Length returns the length of the interval.
func (i *Interval) Length() *Rational {
	tmp := getScratch()
	defer putScratch(tmp)
	return i.length(tmp)
}

func (i *Interval) length(tmp *scratch) *Rational {
	return NewRational(0, 1).sub(i.b, i.a, tmp)
}

// Mul multiplies one interval by another. Or, to put it differently, it returns interval a inside interval b.
func (i *Interval) Mul(b *Interval) *Interval {
	tmp := getScratch()
	defer putScratch(tmp)
	return i.mul(b, tmp)
}

func (i *Interval) mul(b *Interval, tmp *scratch) *Interval {
	l := i.length(tmp)
	tmp.a0.Set(i.a)
	tmp.ba.Set(b.a)
	tmp.bb.Set(b.b)
	i.a.add(tmp.a0, tmp.ba.mul(tmp.ba, l, tmp), tmp)
	i.b.add(tmp.a0, tmp.bb.mul(tmp.bb, l, tmp), tmp)

	return i
}

// Cover returns true if x falls inside interval a (a.a <= x < a.b.).
func (i *Interval) Cover(x *Rational) bool {
	tmp := getScratch()
	defer putScratch(tmp)
	return i.cover(x, tmp)
}

func (i *Interval) cover(x *Rational, tmp *scratch) bool {
	return x.cmp(i.a, tmp) != -1 && x.cmp(i.b, tmp) == -1
}

// String renders interval to string.
func (i *Interval) String() string {
	return fmt.Sprintf("[%s-%s)", i.a.String(), i.b.String())
}

// ToBits returns the shortest binary code that represents a Rational fraction inside interval a.
func (i *Interval) ToBits() []bool {
	tmp := getScratch()
	defer putScratch(tmp)
	return i.toBits(tmp)
}

func (i *Interval) toBits(tmp *scratch) []bool {
	x, y := big.NewInt(0), big.NewInt(1)
	x1, one := big.NewInt(0), big.NewInt(1)
	tmp.c.Num.Set(x)
	tmp.c.Denom.Set(y)
	for tmp.c.cmp(i.a, tmp) == -1 || tmp.c.cmp(i.b, tmp) != -1 {
		x.Lsh(x, 1)
		y.Lsh(y, 1)
		x1 = x1.Or(x, one)
		tmp.d.Num.Set(x1)
		tmp.d.Denom.Set(y)
		if tmp.c.cmp(i.a, tmp) == -1 && tmp.d.cmp(i.b, tmp) == -1 {
			x.Set(x1)
		}
		tmp.c.Num.Set(x)
		tmp.c.Denom.Set(y)
	}

	bitCount := y.BitLen() - 1 // The MSB is always 0.
	out := make([]bool, bitCount)
	for i := range bitCount {
		out[i] = x.Bit(i) == 1
	}

	return out
}
