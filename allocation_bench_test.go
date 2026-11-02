package decimal

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

var (
	allocationBool    bool
	allocationInt     int64
	allocationUint    uint64
	allocationBigInt  *big.Int
	allocationFloat64 float64
	allocationFloat32 float32
)

func BenchmarkIntegerInspection(b *testing.B) {
	inputs := []struct {
		name  string
		value Decimal
	}{
		{"integer", New(123456789)},
		{"scaled-integer", MustFromString("123456789.0000")},
		{"fraction", MustFromString("-12345.6789")},
		{"wide-integer", MustFromString(strings.Repeat("1234567890", 20))},
		{"tiny", NewWithPrec(1, 1000)},
		{"scaled-zero", NewWithPrec(0, 1000)},
	}
	for _, input := range inputs {
		for _, op := range []string{"IsInteger", "IntPart", "Int64", "Uint64"} {
			b.Run(input.name+"/"+op, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					switch op {
					case "IsInteger":
						allocationBool = input.value.IsInteger()
					case "IntPart":
						allocationBigInt = input.value.IntPart()
					case "Int64":
						allocationInt, allocationBool = input.value.Int64()
					case "Uint64":
						allocationUint, allocationBool = input.value.Uint64()
					}
				}
			})
		}
	}
}

func BenchmarkNativeFloatConversion(b *testing.B) {
	inputs := []struct {
		name  string
		value Decimal
	}{
		{"integer", New(123456789)},
		{"wide-integer", MustFromString(strings.Repeat("1234567890", 20))},
		{"binary-midpoint", NewFromBigInt(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 80), new(big.Int).Lsh(big.NewInt(1), 56)))},
		{"fraction", MustFromString("12345.6789")},
		{"scaled-zero", NewWithPrec(0, 1000)},
	}
	for _, input := range inputs {
		for _, bits := range []int{32, 64} {
			b.Run(fmt.Sprintf("%s/float%d", input.name, bits), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if bits == 32 {
						allocationFloat32, allocationBool = input.value.Float32()
					} else {
						allocationFloat64, allocationBool = input.value.Float64()
					}
				}
			})
		}
	}
}

func BenchmarkRationalConstruction(b *testing.B) {
	large, _ := new(big.Int).SetString(strings.Repeat("1234567890", 20), 10)
	inputs := []struct {
		name  string
		value *big.Rat
	}{
		{"integer", big.NewRat(123456789, 1)},
		{"wide-integer", new(big.Rat).SetInt(large)},
		{"terminating", big.NewRat(-123456789, 10000)},
		{"repeating", big.NewRat(1, 3)},
		{"zero", new(big.Rat)},
	}
	for _, input := range inputs {
		if input.name != "repeating" {
			b.Run(input.name+"/exact", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var err error
					benchDecimalResult, err = NewFromBigRat(input.value)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
		b.Run(input.name+"/rounded", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var err error
				benchDecimalResult, err = NewFromBigRatWithPrec(input.value, 20, RoundHalfEven)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkIntegerFormatting(b *testing.B) {
	for _, input := range []struct {
		name  string
		value Decimal
	}{{"integer", New(123456789)}, {"scaled-integer", MustFromString("123456789.0000")}} {
		b.Run(input.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchStringResult = fmt.Sprintf("%d", input.value)
			}
		})
	}
}
