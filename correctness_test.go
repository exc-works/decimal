package decimal

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

func TestPowerNegativeWithoutIntermediateRounding(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		power       int64
	}{
		{"0.1", "100", -2}, {"0.2", "25", -2}, {"0.3", "11.2", -2},
		{"-0.3", "-37.1", -3}, {"1.5", "0.5", -2},
		{"1.00", "1", math.MinInt64}, {"-1.0", "1", math.MinInt64},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got := MustFromString(tc.input).Power(tc.power)
			if !got.Equal(MustFromString(tc.want)) {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestRootExactHalfEvenBoundaries(t *testing.T) {
	for root := int64(2); root <= 9; root++ {
		for prec := 0; prec <= 3; prec++ {
			for k := int64(0); k < 20; k++ {
				midpoint := NewWithPrec(2*k+1, prec).QuoWithPrec(New(2), prec+1, RoundUnnecessary)
				x := One
				for i := int64(0); i < root; i++ {
					x = x.MulExact(midpoint)
				}
				want := NewWithPrec(k+k%2, prec)
				for _, sign := range []int64{1, -1} {
					if sign < 0 && root%2 == 0 {
						continue
					}
					input := x.MulExact(New(sign))
					got, err := input.ApproxRootWithPrec(root, prec)
					if err != nil || !got.Equal(want.MulExact(New(sign))) {
						t.Fatalf("root(%s, %d), prec %d: got %s, err %v, want %s", input, root, prec, got, err, want.MulExact(New(sign)))
					}
				}
			}
		}
	}
}

func TestYAMLScalarPrecision(t *testing.T) {
	for _, input := range []string{"1.23456789012345678901234567890", "9007199254740993.0", "1e1000", "123456789012345678901234567890", "!!float 1.23456789012345678901234567890", "1.23456789012345678901234567890 # comment"} {
		t.Run(input, func(t *testing.T) {
			var value Decimal
			if err := yaml.Unmarshal([]byte(input), &value); err != nil {
				t.Fatal(err)
			}
			var nullable NullDecimal
			if err := yaml.Unmarshal([]byte(input), &nullable); err != nil {
				t.Fatal(err)
			}
			want := input
			want = strings.TrimPrefix(want, "!!float ")
			want, _, _ = strings.Cut(want, " #")
			if !value.Equal(MustFromString(want)) || !nullable.Valid || !nullable.Decimal.Equal(value) {
				t.Fatalf("input %q: Decimal=%s NullDecimal=%s", input, value, nullable)
			}
		})
	}
}

func TestBinaryPrecisionBounds(t *testing.T) {
	if _, err := NewWithPrec(0, int(^uint(0)>>1)).MarshalBinary(); err != nil {
		t.Fatal(err)
	}
	for _, prec := range []int{maxParsedPrecision + 1, math.MaxInt32} {
		if _, err := NewWithPrec(1, prec).MarshalBinary(); !errors.Is(err, ErrInvalidPrecision) {
			t.Fatalf("precision %d: got %v, want ErrInvalidPrecision", prec, err)
		}
	}
	if big.NewInt(int64(^uint(0)>>1)).BitLen() > 32 {
		prec := uint64(1) << 32
		if _, err := NewWithPrec(1, int(prec)).MarshalBinary(); !errors.Is(err, ErrInvalidPrecision) {
			t.Fatalf("overflowing precision: %v", err)
		}
	}
	input := NewWithPrec(1, maxParsedPrecision)
	data, err := input.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var output Decimal
	if err := output.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}
	if output.Precision() != input.Precision() || output.BigInt().Cmp(input.BigInt()) != 0 {
		t.Fatal("binary round trip changed the value")
	}
}

// Exercise real decoder interfaces rather than assuming an *any callback target.
func yamlTestDecoder(value any, decodeErr error) func(any) error {
	data, err := yaml.Marshal(value)
	return func(target any) error {
		if decodeErr != nil {
			return decodeErr
		}
		if err != nil {
			return err
		}
		return yaml.Unmarshal(data, target)
	}
}
