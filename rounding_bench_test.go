package decimal

import "testing"

func BenchmarkRescaleRounding(b *testing.B) {
	inputs := []struct {
		name  string
		value Decimal
	}{
		{"positive-tie", MustFromString("12345.685")},
		{"negative-tie", MustFromString("-12345.685")},
		{"negative-exact", MustFromString("-12345.680")},
	}
	for _, input := range inputs {
		for _, mode := range []struct {
			name  string
			value RoundingMode
		}{{"half-even", RoundHalfEven}, {"up", RoundUp}, {"down", RoundDown}} {
			b.Run(input.name+"/"+mode.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					benchDecimalResult = input.value.Rescale(2, mode.value)
				}
			})
		}
	}
}

func BenchmarkSignificantFiguresRounding(b *testing.B) {
	for _, input := range []struct {
		name  string
		value Decimal
	}{{"positive", MustFromString("123456.7")}, {"negative", MustFromString("-123456.7")}} {
		b.Run(input.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchDecimalResult = input.value.SignificantFigures(3, RoundHalfEven)
			}
		})
	}
}
