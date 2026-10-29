package decimal

import (
	"fmt"
	"math/big"
)

// fixedInterval encloses a real value in [lo, hi] / 2^bits. All arithmetic
// rounds outwards. Operands are read-only; every result owns its integers.
type fixedInterval struct{ lo, hi *big.Int }

func ceilQuotient(a, b *big.Int) *big.Int {
	return ceilQuotientTo(new(big.Int), a, b)
}

// ceilQuotientTo may overwrite a when the caller owns it. b is positive.
func ceilQuotientTo(quotient, a, b *big.Int) *big.Int {
	quotient.Neg(a)
	quotient.Div(quotient, b)
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
	dst.lo.Mul(a.lo, b.lo).Div(dst.lo, scale)
	dst.hi.Mul(a.hi, b.hi)
	ceilQuotientTo(dst.hi, dst.hi, scale)
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

// roundBounds succeeds only if every real value in the enclosure has the
// same RoundHalfEven result at the requested decimal scale.
func roundBounds(value fixedInterval, scale *big.Int, prec int) (Decimal, bool) {
	multiplier := safeGetPrecisionMultiplier(prec)
	lo := new(big.Int)
	var hi, remainder big.Int
	ok := roundBoundsTo(lo, &hi, &remainder, value, scale, multiplier)
	return Decimal{i: lo, prec: prec}, ok
}

func roundBoundsTo(lo, hi, remainder *big.Int, value fixedInterval, scale, multiplier *big.Int) bool {
	round := func(quotient, bound *big.Int) {
		quotient.Mul(bound, multiplier)
		quotient.QuoRem(quotient, scale, remainder)
		if remainder.Sign() != 0 {
			applyDivisionRounding(quotient, remainder, scale, RoundHalfEven)
		}
	}
	round(lo, value.lo)
	round(hi, value.hi)
	return lo.Cmp(hi) == 0
}

// 3.322 is a rational upper bound for log2(10); no floating-point estimate
// enters the numeric calculation or its precision budget.
func decimalBits(prec int) int { return (prec*3322 + 999) / 1000 }

func certifyDecimal(prec, extraDigits int, evaluate func(*big.Int, int) (fixedInterval, error)) (Decimal, error) {
	if prec > maxParsedPrecision || extraDigits > maxParsedPrecision-prec {
		return Decimal{}, fmt.Errorf("transcendental precision budget exceeded: %w", ErrNoConvergence)
	}
	baseBits := decimalBits(prec + extraDigits)
	maxBits := decimalBits(maxParsedPrecision)
	multiplier := safeGetPrecisionMultiplier(prec)
	coefficient := new(big.Int)
	var scale, upper, remainder big.Int
	for guard := 64; guard <= 8192; guard *= 2 {
		bits := baseBits + guard
		if bits > maxBits {
			break
		}
		scale.SetInt64(1).Lsh(&scale, uint(bits))
		bounds, err := evaluate(&scale, bits)
		if err != nil {
			return Decimal{}, err
		}
		if roundBoundsTo(coefficient, &upper, &remainder, bounds, &scale, multiplier) {
			return Decimal{i: coefficient, prec: prec}, nil
		}
	}
	return Decimal{}, fmt.Errorf("cannot certify rounding within precision budget: %w", ErrNoConvergence)
}

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
		// read-only bounds between the argument and base logarithms.
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
			nextUpper.Lsh(&nextUpper, 1)
			hi.Add(hi, &nextUpper)
			return fixedInterval{lo, hi}, nil
		}
	}
	return fixedInterval{}, fmt.Errorf("exponential series budget exhausted: %w", ErrNoConvergence)
}

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
	// Since e > 2 and 16^(prec+1) > 2*10^prec, this certifies rounding to
	// zero for very negative inputs without constructing an enormous e^|x|.
	if negative && r.Cmp(new(big.Rat).SetInt64(int64(4*(prec+1)))) >= 0 {
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
