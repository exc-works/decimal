package decimal

import (
	"math/big"
)

func (d Decimal) round(mode RoundingMode) Decimal {
	switch mode {
	case RoundDown:
		return d.roundDown()
	case RoundUp:
		return d.roundUp()
	case RoundCeiling:
		return d.roundCeiling()
	case RoundHalfUp:
		return d.roundHalfUp()
	case RoundHalfDown:
		return d.roundHalfDown()
	case RoundHalfEven:
		return d.roundHalfEven()
	case RoundUnnecessary:
		return d.roundUnnecessary()
	default:
		panic("invalid rounding mode")
	}
}

func (d Decimal) roundDown() Decimal {
	value := new(big.Int).Quo(d.i, safeGetPrecisionMultiplier(d.prec))
	return Decimal{
		i:    value,
		prec: d.prec,
	}
}

func (d Decimal) roundTruncate() Decimal {
	return d.roundDown()
}

func (d Decimal) roundUp() Decimal {
	if d.IsNegative() {
		// Make d positive
		abs := d.Neg()
		abs = abs.roundUp()
		return abs.Neg()
	}

	// Get the truncated quotient and remainder
	quo, rem := new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
	if rem.Sign() == 0 {
		return Decimal{
			i:    quo,
			prec: d.prec,
		}
	}

	return Decimal{
		i:    quo.Add(quo, oneInt),
		prec: d.prec,
	}
}

func (d Decimal) roundCeiling() Decimal {
	if d.IsNegative() {
		return d.roundDown()
	}
	return d.roundUp()
}

func (d Decimal) roundHalfUp() Decimal {
	if d.prec == 0 {
		return d
	}
	if d.IsNegative() {
		// Make a positive
		abs := d.Neg()
		abs = abs.roundHalfUp()
		return abs.Neg()
	}
	quo, rem := new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
	fivePrecision := new(big.Int).Mul(fiveInt, safeGetPrecisionMultiplier(d.prec-1))
	cmp := rem.Cmp(fivePrecision)
	if cmp < 0 {
		return Decimal{
			i:    quo,
			prec: d.prec,
		}
	} else {
		return Decimal{
			i:    quo.Add(quo, oneInt),
			prec: d.prec,
		}
	}
}

func (d Decimal) roundHalfDown() Decimal {
	if d.prec == 0 {
		return d
	}
	if d.IsNegative() {
		// Make a positive
		abs := d.Neg()
		abs = abs.roundHalfDown()
		return abs.Neg()
	}

	quo, rem := new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
	fivePrecision := new(big.Int).Mul(fiveInt, safeGetPrecisionMultiplier(d.prec-1))
	cmp := rem.Cmp(fivePrecision)
	if cmp <= 0 {
		return Decimal{
			i:    quo,
			prec: d.prec,
		}
	} else {
		return Decimal{
			i:    quo.Add(quo, oneInt),
			prec: d.prec,
		}
	}
}

func (d Decimal) roundHalfEven() Decimal {
	if d.prec == 0 {
		return d
	}
	if d.IsNegative() {
		// Make d positive
		abs := d.Neg()
		abs = abs.roundHalfEven()
		return abs.Neg()
	}

	quo, rem := new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
	fivePrecision := new(big.Int).Mul(fiveInt, safeGetPrecisionMultiplier(d.prec-1))
	cmp := rem.Cmp(fivePrecision)
	var resultD Decimal
	if cmp < 0 {
		resultD = Decimal{
			i:    quo,
			prec: d.prec,
		}
	} else if cmp > 0 {
		resultD = Decimal{
			i:    quo.Add(quo, oneInt),
			prec: d.prec,
		}
	} else {
		// Bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			resultD = Decimal{
				i:    quo,
				prec: d.prec,
			}
		} else {
			resultD = Decimal{
				i:    quo.Add(quo, oneInt),
				prec: d.prec,
			}
		}
	}

	return resultD
}

func (d Decimal) roundUnnecessary() Decimal {
	if d.IsNegative() {
		// Make d positive
		abs := d.Neg()
		abs = abs.roundUnnecessary()
		return abs.Neg()
	}

	quo, rem := new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
	if rem.Sign() != 0 {
		panic("expected 0 remainder")
	}
	return Decimal{
		i:    quo,
		prec: d.prec,
	}
}

// applyDivisionRounding adjusts quo (the truncated-toward-zero quotient of an
// integer division) in place, given the corresponding non-zero remainder rem
// and the divisor. divisor may be negative; only its magnitude is used for
// the halfway comparison. rem carries the sign of the original numerator
// (Go's QuoRem convention).
//
// Caller must check rem.Sign() != 0 before invoking this helper. quo and rem
// must be freshly-owned big.Ints; this function mutates quo and may mutate
// rem (it doubles |rem| for the halfway check). divisor is read-only and is
// allowed to alias a shared/cached big.Int.
func applyDivisionRounding(quo, rem, divisor *big.Int, mode RoundingMode) {
	// The true quotient q = quo + rem/divisor has the sign of num*divisor;
	// since rem inherits num's sign, sign(q) == sign(rem) * sign(divisor).
	resultPositive := (rem.Sign() > 0) == (divisor.Sign() > 0)

	awayFromZero := func() {
		if resultPositive {
			quo.Add(quo, oneInt)
		} else {
			quo.Sub(quo, oneInt)
		}
	}

	switch mode {
	case RoundDown:
		// already truncated toward zero
	case RoundUp:
		awayFromZero()
	case RoundCeiling:
		if resultPositive {
			awayFromZero()
		}
	case RoundHalfUp, RoundHalfDown, RoundHalfEven:
		// Compare |2*rem| against |divisor|.
		twiceAbsRem := rem.Abs(rem)
		twiceAbsRem.Lsh(twiceAbsRem, 1)
		absDivisor := divisor
		if divisor.Sign() < 0 {
			absDivisor = new(big.Int).Neg(divisor)
		}
		cmp := twiceAbsRem.Cmp(absDivisor)
		switch mode {
		case RoundHalfUp:
			if cmp >= 0 {
				awayFromZero()
			}
		case RoundHalfDown:
			if cmp > 0 {
				awayFromZero()
			}
		case RoundHalfEven:
			if cmp > 0 {
				awayFromZero()
			} else if cmp == 0 && quo.Bit(0) != 0 {
				awayFromZero()
			}
		}
	case RoundUnnecessary:
		panic("inexact conversion")
	default:
		panic("invalid rounding mode")
	}
}
