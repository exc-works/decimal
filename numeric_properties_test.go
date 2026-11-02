package decimal

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

func checkRationalRounding(t *testing.T, exact *big.Rat, prec int, mode RoundingMode, compute func() Decimal) {
	t.Helper()
	want, allowed := roundedRationalReference(exact, prec, mode)
	var got Decimal
	var failure any
	func() {
		defer func() { failure = recover() }()
		got = compute()
	}()
	if !allowed {
		err, ok := failure.(error)
		if !ok || !errors.Is(err, ErrRoundUnnecessary) {
			t.Fatalf("required rounding was not reported: %v", failure)
		}
		return
	}
	if failure != nil || got.i == nil || got.prec != prec || got.i.Cmp(want) != 0 {
		t.Fatalf("wrong rounding at scale %d in mode %d (panic=%v)", prec, mode, failure)
	}
}

func FuzzArithmeticCorrectRounding(f *testing.F) {
	for _, pair := range [][2]int64{{123, 234}, {12451, 1}, {-12451, 1}, {0, 3}, {math.MinInt64, math.MaxInt64}} {
		for mode := uint8(0); mode < 7; mode++ {
			f.Add(pair[0], pair[1], uint8(2), uint8(2), uint8(3), mode)
		}
	}
	f.Fuzz(func(t *testing.T, ac, bc int64, sa, sb, target, index uint8) {
		as, bs, prec := int(sa%13), int(sb%13), int(target%13)
		modes := [...]RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundUnnecessary}
		mode := modes[int(index)%len(modes)]
		power10 := func(p int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(p)), nil) }
		ar := new(big.Rat).SetFrac(big.NewInt(ac), power10(as))
		br := new(big.Rat).SetFrac(big.NewInt(bc), power10(bs))
		a, b := NewWithPrec(ac, as), NewWithPrec(bc, bs)
		beforeA, beforeB := captureDecimalState(a), captureDecimalState(b)
		if a.Cmp(b) != ar.Cmp(br) {
			t.Fatal("comparison disagrees with independent rationals")
		}
		checkRationalRounding(t, new(big.Rat).Mul(ar, br), prec, mode, func() Decimal {
			return a.MulWithPrec(b, prec, mode)
		})
		if bc != 0 {
			checkRationalRounding(t, new(big.Rat).Quo(ar, br), prec, mode, func() Decimal {
				return a.QuoWithPrec(b, prec, mode)
			})
		}
		if captureDecimalState(a) != beforeA || captureDecimalState(b) != beforeB {
			t.Fatal("arithmetic mutated an input")
		}
	})
}

func FuzzRootCorrectRounding(f *testing.F) {
	f.Add(int64(25), uint8(2), uint8(0), uint8(0))
	f.Add(int64(15625), uint8(4), uint8(0), uint8(1))
	f.Add(int64(-1953125), uint8(6), uint8(1), uint8(1))
	f.Add(int64(math.MaxInt64), uint8(0), uint8(7), uint8(8))
	f.Fuzz(func(t *testing.T, coefficient int64, rawScale, rawDegree, rawPrec uint8) {
		scale, degree, prec := int(rawScale%7), int64(2+rawDegree%8), int(rawPrec%9)
		input := NewWithPrec(coefficient, scale)
		before := captureDecimalState(input)
		got, err := input.ApproxRootWithPrec(degree, prec)
		if captureDecimalState(input) != before {
			t.Fatal("root mutated its input")
		}
		if coefficient < 0 && degree%2 == 0 {
			if !errors.Is(err, ErrNegativeRoot) {
				t.Fatalf("wrong root domain error: %v", err)
			}
			return
		}
		if err != nil {
			t.Fatalf("bounded valid root failed: %v", err)
		}
		den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		certifyRootRounding(t, new(big.Rat).SetFrac(big.NewInt(coefficient), den), degree, prec, got)
	})
}

func TestIndependentFloatBoundaryRounding(t *testing.T) {
	for _, tc := range []struct {
		input string
		bits  uint64
	}{
		{"9007199254740993", 0x4340000000000000},
		{"9007199254740995", 0x4340000000000002},
		{"-9007199254740993", 0xc340000000000000},
		{"1e400", 0x7ff0000000000000},
		{"-1e400", 0xfff0000000000000},
	} {
		got, exact := MustFromString(tc.input).Float64()
		if math.Float64bits(got) != tc.bits || exact {
			t.Fatalf("float64 boundary %s: bits=%x exact=%v", tc.input, math.Float64bits(got), exact)
		}
	}
	for _, tc := range []struct {
		input string
		bits  uint32
	}{
		{"16777217", 0x4b800000}, {"16777219", 0x4b800002}, {"-16777217", 0xcb800000},
	} {
		got, exact := MustFromString(tc.input).Float32()
		if math.Float32bits(got) != tc.bits || exact {
			t.Fatalf("float32 boundary %s: bits=%x exact=%v", tc.input, math.Float32bits(got), exact)
		}
	}
	for _, tc := range []struct {
		numerator int64
		exponent  uint
		bits      uint64
		exact     bool
	}{
		{1, 1074, 1, true}, {1, 1075, 0, false}, {3, 1075, 2, false},
		{-1, 1075, 0x8000000000000000, false},
	} {
		rat := new(big.Rat).SetFrac(big.NewInt(tc.numerator), new(big.Int).Lsh(big.NewInt(1), tc.exponent))
		d, err := NewFromBigRat(rat)
		if err != nil {
			t.Fatal(err)
		}
		got, exact := d.Float64()
		if math.Float64bits(got) != tc.bits || exact != tc.exact {
			t.Fatalf("subnormal boundary: bits=%x exact=%v", math.Float64bits(got), exact)
		}
	}
	for _, tc := range []struct {
		numerator int64
		exponent  uint
		bits      uint32
		exact     bool
	}{{1, 149, 1, true}, {1, 150, 0, false}, {3, 150, 2, false}, {-1, 150, 0x80000000, false}} {
		rat := new(big.Rat).SetFrac(big.NewInt(tc.numerator), new(big.Int).Lsh(big.NewInt(1), tc.exponent))
		d, err := NewFromBigRat(rat)
		if err != nil {
			t.Fatal(err)
		}
		got, exact := d.Float32()
		if math.Float32bits(got) != tc.bits || exact != tc.exact {
			t.Fatalf("float32 subnormal: bits=%x exact=%v", math.Float32bits(got), exact)
		}
	}
	// Values above the largest finite float need not overflow: a quarter ulp
	// still rounds back to MaxFloat64. A half ulp reaches the infinity boundary.
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 1024), new(big.Int).Lsh(big.NewInt(1), 971))
	near := new(big.Int).Add(max, new(big.Int).Lsh(big.NewInt(1), 969))
	got, exact := NewFromBigInt(near).Float64()
	if math.Float64bits(got) != 0x7fefffffffffffff || exact {
		t.Fatal("premature floating overflow")
	}
	midpoint := new(big.Int).Add(max, new(big.Int).Lsh(big.NewInt(1), 970))
	got, exact = NewFromBigInt(midpoint).Float64()
	if math.Float64bits(got) != 0x7ff0000000000000 || exact {
		t.Fatal("wrong infinity-boundary rounding")
	}
}
