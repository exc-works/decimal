package decimal

import (
	"fmt"
	"math/big"
	"math/rand/v2"
	"testing"
)

func TestDirectedBinaryScaling(t *testing.T) {
	rng := rand.New(rand.NewPCG(987, 321))
	for _, bits := range []uint{0, 1, 63, 64, 512, 3322} {
		for _, binary := range []bool{false, true} {
			scale := new(big.Int).Lsh(big.NewInt(1), bits)
			if !binary {
				scale.Add(scale, big.NewInt(10))
			}
			for range 40 {
				a := new(big.Int).SetUint64(rng.Uint64())
				a.Lsh(a, bits/2)
				if rng.IntN(2) == 0 {
					a.Neg(a)
				}
				q, rem := new(big.Int).QuoRem(a, scale, new(big.Int))
				if rem.Sign() > 0 {
					q.Add(q, big.NewInt(1))
				}
				before := new(big.Int).Set(a)
				if got := ceilQuotient(a, scale); got.Cmp(q) != 0 || a.Cmp(before) != 0 {
					t.Fatalf("ceil at bits %d", bits)
				}
				owned := new(big.Int).Set(a)
				if got := ceilQuotientTo(owned, owned, scale); got.Cmp(q) != 0 {
					t.Fatal("aliased ceiling")
				}
				lo1, lo2 := new(big.Int).Abs(a), new(big.Int).SetUint64(rng.Uint64())
				x := fixedInterval{lo1, new(big.Int).Add(lo1, big.NewInt(17))}
				y := fixedInterval{lo2, new(big.Int).Add(lo2, big.NewInt(31))}
				wantLo := new(big.Int).Mul(x.lo, y.lo)
				wantLo.Div(wantLo, scale)
				wantHi, tail := new(big.Int).QuoRem(new(big.Int).Mul(x.hi, y.hi), scale, new(big.Int))
				if tail.Sign() != 0 {
					wantHi.Add(wantHi, big.NewInt(1))
				}
				got := multiplyPositiveBounds(x, y, scale)
				if got.lo.Cmp(wantLo) != 0 || got.hi.Cmp(wantHi) != 0 {
					t.Fatalf("multiply at bits %d", bits)
				}
			}
		}
	}
}

func TestLogarithmsNearOne(t *testing.T) {
	for _, prec := range []int{0, 30, 1000} {
		step := NewWithPrec(1, prec+100)
		for _, input := range []Decimal{One.Add(step), One.Sub(step)} {
			for _, call := range []func(int) (Decimal, error){input.LnWithPrec, input.Log2WithPrec, input.Log10WithPrec} {
				got, err := call(prec)
				if err != nil || !got.IsZero() || got.prec != prec {
					t.Fatalf("near-one logarithm at precision %d: %s %v", prec, got, err)
				}
			}
		}
	}
}

func TestLogReductionIdentityAndSeriesCost(t *testing.T) {
	rng := rand.New(rand.NewPCG(511, 73))
	for range 300 {
		var numeratorBytes, denominatorBytes [64]byte
		for i := range numeratorBytes {
			numeratorBytes[i] = byte(rng.Uint64())
			denominatorBytes[i] = byte(rng.Uint64())
		}
		num := new(big.Int).SetBytes(numeratorBytes[:])
		den := new(big.Int).SetBytes(denominatorBytes[:])
		// Vary binary exponents independently of the significand digits.
		num.Lsh(num, uint(rng.IntN(100)))
		den.Lsh(den, uint(rng.IntN(100)))
		r := new(big.Rat).SetFrac(num, den)
		beforeNum, beforeDen := new(big.Int).Set(num), new(big.Int).Set(den)
		for _, base := range []int{0, 2, 10} {
			argument, reciprocal := reduceLogArgument(num, den, base)
			mantissa := new(big.Rat).SetFrac(argument.num, argument.den)
			if mantissa.Cmp(big.NewRat(1, 1)) < 0 || mantissa.Cmp(big.NewRat(2, 1)) >= 0 {
				t.Fatal("mantissa outside [1,2)")
			}
			value := new(big.Rat).Set(mantissa)
			power := new(big.Int).Lsh(big.NewInt(1), uint(absLogExponent(argument.exponent)))
			if argument.exponent >= 0 {
				value.Mul(value, new(big.Rat).SetInt(power))
			} else {
				value.Quo(value, new(big.Rat).SetInt(power))
			}
			if reciprocal {
				value.Inv(value)
			}
			if value.Cmp(r) != 0 {
				t.Fatal("log reduction changed the exact argument")
			}
			// Construct both candidates independently. Prefer avoiding a new
			// ln(2) evaluation, then minimize the positive atanh series input.
			for _, inverse := range []bool{false, true} {
				n, d := num, den
				if inverse {
					n, d = d, n
				}
				n, d, exponent := binaryMantissa(n, d)
				chosenConstant := base == 0 && argument.exponent != 0
				otherConstant := base == 0 && exponent != 0
				if chosenConstant && !otherConstant {
					t.Fatal("unnecessary ln(2) evaluation")
				}
				otherMantissa := new(big.Rat).SetFrac(n, d)
				// A 64-bit lower bound can leave directions indistinguishable
				// near sqrt(2). The chosen mantissa remains within 2^-60 of
				// the better ratio; exact output accuracy does not depend on it.
				allowance := new(big.Rat).Add(big.NewRat(1, 1), new(big.Rat).SetFrac(oneInt, new(big.Int).Lsh(big.NewInt(1), 60)))
				if chosenConstant == otherConstant && new(big.Rat).Quo(mantissa, otherMantissa).Cmp(allowance) > 0 {
					t.Fatal("slower mantissa selected beyond the bounded decision tolerance")
				}
			}
			argument.num.SetInt64(91)
			argument.den.SetInt64(93)
			if num.Cmp(beforeNum) != 0 || den.Cmp(beforeDen) != 0 {
				t.Fatal("reduction mutated caller inputs")
			}
		}
	}
}

func BenchmarkLogReductionLargeInput(b *testing.B) {
	input := MustFromString("0.25").Add(NewWithPrec(1, 100000))
	for _, prec := range []int{0, 18} {
		b.Run(fmt.Sprintf("scale100000/prec%d", prec), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var err error
				benchDecimalResult, err = input.LnWithPrec(prec)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func absLogExponent(exponent int64) int64 {
	if exponent < 0 {
		return -exponent
	}
	return exponent
}

func TestLogReductionUnitBinadesAndPowers(t *testing.T) {
	for _, tc := range []struct {
		input      string
		reciprocal bool
	}{
		{"0.25", false}, {"0.250000000000000000001", false},
		{"0.249999999999999999999", true}, {"0.125000000000000000001", false},
		{"0.6", true}, {"0.999999999999999999999", true},
		{"1.000000000000000000001", false}, {"1.9", false},
	} {
		r := MustFromString(tc.input).BigRat()
		_, reciprocal := reduceLogArgument(r.Num(), r.Denom(), 0)
		if reciprocal != tc.reciprocal {
			t.Fatalf("Ln reduction for %s: reciprocal=%v", tc.input, reciprocal)
		}
	}
}

func TestNegativeExponentialCertifiedZero(t *testing.T) {
	for _, tc := range []struct {
		x    int64
		prec int
	}{{-300000, 100000}, {-327683, 131072}, {-5, 1}, {-3, 0}} {
		got, err := New(tc.x).ExpWithPrec(tc.prec)
		if err != nil || !got.IsZero() || got.prec != tc.prec {
			t.Fatalf("Exp(%d), scale %d: %s %v", tc.x, tc.prec, got, err)
		}
	}
}

func BenchmarkOptimizedLogarithms(b *testing.B) {
	for _, prec := range []int{1000, 10000} {
		for _, tc := range []struct {
			name  string
			input Decimal
		}{
			{"three", New(3)},
			{"below-one", One.Sub(NewWithPrec(1, prec+100))},
			{"above-one", One.Add(NewWithPrec(1, prec+100))},
			{"above-quarter", MustFromString("0.25").Add(NewWithPrec(1, prec+100))},
			{"below-quarter", MustFromString("0.25").Sub(NewWithPrec(1, prec+100))},
			{"above-eighth", MustFromString("0.125").Add(NewWithPrec(1, prec+100))},
			{"below-eighth", MustFromString("0.125").Sub(NewWithPrec(1, prec+100))},
		} {
			b.Run(fmt.Sprintf("%s/%d", tc.name, prec), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var err error
					benchDecimalResult, err = tc.input.LnWithPrec(prec)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func TestLogReductionPrefixThreshold(t *testing.T) {
	allowance := new(big.Rat).Add(big.NewRat(1, 1), new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 60)))
	for _, bits := range []uint{63, 64, 65, 100, 1000, 20000} {
		den := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(71))
		twiceSquared := new(big.Int).Mul(den, den)
		twiceSquared.Lsh(twiceSquared, 1)
		center := new(big.Int).Sqrt(twiceSquared)
		delta := big.NewInt(1)
		if bits > 64 {
			delta.Lsh(delta, bits-64)
		}
		for _, sign := range []int64{-1, 0, 1} {
			num := new(big.Int).Add(center, new(big.Int).Mul(delta, big.NewInt(sign)))
			beforeNum, beforeDen := new(big.Int).Set(num), new(big.Int).Set(den)
			argument, inverse := reduceLogArgument(num, den, 2)
			chosen := new(big.Rat).SetFrac(argument.num, argument.den)
			original := new(big.Rat).SetFrac(num, den)
			other := new(big.Rat).Quo(big.NewRat(2, 1), original)
			if inverse {
				if new(big.Int).Mul(num, num).Cmp(twiceSquared) <= 0 {
					t.Fatal("false positive inverse")
				}
				original, other = other, original
			}
			if chosen.Cmp(original) != 0 {
				t.Fatal("truncated enclosure coefficients")
			}
			ratio := new(big.Rat).Quo(chosen, other)
			if ratio.Cmp(allowance) > 0 {
				t.Fatalf("cost bound exceeds tolerance at %d bits", bits)
			}
			argument.num.SetInt64(1)
			argument.den.SetInt64(1)
			if num.Cmp(beforeNum) != 0 || den.Cmp(beforeDen) != 0 {
				t.Fatal("input alias")
			}
		}
	}
}
