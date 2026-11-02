package decimal

import (
	"errors"
	"math/big"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestRescaleIndependentRounding(t *testing.T) {
	rng := rand.New(rand.NewPCG(998, 2026))
	for range 500 {
		d := NewWithPrec(rng.Int64N(200001)-100000, rng.IntN(8))
		prec := rng.IntN(10)
		for _, mode := range []RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundUnnecessary} {
			want, allowed := roundedRationalReference(d.BigRat(), prec, mode)
			checkRoundedResult(t, d, prec, want, allowed, func() Decimal { return d.Rescale(prec, mode) })
		}
	}
}

func TestSignificantFiguresIndependentRounding(t *testing.T) {
	for _, text := range []string{"123456.7", "123500.0", "124500.0", "999500.0", "100000.0", "12.345", "0.001235", "0.001245"} {
		for _, sign := range []string{"", "-"} {
			d := MustFromString(sign + text)
			for _, figures := range []int{1, 3, 6, 12} {
				digits := len(strings.TrimPrefix(d.i.String(), "-"))
				target := figures - (digits - d.prec)
				for _, mode := range []RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundUnnecessary} {
					r := d.BigRat()
					prec := min(d.prec, max(0, target))
					unit := big.NewInt(1)
					if target < 0 {
						unit.Exp(big.NewInt(10), big.NewInt(int64(-target)), nil)
						r.Quo(r, new(big.Rat).SetInt(unit))
					}
					want, allowed := roundedRationalReference(r, prec, mode)
					if allowed {
						want.Mul(want, unit)
					}
					checkRoundedResult(t, d, prec, want, allowed, func() Decimal { return d.SignificantFigures(figures, mode) })
				}
			}
		}
	}
}

func checkRoundedResult(t *testing.T, input Decimal, prec int, want *big.Int, allowed bool, call func() Decimal) {
	t.Helper()
	before := captureDecimalState(input)
	defer func() {
		if p := recover(); p != nil {
			err, ok := p.(error)
			if allowed || !ok || !errors.Is(err, ErrRoundUnnecessary) {
				t.Fatalf("unexpected panic: %v", p)
			}
		} else if !allowed {
			t.Fatal("missing rounding panic")
		}
		if captureDecimalState(input) != before {
			t.Fatal("rounding mutated input")
		}
	}()
	got := call()
	if !allowed || got.prec != prec || got.i.Cmp(want) != 0 {
		t.Fatalf("input=%v got=%v scale=%d want coefficient=%v scale=%d", input, got, got.prec, want, prec)
	}
}
