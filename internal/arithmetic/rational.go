package arithmetic

import (
	"fmt"
	"math/big"
)

// A Rational represents a ratio Num/Denom of arbitrary precision.
type Rational struct {
	Num, Denom *big.Int
}

var (
	// Temporary variables for calculations, to avoid allocating big.Ints on every operation.
	a = big.NewInt(0)
	b = big.NewInt(0)
)

// NewRational creates a new Rational with numerator num and denominator denom.
func NewRational(num, denom int64) *Rational {
	return &Rational{
		Num:   big.NewInt(num),
		Denom: big.NewInt(denom),
	}
}

// NewBigRational creates a new Rational with numerator num and denominator denom.
func NewBigRational(num, denom *big.Int) *Rational {
	a, b := big.NewInt(0), big.NewInt(0)
	return &Rational{
		Num:   a.Set(num),
		Denom: b.Set(denom),
	}
}

// NewRationalFromBits creates a new Rational with numerator 0.<bits-LSB-first> and denominator 1.0.
func NewRationalFromBits(bits []bool) *Rational {
	z := NewRational(0, 1)
	for i, bit := range bits {
		if bit {
			z.Num.SetBit(z.Num, i, 1)
		}
	}
	z.Denom.Lsh(z.Denom, uint(len(bits)))

	return z
}

// Set sets z to y.
func (z *Rational) Set(y *Rational) {
	z.Num.Set(y.Num)
	z.Denom.Set(y.Denom)
}

// Add sets z to the sum x+y and returns z.
func (z *Rational) Add(x, y *Rational) *Rational {
	a.Mul(x.Num, y.Denom)
	b.Mul(y.Num, x.Denom)
	z.Num.Add(a, b)
	z.Denom.Mul(x.Denom, y.Denom)
	z.Reduce()
	return z
}

// Sub sets z to the difference x-y and returns z.
func (z *Rational) Sub(x, y *Rational) *Rational {
	a.Mul(x.Num, y.Denom)
	b.Mul(y.Num, x.Denom)
	z.Num.Sub(a, b)
	z.Denom.Mul(x.Denom, y.Denom)
	z.Reduce()
	return z
}

// Mul sets z to the product x*y and returns z.
func (z *Rational) Mul(x, y *Rational) *Rational {
	z.Num.Mul(x.Num, y.Num)
	z.Denom.Mul(x.Denom, y.Denom)
	z.Reduce()
	return z
}

// Div sets z to the ratio x/y and returns z.
func (z *Rational) Div(x, y *Rational) *Rational {
	z.Num.Mul(x.Num, y.Denom)
	z.Denom.Mul(x.Denom, y.Num)
	z.Reduce()
	return z
}

// Reduce reduces z by dividing z.Num and z.Denom by their greatest common denominator, so that z.Num and z.Denom become coprime, and the ratio cannot be further reduced.
func (z *Rational) Reduce() {
	a.GCD(nil, nil, z.Num, z.Denom)
	z.Num.Div(z.Num, a)
	z.Denom.Div(z.Denom, a)
}

// String returns the ratio as string "<num>/<denom>".
func (z *Rational) String() string {
	return fmt.Sprintf("%s/%s", z.Num.String(), z.Denom.String())
}

// Cmp compares x and y and returns:
//   - -1 if x < y;
//   - 0 if x == y;
//   - +1 if x > y.
func (x *Rational) Cmp(y *Rational) int {
	a.Mul(x.Num, y.Denom)
	b.Mul(y.Num, x.Denom)

	return a.Cmp(b)
}

// ScaleWithInterval performs z = (z - i.a) / i.Length(). z, i.a, and i.b are all in range [0, 1).
func (z *Rational) ScaleWithInterval(i *Interval) *Rational {
	z.Sub(z, i.a)
	z.Div(z, i.Length())

	return z
}
