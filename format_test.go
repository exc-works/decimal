package decimal

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func TestFormatVerbs(t *testing.T) {
	d := MustFromString("12345.678")

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{"verb_v", "%v", "12345.678"},
		{"verb_s", "%s", "12345.678"},
		{"verb_f_default", "%f", "12345.678"},
		{"verb_f_prec2", "%.2f", "12345.68"},
		{"verb_f_prec0", "%.0f", "12346"},
		{"verb_f_prec5", "%.5f", "12345.67800"},
		{"verb_e_default", "%e", "1.234568e+04"},
		{"verb_e_prec2", "%.2e", "1.23e+04"},
		{"verb_e_prec0", "%.0e", "1e+04"},
		{"verb_E_default", "%E", "1.234568E+04"},
		{"verb_g_default", "%g", "12345.678"},
		{"verb_g_prec2_scientific", "%.2g", "1.2e+04"},
		{"verb_G_prec2_scientific", "%.2G", "1.2E+04"},
		{"verb_g_prec0", "%.0g", "1e+04"},
		{"verb_g_prec5", "%.5g", "12346"},
		{"verb_q", "%q", `"12345.678"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fmt.Sprintf(tc.format, d)
			if got != tc.want {
				t.Fatalf("Sprintf(%q, d) = %q, want %q", tc.format, got, tc.want)
			}
		})
	}
}

func TestFormatVerbDInteger(t *testing.T) {
	d := MustFromString("12345")
	if got := fmt.Sprintf("%d", d); got != "12345" {
		t.Fatalf("Sprintf(%%d, 12345) = %q, want %q", got, "12345")
	}

	neg := MustFromString("-9876")
	if got := fmt.Sprintf("%d", neg); got != "-9876" {
		t.Fatalf("Sprintf(%%d, -9876) = %q, want %q", got, "-9876")
	}
}

func TestFormatVerbBinary(t *testing.T) {
	d := MustFromString("12345.67")
	got := fmt.Sprintf("%b", d)
	// Unscaled integer is 1234567, binary "100101101011010000111", scale 2.
	want := "100101101011010000111p-2"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatVerbDPrecision(t *testing.T) {
	tests := []struct {
		format string
		value  string
		want   string
	}{
		{"%.4d", "12", "0012"},
		{"%.4d", "-12", "-0012"},
		{"%.0d", "0", ""},
		{"%.0d", "5", "5"},
		{"%.3d", "12345", "12345"},
		{"%+.4d", "12", "+0012"},
		{"%+.4d", "-12", "-0012"},
		{"% .4d", "12", " 0012"},
		{"%5.3d", "12", "  012"},
		{"%-5.3d", "12", "012  "},
		{"%05.3d", "12", "  012"},
		{"%+.0d", "0", ""},
	}
	for _, tc := range tests {
		d := MustFromString(tc.value)
		if got := fmt.Sprintf(tc.format, d); got != tc.want {
			t.Errorf("Sprintf(%q, %s) = %q, want %q", tc.format, tc.value, got, tc.want)
		}
	}
}

func TestFormatVerbBinarySignFlags(t *testing.T) {
	d := MustFromString("3.5")
	// Unscaled integer 35, binary "100011", scale 1.
	if got := fmt.Sprintf("%b", d); got != "100011p-1" {
		t.Fatalf("%%b got %q", got)
	}
	if got := fmt.Sprintf("%+b", d); got != "+100011p-1" {
		t.Fatalf("%%+b got %q", got)
	}
	if got := fmt.Sprintf("% b", d); got != " 100011p-1" {
		t.Fatalf("%% b got %q", got)
	}
	// Negative values keep their sign, not doubled.
	neg := MustFromString("-3.5")
	if got := fmt.Sprintf("%+b", neg); got != "-100011p-1" {
		t.Fatalf("%%+b neg got %q", got)
	}
}

func TestFormatNegativeValues(t *testing.T) {
	d := MustFromString("-9876.54")
	tests := []struct {
		format, want string
	}{
		{"%v", "-9876.54"},
		{"%s", "-9876.54"},
		{"%.2f", "-9876.54"},
		{"%.3e", "-9.877e+03"},
	}
	for _, tc := range tests {
		got := fmt.Sprintf(tc.format, d)
		if got != tc.want {
			t.Errorf("Sprintf(%q, -9876.54) = %q, want %q", tc.format, got, tc.want)
		}
	}
}

func TestFormatWidthAndPrecision(t *testing.T) {
	// 12.355 -> 12.36 under RoundHalfEven (5 rounds away from even 5 -> 6).
	d := MustFromString("12.355")

	tests := []struct {
		format, want string
	}{
		{"%10.2f", "     12.36"},
		{"%-10.2f", "12.36     "},
		{"%010.2f", "0000012.36"},
		{"%+.2f", "+12.36"},
		{"% .2f", " 12.36"},
		{"%+10.2f", "    +12.36"},
		{"%+010.2f", "+000012.36"},
	}
	for _, tc := range tests {
		got := fmt.Sprintf(tc.format, d)
		if got != tc.want {
			t.Errorf("Sprintf(%q, 12.355) = %q, want %q", tc.format, got, tc.want)
		}
	}
}

func TestFormatSignFlagsNegative(t *testing.T) {
	d := MustFromString("-12.355")
	// For negatives, '+' and ' ' should not double up the sign.
	if got := fmt.Sprintf("%+.2f", d); got != "-12.36" {
		t.Errorf("got %q, want %q", got, "-12.36")
	}
	if got := fmt.Sprintf("% .2f", d); got != "-12.36" {
		t.Errorf("got %q, want %q", got, "-12.36")
	}
}

func TestFormatUnknownVerb(t *testing.T) {
	d := MustFromString("1.23")
	got := fmt.Sprintf("%y", d)
	if !strings.HasPrefix(got, "%!") {
		t.Fatalf("expected unknown-verb error prefix, got %q", got)
	}
	if !strings.Contains(got, "decimal.Decimal=") {
		t.Fatalf("expected decimal.Decimal= tag in unknown verb output, got %q", got)
	}
	if !strings.Contains(got, "1.23") {
		t.Fatalf("expected string form in unknown verb output, got %q", got)
	}
}

func TestFormatZero(t *testing.T) {
	d := Zero
	if got := fmt.Sprintf("%v", d); got != "0" {
		t.Errorf("Sprintf(%%v, 0) = %q", got)
	}
	if got := fmt.Sprintf("%.3e", d); got != "0.000e+00" {
		t.Errorf("Sprintf(%%.3e, 0) = %q", got)
	}
	if got := fmt.Sprintf("%.2f", d); got != "0.00" {
		t.Errorf("Sprintf(%%.2f, 0) = %q", got)
	}
}

func TestClone(t *testing.T) {
	raw := big.NewInt(123456)
	d := NewFromBigIntWithPrec(raw, 2)
	clone := d.Clone()

	// Mutating the external big.Int must not affect the clone or the
	// original Decimal, since both took defensive copies.
	raw.SetInt64(999999)

	if clone.String() != "1234.56" {
		t.Errorf("clone changed after raw mutation: %q", clone.String())
	}
	if d.String() != "1234.56" {
		t.Errorf("original changed after raw mutation: %q", d.String())
	}

	// Confirm internal big.Int pointers differ, i.e. clone is independent.
	if d.i == clone.i {
		t.Errorf("Clone returned shared big.Int pointer")
	}
}

func TestCloneNilReceiver(t *testing.T) {
	var d Decimal
	clone := d.Clone()
	if clone.i != nil {
		t.Errorf("expected nil clone for nil receiver, got %+v", clone)
	}
}

func TestNewFromDecimalRoundTrip(t *testing.T) {
	original := MustFromString("3.14159")
	copyDec := NewFromDecimal(original)

	if !copyDec.Equal(original) {
		t.Errorf("copy not equal to original: %q vs %q", copyDec.String(), original.String())
	}
	if copyDec.i == original.i {
		t.Errorf("NewFromDecimal should return independent big.Int")
	}
}

func TestFormatWithSeparators(t *testing.T) {
	tests := []struct {
		value     string
		thousands rune
		decimal   rune
		want      string
	}{
		{"12345.678", ',', '.', "12,345.678"},
		{"1000000", ' ', '.', "1 000 000"},
		{"-9876.54", ',', '.', "-9,876.54"},
		{"0", ',', '.', "0"},
		{"12.3", ',', '.', "12.3"},
		{"1234567.89", '.', ',', "1.234.567,89"},
		{"999", ',', '.', "999"},
		{"1234", ',', '.', "1,234"},
		{"-1234567", ',', '.', "-1,234,567"},
		{"1234567", 0, '.', "1234567"},
	}
	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			d := MustFromString(tc.value)
			got := d.FormatWithSeparators(tc.thousands, tc.decimal)
			if got != tc.want {
				t.Errorf("FormatWithSeparators(%q, %q, %q) = %q, want %q",
					tc.value, tc.thousands, tc.decimal, got, tc.want)
			}
		})
	}
}

func TestFormatVerbDRejectsNonInteger(t *testing.T) {
	d := MustFromString("1.5")
	got := fmt.Sprintf("%d", d)
	want := "%!d(decimal.Decimal=1.5)"
	if got != want {
		t.Fatalf("Sprintf(%%d, 1.5) = %q, want %q", got, want)
	}
}

func TestFormatG_MatchesFloat(t *testing.T) {
	cases := []struct {
		format string
		input  string
		ref    float64
	}{
		{"%.2g", "12345.678", 12345.678},
		{"%.2G", "12345.678", 12345.678},
		{"%.3g", "0.00012345", 0.00012345},
		{"%.4g", "1", 1},
		{"%.4g", "123456", 123456},
		{"%.2g", "0.5", 0.5},
		{"%.2g", "1.0", 1.0},
		{"%.1g", "100", 100},
		{"%.0g", "12345.678", 12345.678},
		{"%g", "1234.5678", 1234.5678},
		{"%g", "100", 100},
		{"%g", "1234.5600", 1234.56},
		{"%g", "0.00001", 0.00001},
		{"%g", "0", 0},
		{"%g", "-100", -100},
		{"%.5g", "0.0001", 0.0001},
		{"%.3G", "12345", 12345},
		{"%g", "1000000", 1000000},
		{"%g", "10000000", 10000000},
		{"%g", "12345678", 12345678},
		{"%g", "100000", 100000},
		{"%g", "999999", 999999},
		{"%g", "1230", 1230},
		{"%g", "0.0001", 0.0001},
		{"%g", "-1000000", -1000000},
	}
	for _, tc := range cases {
		t.Run(tc.format+"_"+tc.input, func(t *testing.T) {
			d := MustFromString(tc.input)
			got := fmt.Sprintf(tc.format, d)
			want := fmt.Sprintf(tc.format, tc.ref)
			if got != want {
				t.Errorf("%s(%s) = %q, want %q (float reference)", tc.format, tc.input, got, want)
			}
		})
	}
}
