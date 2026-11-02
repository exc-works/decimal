package decimal

import (
	"math/big"
)

// round rounds d.i/10^d.prec to an integer coefficient. Here d.prec is the
// number of discarded places; callers replace the result's scale afterward.
// The input is read-only and the returned coefficient is independently owned.
func (d Decimal) round(mode RoundingMode) Decimal {
	validateRoundingMode(mode)
	divisor := safeGetPrecisionMultiplier(d.prec)
	coefficient := new(big.Int)
	if mode == RoundDown {
		coefficient.Quo(d.i, divisor)
	} else {
		var remainder big.Int
		coefficient.QuoRem(d.i, divisor, &remainder)
		if remainder.Sign() != 0 {
			applyDivisionRounding(coefficient, &remainder, divisor, mode)
		}
	}
	return Decimal{i: coefficient, prec: d.prec}
}

// validateRoundingMode panics when mode is not one of the documented constants.
// Callers that have a fast path bypassing applyDivisionRounding (e.g. Quo on
// exact division) must invoke this so the public contract — "panics on invalid
// rounding mode" — is honored regardless of input shape.
func validateRoundingMode(mode RoundingMode) {
	switch mode {
	case RoundDown, RoundUp, RoundCeiling,
		RoundHalfUp, RoundHalfDown, RoundHalfEven,
		RoundUnnecessary:
		return
	default:
		panic("invalid rounding mode")
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
		panic(ErrRoundUnnecessary)
	default:
		panic("invalid rounding mode")
	}
}
