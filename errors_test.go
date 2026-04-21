package decimal

import (
	"encoding/json"
	"errors"
	"math/big"
	"testing"
)

func TestErrInvalidFormat_NewFromString(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"not a number", "not a number"},
		{"multiple dots", "1.2.3"},
		{"trailing dot", "1."},
		{"leading dot", ".1"},
		{"bare minus", "-"},
		{"dangling exponent", "1e"},
		{"missing exponent value", "1e+"},
		{"non-numeric exponent", "1eabc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewFromString(tc.in)
			if err == nil {
				t.Fatalf("NewFromString(%q) expected error, got nil", tc.in)
			}
			if !errors.Is(err, ErrInvalidFormat) {
				t.Fatalf("NewFromString(%q): expected ErrInvalidFormat, got %v", tc.in, err)
			}
		})
	}
}

func TestErrInvalidFormat_ExponentOverflow(t *testing.T) {
	_, err := NewFromString("1e99999999999999999999")
	if err == nil {
		t.Fatal("expected error for exponent overflow")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestErrInvalidFormat_NewFromBigRat(t *testing.T) {
	_, err := NewFromBigRat(nil)
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat for nil big.Rat, got %v", err)
	}

	// 1/3 is non-terminating in base 10.
	_, err = NewFromBigRat(big.NewRat(1, 3))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat for non-terminating rat, got %v", err)
	}
}

func TestErrInvalidFormat_NewFromBigRatWithPrec(t *testing.T) {
	_, err := NewFromBigRatWithPrec(nil, 2, RoundHalfEven)
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat for nil big.Rat, got %v", err)
	}
}

func TestErrNegativeRoot_Sqrt(t *testing.T) {
	d := New(-4)
	_, err := d.Sqrt()
	if err == nil {
		t.Fatal("expected error for Sqrt of negative")
	}
	if !errors.Is(err, ErrNegativeRoot) {
		t.Fatalf("expected ErrNegativeRoot, got %v", err)
	}
}

func TestErrNegativeRoot_ApproxRootEven(t *testing.T) {
	d := New(-16)
	_, err := d.ApproxRoot(4)
	if err == nil {
		t.Fatal("expected error for even root of negative")
	}
	if !errors.Is(err, ErrNegativeRoot) {
		t.Fatalf("expected ErrNegativeRoot, got %v", err)
	}
}

func TestErrInvalidRoot(t *testing.T) {
	_, err := New(16).ApproxRoot(0)
	if !errors.Is(err, ErrInvalidRoot) {
		t.Fatalf("expected ErrInvalidRoot for zero root, got %v", err)
	}
	_, err = New(16).ApproxRoot(-2)
	if !errors.Is(err, ErrInvalidRoot) {
		t.Fatalf("expected ErrInvalidRoot for negative root, got %v", err)
	}
}

func TestErrUnmarshal_Binary(t *testing.T) {
	var d Decimal
	// Fewer than PrecisionFixedSize bytes triggers the wrapped error.
	err := d.UnmarshalBinary([]byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for short binary data")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
}

func TestErrInvalidFormat_UnmarshalText(t *testing.T) {
	var d Decimal
	err := d.UnmarshalText([]byte("not a number"))
	if err == nil {
		t.Fatal("expected error")
	}
	// UnmarshalText forwards NewFromString's error, which wraps ErrInvalidFormat.
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestErrInvalidFormat_UnmarshalJSON(t *testing.T) {
	var d Decimal
	err := json.Unmarshal([]byte(`"not a number"`), &d)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestSentinelsDistinct(t *testing.T) {
	all := []error{
		ErrOverflow,
		ErrDivideByZero,
		ErrInvalidPrecision,
		ErrInvalidFormat,
		ErrNegativeRoot,
		ErrInvalidRoot,
		ErrInvalidLog,
		ErrRoundUnnecessary,
		ErrUnmarshal,
		ErrInvalidArgument,
	}
	for i, a := range all {
		if a == nil {
			t.Fatalf("sentinel at index %d is nil", i)
		}
		for j, b := range all {
			if i == j {
				continue
			}
			if errors.Is(a, b) {
				t.Fatalf("sentinels should be distinct: %v Is %v", a, b)
			}
		}
	}
}

func TestErrInvalidArgument_Validator(t *testing.T) {
	err := RegisterGoPlaygroundValidator(nil)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("RegisterGoPlaygroundValidator(nil): expected ErrInvalidArgument, got %v", err)
	}
}

func TestErrUnmarshal_Scan_UnsupportedType(t *testing.T) {
	var d Decimal
	err := d.Scan(struct{}{})
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("Scan(struct{}{}): expected ErrUnmarshal, got %v", err)
	}
}
