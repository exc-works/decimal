package decimal

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

// ln2Literal is a 70-place reference for tests. Runtime logarithms compute
// ln(2) at their working precision instead of using this finite literal.
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

// Log10 returns log10(d) rounded with RoundHalfEven at max(d.Precision(), 30).
// It returns ErrInvalidLog for non-positive inputs or ErrNoConvergence if
// rounding cannot be certified within the calculation budget.
func (d Decimal) Log10() (Decimal, error) { return d.Log10WithPrec(workingLogPrec(d)) }

// Log10WithPrec returns log10(d) correctly rounded to prec decimal places.
// It returns ErrInvalidLog for non-positive inputs or ErrNoConvergence if
// rounding cannot be certified within the calculation budget. It panics if
// prec is negative.
func (d Decimal) Log10WithPrec(prec int) (Decimal, error) { return certifiedLog(d, prec, 10) }

// Ln returns ln(d) rounded with RoundHalfEven at max(d.Precision(), 30).
// Its error behavior is the same as LnWithPrec.
func (d Decimal) Ln() (Decimal, error) { return d.LnWithPrec(workingLogPrec(d)) }

// LnWithPrec returns ln(d) correctly rounded to prec decimal places using
// RoundHalfEven. It computes constants at the working precision and certifies
// the final rounding with outward-rounded intervals. It returns ErrInvalidLog
// for non-positive inputs or ErrNoConvergence if the calculation budget is
// exhausted. It panics if prec is negative.
func (d Decimal) LnWithPrec(prec int) (Decimal, error) { return certifiedLog(d, prec, 0) }

// Exp returns e^d rounded with RoundHalfEven at max(d.Precision(), 30).
// Its error behavior is the same as ExpWithPrec.
func (d Decimal) Exp() (Decimal, error) { return d.ExpWithPrec(workingLogPrec(d)) }

// ExpWithPrec returns e^d correctly rounded to prec decimal places using
// RoundHalfEven. It bounds Taylor truncation and all argument-reduction errors,
// and returns ErrNoConvergence if rounding cannot be certified within the
// calculation budget. It panics if prec is negative.
func (d Decimal) ExpWithPrec(prec int) (Decimal, error) { return certifiedExp(d, prec) }
