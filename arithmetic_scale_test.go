package decimal

import (
	"errors"
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

// Round a rational by comparing its distance from the neighboring integers.
// This reference does not use Decimal rounding or its interval primitives.
func roundedRationalReference(r *big.Rat, prec int, mode RoundingMode) (*big.Int, bool) {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(prec)), nil)
	x := new(big.Rat).Mul(r, new(big.Rat).SetInt(scale))
	lo := new(big.Int).Div(x.Num(), x.Denom())
	if new(big.Rat).SetInt(lo).Cmp(x) == 0 {
		return lo, true
	}
	if mode == RoundUnnecessary {
		return nil, false
	}
	hi := new(big.Int).Add(lo, big.NewInt(1))
	switch mode {
	case RoundDown:
		if x.Sign() < 0 {
			return hi, true
		}
		return lo, true
	case RoundUp:
		if x.Sign() < 0 {
			return lo, true
		}
		return hi, true
	case RoundCeiling:
		return hi, true
	}
	distance := new(big.Rat).Sub(x, new(big.Rat).SetInt(lo))
	cmp := distance.Cmp(big.NewRat(1, 2))
	if cmp < 0 {
		return lo, true
	}
	if cmp > 0 {
		return hi, true
	}
	if mode == RoundHalfEven {
		if lo.Bit(0) == 0 {
			return lo, true
		}
		return hi, true
	}
	if (mode == RoundHalfUp) == (x.Sign() > 0) {
		return hi, true
	}
	return lo, true
}

func TestDirectArithmeticRationalOracle(t *testing.T) {
	rng := rand.New(rand.NewPCG(123, 456))
	for range 500 {
		a := NewWithPrec(rng.Int64N(20001)-10000, rng.IntN(8))
		b := NewWithPrec(rng.Int64N(20000)+1, rng.IntN(8))
		if rng.IntN(2) == 0 {
			b = b.Neg()
		}
		ar, br := a.BigRat(), b.BigRat()
		if got, want := a.Cmp(b), ar.Cmp(br); got != want {
			t.Fatalf("comparison %s/%s: %d != %d", a, b, got, want)
		}
		for _, mode := range []RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundUnnecessary} {
			for _, multiply := range []bool{false, true} {
				exact := new(big.Rat)
				if multiply {
					exact.Mul(ar, br)
				} else {
					exact.Quo(ar, br)
				}
				prec := max(a.prec, b.prec)
				want, allowed := roundedRationalReference(exact, prec, mode)
				func() {
					defer func() {
						if p := recover(); p != nil {
							e, ok := p.(error)
							if allowed || !ok || !errors.Is(e, ErrRoundUnnecessary) {
								t.Fatalf("unexpected panic %v", p)
							}
						} else if !allowed {
							t.Fatal("required rounding did not panic")
						}
					}()
					var got Decimal
					if multiply {
						got = a.Mul(b, mode)
					} else {
						got = a.Quo(b, mode)
					}
					if allowed && (got.prec != prec || got.i.Cmp(want) != 0) {
						t.Fatalf("a=%s b=%s multiply=%v mode=%d: got %s coefficient %s, want %s", a, b, multiply, mode, got, got.i, want)
					}
				}()
			}
		}
	}
}

func TestComparisonExtremeScales(t *testing.T) {
	for _, scale := range []int{100000, math.MaxInt} {
		for _, pair := range []struct {
			a, b Decimal
			want int
		}{{NewWithPrec(1, scale), Zero, 1}, {NewWithPrec(1, scale), New(-1), 1}, {NewWithPrec(1, scale), One, -1}, {NewWithPrec(-1, scale), New(-1), 1}, {NewWithPrec(0, scale), Decimal{}, 0}} {
			if got := pair.a.Cmp(pair.b); got != pair.want {
				t.Fatalf("scale %d: comparison %d, want %d", scale, got, pair.want)
			}
			if got := pair.b.Cmp(pair.a); got != -pair.want {
				t.Fatal("asymmetric comparison")
			}
		}
	}
}

func TestTerminatingRationalFactorization(t *testing.T) {
	for _, twos := range []int{0, 1, 7, 1000, 10000} {
		for _, fives := range []int{0, 1, 9, 1000} {
			den := new(big.Int).Lsh(big.NewInt(1), uint(twos))
			den.Mul(den, new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(fives)), nil))
			for _, numerator := range []int64{7, -7} {
				r := new(big.Rat).SetFrac(big.NewInt(numerator), den)
				before := new(big.Rat).Set(r)
				d, err := NewFromBigRat(r)
				if err != nil || d.prec != max(twos, fives) || d.BigRat().Cmp(r) != 0 || r.Cmp(before) != 0 {
					t.Fatalf("factors %d/%d: conversion mismatch: %v", twos, fives, err)
				}
			}
		}
	}
	if _, err := NewFromBigRat(big.NewRat(1, 21)); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("nonterminating accepted: %v", err)
	}
}

func TestDirectArithmeticResultOwnership(t *testing.T) {
	a, b := MustFromString("-1.250"), MustFromString("2.5")
	for _, result := range []Decimal{a.Mul(b, RoundHalfEven), a.Quo(b, RoundHalfEven)} {
		result.i.SetInt64(99)
	}
	if a.StringWithTrailingZeros() != "-1.250" || b.StringWithTrailingZeros() != "2.5" {
		t.Fatal("mutated arithmetic input")
	}
}

func TestDirectArithmeticExtremeScale(t *testing.T) {
	if got := NewWithPrec(3, math.MaxInt).Mul(New(2), RoundUnnecessary); got.prec != math.MaxInt || got.i.Int64() != 6 {
		t.Fatal("multiplication unnecessarily added scales")
	}
	if got := Zero.QuoWithPrec(NewWithPrec(1, math.MaxInt), math.MaxInt, RoundUnnecessary); got.prec != math.MaxInt || !got.IsZero() {
		t.Fatal("zero quotient unnecessarily expanded its scale")
	}
	defer func() {
		p := recover()
		err, ok := p.(error)
		if !ok || !errors.Is(err, ErrOverflow) {
			t.Fatalf("unexpected division overflow: %v", p)
		}
	}()
	One.QuoWithPrec(NewWithPrec(1, math.MaxInt), math.MaxInt, RoundDown)
}

func BenchmarkDirectArithmetic(b *testing.B) {
	x, y := NewWithPrec(5, 10000), New(2)
	b.Run("Mul", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchDecimalResult = x.Mul(y, RoundHalfEven)
		}
	})
	b.Run("Quo", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchDecimalResult = x.Quo(y, RoundHalfEven)
		}
	})
	b.Run("Cmp", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if x.Cmp(y) >= 0 {
				b.Fatal("incorrect comparison")
			}
		}
	})
	r := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 100000))
	b.Run("BigRat", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchDecimalResult, err = NewFromBigRat(r)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
