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

// Copy returns a copy of interval a.
func (a *Interval) Copy() *Interval {
	return NewInterval(a.a, a.b)
}

// Length returns the length of the interval.
func (a *Interval) Length() *Rational {
	l := NewRational(1, 1)

	return l.Sub(a.b, a.a)
}

// Mul multiplies one interval by another. Or, to put it differently, it returns interval a inside interval b.
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

// Cover returns true if x falls inside interval a (a.a <= x < a.b.).
func (a *Interval) Cover(x *Rational) bool {
	return x.Cmp(a.a) != -1 && x.Cmp(a.b) == -1
}

// String renders interval to string.
func (a *Interval) String() string {
	return fmt.Sprintf("[%s-%s)", a.a.String(), a.b.String())
}

// ToBits returns the shortest binary code that represents a Rational fraction inside interval a.
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

	bitCount := y.BitLen() - 1 // The MSB is always 0.
	out := make([]bool, bitCount)
	for i := range bitCount {
		out[i] = x.Bit(i) == 1
	}

	return out
}
