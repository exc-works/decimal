package decimal

import (
	"math/big"
	"testing"
)

// Derive extrema by evaluating every exact rational corner independently of
// the sign-based division shortcut. Also check that scratch reuse never leaks
// into caller-owned operands or the other result endpoint.
func TestIntervalArithmeticOwnership(t *testing.T) {
	check := func(name string, inputs []*big.Int, call func() fixedInterval, corners []*big.Rat) {
		t.Helper()
		before := make([]*big.Int, len(inputs))
		for i, input := range inputs {
			before[i] = new(big.Int).Set(input)
		}
		minimum, maximum := corners[0], corners[0]
		for _, corner := range corners[1:] {
			if corner.Cmp(minimum) < 0 {
				minimum = corner
			}
			if corner.Cmp(maximum) > 0 {
				maximum = corner
			}
		}
		floorAndCeil := func(r *big.Rat) (*big.Int, *big.Int) {
			q, remainder := new(big.Int).QuoRem(r.Num(), r.Denom(), new(big.Int))
			lo, hi := new(big.Int).Set(q), new(big.Int).Set(q)
			if remainder.Sign() < 0 {
				lo.Sub(lo, big.NewInt(1))
			}
			if remainder.Sign() > 0 {
				hi.Add(hi, big.NewInt(1))
			}
			return lo, hi
		}
		wantLo, _ := floorAndCeil(minimum)
		_, wantHi := floorAndCeil(maximum)
		result := call()
		if result.lo.Cmp(wantLo) != 0 || result.hi.Cmp(wantHi) != 0 {
			t.Fatalf("%s: got [%s,%s], want [%s,%s]", name, result.lo, result.hi, wantLo, wantHi)
		}
		if result.lo == result.hi {
			t.Fatalf("%s: endpoints share the same integer", name)
		}
		result.lo.SetInt64(991)
		if result.hi.Cmp(wantHi) != 0 {
			t.Fatalf("%s: endpoints share mutable storage", name)
		}
		result.hi.SetInt64(-997)
		if result.lo.Int64() != 991 {
			t.Fatalf("%s: upper mutation changed lower endpoint", name)
		}
		for i, input := range inputs {
			if input.Cmp(before[i]) != 0 {
				t.Fatalf("%s: changed operand %d", name, i)
			}
		}
	}
	for _, scale := range []*big.Int{big.NewInt(1), big.NewInt(11), new(big.Int).Lsh(big.NewInt(1), 256)} {
		for _, endpoints := range [][2]int64{{-9, -4}, {-9, 0}, {-9, 4}, {0, 0}, {0, 4}, {3, 9}} {
			a := fixedInterval{big.NewInt(endpoints[0]), big.NewInt(endpoints[1])}
			for _, denominators := range [][2]int64{{2, 2}, {2, 7}} {
				b := fixedInterval{big.NewInt(denominators[0]), big.NewInt(denominators[1])}
				var corners []*big.Rat
				for _, x := range []*big.Int{a.lo, a.hi} {
					for _, y := range []*big.Int{b.lo, b.hi} {
						corners = append(corners, new(big.Rat).SetFrac(new(big.Int).Mul(x, scale), y))
					}
				}
				check("divide", []*big.Int{a.lo, a.hi, b.lo, b.hi, scale}, func() fixedInterval { return divideBounds(a, b, scale) }, corners)
				if a.lo.Sign() >= 0 {
					corners = nil
					for _, x := range []*big.Int{a.lo, a.hi} {
						for _, y := range []*big.Int{b.lo, b.hi} {
							corners = append(corners, new(big.Rat).SetFrac(new(big.Int).Mul(x, y), scale))
						}
					}
					check("multiply", []*big.Int{a.lo, a.hi, b.lo, b.hi, scale}, func() fixedInterval { return multiplyPositiveBounds(a, b, scale) }, corners)
				}
			}
			for _, k := range []int64{-7, 0, 7} {
				corners := []*big.Rat{new(big.Rat).SetInt(new(big.Int).Mul(a.lo, big.NewInt(k))), new(big.Rat).SetInt(new(big.Int).Mul(a.hi, big.NewInt(k)))}
				check("multiply-int", []*big.Int{a.lo, a.hi}, func() fixedInterval { return multiplyBoundsByInt(a, k) }, corners)
			}
		}
	}
}

func TestTranscendentalInputOwnership(t *testing.T) {
	for _, text := range []string{"0.7", "3.14", "-3.14"} {
		input := MustFromString(text)
		coefficient, precision := input.BigInt(), input.Precision()
		for _, call := range []func(int) (Decimal, error){input.LnWithPrec, input.Log2WithPrec, input.Log10WithPrec, input.ExpWithPrec} {
			result, err := call(30)
			if err == nil {
				result.i.SetInt64(991)
			}
			if input.i.Cmp(coefficient) != 0 || input.prec != precision {
				t.Fatalf("calculation or result mutation changed input %s", text)
			}
		}
	}
}
