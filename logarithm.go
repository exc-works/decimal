package decimal

import (
	"fmt"
	"math/big"
)

// certifiedLog prepares exact arguments, then lets the Ziv driver certify the
// final rounding. The evaluator uses ln(x)=ln(m)+k*ln(2), with 1<=m<2;
// base-2/base-10 logarithms divide by a separately enclosed ln(base).
func certifiedLog(d Decimal, prec, base int) (Decimal, error) {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	if d.Sign() <= 0 {
		return Decimal{}, fmt.Errorf("logarithm of non-positive value: %w", ErrInvalidLog)
	}
	if d.prec > maxParsedPrecision || prec > maxParsedPrecision {
		return Decimal{}, fmt.Errorf("logarithm precision budget exceeded: %w", ErrNoConvergence)
	}
	r := d.BigRat()
	if r.Num().Cmp(r.Denom()) == 0 {
		return NewWithPrec(0, prec), nil
	}
	// Exact logarithms avoid unnecessary interval work, including requests at
	// the precision budget boundary where no guard digits would fit.
	if base == 2 && r.Num().TrailingZeroBits() == uint(r.Num().BitLen()-1) && r.Denom().TrailingZeroBits() == uint(r.Denom().BitLen()-1) {
		return NewWithAppendPrec(int64(r.Num().BitLen()-r.Denom().BitLen()), prec), nil
	}
	if base == 10 {
		num, den := r.Num().String(), r.Denom().String()
		powerOfTen := func(s string) bool {
			if s[0] != '1' {
				return false
			}
			for i := 1; i < len(s); i++ {
				if s[i] != '0' {
					return false
				}
			}
			return true
		}
		if powerOfTen(num) && powerOfTen(den) {
			return NewWithAppendPrec(int64(len(num)-len(den)), prec), nil
		}
	}
	// Normalization is independent of working precision, so retain its owned
	// read-only integers across all certification attempts.
	num, den, exponent := binaryMantissa(r.Num(), r.Denom())
	argument := logArgument{num, den, exponent}
	var baseArgument logArgument
	if base != 0 {
		var baseInteger big.Int
		baseInteger.SetInt64(int64(base))
		num, den, exponent := binaryMantissa(&baseInteger, oneInt)
		baseArgument = logArgument{num, den, exponent}
	}
	return certifyDecimal(prec, 0, func(scale *big.Int, bits int) (fixedInterval, error) {
		// Each precision attempt computes ln(2) at most once, sharing the
		// read-only bounds between the argument and base logarithms. This
		// cache belongs to one attempt: its binary scale changes on retry.
		var ln2 fixedInterval
		value, err := naturalLogBounds(argument, scale, bits, &ln2)
		if err != nil || base == 0 {
			return value, err
		}
		denominator, err := naturalLogBounds(baseArgument, scale, bits, &ln2)
		if err != nil {
			return fixedInterval{}, err
		}
		return divideBounds(value, denominator, scale), nil
	})
}

// logMantissaBounds evaluates ln(m) = 2 sum z^(2j+1)/(2j+1), where
// z=(m-1)/(m+1) and 1 <= m <= 2. After the current term the exact tail is
// at most 3*z^(2j+3)/(2j+3), since z^2 <= 1/9. This bounds truncation as
// well as all fixed-point rounding errors, rather than testing term size alone.
func logMantissaBounds(num, den, scale *big.Int, maxTerms int) (fixedInterval, error) {
	if num.Cmp(den) == 0 {
		return fixedInterval{new(big.Int), new(big.Int)}, nil
	}
	var numerator, denominator big.Int
	numerator.Sub(num, den)
	denominator.Add(num, den)
	z := fractionBounds(&numerator, &denominator, scale)
	zSquared := multiplyPositiveBounds(z, z, scale)
	power := z
	next := fixedInterval{new(big.Int), new(big.Int)}
	sumLo, sumHi := new(big.Int), new(big.Int)
	var odd, lowerTerm, upperTerm, tail, two, three big.Int
	two.SetInt64(2)
	three.SetInt64(3)
	for j := 0; j < maxTerms; j++ {
		odd.SetInt64(int64(2*j + 1))
		lowerTerm.Div(power.lo, &odd)
		ceilQuotientTo(&upperTerm, power.hi, &odd)
		sumLo.Add(sumLo, &lowerTerm)
		sumHi.Add(sumHi, &upperTerm)
		multiplyPositiveBoundsTo(next, power, zSquared, scale)
		if next.hi.Cmp(&two) <= 0 {
			// A small next term is a stopping threshold for this series, not
			// proof of final decimal rounding. Include its bounded tail; the
			// outer Ziv driver performs the separate rounding acceptance test.
			odd.SetInt64(int64(2*j + 3))
			tail.Mul(next.hi, &three)
			ceilQuotientTo(&tail, &tail, &odd)
			sumLo.Lsh(sumLo, 1)
			sumHi.Lsh(sumHi, 1).Add(sumHi, &tail)
			return fixedInterval{sumLo, sumHi}, nil
		}
		power, next = next, power
	}
	return fixedInterval{}, fmt.Errorf("logarithm series budget exhausted: %w", ErrNoConvergence)
}

// binaryMantissa writes num/den = m*2^k with exact 1 <= m < 2.
func binaryMantissa(num, den *big.Int) (*big.Int, *big.Int, int64) {
	n, q := new(big.Int).Set(num), new(big.Int).Set(den)
	k := int64(n.BitLen()) - int64(q.BitLen())
	if k >= 0 {
		q.Lsh(q, uint(k))
	} else {
		n.Lsh(n, uint(-k))
	}
	if n.Cmp(q) < 0 {
		n.Lsh(n, 1)
		k--
	}
	return n, q, k
}

type logArgument struct {
	num, den *big.Int
	exponent int64
}

func naturalLogBounds(argument logArgument, scale *big.Int, maxTerms int, ln2 *fixedInterval) (fixedInterval, error) {
	mantissa, err := logMantissaBounds(argument.num, argument.den, scale, maxTerms)
	if err != nil {
		return fixedInterval{}, err
	}
	if argument.exponent == 0 {
		return mantissa, nil
	}
	if ln2.lo == nil {
		var two big.Int
		two.SetInt64(2)
		*ln2, err = logMantissaBounds(&two, oneInt, scale, maxTerms)
		if err != nil {
			return fixedInterval{}, err
		}
	}
	exponent := multiplyBoundsByInt(*ln2, argument.exponent)
	mantissa.lo.Add(mantissa.lo, exponent.lo)
	mantissa.hi.Add(mantissa.hi, exponent.hi)
	return mantissa, nil
}
