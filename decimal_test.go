package decimal

import (
	"math/big"
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
