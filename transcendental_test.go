package decimal

import (
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"testing"
)

func TestTranscendentalIndependentReferences(t *testing.T) {
	// Fixtures use Python's Decimal at 1500 significant digits, independently
	// of the fixed-point interval algorithms. Include 1000-place requests,
	// large outputs, and inputs within 1e-100 of rounding transitions.
	data, err := os.ReadFile("testdata/transcendental.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Op, Input, Want string
		Prec            int
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Op+"/"+tc.Input+"/"+big.NewInt(int64(tc.Prec)).String(), func(t *testing.T) {
			input := MustFromString(tc.Input)
			var got Decimal
			var err error
			switch tc.Op {
			case "ln":
				got, err = input.LnWithPrec(tc.Prec)
			case "log2":
				got, err = input.Log2WithPrec(tc.Prec)
			case "log10":
				got, err = input.Log10WithPrec(tc.Prec)
			case "exp":
				got, err = input.ExpWithPrec(tc.Prec)
			default:
				t.Fatalf("unknown operation %q", tc.Op)
			}
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(MustFromString(tc.Want)) || got.Precision() != tc.Prec {
				t.Fatalf("got %s (scale %d), want %s (scale %d)", got, got.Precision(), tc.Want, tc.Prec)
			}
		})
	}
}

func TestTranscendentalBudgets(t *testing.T) {
	for _, operation := range []func(int) (Decimal, error){One.LnWithPrec, One.Log2WithPrec, One.Log10WithPrec} {
		got, err := operation(maxParsedPrecision)
		if err != nil || !got.IsZero() || got.Precision() != maxParsedPrecision {
			t.Fatalf("exact zero at precision boundary: %v", err)
		}
	}
	for _, operation := range []func(int) (Decimal, error){One.LnWithPrec, One.Log2WithPrec, One.Log10WithPrec, One.ExpWithPrec} {
		_, err := operation(maxParsedPrecision + 1)
		if !errors.Is(err, ErrNoConvergence) {
			t.Fatalf("excess precision: got %v", err)
		}
	}
	if _, err := New(maxParsedPrecision + 1).ExpWithPrec(0); !errors.Is(err, ErrNoConvergence) {
		t.Fatalf("oversized exponential: %v", err)
	}
	// Far negative arguments must certify zero without constructing e^|x|.
	got, err := MustFromString("-1e1000").ExpWithPrec(1000)
	if err != nil || !got.IsZero() || got.Precision() != 1000 {
		t.Fatalf("underflow: %s, %v", got, err)
	}
	scale := new(big.Int).Lsh(big.NewInt(1), 256)
	if _, err := logMantissaBounds(big.NewInt(2), oneInt, scale, 1); !errors.Is(err, ErrNoConvergence) {
		t.Fatalf("log series budget: %v", err)
	}
	if _, err := expSeriesBounds(fractionBounds(oneInt, big.NewInt(2), scale), scale, 1); !errors.Is(err, ErrNoConvergence) {
		t.Fatalf("exp series budget: %v", err)
	}
}

func TestLog2CorrectRounding(t *testing.T) {
	for _, tc := range []struct{ input, want string }{{"1.9", "0.9"}, {"0.3", "-1.7"}, {"3.000000000000000000", "1.584962500721156181"}} {
		if got := MustFromString(tc.input).Log2(); !got.Equal(MustFromString(tc.want)) {
			t.Fatalf("Log2(%s)=%s, want %s", tc.input, got, tc.want)
		}
	}
	if _, err := Zero.Log2WithPrec(18); !errors.Is(err, ErrInvalidLog) {
		t.Fatalf("domain error: %v", err)
	}
}
