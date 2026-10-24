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
func (a *Interval) Length() *Rational {
	tmp := getScratch()
	defer putScratch(tmp)
	return a.length(tmp)
}

func (a *Interval) length(tmp *scratch) *Rational {
	return NewRational(0, 1).sub(a.b, a.a, tmp)
}

// Mul multiplies one interval by another. Or, to put it differently, it returns interval a inside interval b.
func (a *Interval) Mul(b *Interval) *Interval {
	tmp := getScratch()
	defer putScratch(tmp)
	return a.mul(b, tmp)
}

func (a *Interval) mul(b *Interval, tmp *scratch) *Interval {
	l := a.length(tmp)
	tmp.a0.Set(a.a)
	tmp.ba.Set(b.a)
	tmp.bb.Set(b.b)
	a.a.add(tmp.a0, tmp.ba.mul(tmp.ba, l, tmp), tmp)
	a.b.add(tmp.a0, tmp.bb.mul(tmp.bb, l, tmp), tmp)

	return a
}

// Cover returns true if x falls inside interval a (a.a <= x < a.b.).
func (a *Interval) Cover(x *Rational) bool {
	tmp := getScratch()
	defer putScratch(tmp)
	return a.cover(x, tmp)
}

func (a *Interval) cover(x *Rational, tmp *scratch) bool {
	return x.cmp(a.a, tmp) != -1 && x.cmp(a.b, tmp) == -1
}

// String renders interval to string.
func (a *Interval) String() string {
	return fmt.Sprintf("[%s-%s)", a.a.String(), a.b.String())
}

// ToBits returns the shortest binary code that represents a Rational fraction inside interval a.
func (a *Interval) ToBits() []bool {
	tmp := getScratch()
	defer putScratch(tmp)
	return a.toBits(tmp)
}

func (a *Interval) toBits(tmp *scratch) []bool {
	x, y := big.NewInt(0), big.NewInt(1)
	x1, one := big.NewInt(0), big.NewInt(1)
	tmp.c.Num.Set(x)
	tmp.c.Denom.Set(y)
	for tmp.c.cmp(a.a, tmp) == -1 || tmp.c.cmp(a.b, tmp) != -1 {
		x.Lsh(x, 1)
		y.Lsh(y, 1)
		x1 = x1.Or(x, one)
		tmp.d.Num.Set(x1)
		tmp.d.Denom.Set(y)
		if tmp.c.cmp(a.a, tmp) == -1 && tmp.d.cmp(a.b, tmp) == -1 {
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
