package decimal

import (
	"fmt"
	"math/big"
)

// certifiedExp reduces |x| to y=|x|/2^k <= 1/2, encloses exp(y), and restores
// the argument by k squarings (plus a reciprocal for negative x). Every step
// propagates interval bounds; the Ziv driver certifies the final rounding.
func certifiedExp(d Decimal, prec int) (Decimal, error) {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	if prec > maxParsedPrecision || d.prec > maxParsedPrecision {
		return Decimal{}, fmt.Errorf("exponential precision budget exceeded: %w", ErrNoConvergence)
	}
	if d.IsZero() {
		return NewWithAppendPrec(1, prec), nil
	}
	negative := d.IsNegative()
	r := d.BigRat()
	r.Abs(r)
	// The first five positive Taylor terms give exp(5/2) > 10. Hence
	// |x| >= (5/2)*(prec+1) implies exp(-|x|) < 10^-(prec+1), strictly
	// below half a decimal unit. Certify zero without constructing e^|x|.
	if negative && r.Cmp(new(big.Rat).SetFrac64(int64(5*(prec+1)), 2)) >= 0 {
		return NewWithPrec(0, prec), nil
	}
	// e < 10 implies e^|x| < 10^ceil(|x|). Reserve these integer digits as
	// additional working precision; final acceptance still requires enclosure
	// certification, including every argument-reduction squaring error.
	var digits big.Int
	ceilQuotientTo(&digits, r.Num(), r.Denom())
	if !digits.IsInt64() || digits.Int64() > maxParsedPrecision {
		return Decimal{}, fmt.Errorf("exponential result exceeds precision budget: %w", ErrNoConvergence)
	}
	reducedDen := new(big.Int).Set(r.Denom())
	twiceNum := new(big.Int).Lsh(r.Num(), 1)
	reductions := 0
	for twiceNum.Cmp(reducedDen) > 0 {
		reducedDen.Lsh(reducedDen, 1)
		reductions++
	}
	return certifyDecimal(prec, int(digits.Int64()), func(scale *big.Int, bits int) (fixedInterval, error) {
		x := fractionBounds(r.Num(), reducedDen, scale)
		value, err := expSeriesBounds(x, scale, bits)
		if err != nil {
			return fixedInterval{}, err
		}
		if reductions > 0 {
			// Alternating owned buffers keep the squaring inputs intact and
			// reuse their storage across all argument-restoration steps.
			scratch := fixedInterval{new(big.Int), new(big.Int)}
			for range reductions {
				multiplyPositiveBoundsTo(scratch, value, value, scale)
				value, scratch = scratch, value
			}
		}
		if negative {
			value = divideBounds(fixedInterval{scale, scale}, value, scale)
		}
		return value, nil
	})
}

// expSeriesBounds sums nonnegative terms for 0 <= x <= 1/2. The tail after
// term n is <= 2*term(n+1), because subsequent term ratios are <= 1/4.
func expSeriesBounds(x fixedInterval, scale *big.Int, maxTerms int) (fixedInterval, error) {
	term := fixedInterval{new(big.Int).Set(scale), new(big.Int).Set(scale)}
	product := fixedInterval{new(big.Int), new(big.Int)}
	lo, hi := new(big.Int).Set(scale), new(big.Int).Set(scale)
	var divisor, nextUpper, two big.Int
	two.SetInt64(2)
	for n := 1; n <= maxTerms; n++ {
		multiplyPositiveBoundsTo(product, term, x, scale)
		divisor.SetInt64(int64(n))
		product.lo.Div(product.lo, &divisor)
		ceilQuotientTo(product.hi, product.hi, &divisor)
		term, product = product, term
		lo.Add(lo, term.lo)
		hi.Add(hi, term.hi)
		// Only the upper next-term bound is needed for the remainder.
		nextUpper.Mul(term.hi, x.hi)
		ceilQuotientTo(&nextUpper, &nextUpper, scale)
		divisor.SetInt64(int64(n + 1))
		ceilQuotientTo(&nextUpper, &nextUpper, &divisor)
		if nextUpper.Cmp(&two) <= 0 {
			// Include the full omitted tail before handing this enclosure to
			// Ziv's endpoint-rounding test; a small term alone is insufficient.
			nextUpper.Lsh(&nextUpper, 1)
			hi.Add(hi, &nextUpper)
			return fixedInterval{lo, hi}, nil
		}
	}
	return fixedInterval{}, fmt.Errorf("exponential series budget exhausted: %w", ErrNoConvergence)
}
