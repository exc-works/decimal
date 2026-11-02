package decimal

import (
	"errors"
	"math"
	"math/big"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestMulWithPrecRationalOracle(t *testing.T) {
	rng := rand.New(rand.NewPCG(20261102, 7103))
	for range 500 {
		ac, bc := rng.Int64N(20001)-10000, rng.Int64N(20001)-10000
		as, bs, target := rng.IntN(9), rng.IntN(9), rng.IntN(15)
		a, b := NewWithPrec(ac, as), NewWithPrec(bc, bs)
		// Build the rational directly from the source coefficients and scales,
		// without Decimal conversion, multiplication, or rescaling helpers.
		exact := new(big.Rat).SetFrac(
			new(big.Int).Mul(big.NewInt(ac), big.NewInt(bc)),
			new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(as+bs)), nil),
		)
		for _, mode := range []RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundUnnecessary} {
			want, allowed := roundedRationalReference(exact, target, mode)
			func() {
				defer func() {
					if p := recover(); p != nil {
						err, ok := p.(error)
						if allowed || !ok || !errors.Is(err, ErrRoundUnnecessary) {
							t.Fatalf("unexpected panic %v", p)
						}
					} else if !allowed {
						t.Fatal("inexact RoundUnnecessary did not panic")
					}
				}()
				got := a.MulWithPrec(b, target, mode)
				if got.prec != target || got.i.Cmp(want) != 0 {
					t.Fatalf("%d*10^-%d * %d*10^-%d at %d mode %d: coefficient %s, want %s", ac, as, bc, bs, target, mode, got.i, want)
				}
			}()
		}
		if a.i.Int64() != ac || b.i.Int64() != bc || a.prec != as || b.prec != bs {
			t.Fatal("MulWithPrec mutated an operand")
		}
	}
	for _, tc := range []struct {
		a, b, want string
		prec       int
	}{
		{"1.23", "2.34", "2.878", 3},
		{"1.23", "2.34", "2.87820", 5},
		{"1.2451", "1", "1.25", 2},
		{"-1.2451", "1", "-1.25", 2},
		{"1.245", "1", "1.24", 2},
		{"1.255", "1", "1.26", 2},
		{"0", "1.23", "0.0000", 4},
	} {
		got := MustFromString(tc.a).MulWithPrec(MustFromString(tc.b), tc.prec, RoundHalfEven)
		if got.StringWithTrailingZeros() != tc.want || got.prec != tc.prec {
			t.Fatalf("%s * %s at %d: %s, want %s", tc.a, tc.b, tc.prec, got.StringWithTrailingZeros(), tc.want)
		}
	}
}

func TestMulWithPrecExtremeScales(t *testing.T) {
	for _, sign := range []int64{-1, 1} {
		for _, target := range []int{0, math.MaxInt} {
			for _, mode := range []RoundingMode{RoundDown, RoundUp, RoundCeiling, RoundHalfUp, RoundHalfDown, RoundHalfEven} {
				want := int64(0)
				if mode == RoundUp || mode == RoundCeiling && sign > 0 {
					want = sign
				}
				got := NewWithPrec(sign, math.MaxInt).MulWithPrec(NewWithPrec(1, math.MaxInt), target, mode)
				if got.prec != target || got.i.Cmp(big.NewInt(want)) != 0 {
					t.Fatalf("sign %d target %d mode %d: coefficient %s, want %d", sign, target, mode, got.i, want)
				}
			}
		}
	}
	got := NewWithPrec(0, math.MaxInt).MulWithPrec(NewWithPrec(1, math.MaxInt), 4, RoundUnnecessary)
	if !got.IsZero() || got.prec != 4 {
		t.Fatal("exact zero did not preserve requested scale")
	}
}

func TestContractDefaultMathPrecision(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		call              func(Decimal) (Decimal, error)
	}{
		{"sqrt", "2", "1.41421356237309504880", Decimal.Sqrt},
		{"root", "8", "2.00000000000000000000", func(d Decimal) (Decimal, error) { return d.ApproxRoot(3) }},
		{"ln", "2", "0.69314718055994530942", Decimal.Ln},
		{"log2", "3", "1.58496250072115618145", func(d Decimal) (Decimal, error) { return d.Log2(), nil }},
		{"log10", "2", "0.30102999566398119521", Decimal.Log10},
		{"exp", "1", "2.71828182845904523536", Decimal.Exp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.call(MustFromString(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if got.Precision() != 20 || got.StringWithTrailingZeros() != tc.want {
				t.Fatalf("got %s (scale %d), want %s (scale 20)", got.StringWithTrailingZeros(), got.Precision(), tc.want)
			}
			input := MustFromString(tc.input + "." + strings.Repeat("0", 25))
			got, err = tc.call(input)
			if err != nil || got.Precision() != 25 {
				t.Fatalf("high input scale: got scale %d, error %v", got.Precision(), err)
			}
		})
	}
}

func TestContractEqualExtremaPreserveFirstOperand(t *testing.T) {
	for _, pair := range [][2]Decimal{
		{MustFromString("1.0"), MustFromString("1.00")},
		{Decimal{}, NewWithPrec(0, 4)},
		{NewWithPrec(0, 4), Decimal{}},
	} {
		a, b := pair[0], pair[1]
		for _, got := range []Decimal{a.Max(b), a.Min(b), Max(a, b), Min(a, b)} {
			if got.i != a.i || got.prec != a.prec {
				t.Fatalf("equal extrema changed first operand: got scale=%d nil=%v, want scale=%d nil=%v", got.prec, got.IsNil(), a.prec, a.IsNil())
			}
		}
	}
}

func TestContractPowerHalfEvenRationalOracle(t *testing.T) {
	for _, coeff := range []int64{-30, -25, -20, -15, -3, -2, 0, 2, 3, 15, 20, 25, 30} {
		for _, scale := range []int{0, 1, 2, 4} {
			den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
			for _, power := range []int64{-3, -2, -1, 0, 1, 2, 3} {
				if coeff == 0 && power < 0 {
					continue
				}
				magnitude := power
				if magnitude < 0 {
					magnitude = -magnitude
				}
				n := big.NewInt(magnitude)
				exact := new(big.Rat).SetFrac(
					new(big.Int).Exp(big.NewInt(coeff), n, nil),
					new(big.Int).Exp(den, n, nil),
				)
				if power < 0 {
					exact.Inv(exact)
				}
				want, _ := roundedRationalReference(exact, scale, RoundHalfEven)
				input := NewWithPrec(coeff, scale)
				got := input.Power(power)
				if got.prec != scale || got.i.Cmp(want) != 0 {
					t.Fatalf("coefficient=%d scale=%d power=%d: got coefficient %s scale %d, want coefficient %s scale %d", coeff, scale, power, got.i, got.prec, want, scale)
				}
				if input.i.Cmp(big.NewInt(coeff)) != 0 || input.prec != scale {
					t.Fatal("Power mutated its input")
				}
			}
		}
	}
}
