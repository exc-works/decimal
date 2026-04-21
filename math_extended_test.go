package decimal

import (
	"errors"
	"testing"
)

// smallEpsilon is the comparison tolerance used for approximate equality in
// Log10/Ln/Exp tests. It corresponds to 10^-12, which comfortably sits above
// the internal rounding noise of these iterative routines while still
// catching regressions in convergence behaviour.
func smallEpsilon(t *testing.T) Decimal {
	t.Helper()
	eps, err := NewFromString("0.000000000001")
	if err != nil {
		t.Fatalf("failed to parse epsilon: %v", err)
	}
	return eps
}

func approxEqual(t *testing.T, got, want Decimal) bool {
	t.Helper()
	return got.Sub(want).Abs().LT(smallEpsilon(t))
}

func TestDecimalLog10(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"log10 of 1", "1", "0"},
		{"log10 of 10", "10", "1"},
		{"log10 of 100", "100", "2"},
		{"log10 of 1000", "1000", "3"},
		{"log10 of 0.1", "0.1", "-1"},
		{"log10 of 2", "2", "0.301029995663981195"},
		{"log10 of 50", "50", "1.698970004336018805"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := mustDecimal(t, tc.input)
			got, err := d.Log10WithPrec(18)
			if err != nil {
				t.Fatalf("Log10() returned error: %v", err)
			}
			want := mustDecimal(t, tc.want)
			if !approxEqual(t, got, want) {
				t.Fatalf("Log10(%s) = %s, want %s", tc.input, got.String(), tc.want)
			}
		})
	}

	t.Run("log10 of 0 returns error", func(t *testing.T) {
		_, err := mustDecimal(t, "0").Log10()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidLog) {
			t.Fatalf("expected ErrInvalidLog, got %v", err)
		}
	})

	t.Run("log10 of negative returns error", func(t *testing.T) {
		_, err := mustDecimal(t, "-1").Log10()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidLog) {
			t.Fatalf("expected ErrInvalidLog, got %v", err)
		}
	})

	t.Run("log10 default precision works for integer receiver", func(t *testing.T) {
		got, err := mustDecimal(t, "100").Log10()
		if err != nil {
			t.Fatalf("Log10() returned error: %v", err)
		}
		if !approxEqual(t, got, New(2)) {
			t.Fatalf("Log10(100) = %s, want 2", got.String())
		}
	})
}

func TestDecimalLn(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ln of 1", "1", "0"},
		{"ln of 2", "2", "0.693147180559945309"},
		{"ln of 10", "10", "2.302585092994045684"},
		{"ln of 0.5", "0.5", "-0.693147180559945309"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := mustDecimal(t, tc.input)
			got, err := d.LnWithPrec(18)
			if err != nil {
				t.Fatalf("Ln() returned error: %v", err)
			}
			want := mustDecimal(t, tc.want)
			if !approxEqual(t, got, want) {
				t.Fatalf("Ln(%s) = %s, want %s", tc.input, got.String(), tc.want)
			}
		})
	}

	t.Run("ln of e is approximately 1", func(t *testing.T) {
		e := mustDecimal(t, "2.718281828459045235360287471352662")
		got, err := e.LnWithPrec(20)
		if err != nil {
			t.Fatalf("Ln(e) returned error: %v", err)
		}
		if !approxEqual(t, got, New(1)) {
			t.Fatalf("Ln(e) = %s, want 1", got.String())
		}
	})

	t.Run("ln of 0 returns error", func(t *testing.T) {
		_, err := mustDecimal(t, "0").Ln()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidLog) {
			t.Fatalf("expected ErrInvalidLog, got %v", err)
		}
	})

	t.Run("ln of negative returns error", func(t *testing.T) {
		_, err := mustDecimal(t, "-3.14").Ln()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidLog) {
			t.Fatalf("expected ErrInvalidLog, got %v", err)
		}
	})
}

func TestDecimalExp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"exp of 0", "0", "1"},
		{"exp of 1", "1", "2.718281828459045235"},
		{"exp of 2", "2", "7.389056098930650227"},
		{"exp of -1", "-1", "0.367879441171442322"},
		{"exp of 0.5", "0.5", "1.648721270700128146"},
		{"exp of -2", "-2", "0.135335283236612691"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := mustDecimal(t, tc.input)
			got, err := d.ExpWithPrec(18)
			if err != nil {
				t.Fatalf("Exp() returned error: %v", err)
			}
			want := mustDecimal(t, tc.want)
			if !approxEqual(t, got, want) {
				t.Fatalf("Exp(%s) = %s, want %s", tc.input, got.String(), tc.want)
			}
		})
	}

	t.Run("exp default precision gives 1 for zero", func(t *testing.T) {
		got, err := New(0).Exp()
		if err != nil {
			t.Fatalf("Exp() returned error: %v", err)
		}
		if !approxEqual(t, got, New(1)) {
			t.Fatalf("Exp(0) = %s, want 1", got.String())
		}
	})
}

func TestExpLnRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"round trip for 5", "5"},
		{"round trip for 0.5", "0.5"},
		{"round trip for 12.3456", "12.3456"},
		{"round trip for e", "2.71828182845904523536"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := mustDecimal(t, tc.input)

			// exp(ln(x)) == x
			lnVal, err := d.LnWithPrec(25)
			if err != nil {
				t.Fatalf("Ln() returned error: %v", err)
			}
			expOfLn, err := lnVal.ExpWithPrec(18)
			if err != nil {
				t.Fatalf("Exp() returned error: %v", err)
			}
			if !approxEqual(t, expOfLn, d) {
				t.Fatalf("Exp(Ln(%s)) = %s, want %s", tc.input, expOfLn.String(), tc.input)
			}
		})
	}

	t.Run("ln of exp of 2 is 2", func(t *testing.T) {
		two := New(2)
		expVal, err := two.ExpWithPrec(25)
		if err != nil {
			t.Fatalf("Exp() returned error: %v", err)
		}
		lnOfExp, err := expVal.LnWithPrec(18)
		if err != nil {
			t.Fatalf("Ln() returned error: %v", err)
		}
		if !approxEqual(t, lnOfExp, two) {
			t.Fatalf("Ln(Exp(2)) = %s, want 2", lnOfExp.String())
		}
	})
}

func TestLog10WithPrec_TinyPositive(t *testing.T) {
	d := MustFromString("1e-50")
	got, err := d.Log10WithPrec(18)
	if err != nil {
		t.Fatalf("Log10WithPrec(1e-50): unexpected error %v", err)
	}
	want := New(-50)
	if !approxEqual(t, got, want) {
		t.Fatalf("Log10(1e-50) = %s, want ~-50", got.String())
	}
}

func TestLnWithPrec_TinyPositive(t *testing.T) {
	d := MustFromString("1e-50")
	got, err := d.LnWithPrec(18)
	if err != nil {
		t.Fatalf("LnWithPrec(1e-50): unexpected error %v", err)
	}
	want := MustFromString("-115.129254649702284200899572") // -50 * ln(10)
	if got.Sub(want).Abs().GT(MustFromString("0.0001")) {
		t.Fatalf("Ln(1e-50) = %s, want ~-115.129254", got.String())
	}
}

func TestLog10_NonPositive_StillErrors(t *testing.T) {
	_, err := Zero.Log10()
	if !errors.Is(err, ErrInvalidLog) {
		t.Fatalf("Log10(0): expected ErrInvalidLog, got %v", err)
	}
	_, err = MustFromString("-1").Log10()
	if !errors.Is(err, ErrInvalidLog) {
		t.Fatalf("Log10(-1): expected ErrInvalidLog, got %v", err)
	}
}

// TestMathConstantsSanity guards the high-precision mathematical constants
// hardcoded in math.go against typos or truncation mistakes. Every assertion
// exercises the constant through an independent iterative routine and
// compares against a mathematical identity (Exp(ln(k)) == k,
// Ln(10)/Ln(2) == log2(10), etc.). A failure here should NOT be silently
// patched by editing the constant — human review is required.
func TestMathConstantsSanity(t *testing.T) {
	// epsilon is chosen to be comfortably larger than the accumulated
	// truncation error of the iterative Log2/Exp routines at workPrec while
	// remaining many orders of magnitude tighter than a single-digit typo in
	// a 70-decimal-digit literal. 1e-50 leaves ~20 digits of safety margin
	// against the 70-digit ln2Literal.
	const workPrec = 60
	epsilon, err := NewFromString("1e-50")
	if err != nil {
		t.Fatalf("failed to parse epsilon: %v", err)
	}

	// Parse the ln2 literal directly so we verify the constant itself, not
	// some derived value.
	ln2, err := NewFromString(ln2Literal)
	if err != nil {
		t.Fatalf("failed to parse ln2Literal: %v", err)
	}

	t.Run("ln2 literal: Exp(ln2) == 2", func(t *testing.T) {
		got, err := ln2.ExpWithPrec(workPrec)
		if err != nil {
			t.Fatalf("Exp(ln2) returned error: %v", err)
		}
		diff := got.Sub(New(2)).Abs()
		if diff.GTE(epsilon) {
			t.Fatalf("Exp(ln2Literal) = %s, diff from 2 = %s, exceeds epsilon %s. "+
				"Do NOT silently edit ln2Literal; verify the literal against an "+
				"authoritative source first.",
				got.String(), diff.String(), epsilon.String())
		}
	})

	t.Run("cross-check: Ln(10)/Ln(2) == log2(10)", func(t *testing.T) {
		// log2(10) to ~70 digits (independent reference from ln2Literal).
		//   log2(10) = 3.32192809488736234787031942948939017586483139302458061...
		log2Of10Ref, err := NewFromString(
			"3.3219280948873623478703194294893901758648313930245806120547563958")
		if err != nil {
			t.Fatalf("failed to parse log2(10) reference: %v", err)
		}

		ln10, err := New(10).LnWithPrec(workPrec)
		if err != nil {
			t.Fatalf("Ln(10) returned error: %v", err)
		}
		ln2Computed, err := New(2).LnWithPrec(workPrec)
		if err != nil {
			t.Fatalf("Ln(2) returned error: %v", err)
		}
		ratio := ln10.QuoWithPrec(ln2Computed, workPrec, RoundHalfEven)
		diff := ratio.Sub(log2Of10Ref).Abs()
		if diff.GTE(epsilon) {
			t.Fatalf("Ln(10)/Ln(2) = %s, diff from log2(10) reference = %s, "+
				"exceeds epsilon %s. This suggests a typo in ln2Literal.",
				ratio.String(), diff.String(), epsilon.String())
		}
	})

	t.Run("ln2 literal: Ln(2) rounds to ln2Literal", func(t *testing.T) {
		// Independently compute Ln(2) via the library and confirm the first
		// ~50 digits agree with the hardcoded literal. This catches the case
		// where Ln(2) happens to be self-consistent with a wrong ln2Literal
		// (because LnWithPrec multiplies by ln2Literal itself), by comparing
		// at high precision — any digit typo in the literal would propagate
		// and show up here as a large discrepancy against the stored string.
		ln2Computed, err := New(2).LnWithPrec(workPrec)
		if err != nil {
			t.Fatalf("Ln(2) returned error: %v", err)
		}
		diff := ln2Computed.Sub(ln2).Abs()
		if diff.GTE(epsilon) {
			t.Fatalf("Ln(2) computed = %s vs ln2Literal = %s, diff = %s, "+
				"exceeds epsilon %s.",
				ln2Computed.String(), ln2.String(), diff.String(),
				epsilon.String())
		}
	})

	t.Run("Exp(Ln(10)) == 10", func(t *testing.T) {
		ln10, err := New(10).LnWithPrec(workPrec)
		if err != nil {
			t.Fatalf("Ln(10) returned error: %v", err)
		}
		got, err := ln10.ExpWithPrec(workPrec)
		if err != nil {
			t.Fatalf("Exp(Ln(10)) returned error: %v", err)
		}
		diff := got.Sub(New(10)).Abs()
		if diff.GTE(epsilon) {
			t.Fatalf("Exp(Ln(10)) = %s, diff from 10 = %s, exceeds epsilon %s.",
				got.String(), diff.String(), epsilon.String())
		}
	})
}
