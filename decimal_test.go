package decimal

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

func mustDecimal(t *testing.T, input string) Decimal {
	t.Helper()

	d, err := NewFromString(input)
	if err != nil {
		t.Fatalf("NewFromString(%q) returned error: %v", input, err)
	}
	return d
}

func assertDecimalEqual(t *testing.T, got, want Decimal) {
	t.Helper()

	if !got.Equal(want) {
		t.Fatalf("got %s (prec=%d), want %s (prec=%d)", got.String(), got.Precision(), want.String(), want.Precision())
	}
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got nil")
		}
	}()
	fn()
}

func TestNewFromString(t *testing.T) {
	largeZeroFraction := "0." + strings.Repeat("0", 1000)

	tests := []struct {
		name          string
		input         string
		want          string
		wantPrecision int
		wantNeg       bool
		wantErr       bool
	}{
		{name: "empty", input: "", wantErr: true},
		{name: "trailing dot", input: "1.", wantErr: true},
		{name: "leading dot", input: ".1", wantErr: true},
		{name: "sign only", input: "-", wantErr: true},
		{name: "negative trailing dot", input: "-1.", wantErr: true},
		{name: "negative leading dot", input: "-.1", wantErr: true},
		{name: "large fractional zero precision", input: largeZeroFraction, want: "0", wantPrecision: 0, wantNeg: false},
		{name: "integer", input: "1", want: "1", wantPrecision: 0, wantNeg: false},
		{name: "zero", input: "0", want: "0", wantPrecision: 0, wantNeg: false},
		{name: "negative integer", input: "-1", want: "-1", wantPrecision: 0, wantNeg: true},
		{name: "fraction", input: "1.0001", want: "1.0001", wantPrecision: 4, wantNeg: false},
		{name: "negative fraction", input: "-1.0001", want: "-1.0001", wantPrecision: 4, wantNeg: true},
		{name: "scientific lower", input: "3.7154500000000011e-15", want: "0.0000000000000037154500000000011", wantPrecision: 31, wantNeg: false},
		{name: "scientific upper", input: "3.7154500000000011E-15", want: "0.0000000000000037154500000000011", wantPrecision: 31, wantNeg: false},
		{name: "scientific positive exponent", input: "3.7154e3", want: "3715.4", wantPrecision: 1, wantNeg: false},
		{name: "scientific negative exponent", input: "-3.7154e5", want: "-371540", wantPrecision: 0, wantNeg: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewFromString(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NewFromString(%q) expected error, got %s", tc.input, got.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFromString(%q) returned error: %v", tc.input, err)
			}
			if got.Precision() != tc.wantPrecision {
				t.Fatalf("NewFromString(%q) precision = %d, want %d", tc.input, got.Precision(), tc.wantPrecision)
			}
			if got.IsNegative() != tc.wantNeg {
				t.Fatalf("NewFromString(%q) negative = %v, want %v", tc.input, got.IsNegative(), tc.wantNeg)
			}
			if got.String() != tc.want {
				t.Fatalf("NewFromString(%q) = %s, want %s", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewFromFloat64(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{name: "zero", value: 0, want: "0"},
		{name: "one", value: 1, want: "1"},
		{name: "one point one", value: 1.1, want: "1.1"},
		{name: "one point zero one", value: 1.01, want: "1.01"},
		{name: "one point zero zero one", value: 1.001, want: "1.001"},
		{name: "one point one billionth", value: 1.000000001, want: "1.000000001"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewFromFloat64(tc.value)
			want := mustDecimal(t, tc.want)
			assertDecimalEqual(t, got, want)
		})
	}
}

func TestNewFromFloat32(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  string
	}{
		{name: "zero", value: 0, want: "0"},
		{name: "one", value: 1, want: "1"},
		{name: "half", value: 0.5, want: "0.5"},
		{name: "fraction", value: 12.34, want: "12.34"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewFromFloat32(tc.value)
			want := mustDecimal(t, tc.want)
			assertDecimalEqual(t, got, want)
		})
	}
}

func TestNewFromBigRat(t *testing.T) {
	tests := []struct {
		name      string
		value     *big.Rat
		want      string
		wantPrec  int
		wantError bool
	}{
		{name: "half", value: big.NewRat(1, 2), want: "0.5", wantPrec: 1},
		{name: "negative fraction", value: big.NewRat(-7, 4), want: "-1.75", wantPrec: 2},
		{name: "integer", value: big.NewRat(5, 1), want: "5", wantPrec: 0},
		{name: "zero", value: big.NewRat(0, 7), want: "0", wantPrec: 0},
		{name: "non terminating", value: big.NewRat(1, 3), wantError: true},
		{name: "nil", value: nil, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewFromBigRat(tc.value)
			if tc.wantError {
				if err == nil {
					t.Fatalf("NewFromBigRat(%v) expected error, got %s", tc.value, got.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFromBigRat(%v) returned error: %v", tc.value, err)
			}

			want := mustDecimal(t, tc.want)
			assertDecimalEqual(t, got, want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("NewFromBigRat(%v) precision = %d, want %d", tc.value, got.Precision(), tc.wantPrec)
			}
		})
	}
}

func TestNewFromBigRatWithPrec(t *testing.T) {
	tests := []struct {
		name      string
		value     *big.Rat
		prec      int
		mode      RoundingMode
		want      string
		wantPrec  int
		wantError bool
	}{
		{name: "non terminating round down", value: big.NewRat(1, 3), prec: 2, mode: RoundDown, want: "0.33", wantPrec: 2},
		{name: "non terminating round up", value: big.NewRat(1, 3), prec: 2, mode: RoundUp, want: "0.34", wantPrec: 2},
		{name: "negative non terminating round down", value: big.NewRat(-1, 3), prec: 2, mode: RoundDown, want: "-0.33", wantPrec: 2},
		{name: "negative non terminating round up", value: big.NewRat(-1, 3), prec: 2, mode: RoundUp, want: "-0.34", wantPrec: 2},
		{name: "half even tie", value: big.NewRat(1, 8), prec: 2, mode: RoundHalfEven, want: "0.12", wantPrec: 2},
		{name: "round up with tail beyond guard digit", value: big.NewRat(3300001, 10000000), prec: 2, mode: RoundUp, want: "0.34", wantPrec: 2},
		{name: "half even with tail beyond guard digit", value: big.NewRat(3250001, 10000000), prec: 2, mode: RoundHalfEven, want: "0.33", wantPrec: 2},
		{name: "keep zero scale", value: big.NewRat(0, 7), prec: 4, mode: RoundHalfEven, want: "0", wantPrec: 4},
		{name: "nil", value: nil, prec: 2, mode: RoundHalfEven, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewFromBigRatWithPrec(tc.value, tc.prec, tc.mode)
			if tc.wantError {
				if err == nil {
					t.Fatalf("NewFromBigRatWithPrec(%v, %d, %v) expected error, got %s", tc.value, tc.prec, tc.mode, got.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFromBigRatWithPrec(%v, %d, %v) returned error: %v", tc.value, tc.prec, tc.mode, err)
			}

			want := mustDecimal(t, tc.want)
			assertDecimalEqual(t, got, want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("NewFromBigRatWithPrec(%v, %d, %v) precision = %d, want %d",
					tc.value, tc.prec, tc.mode, got.Precision(), tc.wantPrec)
			}
		})
	}

	t.Run("negative precision panics", func(t *testing.T) {
		assertPanic(t, func() {
			_, _ = NewFromBigRatWithPrec(big.NewRat(1, 3), -1, RoundHalfEven)
		})
	})
}

func TestNewFromBigRatWithPrecMatchesRescale(t *testing.T) {
	inputs := []string{
		"0",
		"1.25",
		"-1.25",
		"2.345",
		"-2.345",
		"9.995",
		"-9.995",
		"123456789.5000",
	}

	targetPrecs := []int{0, 1, 2, 3, 4}
	modes := []RoundingMode{
		RoundDown,
		RoundUp,
		RoundCeiling,
		RoundHalfUp,
		RoundHalfDown,
		RoundHalfEven,
		RoundUnnecessary,
	}

	callRescale := func(v Decimal, prec int, mode RoundingMode) (got Decimal, panicked bool) {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		got = v.Rescale(prec, mode)
		return
	}

	callFromRat := func(v *big.Rat, prec int, mode RoundingMode) (got Decimal, panicked bool) {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		var err error
		got, err = NewFromBigRatWithPrec(v, prec, mode)
		if err != nil {
			t.Fatalf("NewFromBigRatWithPrec(%s, %d, %v) returned error: %v", v.RatString(), prec, mode, err)
		}
		return
	}

	for _, input := range inputs {
		exact := mustDecimal(t, input)
		rat := new(big.Rat).SetFrac(new(big.Int).Set(exact.i), safeGetPrecisionMultiplier(exact.prec))

		for _, prec := range targetPrecs {
			for _, mode := range modes {
				name := input + "/prec=" + strconv.Itoa(prec) + "/mode=" + strconv.Itoa(int(mode))
				t.Run(name, func(t *testing.T) {
					want, wantPanic := callRescale(exact, prec, mode)
					got, gotPanic := callFromRat(rat, prec, mode)

					if gotPanic != wantPanic {
						t.Fatalf("panic mismatch: got=%v want=%v", gotPanic, wantPanic)
					}
					if wantPanic {
						return
					}

					assertDecimalEqual(t, got, want)
					if got.Precision() != want.Precision() {
						t.Fatalf("precision mismatch: got=%d want=%d", got.Precision(), want.Precision())
					}
				})
			}
		}
	}
}

func TestNewWithPrec(t *testing.T) {
	got := NewWithPrec(0, 18)
	if got.Precision() != 18 {
		t.Fatalf("Precision() = %d, want 18", got.Precision())
	}
	if got.Sign() != 0 {
		t.Fatalf("Sign() = %d, want 0", got.Sign())
	}
	if got.String() != "0" {
		t.Fatalf("String() = %s, want 0", got.String())
	}
	if got.StringWithTrailingZeros() != "0.000000000000000000" {
		t.Fatalf("StringWithTrailingZeros() = %s, want 0.000000000000000000", got.StringWithTrailingZeros())
	}
}

func TestNewFromUintWithAppendPrec(t *testing.T) {
	got := NewFromUintWithAppendPrec(1, 4)
	if got.Precision() != 4 {
		t.Fatalf("Precision() = %d, want 4", got.Precision())
	}
	if got.String() != "1" {
		t.Fatalf("String() = %s, want 1", got.String())
	}
	if got.StringWithTrailingZeros() != "1.0000" {
		t.Fatalf("StringWithTrailingZeros() = %s, want 1.0000", got.StringWithTrailingZeros())
	}
}

func TestNegativePrecisionPanics(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{name: "new with prec", fn: func() { NewWithPrec(1, -1) }},
		{name: "new from int64", fn: func() { NewFromInt64(1, -1) }},
		{name: "new from uint64", fn: func() { NewFromUint64(1, -1) }},
		{name: "new from big int with prec", fn: func() { NewFromBigIntWithPrec(big.NewInt(1), -1) }},
		{name: "rescale", fn: func() { New(1).Rescale(-1, RoundDown) }},
		{name: "floor with prec", fn: func() { New(1).FloorWithPrec(-1) }},
		{name: "ceil with prec", fn: func() { New(1).CeilWithPrec(-1) }},
		{name: "truncate with prec", fn: func() { New(1).TruncateWithPrec(-1) }},
		{name: "round with prec", fn: func() { New(1).RoundWithPrec(-1) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertPanic(t, tc.fn)
		})
	}
}

func TestDecimalAdd(t *testing.T) {
	tests := []struct {
		name  string
		left  Decimal
		right Decimal
		want  Decimal
	}{
		{
			name:  "mixed precision",
			left:  NewFromUint64(1, 0),
			right: NewFromUint64(50, 1),
			want:  NewFromUint64(60, 1),
		},
		{
			name:  "same precision",
			left:  NewFromUint64(10, 1),
			right: NewFromUint64(5, 1),
			want:  NewFromUint64(15, 1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertDecimalEqual(t, tc.left.Add(tc.right), tc.want)
		})
	}
}

func TestDecimalSafeAdd(t *testing.T) {
	tests := []struct {
		name      string
		left      Decimal
		right     Decimal
		want      Decimal
		wantPanic bool
	}{
		{
			name:  "positive result",
			left:  NewFromInt64(1, 0),
			right: NewFromUint64(5, 1),
			want:  NewFromUint64(15, 1),
		},
		{
			name:  "zero result",
			left:  NewFromInt64(1, 0),
			right: NewFromInt64(-10, 1),
			want:  New(0),
		},
		{
			name:      "negative result panics",
			left:      NewFromInt64(1, 0),
			right:     NewFromInt64(-20, 1),
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				assertPanic(t, func() {
					_ = tc.left.SafeAdd(tc.right)
				})
				return
			}
			assertDecimalEqual(t, tc.left.SafeAdd(tc.right), tc.want)
		})
	}
}

func TestDecimalSafeSub(t *testing.T) {
	tests := []struct {
		name      string
		left      Decimal
		right     Decimal
		want      Decimal
		wantPanic bool
	}{
		{
			name:  "positive result",
			left:  NewFromInt64(1, 0),
			right: NewFromUint64(5, 1),
			want:  NewFromUint64(5, 1),
		},
		{
			name:  "negative operand",
			left:  NewFromInt64(1, 0),
			right: NewFromInt64(-10, 1),
			want:  NewFromUint64(20, 1),
		},
		{
			name:      "negative result panics",
			left:      NewFromInt64(1, 0),
			right:     NewFromUint64(20, 1),
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				assertPanic(t, func() {
					_ = tc.left.SafeSub(tc.right)
				})
				return
			}
			assertDecimalEqual(t, tc.left.SafeSub(tc.right), tc.want)
		})
	}
}

func TestDecimalMul(t *testing.T) {
	tests := []struct {
		name     string
		left     Decimal
		right    Decimal
		mode     RoundingMode
		want     Decimal
		wantPrec int
	}{
		{
			name:     "round down",
			left:     mustDecimal(t, "1.111"),
			right:    mustDecimal(t, "1.111"),
			mode:     RoundDown,
			want:     mustDecimal(t, "1.234"),
			wantPrec: 3,
		},
		{
			name:     "round up",
			left:     mustDecimal(t, "1.111"),
			right:    mustDecimal(t, "1.111"),
			mode:     RoundUp,
			want:     mustDecimal(t, "1.235"),
			wantPrec: 3,
		},
		{
			name:     "negative result",
			left:     mustDecimal(t, "-1.333"),
			right:    mustDecimal(t, "1.333"),
			mode:     RoundUp,
			want:     mustDecimal(t, "-1.777"),
			wantPrec: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.left.Mul(tc.right, tc.mode)
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}
}

func TestDecimalMulExact(t *testing.T) {
	tests := []struct {
		name     string
		left     Decimal
		right    Decimal
		want     Decimal
		wantPrec int
	}{
		{
			name:     "precision is sum of inputs",
			left:     mustDecimal(t, "1.20"),
			right:    mustDecimal(t, "2.30"),
			want:     mustDecimal(t, "2.7600"),
			wantPrec: 4,
		},
		{
			name:     "mixed precision",
			left:     mustDecimal(t, "1.234"),
			right:    mustDecimal(t, "2.5"),
			want:     mustDecimal(t, "3.0850"),
			wantPrec: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.left.MulExact(tc.right)
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("MulExact Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}

			deprecated := tc.left.Mul2(tc.right)
			assertDecimalEqual(t, deprecated, got)
			if deprecated.Precision() != got.Precision() {
				t.Fatalf("Mul2 Precision() = %d, want %d", deprecated.Precision(), got.Precision())
			}
		})
	}
}

func TestDecimalQuo(t *testing.T) {
	tests := []struct {
		name     string
		left     Decimal
		right    Decimal
		want     Decimal
		wantPrec int
	}{
		{
			name:     "integer division rounds up",
			left:     mustDecimal(t, "5"),
			right:    mustDecimal(t, "2"),
			want:     mustDecimal(t, "3"),
			wantPrec: 0,
		},
		{
			name:     "negative over negative",
			left:     mustDecimal(t, "-5"),
			right:    mustDecimal(t, "-2"),
			want:     mustDecimal(t, "3"),
			wantPrec: 0,
		},
		{
			name:     "fractional division",
			left:     mustDecimal(t, "55"),
			right:    mustDecimal(t, "100.0"),
			want:     mustDecimal(t, "0.6"),
			wantPrec: 1,
		},
		{
			name:     "negative fractional division",
			left:     mustDecimal(t, "-55"),
			right:    mustDecimal(t, "100.0"),
			want:     mustDecimal(t, "-0.6"),
			wantPrec: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.left.Quo(tc.right, RoundUp)
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}
}

func TestDecimalPower(t *testing.T) {
	tests := []struct {
		name     string
		value    Decimal
		power    int64
		want     Decimal
		wantPrec int
	}{
		{
			name:     "zero power",
			value:    mustDecimal(t, "2"),
			power:    0,
			want:     mustDecimal(t, "1"),
			wantPrec: 0,
		},
		{
			name:     "positive power",
			value:    mustDecimal(t, "2"),
			power:    3,
			want:     mustDecimal(t, "8"),
			wantPrec: 0,
		},
		{
			name:     "negative odd base",
			value:    mustDecimal(t, "-2"),
			power:    3,
			want:     mustDecimal(t, "-8"),
			wantPrec: 0,
		},
		{
			name:     "fractional base",
			value:    mustDecimal(t, "2.5"),
			power:    3,
			want:     mustDecimal(t, "15.6"),
			wantPrec: 1,
		},
		{
			name:     "negative power",
			value:    NewWithAppendPrec(2, 18),
			power:    -1,
			want:     mustDecimal(t, "0.5"),
			wantPrec: 18,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.Power(tc.power)
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}
}

func TestDecimalSqrtAndApproxRoot(t *testing.T) {
	sqrtTests := []struct {
		name     string
		value    Decimal
		want     Decimal
		wantPrec int
	}{
		{
			name:     "integer input retains source behavior",
			value:    mustDecimal(t, "4"),
			want:     mustDecimal(t, "1"),
			wantPrec: 0,
		},
		{
			name:     "scaled integer input",
			value:    mustDecimal(t, "4.0000"),
			want:     mustDecimal(t, "2"),
			wantPrec: 4,
		},
		{
			name:     "decimal input",
			value:    mustDecimal(t, "16.0"),
			want:     mustDecimal(t, "4.0"),
			wantPrec: 1,
		},
		{
			name:     "fractional input",
			value:    NewWithPrec(25, 2),
			want:     NewWithPrec(5, 1),
			wantPrec: 2,
		},
		{
			name:     "high precision input",
			value:    NewWithAppendPrec(2, 18),
			want:     NewWithPrec(1414213562373095049, 18),
			wantPrec: 18,
		},
	}

	for _, tc := range sqrtTests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.value.Sqrt()
			if err != nil {
				t.Fatalf("Sqrt() returned error: %v", err)
			}
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}

	t.Run("negative input returns error", func(t *testing.T) {
		_, err := mustDecimal(t, "-4").Sqrt()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	approxRootTests := []struct {
		name     string
		value    Decimal
		root     int64
		want     Decimal
		wantPrec int
	}{
		{
			name:     "fifth root of 3125",
			value:    mustDecimal(t, "3125.0000"),
			root:     5,
			want:     mustDecimal(t, "5.0000"),
			wantPrec: 4,
		},
		{
			name:     "fifth root of 100000",
			value:    mustDecimal(t, "100000.0000"),
			root:     5,
			want:     mustDecimal(t, "10.0000"),
			wantPrec: 4,
		},
	}

	for _, tc := range approxRootTests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.value.ApproxRoot(tc.root)
			if err != nil {
				t.Fatalf("ApproxRoot(%d) returned error: %v", tc.root, err)
			}
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}

	t.Run("invalid root", func(t *testing.T) {
		if _, err := mustDecimal(t, "16.0").ApproxRoot(0); err == nil {
			t.Fatal("expected error for zero root")
		}
		if _, err := mustDecimal(t, "16.0").ApproxRoot(-2); err == nil {
			t.Fatal("expected error for negative root")
		}
	})

	t.Run("even root of negative value", func(t *testing.T) {
		if _, err := mustDecimal(t, "-16.0").ApproxRoot(2); err == nil {
			t.Fatal("expected error for even root of negative value")
		}
	})
}

func TestDecimalLog2(t *testing.T) {
	bigLogInput := new(big.Int)
	if _, ok := bigLogInput.SetString("200000000000000000", 10); !ok {
		t.Fatal("failed to construct test big.Int")
	}

	tests := []struct {
		name     string
		value    Decimal
		want     Decimal
		wantPrec int
	}{
		{
			name:     "log2 of 1",
			value:    NewWithAppendPrec(1, 18),
			want:     New(0),
			wantPrec: 18,
		},
		{
			name:     "log2 of 2",
			value:    NewWithAppendPrec(2, 18),
			want:     New(1),
			wantPrec: 18,
		},
		{
			name:     "log2 of 33",
			value:    NewWithAppendPrec(33, 18),
			want:     mustDecimal(t, "5.044394119358453436"),
			wantPrec: 18,
		},
		{
			name:     "log2 of 2.12345678",
			value:    mustDecimal(t, "2.12345678"),
			want:     mustDecimal(t, "1.08641474"),
			wantPrec: 8,
		},
		{
			name:     "log2 of 0.2",
			value:    NewFromBigIntWithPrec(bigLogInput, 18),
			want:     mustDecimal(t, "-2.321928094887362348"),
			wantPrec: 18,
		},
		{
			name:     "log2 of tiny number",
			value:    NewWithPrec(2, 18),
			want:     mustDecimal(t, "-58.794705707972522263"),
			wantPrec: 18,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.Log2()
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}
}

func TestDecimalRescale(t *testing.T) {
	tests := []struct {
		name     string
		value    Decimal
		prec     int
		mode     RoundingMode
		want     Decimal
		wantPrec int
	}{
		{
			name:     "increase precision",
			value:    New(1),
			prec:     2,
			mode:     RoundDown,
			want:     New(1),
			wantPrec: 2,
		},
		{
			name:     "round down",
			value:    NewWithPrec(55, 1),
			prec:     0,
			mode:     RoundDown,
			want:     New(5),
			wantPrec: 0,
		},
		{
			name:     "round up",
			value:    NewWithPrec(55, 1),
			prec:     0,
			mode:     RoundUp,
			want:     New(6),
			wantPrec: 0,
		},
		{
			name:     "round ceiling with negative value",
			value:    NewWithPrec(-55, 1),
			prec:     0,
			mode:     RoundCeiling,
			want:     New(-5),
			wantPrec: 0,
		},
		{
			name:     "round half up",
			value:    NewWithPrec(25, 1),
			prec:     0,
			mode:     RoundHalfUp,
			want:     New(3),
			wantPrec: 0,
		},
		{
			name:     "round half down",
			value:    NewWithPrec(25, 1),
			prec:     0,
			mode:     RoundHalfDown,
			want:     New(2),
			wantPrec: 0,
		},
		{
			name:     "round half even",
			value:    NewWithPrec(25, 1),
			prec:     0,
			mode:     RoundHalfEven,
			want:     New(2),
			wantPrec: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.Rescale(tc.prec, tc.mode)
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}

	t.Run("round unnecessary panics on inexact value", func(t *testing.T) {
		assertPanic(t, func() {
			NewWithPrec(10001, 4).Rescale(0, RoundUnnecessary)
		})
	})

	t.Run("round unnecessary succeeds on exact value", func(t *testing.T) {
		got := NewWithPrec(10000, 4).Rescale(0, RoundUnnecessary)
		assertDecimalEqual(t, got, New(1))
		if got.Precision() != 0 {
			t.Fatalf("Precision() = %d, want 0", got.Precision())
		}
	})
}

func TestDecimalShift(t *testing.T) {
	tests := []struct {
		name     string
		input    Decimal
		places   int
		want     Decimal
		wantPrec int
	}{
		{
			name:     "shift right within precision",
			input:    mustDecimal(t, "12.34"),
			places:   1,
			want:     mustDecimal(t, "123.4"),
			wantPrec: 1,
		},
		{
			name:     "shift right beyond precision",
			input:    mustDecimal(t, "12.34"),
			places:   4,
			want:     mustDecimal(t, "123400"),
			wantPrec: 0,
		},
		{
			name:     "shift left",
			input:    mustDecimal(t, "12.34"),
			places:   -3,
			want:     mustDecimal(t, "0.01234"),
			wantPrec: 5,
		},
		{
			name:     "negative value",
			input:    mustDecimal(t, "-12.34"),
			places:   2,
			want:     mustDecimal(t, "-1234"),
			wantPrec: 0,
		},
		{
			name:     "zero value",
			input:    NewWithPrec(0, 2),
			places:   -4,
			want:     NewWithPrec(0, 6),
			wantPrec: 6,
		},
		{
			name:     "zero places",
			input:    mustDecimal(t, "12.34"),
			places:   0,
			want:     mustDecimal(t, "12.34"),
			wantPrec: 2,
		},
		{
			name:     "nil decimal",
			input:    Decimal{},
			places:   3,
			want:     Zero,
			wantPrec: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Shift(tc.places)
			assertDecimalEqual(t, got, tc.want)
			if got.Precision() != tc.wantPrec {
				t.Fatalf("Precision() = %d, want %d", got.Precision(), tc.wantPrec)
			}
		})
	}
}

func TestDecimalIntegerRoundingMethods(t *testing.T) {
	tests := []struct {
		name         string
		input        Decimal
		wantFloor    string
		wantCeil     string
		wantTruncate string
		wantRound    string
	}{
		{
			name:         "positive fractional",
			input:        mustDecimal(t, "1.9"),
			wantFloor:    "1",
			wantCeil:     "2",
			wantTruncate: "1",
			wantRound:    "2",
		},
		{
			name:         "negative fractional",
			input:        mustDecimal(t, "-1.9"),
			wantFloor:    "-2",
			wantCeil:     "-1",
			wantTruncate: "-1",
			wantRound:    "-2",
		},
		{
			name:         "half even tie rounds down to even",
			input:        mustDecimal(t, "2.5"),
			wantFloor:    "2",
			wantCeil:     "3",
			wantTruncate: "2",
			wantRound:    "2",
		},
		{
			name:         "half even tie rounds up to even",
			input:        mustDecimal(t, "3.5"),
			wantFloor:    "3",
			wantCeil:     "4",
			wantTruncate: "3",
			wantRound:    "4",
		},
		{
			name:         "negative half even tie",
			input:        mustDecimal(t, "-2.5"),
			wantFloor:    "-3",
			wantCeil:     "-2",
			wantTruncate: "-2",
			wantRound:    "-2",
		},
		{
			name:         "trailing zeros integer",
			input:        mustDecimal(t, "2.000"),
			wantFloor:    "2",
			wantCeil:     "2",
			wantTruncate: "2",
			wantRound:    "2",
		},
		{
			name:         "zero value",
			input:        Decimal{},
			wantFloor:    "0",
			wantCeil:     "0",
			wantTruncate: "0",
			wantRound:    "0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cases := []struct {
				name string
				got  Decimal
				want string
			}{
				{name: "Floor", got: tc.input.Floor(), want: tc.wantFloor},
				{name: "Ceil", got: tc.input.Ceil(), want: tc.wantCeil},
				{name: "Truncate", got: tc.input.Truncate(), want: tc.wantTruncate},
				{name: "Round", got: tc.input.Round(), want: tc.wantRound},
			}

			for _, method := range cases {
				t.Run(method.name, func(t *testing.T) {
					want := mustDecimal(t, method.want)
					assertDecimalEqual(t, method.got, want)
					if method.got.Precision() != 0 {
						t.Fatalf("%s precision = %d, want 0", method.name, method.got.Precision())
					}
				})
			}
		})
	}
}

func TestDecimalRoundingMethodsWithPrecision(t *testing.T) {
	tests := []struct {
		name         string
		input        Decimal
		prec         int
		wantFloor    string
		wantCeil     string
		wantTruncate string
		wantRound    string
		wantPrec     int
	}{
		{
			name:         "positive value",
			input:        mustDecimal(t, "1.239"),
			prec:         2,
			wantFloor:    "1.23",
			wantCeil:     "1.24",
			wantTruncate: "1.23",
			wantRound:    "1.24",
			wantPrec:     2,
		},
		{
			name:         "negative value",
			input:        mustDecimal(t, "-1.239"),
			prec:         2,
			wantFloor:    "-1.24",
			wantCeil:     "-1.23",
			wantTruncate: "-1.23",
			wantRound:    "-1.24",
			wantPrec:     2,
		},
		{
			name:         "half even tie",
			input:        mustDecimal(t, "1.245"),
			prec:         2,
			wantFloor:    "1.24",
			wantCeil:     "1.25",
			wantTruncate: "1.24",
			wantRound:    "1.24",
			wantPrec:     2,
		},
		{
			name:         "increase precision",
			input:        mustDecimal(t, "1.2"),
			prec:         4,
			wantFloor:    "1.2",
			wantCeil:     "1.2",
			wantTruncate: "1.2",
			wantRound:    "1.2",
			wantPrec:     4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cases := []struct {
				name string
				got  Decimal
				want string
			}{
				{name: "FloorWithPrec", got: tc.input.FloorWithPrec(tc.prec), want: tc.wantFloor},
				{name: "CeilWithPrec", got: tc.input.CeilWithPrec(tc.prec), want: tc.wantCeil},
				{name: "TruncateWithPrec", got: tc.input.TruncateWithPrec(tc.prec), want: tc.wantTruncate},
				{name: "RoundWithPrec", got: tc.input.RoundWithPrec(tc.prec), want: tc.wantRound},
			}

			for _, method := range cases {
				t.Run(method.name, func(t *testing.T) {
					want := mustDecimal(t, method.want)
					assertDecimalEqual(t, method.got, want)
					if method.got.Precision() != tc.wantPrec {
						t.Fatalf("%s precision = %d, want %d", method.name, method.got.Precision(), tc.wantPrec)
					}
				})
			}
		})
	}
}

func TestDecimalIsIntegerAndHasFraction(t *testing.T) {
	tests := []struct {
		name         string
		input        Decimal
		wantInteger  bool
		wantFraction bool
	}{
		{name: "integer", input: New(1), wantInteger: true, wantFraction: false},
		{name: "trailing zeros", input: mustDecimal(t, "1.000"), wantInteger: true, wantFraction: false},
		{name: "fraction", input: mustDecimal(t, "1.2"), wantInteger: false, wantFraction: true},
		{name: "negative trailing zeros", input: mustDecimal(t, "-2.0"), wantInteger: true, wantFraction: false},
		{name: "zero value", input: Decimal{}, wantInteger: true, wantFraction: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.input.IsInteger(); got != tc.wantInteger {
				t.Fatalf("IsInteger() = %v, want %v", got, tc.wantInteger)
			}
			if got := tc.input.HasFraction(); got != tc.wantFraction {
				t.Fatalf("HasFraction() = %v, want %v", got, tc.wantFraction)
			}
			if tc.input.HasFraction() == tc.input.IsInteger() {
				t.Fatal("HasFraction() should be the inverse of IsInteger()")
			}
		})
	}
}

func TestDecimalQuoRem(t *testing.T) {
	tests := []struct {
		name      string
		left      Decimal
		right     Decimal
		wantQuo   string
		wantRem   string
		wantRecom string
	}{
		{
			name:      "positive integers",
			left:      New(7),
			right:     New(3),
			wantQuo:   "2",
			wantRem:   "1",
			wantRecom: "7",
		},
		{
			name:      "positive over negative",
			left:      New(7),
			right:     New(-3),
			wantQuo:   "-2",
			wantRem:   "1",
			wantRecom: "7",
		},
		{
			name:      "negative over positive",
			left:      New(-7),
			right:     New(3),
			wantQuo:   "-2",
			wantRem:   "-1",
			wantRecom: "-7",
		},
		{
			name:      "negative integers",
			left:      New(-7),
			right:     New(-3),
			wantQuo:   "2",
			wantRem:   "-1",
			wantRecom: "-7",
		},
		{
			name:      "fractional operands",
			left:      mustDecimal(t, "5.5"),
			right:     mustDecimal(t, "2.0"),
			wantQuo:   "2",
			wantRem:   "1.5",
			wantRecom: "5.5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotQuo, gotRem := tc.left.QuoRem(tc.right)

			assertDecimalEqual(t, gotQuo, mustDecimal(t, tc.wantQuo))
			assertDecimalEqual(t, gotRem, mustDecimal(t, tc.wantRem))

			recombined := gotQuo.MulExact(tc.right).Add(gotRem)
			assertDecimalEqual(t, recombined, mustDecimal(t, tc.wantRecom))
		})
	}
}

func TestDecimalQuoRemPanicsOnZeroDivisor(t *testing.T) {
	assertPanic(t, func() {
		_, _ = New(1).QuoRem(Zero)
	})
}

func TestDecimalMod(t *testing.T) {
	tests := []struct {
		name  string
		left  Decimal
		right Decimal
		want  string
	}{
		{name: "positive integers", left: New(7), right: New(3), want: "1"},
		{name: "positive over negative", left: New(7), right: New(-3), want: "1"},
		{name: "negative over positive", left: New(-7), right: New(3), want: "-1"},
		{name: "negative integers", left: New(-7), right: New(-3), want: "-1"},
		{name: "fractional operands", left: mustDecimal(t, "5.5"), right: mustDecimal(t, "2.0"), want: "1.5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertDecimalEqual(t, tc.left.Mod(tc.right), mustDecimal(t, tc.want))
		})
	}
}

func TestDecimalModPanicsOnZeroDivisor(t *testing.T) {
	assertPanic(t, func() {
		_ = New(1).Mod(Zero)
	})
}

func TestDecimalFloat64(t *testing.T) {
	overflow := NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(400), nil))

	tests := []struct {
		name      string
		input     Decimal
		want      float64
		wantExact bool
		checkInf  bool
		infSign   int
	}{
		{name: "exact integer", input: New(42), want: 42, wantExact: true},
		{name: "exact fraction", input: mustDecimal(t, "0.5"), want: 0.5, wantExact: true},
		{name: "inexact fraction", input: mustDecimal(t, "0.1"), want: 0.1, wantExact: false},
		{name: "zero value", input: Decimal{}, want: 0, wantExact: true},
		{name: "positive overflow", input: overflow, want: math.Inf(1), wantExact: false, checkInf: true, infSign: 1},
		{name: "negative overflow", input: overflow.Neg(), want: math.Inf(-1), wantExact: false, checkInf: true, infSign: -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, exact := tc.input.Float64()
			if tc.checkInf {
				if !math.IsInf(got, tc.infSign) {
					t.Fatalf("Float64() = %v, want %v", got, tc.want)
				}
			} else if got != tc.want {
				t.Fatalf("Float64() = %v, want %v", got, tc.want)
			}
			if exact != tc.wantExact {
				t.Fatalf("Float64() exact = %v, want %v", exact, tc.wantExact)
			}
		})
	}
}

func TestDecimalFloat32(t *testing.T) {
	overflow := NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(400), nil))

	tests := []struct {
		name      string
		input     Decimal
		want      float32
		wantExact bool
		checkInf  bool
		infSign   int
	}{
		{name: "exact integer", input: New(42), want: 42, wantExact: true},
		{name: "exact fraction", input: mustDecimal(t, "0.5"), want: 0.5, wantExact: true},
		{name: "inexact fraction", input: mustDecimal(t, "0.1"), want: 0.1, wantExact: false},
		{name: "zero value", input: Decimal{}, want: 0, wantExact: true},
		{name: "positive overflow", input: overflow, want: float32(math.Inf(1)), wantExact: false, checkInf: true, infSign: 1},
		{name: "negative overflow", input: overflow.Neg(), want: float32(math.Inf(-1)), wantExact: false, checkInf: true, infSign: -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, exact := tc.input.Float32()
			if tc.checkInf {
				if !math.IsInf(float64(got), tc.infSign) {
					t.Fatalf("Float32() = %v, want %v", got, tc.want)
				}
			} else if got != tc.want {
				t.Fatalf("Float32() = %v, want %v", got, tc.want)
			}
			if exact != tc.wantExact {
				t.Fatalf("Float32() exact = %v, want %v", exact, tc.wantExact)
			}
		})
	}
}

func TestDecimalBigRat(t *testing.T) {
	tests := []struct {
		name  string
		input Decimal
		want  *big.Rat
	}{
		{name: "positive integer", input: New(42), want: big.NewRat(42, 1)},
		{name: "negative fraction", input: mustDecimal(t, "-12.34"), want: big.NewRat(-617, 50)},
		{name: "fraction with trailing zeros", input: mustDecimal(t, "1.2300"), want: big.NewRat(123, 100)},
		{name: "zero", input: Zero, want: big.NewRat(0, 1)},
		{name: "zero value", input: Decimal{}, want: big.NewRat(0, 1)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.BigRat()
			if got.Cmp(tc.want) != 0 {
				t.Fatalf("BigRat() = %s, want %s", got.RatString(), tc.want.RatString())
			}
		})
	}

	t.Run("returned rat does not mutate decimal", func(t *testing.T) {
		input := mustDecimal(t, "12.34")

		got := input.BigRat()
		got.Add(got, big.NewRat(1, 3))

		assertDecimalEqual(t, input, mustDecimal(t, "12.34"))

		again := input.BigRat()
		want := big.NewRat(617, 50)
		if again.Cmp(want) != 0 {
			t.Fatalf("BigRat() after mutation = %s, want %s", again.RatString(), want.RatString())
		}
	})
}

func TestDecimalInt64(t *testing.T) {
	tests := []struct {
		name   string
		input  Decimal
		want   int64
		wantOK bool
	}{
		{name: "exact integer", input: New(42), want: 42, wantOK: true},
		{name: "exact trailing zeros", input: mustDecimal(t, "42.0"), want: 42, wantOK: true},
		{name: "negative integer", input: New(-42), want: -42, wantOK: true},
		{name: "max int64", input: mustDecimal(t, "9223372036854775807"), want: math.MaxInt64, wantOK: true},
		{name: "min int64", input: mustDecimal(t, "-9223372036854775808"), want: math.MinInt64, wantOK: true},
		{name: "non integer", input: mustDecimal(t, "42.5"), want: 0, wantOK: false},
		{name: "overflow", input: mustDecimal(t, "9223372036854775808"), want: 0, wantOK: false},
		{name: "zero value", input: Decimal{}, want: 0, wantOK: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.input.Int64()
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("Int64() = (%d, %v), want (%d, %v)", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestDecimalUint64(t *testing.T) {
	tests := []struct {
		name   string
		input  Decimal
		want   uint64
		wantOK bool
	}{
		{name: "exact integer", input: New(42), want: 42, wantOK: true},
		{name: "exact trailing zeros", input: mustDecimal(t, "42.0"), want: 42, wantOK: true},
		{name: "max uint64", input: mustDecimal(t, "18446744073709551615"), want: math.MaxUint64, wantOK: true},
		{name: "negative integer", input: New(-1), want: 0, wantOK: false},
		{name: "non integer", input: mustDecimal(t, "42.5"), want: 0, wantOK: false},
		{name: "overflow", input: mustDecimal(t, "18446744073709551616"), want: 0, wantOK: false},
		{name: "zero value", input: Decimal{}, want: 0, wantOK: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.input.Uint64()
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("Uint64() = (%d, %v), want (%d, %v)", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestDecimalStripTrailingZeros(t *testing.T) {
	tests := []struct {
		name  string
		value Decimal
		want  string
	}{
		{name: "zero", value: mustDecimal(t, "0"), want: "0"},
		{name: "zero with decimals", value: mustDecimal(t, "0.00"), want: "0"},
		{name: "simple fractional", value: mustDecimal(t, "0.10"), want: "0.1"},
		{name: "already stripped", value: mustDecimal(t, "0.11"), want: "0.11"},
		{name: "large trailing zeros", value: mustDecimal(t, "0.110000000000"), want: "0.11"},
		{name: "positive integer scale", value: mustDecimal(t, "1.100"), want: "1.1"},
		{name: "negative zero with decimals", value: mustDecimal(t, "-0.00"), want: "0"},
		{name: "negative fractional", value: mustDecimal(t, "-0.110000000000"), want: "-0.11"},
		{name: "negative integer scale", value: mustDecimal(t, "-1.100"), want: "-1.1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.value.StripTrailingZeros().String(); got != tc.want {
				t.Fatalf("StripTrailingZeros().String() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestDecimalSignificantFigures(t *testing.T) {
	tests := []struct {
		name    string
		value   Decimal
		figures int
		mode    RoundingMode
		want    string
	}{
		{name: "zero", value: mustDecimal(t, "0"), figures: 1, mode: RoundUp, want: "0"},
		{name: "subunitary", value: mustDecimal(t, "0.001001"), figures: 1, mode: RoundUp, want: "0.002"},
		{name: "subunitary two figures", value: mustDecimal(t, "0.001001"), figures: 2, mode: RoundUp, want: "0.0011"},
		{name: "subunitary exact", value: mustDecimal(t, "0.001001"), figures: 4, mode: RoundUp, want: "0.001001"},
		{name: "negative subunitary", value: mustDecimal(t, "-0.001001"), figures: 2, mode: RoundUp, want: "-0.0011"},
		{name: "unitary", value: mustDecimal(t, "1.001001"), figures: 2, mode: RoundUp, want: "1.1"},
		{name: "large unitary", value: mustDecimal(t, "1111.001001"), figures: 5, mode: RoundUp, want: "1111.1"},
		{name: "negative large unitary", value: mustDecimal(t, "-1111.001001"), figures: 5, mode: RoundUp, want: "-1111.1"},
		{name: "round to tens", value: mustDecimal(t, "123.456"), figures: 2, mode: RoundHalfEven, want: "120"},
		{name: "round to hundreds", value: mustDecimal(t, "123.456"), figures: 1, mode: RoundHalfEven, want: "100"},
		{name: "round with carry expansion", value: mustDecimal(t, "999"), figures: 2, mode: RoundHalfEven, want: "1000"},
		{name: "negative round to tens", value: mustDecimal(t, "-123.456"), figures: 2, mode: RoundHalfEven, want: "-120"},
		{name: "round down toward zero at tens", value: mustDecimal(t, "199"), figures: 1, mode: RoundDown, want: "100"},
		{name: "round up away from zero at tens", value: mustDecimal(t, "199"), figures: 1, mode: RoundUp, want: "200"},
		{name: "negative round down toward zero at tens", value: mustDecimal(t, "-199"), figures: 1, mode: RoundDown, want: "-100"},
		{name: "negative round up away from zero at tens", value: mustDecimal(t, "-199"), figures: 1, mode: RoundUp, want: "-200"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.SignificantFigures(tc.figures, tc.mode)
			if got.String() != tc.want {
				t.Fatalf("SignificantFigures(%d, %v) = %s, want %s", tc.figures, tc.mode, got.String(), tc.want)
			}
		})
	}
}
