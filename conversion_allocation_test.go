package decimal

import (
	"errors"
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

func TestRationalConstructionRoundingAndOwnership(t *testing.T) {
	rng := rand.New(rand.NewPCG(321, 2026))
	for range 300 {
		r := big.NewRat(rng.Int64N(20001)-10000, rng.Int64N(100)+1)
		before := r.RatString()
		prec := rng.IntN(8)
		for _, mode := range []RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundUnnecessary} {
			want, allowed := roundedRationalReference(r, prec, mode)
			func() {
				defer func() {
					if p := recover(); p != nil {
						err, ok := p.(error)
						if allowed || !ok || !errors.Is(err, ErrRoundUnnecessary) {
							t.Fatalf("unexpected panic: %v", p)
						}
					} else if !allowed {
						t.Fatal("missing rounding panic")
					}
				}()
				got, err := NewFromBigRatWithPrec(r, prec, mode)
				if err != nil || got.prec != prec || got.i.Cmp(want) != 0 {
					t.Fatalf("r=%s prec=%d mode=%v got=%v err=%v want=%v", r, prec, mode, got, err, want)
				}
				got.i.SetInt64(42)
			}()
			if r.RatString() != before {
				t.Fatal("conversion or result mutation changed rational input")
			}
		}
	}
	for _, r := range []*big.Rat{new(big.Rat), big.NewRat(-123, 1), big.NewRat(123, 8)} {
		before := r.RatString()
		got, err := NewFromBigRat(r)
		if err != nil || got.BigRat().Cmp(r) != 0 {
			t.Fatalf("exact conversion: %v %v", got, err)
		}
		if r.Sign() != 0 { // Zero may share the immutable global zero.
			got.i.SetInt64(42)
		}
		if r.RatString() != before {
			t.Fatal("exact result aliases rational")
		}
	}
}

func TestIntegerFloatConversionAgainstRationals(t *testing.T) {
	rng := rand.New(rand.NewPCG(1123, 2026))
	for range 500 {
		i := new(big.Int).SetUint64(rng.Uint64())
		i.Lsh(i, uint(rng.IntN(1100)))
		i.Add(i, new(big.Int).SetUint64(rng.Uint64()))
		if rng.IntN(2) == 0 {
			i.Neg(i)
		}
		d := NewFromBigInt(i)
		before := captureDecimalState(d)
		r := new(big.Rat).SetInt(i)
		got32, exact32 := d.Float32()
		want32, wantExact32 := r.Float32()
		got64, exact64 := d.Float64()
		want64, wantExact64 := r.Float64()
		if math.Float32bits(got32) != math.Float32bits(want32) || exact32 != wantExact32 || math.Float64bits(got64) != math.Float64bits(want64) || exact64 != wantExact64 {
			t.Fatalf("float conversion differs for %s", i)
		}
		if captureDecimalState(d) != before {
			t.Fatal("float conversion mutated input")
		}
	}
}

func TestIntegerFloat32SingleRounding(t *testing.T) {
	// One above the midpoint between 2^80 and its Float32 successor.
	i := new(big.Int).Lsh(big.NewInt(1), 80)
	i.Add(i, new(big.Int).Lsh(big.NewInt(1), 56))
	i.Add(i, big.NewInt(1))
	for _, negative := range []bool{false, true} {
		wantBits := uint32(0x67800001)
		if negative {
			i.Neg(i)
			wantBits |= 1 << 31
		}
		d := NewFromBigInt(i)
		got, exact := d.Float32()
		if math.Float32bits(got) != wantBits || exact {
			t.Fatalf("double rounding: got %x exact=%v want %x", math.Float32bits(got), exact, wantBits)
		}
	}
}

func TestZeroConversionsExtremeScale(t *testing.T) {
	d := NewWithPrec(0, math.MaxInt)
	if got, exact := d.Float32(); got != 0 || !exact {
		t.Fatal("zero Float32")
	}
	if got, exact := d.Float64(); got != 0 || !exact {
		t.Fatal("zero Float64")
	}
	got, err := NewFromBigRatWithPrec(new(big.Rat), math.MaxInt, RoundUnnecessary)
	if err != nil || got.prec != math.MaxInt || !got.IsZero() {
		t.Fatal("zero rational conversion lost scale")
	}
}
