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
		}{{"three", New(3)}, {"below-one", One.Sub(NewWithPrec(1, prec+100))}, {"above-one", One.Add(NewWithPrec(1, prec+100))}} {
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
