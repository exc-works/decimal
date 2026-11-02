package decimal

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

func TestIntegerInspectionAgainstRationals(t *testing.T) {
	rng := rand.New(rand.NewPCG(112, 2026))
	for range 500 {
		coefficient, scale := rng.Int64(), rng.IntN(16)
		d := NewWithPrec(coefficient, scale)
		before := captureDecimalState(d)
		den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		r := new(big.Rat).SetFrac(big.NewInt(coefficient), den)
		want := new(big.Int).Quo(r.Num(), r.Denom())
		part := d.IntPart()
		if part.Cmp(want) != 0 || d.IsInteger() != r.IsInt() {
			t.Fatalf("inspection differs for coefficient %d scale %d", coefficient, scale)
		}
		gotI, exactI := d.Int64()
		wantI := r.IsInt() && r.Num().IsInt64()
		if exactI != wantI || exactI && gotI != r.Num().Int64() || !exactI && gotI != 0 {
			t.Fatal("wrong Int64 result or exactness")
		}
		gotU, exactU := d.Uint64()
		wantU := r.IsInt() && r.Num().IsUint64()
		if exactU != wantU || exactU && gotU != r.Num().Uint64() || !exactU && gotU != 0 {
			t.Fatal("wrong Uint64 result or exactness")
		}
		part.SetInt64(123)
		if captureDecimalState(d) != before {
			t.Fatal("integer result aliases input or inspection mutated it")
		}
	}
}

func TestIntegerInspectionExtremeScale(t *testing.T) {
	for _, coefficient := range []int64{-1, 0, 1} {
		d := NewWithPrec(coefficient, math.MaxInt)
		before := captureDecimalState(d)
		if d.IsInteger() != (coefficient == 0) || d.IntPart().Sign() != 0 {
			t.Fatal("wrong tiny-value inspection")
		}
		if got, ok := d.Int64(); got != 0 || ok != (coefficient == 0) {
			t.Fatal("wrong tiny Int64")
		}
		if got, ok := d.Uint64(); got != 0 || ok != (coefficient == 0) {
			t.Fatal("wrong tiny Uint64")
		}
		if captureDecimalState(d) != before {
			t.Fatal("inspection changed scale or coefficient")
		}
	}
	for _, d := range []Decimal{Decimal{}, Zero} {
		before := captureDecimalState(d)
		part := d.IntPart()
		part.SetInt64(123)
		if !d.IsInteger() || captureDecimalState(d) != before {
			t.Fatal("zero inspection changed initialization or shared zero")
		}
	}
}
