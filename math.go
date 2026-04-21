package decimal

import (
	"fmt"
	"math/big"
)

// Max returns the greater of a and b.
func Max(a, b Decimal) Decimal {
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

// Min returns the smaller of a and b.
func Min(a, b Decimal) Decimal {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}

// Between reports whether v is within the inclusive range [lower, upper].
func Between(v, lower, upper Decimal) bool {
	return v.Cmp(lower) >= 0 && v.Cmp(upper) <= 0
}

// defaultLogExpPrec is the default working precision used by Log10, Ln, and
// Exp when the receiver's precision is too small to hold a meaningful
// approximation.
const defaultLogExpPrec = 30

// ln2Literal is the natural logarithm of 2 to 70 decimal places. Verified by
// TestMathConstantsSanity. Source: OEIS A002162 / standard mathematical
// references.
//
//	ln 2 = 0.69314718055994530941723212145817656807550013436025525412068000949...
const ln2Literal = "0.6931471805599453094172321214581765680755001343602552541206800094933936"

// workingLogPrec returns a precision suitable for intermediate log/exp
// computations. It uses the receiver's precision but never falls below
// defaultLogExpPrec so that Log10/Ln/Exp remain meaningful for integer
// receivers (prec == 0).
func workingLogPrec(d Decimal) int {
	if d.prec < defaultLogExpPrec {
		return defaultLogExpPrec
	}
	return d.prec
}

// Log10 returns an approximate base-10 logarithm of d.
//
// It returns an error wrapping ErrInvalidLog when d is not strictly positive.
// Internally Log10 is computed as Log2(d) / Log2(10) at the receiver's
// precision (bumped to defaultLogExpPrec when the receiver has fewer
// fractional digits). The result is rescaled to the receiver's precision so
// that callers receive output in a familiar scale.
func (d Decimal) Log10() (Decimal, error) {
	d = initializeIfNeeded(d)
	if d.Sign() <= 0 {
		return Decimal{}, fmt.Errorf("Log10(%s): %w", d.String(), ErrInvalidLog)
	}
	return d.Log10WithPrec(workingLogPrec(d))
}

// Log10WithPrec returns an approximate base-10 logarithm of d rescaled to
// prec decimal places using RoundHalfEven.
//
// It returns an error wrapping ErrInvalidLog when d is not strictly positive.
// It panics if prec is negative.
func (d Decimal) Log10WithPrec(prec int) (Decimal, error) {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	if d.Sign() <= 0 {
		return Decimal{}, fmt.Errorf("Log10(%s): %w", d.String(), ErrInvalidLog)
	}

	// Work at a precision that can absorb rounding from two Log2 calls and a
	// division. A small buffer keeps the final RoundHalfEven clean.
	work := prec + 6
	if work < defaultLogExpPrec {
		work = defaultLogExpPrec
	}

	dScaled := d
	if d.prec < work {
		dScaled = d.Rescale(work, RoundHalfEven)
	}
	logD := dScaled.Log2()

	logTen := NewWithAppendPrec(10, work).Log2()
	result := logD.QuoWithPrec(logTen, work, RoundHalfEven)
	return result.Rescale(prec, RoundHalfEven), nil
}

// Ln returns an approximate natural logarithm of d.
//
// It returns an error wrapping ErrInvalidLog when d is not strictly positive.
// Internally Ln is computed as Log2(d) * ln(2) using a pre-computed
// high-precision constant for ln(2).
func (d Decimal) Ln() (Decimal, error) {
	d = initializeIfNeeded(d)
	if d.Sign() <= 0 {
		return Decimal{}, fmt.Errorf("Ln(%s): %w", d.String(), ErrInvalidLog)
	}
	return d.LnWithPrec(workingLogPrec(d))
}

// LnWithPrec returns an approximate natural logarithm of d rescaled to prec
// decimal places using RoundHalfEven.
//
// It returns an error wrapping ErrInvalidLog when d is not strictly positive.
// It panics if prec is negative.
func (d Decimal) LnWithPrec(prec int) (Decimal, error) {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	if d.Sign() <= 0 {
		return Decimal{}, fmt.Errorf("Ln(%s): %w", d.String(), ErrInvalidLog)
	}

	work := prec + 6
	if work < defaultLogExpPrec {
		work = defaultLogExpPrec
	}

	ln2, err := NewFromString(ln2Literal)
	if err != nil {
		return Decimal{}, fmt.Errorf("failed to parse ln2 constant: %w", err)
	}

	dScaled := d
	if d.prec < work {
		dScaled = d.Rescale(work, RoundHalfEven)
	}
	logD := dScaled.Log2()

	product := logD.Mul(ln2, RoundHalfEven)
	return product.Rescale(prec, RoundHalfEven), nil
}

// Exp returns an approximate value of e raised to the power of d.
//
// It evaluates the Taylor series e^x = sum_{n>=0} x^n / n! with argument
// reduction: e^x = (e^(x/2^k))^(2^k). The reduction shrinks |x| below 0.5,
// which guarantees rapid convergence. The iteration is bounded by
// maxIterations.
func (d Decimal) Exp() (Decimal, error) {
	d = initializeIfNeeded(d)
	return d.ExpWithPrec(workingLogPrec(d))
}

// ExpWithPrec returns e^d rescaled to prec decimal places using
// RoundHalfEven.
//
// It panics if prec is negative.
func (d Decimal) ExpWithPrec(prec int) (Decimal, error) {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)

	work := prec + 10
	if work < defaultLogExpPrec {
		work = defaultLogExpPrec
	}

	if d.IsZero() {
		return NewWithAppendPrec(1, prec), nil
	}

	// Handle negative argument via reciprocal: e^-x = 1 / e^x.
	if d.IsNegative() {
		pos, err := d.Neg().ExpWithPrec(work)
		if err != nil {
			return Decimal{}, err
		}
		one := NewWithAppendPrec(1, work)
		inv := one.QuoWithPrec(pos, work, RoundHalfEven)
		return inv.Rescale(prec, RoundHalfEven), nil
	}

	x := d.Rescale(work, RoundHalfEven)

	// Argument reduction: divide x by 2^k until |x| <= 0.5 so that the Taylor
	// series converges quickly. k is bounded to prevent excessive growth when
	// squaring the result back up.
	half := NewFromBigIntWithPrec(big.NewInt(5), 1).Rescale(work, RoundHalfEven)
	k := 0
	const maxReductions = 60
	two := NewWithAppendPrec(2, work)
	for x.Cmp(half) > 0 && k < maxReductions {
		x = x.QuoWithPrec(two, work, RoundHalfEven)
		k++
	}

	// Taylor series: result = 1 + x + x^2/2! + x^3/3! + ...
	one := NewWithAppendPrec(1, work)
	result := one
	term := one

	// Epsilon at which further terms no longer affect the working precision.
	epsilon := Decimal{
		i:    big.NewInt(1),
		prec: work,
	}

	for n := 1; n < maxIterations; n++ {
		// term_n = term_{n-1} * x / n
		term = term.Mul(x, RoundHalfEven)
		term = term.QuoWithPrec(NewWithAppendPrec(int64(n), work), work, RoundHalfEven)
		result = result.Add(term)
		if term.Abs().LT(epsilon) {
			break
		}
	}

	// Undo the argument reduction by squaring k times.
	for i := 0; i < k; i++ {
		result = result.Mul(result, RoundHalfEven)
	}

	return result.Rescale(prec, RoundHalfEven), nil
}
