package decimal

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

func TestGeneralFormatRoundingBoundaries(t *testing.T) {
	for _, tc := range []struct{ input, format, want string }{
		{"999.9", "%.3g", "1e+03"},
		{"999.9", "%#.3g", "1.00e+03"},
		{"-999.9", "%.3G", "-1E+03"},
		{"0.00009999", "%.3g", "0.0001"},
		{"0.00009999", "%#.3g", "0.000100"},
		{"-0.00009999", "%#.3G", "-0.000100"},
		{"999.5", "%.3g", "1e+03"},
		{"998.5", "%.3g", "998"},
		{"9.5", "%#.1g", "1.e+01"},
		{"1.25", "%.2g", "1.2"},
		{"1.35", "%.2g", "1.4"},
	} {
		if got := fmt.Sprintf(tc.format, MustFromString(tc.input)); got != tc.want {
			t.Errorf("%s(%s) = %q, want %q", tc.format, tc.input, got, tc.want)
		}
	}
}

func TestGeneralFormatExactBinaryValues(t *testing.T) {
	// These rationals are exactly representable by float64. Standard fmt is
	// an independent notation/rounding oracle, including halfway values.
	for coefficient := int64(-8191); coefficient <= 8191; coefficient += 31 {
		r := new(big.Rat).SetFrac(big.NewInt(coefficient), big.NewInt(256))
		d, err := NewFromBigRat(r)
		if err != nil {
			t.Fatal(err)
		}
		f, exact := r.Float64()
		if !exact {
			t.Fatal("inexact oracle input")
		}
		for _, format := range []string{"%.1g", "%.3g", "%#.4G", "%+012.2g"} {
			if got, want := fmt.Sprintf(format, d), fmt.Sprintf(format, f); got != want {
				t.Fatalf("%s(%s): got %q, want %q", format, d, got, want)
			}
		}
	}
}

func TestScientificExtremeScale(t *testing.T) {
	for _, scale := range []int{131072, math.MaxInt} {
		for _, coefficient := range []int64{1, -100} {
			d := NewWithPrec(coefficient, scale)
			want := "1.000000e-" + fmt.Sprint(scale)
			if coefficient < 0 {
				want = "-1.000000e-" + fmt.Sprint(scale-2)
			}
			if got := fmt.Sprintf("%.6e", d); got != want {
				t.Errorf("coefficient %d scale %d: got %s, want %s", coefficient, scale, got, want)
			}
		}
	}
}

func TestStripTrailingZerosExtremeScale(t *testing.T) {
	for _, sign := range []int64{1, -1} {
		coefficient, _ := new(big.Int).SetString("123456789012345678900", 10)
		coefficient.Mul(coefficient, big.NewInt(sign))
		for _, scale := range []int{1, 2, 131072, math.MaxInt} {
			d := NewFromBigIntWithPrec(coefficient, scale)
			got := d.StripTrailingZeros()
			removed := min(2, scale)
			want := new(big.Int).Quo(coefficient, safeGetPrecisionMultiplier(removed))
			if got.prec != scale-removed || got.i.Cmp(want) != 0 || d.i.Cmp(coefficient) != 0 {
				t.Fatalf("incorrect normalization at scale %d", scale)
			}
		}
	}
}

func BenchmarkScientificLargeScale(b *testing.B) {
	d := NewWithPrec(1, 131072)
	b.ReportAllocs()
	for b.Loop() {
		benchStringResult = fmt.Sprintf("%.6e", d)
	}
}

func BenchmarkStripLargeScale(b *testing.B) {
	coefficient, _ := new(big.Int).SetString("123456789012345678900", 10)
	d := NewFromBigIntWithPrec(coefficient, 131072)
	b.ReportAllocs()
	for b.Loop() {
		benchDecimalResult = d.StripTrailingZeros()
	}
}
