package decimal

import (
	"fmt"
	"math/big"
	"testing"
)

var (
	benchIntegerResult *big.Int
	benchBoundsResult  fixedInterval
)

func BenchmarkFixedPointBounds(b *testing.B) {
	for _, bits := range []uint{64, 512} {
		num := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(17))
		den := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), bits/2), big.NewInt(3))
		scale := new(big.Int).Lsh(big.NewInt(1), bits)
		b.Run(fmt.Sprintf("ceil/%d", bits), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				benchIntegerResult = ceilQuotient(num, den)
			}
		})
		b.Run(fmt.Sprintf("fraction/%d", bits), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				benchBoundsResult = fractionBounds(num, den, scale)
			}
		})
	}
}

func BenchmarkIntervalHelpers(b *testing.B) {
	scale := new(big.Int).Lsh(big.NewInt(1), 512)
	a := fractionBounds(big.NewInt(17), big.NewInt(3), scale)
	den := fractionBounds(big.NewInt(7), big.NewInt(3), scale)
	negative := fixedInterval{new(big.Int).Neg(a.hi), new(big.Int).Neg(a.lo)}
	straddling := fixedInterval{negative.lo, a.hi}
	for _, tc := range []struct {
		name string
		call func(*testing.B)
	}{
		{"multiply", func(*testing.B) { benchBoundsResult = multiplyPositiveBounds(a, den, scale) }},
		{"divide-positive", func(*testing.B) { benchBoundsResult = divideBounds(a, den, scale) }},
		{"divide-negative", func(*testing.B) { benchBoundsResult = divideBounds(negative, den, scale) }},
		{"divide-straddling", func(*testing.B) { benchBoundsResult = divideBounds(straddling, den, scale) }},
		{"multiply-int", func(*testing.B) { benchBoundsResult = multiplyBoundsByInt(a, -7) }},
		{"log-mantissa", func(b *testing.B) {
			var err error
			benchBoundsResult, err = logMantissaBounds(big.NewInt(3), big.NewInt(2), scale, 512)
			if err != nil {
				b.Fatal(err)
			}
		}},
		{"round", func(*testing.B) { benchDecimalResult, _ = roundBounds(a, scale, 30) }},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				tc.call(b)
			}
		})
	}
}

func BenchmarkTranscendentalAllocations(b *testing.B) {
	for _, prec := range []int{100, 1000} {
		input := New(3)
		for _, tc := range []struct {
			name string
			call func(int) (Decimal, error)
		}{
			{"ln", input.LnWithPrec}, {"log2", input.Log2WithPrec},
			{"log10", input.Log10WithPrec}, {"exp", input.ExpWithPrec},
		} {
			b.Run(fmt.Sprintf("%s/%d", tc.name, prec), func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					result, err := tc.call(prec)
					if err != nil {
						b.Fatal(err)
					}
					benchDecimalResult = result
				}
			})
		}
	}
	// This input forces precision refinement to decide which side of the
	// half-unit boundary contains log2(x).
	input := MustFromString("1.4142135623730950488016887242096980785696718753769480731766797379907324784621070388503875343276415727")
	b.Run("log2/refinement", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			result, err := input.Log2WithPrec(0)
			if err != nil {
				b.Fatal(err)
			}
			benchDecimalResult = result
		}
	})
}
