package decimal

import (
	"fmt"
	"math/big"
)

// fixedInterval encloses a real value in [lo, hi] / 2^bits. All arithmetic
// rounds outwards. Operands are read-only; every result owns its integers.
type fixedInterval struct{ lo, hi *big.Int }

func ceilQuotient(a, b *big.Int) *big.Int {
	return new(big.Int).Neg(new(big.Int).Div(new(big.Int).Neg(a), b))
}

func fractionBounds(num, den, scale *big.Int) fixedInterval {
	n := new(big.Int).Mul(num, scale)
	return fixedInterval{new(big.Int).Div(n, den), ceilQuotient(n, den)}
}

func multiplyPositiveBounds(a, b fixedInterval, scale *big.Int) fixedInterval {
	return fixedInterval{
		new(big.Int).Div(new(big.Int).Mul(a.lo, b.lo), scale),
		ceilQuotient(new(big.Int).Mul(a.hi, b.hi), scale),
	}
}

func divideBounds(a, b fixedInterval, scale *big.Int) fixedInterval {
	// Denominators in logarithms and exponentials are strictly positive.
	var lo, hi *big.Int
	for _, numerator := range []*big.Int{a.lo, a.hi} {
		for _, denominator := range []*big.Int{b.lo, b.hi} {
			n := new(big.Int).Mul(numerator, scale)
			lower, upper := new(big.Int).Div(n, denominator), ceilQuotient(n, denominator)
			if lo == nil || lower.Cmp(lo) < 0 {
				lo = lower
			}
			if hi == nil || upper.Cmp(hi) > 0 {
				hi = upper
			}
		}
	}
	return fixedInterval{lo, hi}
}

func multiplyBoundsByInt(a fixedInterval, k int64) fixedInterval {
	lower, upper := a.lo, a.hi
	if k < 0 {
		lower, upper = upper, lower
	}
	return fixedInterval{new(big.Int).Mul(lower, big.NewInt(k)), new(big.Int).Mul(upper, big.NewInt(k))}
}

// logMantissaBounds evaluates ln(m) = 2 sum z^(2j+1)/(2j+1), where
// z=(m-1)/(m+1) and 1 <= m <= 2. After the current term the exact tail is
// at most 3*z^(2j+3)/(2j+3), since z^2 <= 1/9. This bounds truncation as
// well as all fixed-point rounding errors, rather than testing term size alone.
func logMantissaBounds(num, den, scale *big.Int, maxTerms int) (fixedInterval, error) {
	if num.Cmp(den) == 0 {
		return fixedInterval{new(big.Int), new(big.Int)}, nil
	}
	z := fractionBounds(new(big.Int).Sub(num, den), new(big.Int).Add(num, den), scale)
	zSquared := multiplyPositiveBounds(z, z, scale)
	power := z
	sumLo, sumHi := new(big.Int), new(big.Int)
	for j := 0; j < maxTerms; j++ {
		odd := big.NewInt(int64(2*j + 1))
		sumLo.Add(sumLo, new(big.Int).Div(power.lo, odd))
		sumHi.Add(sumHi, ceilQuotient(power.hi, odd))
		next := multiplyPositiveBounds(power, zSquared, scale)
		if next.hi.Cmp(big.NewInt(2)) <= 0 {
			tail := ceilQuotient(new(big.Int).Mul(next.hi, big.NewInt(3)), big.NewInt(int64(2*j+3)))
			return fixedInterval{new(big.Int).Lsh(sumLo, 1), new(big.Int).Add(new(big.Int).Lsh(sumHi, 1), tail)}, nil
		}
		power = next
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

func naturalLogBounds(r *big.Rat, scale *big.Int, maxTerms int) (fixedInterval, error) {
	num, den, k := binaryMantissa(r.Num(), r.Denom())
	mantissa, err := logMantissaBounds(num, den, scale, maxTerms)
	if err != nil {
		return fixedInterval{}, err
	}
	ln2, err := logMantissaBounds(big.NewInt(2), oneInt, scale, maxTerms)
	if err != nil {
		return fixedInterval{}, err
	}
	exponent := multiplyBoundsByInt(ln2, k)
	return fixedInterval{new(big.Int).Add(mantissa.lo, exponent.lo), new(big.Int).Add(mantissa.hi, exponent.hi)}, nil
}

// roundBounds succeeds only if every real value in the enclosure has the
// same RoundHalfEven result at the requested decimal scale.
func roundBounds(value fixedInterval, scale *big.Int, prec int) (Decimal, bool) {
	multiplier := safeGetPrecisionMultiplier(prec)
	round := func(bound *big.Int) *big.Int {
		n := new(big.Int).Mul(bound, multiplier)
		q, r := new(big.Int).QuoRem(n, scale, new(big.Int))
		if r.Sign() != 0 {
			applyDivisionRounding(q, r, scale, RoundHalfEven)
		}
		return q
	}
	lo, hi := round(value.lo), round(value.hi)
	return Decimal{i: lo, prec: prec}, lo.Cmp(hi) == 0
}

// 3.322 is a rational upper bound for log2(10); no floating-point estimate
// enters the numeric calculation or its precision budget.
func decimalBits(prec int) int { return (prec*3322 + 999) / 1000 }

func certifyDecimal(prec, extraDigits int, evaluate func(*big.Int, int) (fixedInterval, error)) (Decimal, error) {
	if prec > maxParsedPrecision || extraDigits > maxParsedPrecision-prec {
		return Decimal{}, fmt.Errorf("transcendental precision budget exceeded: %w", ErrNoConvergence)
	}
	for guard := 64; guard <= 8192; guard *= 2 {
		bits := decimalBits(prec+extraDigits) + guard
		if bits > decimalBits(maxParsedPrecision) {
			break
		}
		scale := new(big.Int).Lsh(big.NewInt(1), uint(bits))
		bounds, err := evaluate(scale, bits)
		if err != nil {
			return Decimal{}, err
		}
		if result, ok := roundBounds(bounds, scale, prec); ok {
			return result, nil
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
	return certifyDecimal(prec, 0, func(scale *big.Int, bits int) (fixedInterval, error) {
		value, err := naturalLogBounds(r, scale, bits)
		if err != nil || base == 0 {
			return value, err
		}
		denominator, err := naturalLogBounds(new(big.Rat).SetInt64(int64(base)), scale, bits)
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
	lo, hi := new(big.Int).Set(scale), new(big.Int).Set(scale)
	for n := 1; n <= maxTerms; n++ {
		product := multiplyPositiveBounds(term, x, scale)
		term = fixedInterval{new(big.Int).Div(product.lo, big.NewInt(int64(n))), ceilQuotient(product.hi, big.NewInt(int64(n)))}
		lo.Add(lo, term.lo)
		hi.Add(hi, term.hi)
		next := multiplyPositiveBounds(term, x, scale)
		nextUpper := ceilQuotient(next.hi, big.NewInt(int64(n+1)))
		if nextUpper.Cmp(big.NewInt(2)) <= 0 {
			return fixedInterval{lo, new(big.Int).Add(hi, new(big.Int).Lsh(nextUpper, 1))}, nil
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
	r := d.Abs().BigRat()
	// Since e > 2 and 16^(prec+1) > 2*10^prec, this certifies rounding to
	// zero for very negative inputs without constructing an enormous e^|x|.
	if negative && r.Cmp(new(big.Rat).SetInt64(int64(4*(prec+1)))) >= 0 {
		return NewWithPrec(0, prec), nil
	}
	// e < 10 implies e^|x| < 10^ceil(|x|). Reserve these integer digits as
	// additional working precision; final acceptance still requires enclosure
	// certification, including every argument-reduction squaring error.
	digits := ceilQuotient(r.Num(), r.Denom())
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
		for range reductions {
			value = multiplyPositiveBounds(value, value, scale)
		}
		if negative {
			value = divideBounds(fixedInterval{scale, scale}, value, scale)
		}
		return value, nil
	})
}
