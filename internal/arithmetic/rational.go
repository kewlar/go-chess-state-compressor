package arithmetic

import (
	"fmt"
	"math/big"
)

type Rational struct {
	Num, Denom *big.Int
}

func NewRational(num, denom int64) *Rational {
	return &Rational{
		Num:   big.NewInt(num),
		Denom: big.NewInt(denom),
	}
}

func NewBigRational(num, denom *big.Int) *Rational {
	a, b := big.NewInt(0), big.NewInt(0)
	return &Rational{
		Num:   a.Set(num),
		Denom: b.Set(denom),
	}
}

func NewRationalFromBits(bits []bool) *Rational {
	z := NewRational(0, 1)
	for i, bit := range bits {
		if bit {
			z.Num.SetBit(z.Num, i, 1)
		}
	}
	z.Denom.Lsh(z.Denom, uint(len(bits)-1))

	return z
}

func (z *Rational) Copy() *Rational {
	y := NewRational(0, 1)
	y.Set(z)
	return y
}

func (z *Rational) Set(y *Rational) {
	z.Num.Set(y.Num)
	z.Denom.Set(y.Denom)
}

func (z *Rational) Add(x, y *Rational) *Rational {
	a := big.NewInt(0)
	b := big.NewInt(0)
	a.Mul(x.Num, y.Denom)
	b.Mul(y.Num, x.Denom)
	z.Num.Add(a, b)
	z.Denom.Mul(x.Denom, y.Denom)
	z.Simplify()
	return z
}

func (z *Rational) Sub(x, y *Rational) *Rational {
	a := big.NewInt(0)
	b := big.NewInt(0)
	a.Mul(x.Num, y.Denom)
	b.Mul(y.Num, x.Denom)
	z.Num.Sub(a, b)
	z.Denom.Mul(x.Denom, y.Denom)
	z.Simplify()
	return z
}

func (z *Rational) Mul(x, y *Rational) *Rational {
	z.Num.Mul(x.Num, y.Num)
	z.Denom.Mul(x.Denom, y.Denom)
	z.Simplify()
	return z
}

func (z *Rational) Div(x, y *Rational) *Rational {
	z.Num.Mul(x.Num, y.Denom)
	z.Denom.Mul(x.Denom, y.Num)
	z.Simplify()
	return z
}

func (z *Rational) Simplify() {
	a := big.NewInt(0)
	a.GCD(nil, nil, z.Num, z.Denom)
	z.Num.Div(z.Num, a)
	z.Denom.Div(z.Denom, a)
}

func (z *Rational) String() string {
	return fmt.Sprintf("%s/%s", z.Num.String(), z.Denom.String())
}

func (x *Rational) Cmp(y *Rational) int {
	a := big.NewInt(0)
	b := big.NewInt(0)
	a.Mul(x.Num, y.Denom)
	b.Mul(y.Num, x.Denom)

	return a.Cmp(b)
}

func (z *Rational) ScaleWithInterval(i *Interval) *Rational {
	z.Sub(z, i.a)
	z.Div(z, i.Length())

	return z
}
