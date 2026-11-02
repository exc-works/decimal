package decimal

import "math/big"

// fixedInterval encloses a real value in [lo, hi] / scale, where scale=2^bits
// is supplied by the current Ziv attempt. Arithmetic rounds lower endpoints
// down and upper endpoints up, so the exact value stays inside the interval.
// Operands are read-only; returned bounds own independent integers. Helpers
// ending in To write into explicitly owned scratch storage instead.
type fixedInterval struct{ lo, hi *big.Int }

func ceilQuotient(a, b *big.Int) *big.Int {
	return ceilQuotientTo(new(big.Int), a, b)
}

// ceilQuotientTo may overwrite a when the caller owns it. b is positive.
func ceilQuotientTo(quotient, a, b *big.Int) *big.Int {
	quotient.Neg(a)
	if bits := uint(b.BitLen() - 1); b.Sign() > 0 && b.TrailingZeroBits() == bits {
		// Rsh is floor division, including for negative integers. Thus
		// -floor(-a / 2^bits) is the required ceiling, with no long division.
		quotient.Rsh(quotient, bits)
	} else {
		quotient.Div(quotient, b)
	}
	return quotient.Neg(quotient)
}

func fractionBounds(num, den, scale *big.Int) fixedInterval {
	n := new(big.Int).Mul(num, scale)
	// Derive the upper bound before consuming the owned scaled numerator.
	upper := ceilQuotient(n, den)
	return fixedInterval{n.Div(n, den), upper}
}

func multiplyPositiveBounds(a, b fixedInterval, scale *big.Int) fixedInterval {
	result := fixedInterval{new(big.Int), new(big.Int)}
	multiplyPositiveBoundsTo(result, a, b, scale)
	return result
}

// Destination bounds are independently owned and do not alias the operands.
// Series loops alternate destinations so both objects and limb buffers can
// be reused without overwriting multiplication inputs.
func multiplyPositiveBoundsTo(dst, a, b fixedInterval, scale *big.Int) {
	dst.lo.Mul(a.lo, b.lo)
	dst.hi.Mul(a.hi, b.hi)
	if bits := uint(scale.BitLen() - 1); scale.Sign() > 0 && scale.TrailingZeroBits() == bits {
		dst.lo.Rsh(dst.lo, bits)
		dst.hi.Neg(dst.hi).Rsh(dst.hi, bits).Neg(dst.hi)
	} else {
		// Keep general scales supported by the interval helpers.
		dst.lo.Div(dst.lo, scale)
		ceilQuotientTo(dst.hi, dst.hi, scale)
	}
}

func divideBounds(a, b fixedInterval, scale *big.Int) fixedInterval {
	// Denominators in logarithms and exponentials are strictly positive.
	// Division is increasing in the numerator; its denominator direction
	// depends on the numerator's sign. Select the two extreme corners directly.
	lowerDen, upperDen := b.lo, b.lo
	if a.lo.Sign() >= 0 {
		lowerDen = b.hi
	} else if a.hi.Sign() <= 0 {
		upperDen = b.hi
	}
	lo := new(big.Int).Mul(a.lo, scale)
	lo.Div(lo, lowerDen)
	hi := new(big.Int).Mul(a.hi, scale)
	ceilQuotientTo(hi, hi, upperDen)
	return fixedInterval{lo, hi}
}

func multiplyBoundsByInt(a fixedInterval, k int64) fixedInterval {
	lower, upper := a.lo, a.hi
	if k < 0 {
		lower, upper = upper, lower
	}
	multiplier := big.NewInt(k)
	return fixedInterval{new(big.Int).Mul(lower, multiplier), new(big.Int).Mul(upper, multiplier)}
}

// roundBounds succeeds only if every real value in the enclosure has the
// same RoundHalfEven result at the requested decimal scale.
func roundBounds(value fixedInterval, scale *big.Int, prec int) (Decimal, bool) {
	multiplier := safeGetPrecisionMultiplier(prec)
	lo := new(big.Int)
	var rounding roundingWorkspace
	ok := rounding.roundBoundsTo(lo, value, scale, multiplier)
	return Decimal{i: lo, prec: prec}, ok
}

// roundingWorkspace holds the temporary upper endpoint and remainder. The
// Ziv driver reuses it between attempts; the accepted coefficient is owned
// separately so later scratch mutations cannot change a returned Decimal.
type roundingWorkspace struct {
	upper, remainder big.Int
}

// roundBoundsTo is the Ziv acceptance test: rounding is monotone, so equal
// rounded endpoints imply the same result for every value in the enclosure.
// In particular, this tests the enclosure, not successive approximations.
func (w *roundingWorkspace) roundBoundsTo(coefficient *big.Int, value fixedInterval, scale, multiplier *big.Int) bool {
	round := func(quotient, bound *big.Int) {
		quotient.Mul(bound, multiplier)
		quotient.QuoRem(quotient, scale, &w.remainder)
		if w.remainder.Sign() != 0 {
			applyDivisionRounding(quotient, &w.remainder, scale, RoundHalfEven)
		}
	}
	round(coefficient, value.lo)
	round(&w.upper, value.hi)
	return coefficient.Cmp(&w.upper) == 0
}
