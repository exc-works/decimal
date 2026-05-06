package decimal

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

var (
	benchDecimalResult Decimal
	benchStringResult  string
	benchBytesResult   []byte
)

func BenchmarkDecimal_Add(b *testing.B) {
	tests := []struct {
		left  Decimal
		right Decimal
		name  string
	}{
		{name: "int/int", left: MustFromString("123456789"), right: MustFromString("37")},
		{name: "same-prec", left: MustFromString("12345.6789"), right: MustFromString("0.3701")},
		{name: "mixed-prec", left: MustFromString("1.23456789"), right: MustFromString("123")},
		{name: "high-prec", left: MustFromString("123456789.123456789123456789"), right: MustFromString("0.00000000123456789")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchDecimalResult = tt.left.Add(tt.right)
			}
		})
	}
}

func BenchmarkDecimal_Sub(b *testing.B) {
	tests := []struct {
		left  Decimal
		right Decimal
		name  string
	}{
		{name: "int/int", left: MustFromString("123456789"), right: MustFromString("37")},
		{name: "same-prec", left: MustFromString("12345.6789"), right: MustFromString("0.3701")},
		{name: "mixed-prec", left: MustFromString("1.23456789"), right: MustFromString("123")},
		{name: "high-prec", left: MustFromString("123456789.123456789123456789"), right: MustFromString("0.00000000123456789")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchDecimalResult = tt.left.Sub(tt.right)
			}
		})
	}
}

func BenchmarkDecimal_Mul(b *testing.B) {
	tests := []struct {
		left  Decimal
		right Decimal
		name  string
	}{
		{name: "int/int", left: MustFromString("123456789"), right: MustFromString("37")},
		{name: "same-prec", left: MustFromString("12345.6789"), right: MustFromString("0.3701")},
		{name: "mixed-prec", left: MustFromString("1.23456789"), right: MustFromString("123")},
		{name: "high-prec", left: MustFromString("123456789.123456789123456789"), right: MustFromString("0.00000000123456789")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchDecimalResult = tt.left.Mul(tt.right, RoundHalfEven)
			}
		})
	}
}

func BenchmarkDecimal_String(b *testing.B) {
	tests := []struct {
		input Decimal
		name  string
	}{
		{name: "int", input: MustFromString("123456789")},
		{name: "same-prec", input: MustFromString("12345.6789")},
		{name: "leading-zero-fraction", input: MustFromString("0.00000000123456789")},
		{name: "negative", input: MustFromString("-123456789.123456789")},
		{name: "trailing-zeros", input: MustFromString("123.450000000000")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchStringResult = tt.input.String()
			}
		})
	}
}

func BenchmarkDecimal_MarshalJSON(b *testing.B) {
	tests := []struct {
		input Decimal
		name  string
	}{
		{name: "int", input: MustFromString("123456789")},
		{name: "same-prec", input: MustFromString("12345.6789")},
		{name: "leading-zero-fraction", input: MustFromString("0.00000000123456789")},
		{name: "negative", input: MustFromString("-123456789.123456789")},
		{name: "trailing-zeros", input: MustFromString("123.450000000000")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bz, err := tt.input.MarshalJSON()
				if err != nil {
					b.Fatalf("MarshalJSON returned error: %v", err)
				}
				benchBytesResult = bz
			}
		})
	}
}

func BenchmarkDecimal_NewFromString(b *testing.B) {
	tests := []struct {
		input string
		name  string
	}{
		{name: "int", input: "123456789"},
		{name: "same-prec", input: "12345.6789"},
		{name: "leading-zero-fraction", input: "0.00000000123456789"},
		{name: "negative", input: "-123456789.123456789"},
		{name: "scientific-positive-exp", input: "1.23456789e9"},
		{name: "scientific-negative-exp", input: "-4.56E-9"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				d, err := NewFromString(tt.input)
				if err != nil {
					b.Fatalf("NewFromString(%q) returned error: %v", tt.input, err)
				}
				benchDecimalResult = d
			}
		})
	}
}

func BenchmarkDecimal_StripTrailingZeros(b *testing.B) {
	tests := []struct {
		input string
	}{
		{input: "1"},
		{input: "1.1"},
		{input: "1.10"},
		{input: "1.100"},
		{input: "1.1000"},
		{input: "1.10000"},
		{input: "1.100000"},
		{input: "1.1000000"},
		{input: "1.10000000"},
		{input: "1.100000000"},
		{input: "1.1000000000"},
		{input: "1.10000000000"},
		{input: "1.100000000000"},
		{input: "1.1000000000000"},
		{input: "1.10000000000000"},
		{input: "1.100000000000000"},
		{input: "1.1000000000000000"},
		{input: "1.10000000000000000"},
		{input: "1.100000000000000000"},
		{input: "1.1000000000000000000"},
		{input: "1.10000000000000000000"},
		{input: "1.100000000000000000000"},
		{input: "1.123456789876000000000000"},
		{input: "1.123456789876543212345670"},
		{input: "1.123456789876543212345600"},
		{input: "1.123456789876543212345000"},
		{input: "1.123456789876543212340000"},
	}
	for _, tt := range tests {
		dec := MustFromString(tt.input)
		b.Run(fmt.Sprintf(`%s/bsearch`, tt.input), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec.StripTrailingZeros()
			}
		})
		b.Run(fmt.Sprintf(`%s/number`, tt.input), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				numberDiv(dec)
			}
		})
		b.Run(fmt.Sprintf(`%s/string`, tt.input), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				convertToString(dec)
			}
		})
	}
}

func BenchmarkDecimal_NewWithAppendPrec(b *testing.B) {
	for _, prec := range []int{0, 6, 18, 36, 128} {
		b.Run(fmt.Sprintf("prec=%d", prec), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = NewWithAppendPrec(123456789, prec)
			}
		})
	}
}

func BenchmarkDecimal_NewFromUintWithAppendPrec(b *testing.B) {
	for _, prec := range []int{0, 6, 18, 36, 128} {
		b.Run(fmt.Sprintf("prec=%d", prec), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = NewFromUintWithAppendPrec(123456789, prec)
			}
		})
	}
}

func BenchmarkDecimal_Quo(b *testing.B) {
	tests := []struct {
		left  Decimal
		right Decimal
		name  string
	}{
		{name: "int/int", left: MustFromString("123456789"), right: MustFromString("37")},
		{name: "same-prec", left: MustFromString("12345.6789"), right: MustFromString("0.37")},
		{name: "mixed-prec", left: MustFromString("1.23456789"), right: MustFromString("123")},
		{name: "high-prec", left: MustFromString("123456789.123456789123456789"), right: MustFromString("0.00000000123456789")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tt.left.Quo(tt.right, RoundHalfEven)
			}
		})
	}
}

func BenchmarkDecimal_QuoWithPrec(b *testing.B) {
	tests := []struct {
		left  Decimal
		right Decimal
		prec  int
		name  string
	}{
		{name: "expand-prec", left: MustFromString("1"), right: MustFromString("8"), prec: 18},
		{name: "same-prec", left: MustFromString("12345.6789"), right: MustFromString("0.37"), prec: 4},
		{name: "shrink-prec", left: MustFromString("12345.6789"), right: MustFromString("0.37"), prec: 2},
		{name: "high-prec", left: MustFromString("123456789.123456789123456789"), right: MustFromString("0.00000000123456789"), prec: 18},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tt.left.QuoWithPrec(tt.right, tt.prec, RoundHalfEven)
			}
		})
	}
}

func convertToString(d Decimal) {
	if d.prec == 0 {
		return
	}
	intPart, fractionPart := d.Remainder()
	if intPart.Sign() < 0 {
		intPart = intPart.Neg(intPart)
	}
	if fractionPart.Sign() < 0 {
		fractionPart = fractionPart.Neg(fractionPart)
	}
	fractionPartStr := strings.TrimRight(fractionPart.String(), "0")
	intPart.Mul(intPart, safeGetPrecisionMultiplier(len(fractionPartStr)))
	if len(fractionPartStr) > 0 {
		fractionPart.SetString(fractionPartStr, 10)
		intPart.Add(intPart, fractionPart)
	}
	if d.IsNegative() {
		intPart = intPart.Neg(intPart)
	}
	d = Decimal{
		i:    intPart,
		prec: len(fractionPartStr),
	}
}

func numberDiv(d Decimal) {
	if d.prec == 0 {
		return
	}

	// Create a copy of the internal big.Int to avoid modifying the original.
	value := new(big.Int).Set(d.i)

	// Remove trailing zeros.
	mod := new(big.Int) // Safe to reuse mod (mod always 0).
	for d.prec > 0 {
		value.DivMod(value, tenInt, mod)
		if mod.Cmp(zeroInt) != 0 {
			value = value.Mul(value, tenInt)
			value = value.Add(value, mod) // Restore the last non-zero digit.
			break
		}
		d.prec--
	}

	d = Decimal{
		i:    value,
		prec: d.prec,
	}
}
