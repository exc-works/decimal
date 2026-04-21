package decimal

import (
	"errors"
	"fmt"
	stdmath "math"
	"math/big"
	"strconv"
	"strings"
)

const (
	cacheMaxPrecision = 128

	// max number of iterations in Sqrt, Log2 function
	maxIterations = 300
)

var (
	zeroInt = big.NewInt(0)
	oneInt  = big.NewInt(1)
	twoInt  = big.NewInt(2)
	fiveInt = big.NewInt(5)
	tenInt  = big.NewInt(10)
)

var (
	precisionMultipliers [cacheMaxPrecision + 1]*big.Int

	// Zero is the decimal zero value.
	Zero = New(0)
	// One is the decimal one value.
	One = New(1)
	// Ten is the decimal ten value.
	Ten = New(10)
	// Hundred is the decimal one hundred value.
	Hundred = New(100)
)

// Decimal represents a decimal number with arbitrary precision.
type Decimal struct {
	i    *big.Int
	prec int
}

func init() {
	precisionMultipliers[0] = big.NewInt(1) // 10^0
	for i := 1; i <= cacheMaxPrecision; i++ {
		precisionMultipliers[i] = new(big.Int).Mul(precisionMultipliers[i-1], tenInt)
	}
}

func safeGetPrecisionMultiplier(prec int) *big.Int {
	if prec < 0 {
		panic("negative precision")
	}
	if prec <= cacheMaxPrecision {
		return precisionMultipliers[prec]
	}
	return new(big.Int).Exp(tenInt, big.NewInt(int64(prec)), nil)
}

// New returns a Decimal created from value with precision 0.
func New(value int64) Decimal {
	return NewFromBigInt(big.NewInt(value))
}

// NewFromDecimal returns a deep copy of d. It is provided for API symmetry
// with the other NewFrom* constructors and is equivalent to d.Clone().
func NewFromDecimal(d Decimal) Decimal {
	return d.Clone()
}

// NewFromInt returns a Decimal created from value with precision 0.
func NewFromInt(value int) Decimal {
	return NewFromBigInt(big.NewInt(int64(value)))
}

// NewWithPrec returns a Decimal created from value with the given precision.
// It panics if prec is negative.
func NewWithPrec(value int64, prec int) Decimal {
	requireNonNegativePrecision(prec)
	return NewFromBigIntWithPrec(big.NewInt(value), prec)
}

// NewFromFloat64 returns a Decimal parsed from value.
func NewFromFloat64(value float64) Decimal {
	return MustFromString(strconv.FormatFloat(value, 'f', -1, 64))
}

// NewFromFloat32 returns a Decimal parsed from value.
func NewFromFloat32(value float32) Decimal {
	return MustFromString(strconv.FormatFloat(float64(value), 'f', -1, 32))
}

// NewFromBigRat returns a Decimal converted from value.
//
// It returns an error when value is nil or cannot be represented as a
// terminating decimal (for example 1/3).
func NewFromBigRat(value *big.Rat) (Decimal, error) {
	if value == nil {
		return Decimal{}, fmt.Errorf("big.Rat cannot be nil: %w", ErrInvalidFormat)
	}

	num := new(big.Int).Set(value.Num())
	den := new(big.Int).Set(value.Denom())

	if num.Sign() == 0 {
		return Zero, nil
	}

	twos, rem := countFactor(den, 2)
	fives, rem := countFactor(rem, 5)
	if rem.Cmp(oneInt) != 0 {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: non-terminating decimal: %w", value.RatString(), ErrInvalidFormat)
	}

	prec := twos
	if fives > prec {
		prec = fives
	}

	unscaled := new(big.Int).Mul(num, safeGetPrecisionMultiplier(prec))
	unscaled.Quo(unscaled, den)

	return Decimal{
		i:    unscaled,
		prec: prec,
	}, nil
}

// NewFromBigRatWithPrec returns a Decimal converted from value at precision prec.
//
// The result is rounded according to roundingMode.
// It returns an error when value is nil and panics when prec is negative.
func NewFromBigRatWithPrec(value *big.Rat, prec int, roundingMode RoundingMode) (Decimal, error) {
	requireNonNegativePrecision(prec)
	if value == nil {
		return Decimal{}, fmt.Errorf("big.Rat cannot be nil: %w", ErrInvalidFormat)
	}

	num := new(big.Int).Set(value.Num())
	den := new(big.Int).Set(value.Denom())

	if num.Sign() == 0 {
		return NewWithPrec(0, prec), nil
	}

	scaled := new(big.Int).Mul(num, safeGetPrecisionMultiplier(prec))
	quotient, remainder := new(big.Int).QuoRem(scaled, den, new(big.Int))
	if remainder.Sign() != 0 {
		// Decide whether to adjust quotient based on the exact remainder,
		// so digits beyond prec+1 are also accounted for.
		awayFromZero := func() {
			if remainder.Sign() > 0 {
				quotient.Add(quotient, oneInt)
			} else {
				quotient.Sub(quotient, oneInt)
			}
		}

		switch roundingMode {
		case RoundDown:
			// already truncated toward zero
		case RoundUp:
			awayFromZero()
		case RoundCeiling:
			if remainder.Sign() > 0 {
				quotient.Add(quotient, oneInt)
			}
		case RoundHalfUp, RoundHalfDown, RoundHalfEven:
			twiceAbsRem := new(big.Int).Abs(remainder)
			twiceAbsRem.Mul(twiceAbsRem, twoInt)
			cmp := twiceAbsRem.Cmp(den)

			switch roundingMode {
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
				} else if cmp == 0 {
					if new(big.Int).Abs(quotient).Bit(0) != 0 {
						awayFromZero()
					}
				}
			}
		case RoundUnnecessary:
			panic("inexact conversion")
		default:
			panic("invalid rounding mode")
		}
	}

	return Decimal{
		i:    quotient,
		prec: prec,
	}, nil
}

// NewWithAppendPrec returns a Decimal created from value with prec trailing zeros appended.
// It panics if prec is negative.
func NewWithAppendPrec(value int64, prec int) Decimal {
	requireNonNegativePrecision(prec)
	return Decimal{
		i:    new(big.Int).Mul(big.NewInt(value), safeGetPrecisionMultiplier(prec)),
		prec: prec,
	}
}

// NewFromUintWithAppendPrec returns a Decimal created from value with prec trailing zeros appended.
// It panics if prec is negative.
func NewFromUintWithAppendPrec(value uint64, prec int) Decimal {
	requireNonNegativePrecision(prec)
	return Decimal{
		i:    new(big.Int).Mul(new(big.Int).SetUint64(value), safeGetPrecisionMultiplier(prec)),
		prec: prec,
	}
}

// NewFromBigInt returns a Decimal created from value with precision 0.
func NewFromBigInt(value *big.Int) Decimal {
	return NewFromBigIntWithPrec(value, 0)
}

// NewFromBigIntWithPrec returns a Decimal created from value with the given precision.
// It panics if precision is negative.
func NewFromBigIntWithPrec(value *big.Int, precision int) Decimal {
	requireNonNegativePrecision(precision)
	return Decimal{
		i:    new(big.Int).Set(value),
		prec: precision,
	}
}

// NewFromInt64 returns a Decimal created from value with the given precision.
// It panics if precision is negative.
func NewFromInt64(value int64, precision int) Decimal {
	requireNonNegativePrecision(precision)
	return Decimal{
		i:    new(big.Int).SetInt64(value),
		prec: precision,
	}
}

// NewFromUint64 returns a Decimal created from value with the given precision.
// It panics if precision is negative.
func NewFromUint64(value uint64, precision int) Decimal {
	requireNonNegativePrecision(precision)
	return Decimal{
		i:    new(big.Int).SetUint64(value),
		prec: precision,
	}
}

// NewFromString returns a Decimal parsed from str.
//
// It accepts plain decimal values and scientific notation, and returns an
// error for empty or malformed input.
func NewFromString(str string) (d Decimal, err error) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return Decimal{}, fmt.Errorf("decimal string cannot be empty: %w", ErrInvalidFormat)
	}

	// Parse scientific notation first
	var expOffset int64 = 0
	eIndex := strings.IndexAny(str, "Ee")
	if eIndex != -1 {
		// Parse the exponent
		expStr := str[eIndex+1:]
		if len(expStr) == 0 {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: missing exponent: %w", str, ErrInvalidFormat)
		}

		// Handle optional sign in exponent
		expSign := int64(1)
		switch expStr[0] {
		case '+':
			expStr = expStr[1:]
		case '-':
			expSign = -1
			expStr = expStr[1:]
		}

		if len(expStr) == 0 {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: missing exponent value: %w", str, ErrInvalidFormat)
		}

		expInt, err := strconv.ParseInt(expStr, 10, 32)
		if err != nil {
			if e, ok := errors.AsType[*strconv.NumError](err); ok && errors.Is(e.Err, strconv.ErrRange) {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent too large: %w", str, ErrInvalidFormat)
			}
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric: %w", str, ErrInvalidFormat)
		}

		expOffset = expSign * expInt
		str = str[:eIndex]
	}

	// Extract negative symbol
	neg := false
	if len(str) > 0 && str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string: %w", str, ErrInvalidFormat)
	}

	// Parse the mantissa (number part without exponent)
	var precision int
	strs := strings.Split(str, ".")
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		precision = len(strs[1])
		// Maintain backward compatibility: reject formats like "1." and ".1"
		if len(combinedStr) == 0 || precision == 0 {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string: %w", str, ErrInvalidFormat)
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string: %w", str, ErrInvalidFormat)
	}

	if combinedStr == "" {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string: %w", str, ErrInvalidFormat)
	}

	// Apply exponent offset to precision
	precision -= int(expOffset)

	// Parse the combined string as big.Int first
	combined, ok := new(big.Int).SetString(combinedStr, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("failed to set decimal string %s: %w", combinedStr, ErrInvalidFormat)
	}

	if precision < 0 {
		// Convert to integer by multiplying by 10^(-precision)
		ten := big.NewInt(10)
		multiplier := ten.Exp(ten, big.NewInt(int64(-precision)), nil)
		combined.Mul(combined, multiplier)
		precision = 0
	}

	if neg {
		combined = new(big.Int).Neg(combined)
	}

	// If the result is zero, precision should be 0
	if combined.Sign() == 0 {
		precision = 0
	}

	return Decimal{
		i:    combined,
		prec: precision,
	}, nil
}

// MustFromString returns a Decimal parsed from str and panics if parsing fails.
func MustFromString(str string) Decimal {
	d, err := NewFromString(str)
	if err != nil {
		panic(err)
	}
	return d
}

// Add returns d + d2, rescaled to the larger precision of the two values.
func (d Decimal) Add(d2 Decimal) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Add(d1.i, d2.i),
		prec: maxPrec,
	}
}

// SafeAdd returns d + d2 and panics if the result is negative.
func (d Decimal) SafeAdd(d2 Decimal) Decimal {
	return d.Add(d2).requireNonNegative()
}

// AddRaw returns d + i while preserving d's precision.
func (d Decimal) AddRaw(i int64) Decimal {
	d = initializeIfNeeded(d)
	return Decimal{
		i:    new(big.Int).Add(d.i, big.NewInt(i)),
		prec: d.prec,
	}
}

// Sub returns d - d2, rescaled to the larger precision of the two values.
func (d Decimal) Sub(d2 Decimal) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Sub(d1.i, d2.i),
		prec: maxPrec,
	}
}

// SafeSub returns d - d2 and panics if the result is negative.
func (d Decimal) SafeSub(d2 Decimal) Decimal {
	return d.Sub(d2).requireNonNegative()
}

// SubRaw returns d - i while preserving d's precision.
func (d Decimal) SubRaw(i int64) Decimal {
	d = initializeIfNeeded(d)
	return Decimal{
		i:    new(big.Int).Sub(d.i, big.NewInt(i)),
		prec: d.prec,
	}
}

// Mul returns d * d2 rounded according to roundingMode.
func (d Decimal) Mul(d2 Decimal, roundingMode RoundingMode) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Mul(d1.i, d2.i),
		prec: maxPrec,
	}.round(roundingMode)
}

// MulDown returns d * d2 rounded down.
func (d Decimal) MulDown(d2 Decimal) Decimal {
	return d.Mul(d2, RoundDown)
}

// MulExact returns d * d2 without rounding, using the sum of input precisions.
func (d Decimal) MulExact(d2 Decimal) Decimal {
	d = initializeIfNeeded(d)
	d2 = initializeIfNeeded(d2)
	prec := int64(d.prec) + int64(d2.prec)
	if prec > stdmath.MaxInt32 || prec < stdmath.MinInt32 {
		panic("precision overflow")
	}
	return Decimal{
		i:    new(big.Int).Mul(d.i, d2.i),
		prec: int(prec),
	}
}

// Mul2 returns d * d2 without rounding, using the sum of input precisions.
//
// Deprecated: use MulExact instead.
func (d Decimal) Mul2(d2 Decimal) Decimal {
	return d.MulExact(d2)
}

// QuoWithPrec returns d / d2 rounded to prec decimal places using roundingMode.
// It panics if prec is negative, d2 is zero, or roundingMode is invalid.
func (d Decimal) QuoWithPrec(d2 Decimal, prec int, roundingMode RoundingMode) Decimal {
	d = initializeIfNeeded(d)
	d2 = initializeIfNeeded(d2)
	if prec > d.prec && prec > d2.prec {
		d = d.Rescale(prec, RoundUnnecessary)
		d2 = d2.Rescale(prec, RoundUnnecessary)
		return d.Quo(d2, roundingMode)
	}
	return d.Quo(d2, roundingMode).Rescale(prec, roundingMode)
}

// Quo returns d / d2 rounded according to roundingMode.
// It panics if d2 is zero or roundingMode is invalid.
func (d Decimal) Quo(d2 Decimal, roundingMode RoundingMode) Decimal {
	d = initializeIfNeeded(d)
	d2 = initializeIfNeeded(d2)
	// To adapt to the situation where the precision of both numbers is 0,
	// the precision of both numbers is increased by 1, and the final calculation
	// result is rescaled to 0.
	if d.prec == 0 && d2.prec == 0 {
		d1, d2 := d.RescaleDown(1), d2.RescaleDown(1)
		// multiply precision twice
		d1Twice := new(big.Int).Mul(d1.i, safeGetPrecisionMultiplier(1))
		d1Twice = new(big.Int).Mul(d1Twice, safeGetPrecisionMultiplier(1))

		return Decimal{
			i:    new(big.Int).Quo(d1Twice, d2.i),
			prec: 1 * 2,
		}.Rescale(0, roundingMode)
	}

	d1, d2, maxPrec := rescalePair(d, d2)
	// multiply precision twice
	d1Twice := new(big.Int).Mul(d1.i, safeGetPrecisionMultiplier(maxPrec))
	d1Twice = new(big.Int).Mul(d1Twice, safeGetPrecisionMultiplier(maxPrec))

	return Decimal{
		i:    new(big.Int).Quo(d1Twice, d2.i),
		prec: maxPrec,
	}.round(roundingMode)
}

// QuoDown returns d / d2 rounded down.
func (d Decimal) QuoDown(d2 Decimal) Decimal {
	return d.Quo(d2, RoundDown)
}

// Floor returns the greatest integer value less than or equal to d.
func (d Decimal) Floor() Decimal {
	return d.FloorWithPrec(0)
}

// FloorWithPrec returns d rounded toward negative infinity at the given precision.
// It panics if prec is negative.
func (d Decimal) FloorWithPrec(prec int) Decimal {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	truncated := d.Rescale(prec, RoundDown)
	if d.IsNegative() && d.Cmp(truncated) < 0 {
		return truncated.SubRaw(1)
	}
	return truncated
}

// Ceil returns the least integer value greater than or equal to d.
func (d Decimal) Ceil() Decimal {
	return d.CeilWithPrec(0)
}

// CeilWithPrec returns d rounded toward positive infinity at the given precision.
// It panics if prec is negative.
func (d Decimal) CeilWithPrec(prec int) Decimal {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	truncated := d.Rescale(prec, RoundDown)
	if d.IsPositive() && d.Cmp(truncated) > 0 {
		return truncated.AddRaw(1)
	}
	return truncated
}

// Truncate returns d rounded toward zero to an integer value.
func (d Decimal) Truncate() Decimal {
	return d.TruncateWithPrec(0)
}

// TruncateWithPrec returns d rounded toward zero at the given precision.
// It panics if prec is negative.
func (d Decimal) TruncateWithPrec(prec int) Decimal {
	return d.Rescale(prec, RoundDown)
}

// Round returns d rounded to the nearest integer using RoundHalfEven.
func (d Decimal) Round() Decimal {
	return d.RoundWithPrec(0)
}

// RoundWithPrec returns d rounded to the given precision using RoundHalfEven.
// It panics if prec is negative.
func (d Decimal) RoundWithPrec(prec int) Decimal {
	return d.Rescale(prec, RoundHalfEven)
}

// QuoRem returns the quotient truncated toward zero and the corresponding remainder.
// It panics if d2 is zero.
func (d Decimal) QuoRem(d2 Decimal) (Decimal, Decimal) {
	d1, d2, maxPrec := rescalePair(d, d2)
	if d2.i.Sign() == 0 {
		panic("division by zero")
	}

	quo, rem := new(big.Int).QuoRem(d1.i, d2.i, new(big.Int))
	return Decimal{i: quo, prec: 0}, Decimal{i: rem, prec: maxPrec}
}

// Mod returns the same remainder component as QuoRem (truncated division).
// It panics if d2 is zero.
func (d Decimal) Mod(d2 Decimal) Decimal {
	_, rem := d.QuoRem(d2)
	return rem
}

// IntPart returns the integer part of d.
func (d Decimal) IntPart() *big.Int {
	intPart, _ := d.Remainder()
	return intPart
}

// Remainder returns the integer part and fractional part of d.
func (d Decimal) Remainder() (intPart *big.Int, fractionPart *big.Int) {
	d = initializeIfNeeded(d)
	return new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
}

// Power returns d raised to the given integer power.
func (d Decimal) Power(power int64) Decimal {
	d = initializeIfNeeded(d)
	if power == 0 {
		return One.Rescale(d.prec, RoundUnnecessary)
	}

	if power < 0 {
		// If power is negative, we will return a round up value
		return One.Quo(d.Power(-power), RoundUp)
	}

	tmp, resultD := NewWithAppendPrec(1, d.prec), d
	for i := power; i > 1; {
		if i%2 != 0 {
			tmp = tmp.MulExact(resultD)
		}
		i /= 2
		resultD = resultD.MulExact(resultD)
	}
	return resultD.MulExact(tmp).Rescale(d.prec, RoundHalfEven)
}

// Sqrt returns an approximate square root of d.
//
// Note: since v0.3.0 the output precision is max(d.Precision(), 30), which is
// a breaking change relative to v0.2.x where the result matched the receiver
// precision exactly. Callers that need the previous behavior should use
// SqrtWithPrec(d.Precision()) explicitly.
//
// It returns an error wrapping ErrNegativeRoot when d is negative. The
// output precision is max(d.prec, defaultLogExpPrec) so that integer
// receivers still produce meaningful results.
func (d Decimal) Sqrt() (Decimal, error) {
	d = initializeIfNeeded(d)
	return d.SqrtWithPrec(workingLogPrec(d))
}

// SqrtWithPrec returns an approximate square root of d rescaled to prec
// decimal places using RoundHalfEven.
//
// Note: since v0.3.0 callers that pass prec smaller than 30 still observe a
// lifted working precision internally, but the final result is rescaled to
// the requested prec. Pass d.Precision() to reproduce the v0.2.x behavior of
// matching the receiver precision.
//
// It returns an error wrapping ErrNegativeRoot when d is negative.
// It panics if prec is negative.
func (d Decimal) SqrtWithPrec(prec int) (Decimal, error) {
	return d.ApproxRootWithPrec(2, prec)
}

// ApproxRoot returns an approximate integer-th root of d using iterative
// refinement.
//
// Note: since v0.3.0 the output precision is max(d.Precision(), 30), which is
// a breaking change relative to v0.2.x where the result matched the receiver
// precision exactly. Callers that need the previous behavior should use
// ApproxRootWithPrec(root, d.Precision()) explicitly.
//
// It returns an error wrapping ErrInvalidRoot when root is not strictly
// positive, or ErrNegativeRoot when d is negative and root is even. The
// output precision is max(d.prec, defaultLogExpPrec).
func (d Decimal) ApproxRoot(root int64) (Decimal, error) {
	d = initializeIfNeeded(d)
	return d.ApproxRootWithPrec(root, workingLogPrec(d))
}

// ApproxRootWithPrec returns an approximate integer-th root of d rescaled to
// prec decimal places using RoundHalfEven.
//
// Note: since v0.3.0 callers that pass prec smaller than 30 still observe a
// lifted working precision internally, but the final result is rescaled to
// the requested prec. Pass d.Precision() to reproduce the v0.2.x behavior of
// matching the receiver precision.
//
// It returns an error wrapping ErrInvalidRoot when root is not strictly
// positive, or ErrNegativeRoot when d is negative and root is even.
// It panics if prec is negative.
func (d Decimal) ApproxRootWithPrec(root int64, prec int) (Decimal, error) {
	requireNonNegativePrecision(prec)
	if root <= 0 {
		return Decimal{}, fmt.Errorf("ApproxRoot: root %d must be greater than 0: %w", root, ErrInvalidRoot)
	}

	d = initializeIfNeeded(d)
	if d.IsNegative() {
		if root%2 == 0 {
			return Decimal{}, fmt.Errorf("ApproxRoot(%s, %d): cannot take even root of negative value: %w", d.String(), root, ErrNegativeRoot)
		}
		absRoot, err := d.Neg().ApproxRootWithPrec(root, prec)
		if err != nil {
			return Decimal{}, err
		}
		return absRoot.Neg(), nil
	}

	if root == 1 {
		return d.Rescale(prec, RoundHalfEven), nil
	}
	if d.IsZero() {
		return NewWithAppendPrec(0, prec), nil
	}
	if d.Equal(One) {
		return NewWithAppendPrec(1, prec), nil
	}

	// Compute at a precision that absorbs rounding through the iteration
	// and rescale the input up (losslessly) when needed so intermediate
	// deltas have room to represent fractional digits.
	work := prec + 6
	if work < defaultLogExpPrec {
		work = defaultLogExpPrec
	}
	dWork := d
	if d.prec < work {
		dWork = d.Rescale(work, RoundHalfEven)
	}

	rootInt := big.NewInt(0).SetInt64(root)
	guess := NewWithAppendPrec(1, dWork.prec)
	delta := guess

	for iter := 0; delta.Abs().i.Cmp(oneInt) > 0 && iter < maxIterations; iter++ {
		prev := guess.Power(root - 1)
		if prev.IsZero() {
			prev = One
		}
		delta = dWork.Quo(prev, RoundHalfEven)
		delta = delta.Sub(guess)

		quo := new(big.Int).Quo(delta.i, rootInt)
		delta = Decimal{i: quo, prec: dWork.prec}

		guess = guess.Add(delta)
	}
	return guess.Rescale(prec, RoundHalfEven), nil
}

// Log2 returns an approximate log base 2 of d via iterative refinement.
// The iteration is bounded by maxIterations.
// It panics if d is not greater than 0.
func (d Decimal) Log2() Decimal {
	d = initializeIfNeeded(d)
	if d.Sign() <= 0 {
		panic("value must greater than 0")
	}

	oneDec := NewWithAppendPrec(1, d.prec)
	twoDec := NewWithAppendPrec(2, d.prec)

	lessOne := d.Cmp(oneDec) < 0
	copyD := d
	exp := 4 * d.prec
	if lessOne {
		// Ensure copyD greater than 1
		copyD = copyD.Mul(New(2).Power(int64(exp)), RoundHalfEven)
	}

	intPart, _ := copyD.Remainder()
	n := MostSignificantBit(intPart)
	resultDec := NewFromUintWithAppendPrec(uint64(n), copyD.prec)

	int64N := int64(n)
	if int64N < 0 {
		panic(fmt.Sprintf("Most Significant Bit %d too larger", n))
	}

	remDec := copyD.Quo(New(2).Power(int64N), RoundHalfEven)
	for i := 0; i < maxIterations && remDec.Sign() > 0; i++ {
		if remDec.GTE(twoDec) {
			resultDec = resultDec.Add(oneDec.Quo(twoDec.Power(int64(i)), RoundHalfEven))
			remDec = remDec.Quo(twoDec, RoundHalfEven)
		}
		remDec = remDec.Power(2)
	}

	if lessOne {
		resultDec = resultDec.Sub(New(int64(exp)))
	}
	return resultDec
}

// RescaleDown returns d rescaled to prec decimal places using RoundDown.
// It panics if prec is negative.
func (d Decimal) RescaleDown(prec int) Decimal {
	d = initializeIfNeeded(d)
	return d.Rescale(prec, RoundDown)
}

// Rescale returns d rescaled to prec decimal places using roundingMode.
// It panics if prec is negative or roundingMode is invalid.
func (d Decimal) Rescale(prec int, roundingMode RoundingMode) Decimal {
	requireNonNegativePrecision(prec)
	d = initializeIfNeeded(d)
	if d.prec == prec {
		return d
	}

	diff := d.prec - prec
	var newI = new(big.Int)
	if diff < 0 {
		// Mul never should round
		newI.Mul(d.i, safeGetPrecisionMultiplier(-diff))
	} else {
		roundedDecimal := Decimal{
			i:    d.i,
			prec: diff,
		}.round(roundingMode)
		return Decimal{
			i:    roundedDecimal.i,
			prec: prec,
		}
	}
	return Decimal{
		i:    newI,
		prec: prec,
	}
}

// Shift returns d multiplied by 10^places exactly.
//
// Positive places shift the decimal point to the right; negative places shift
// it to the left.
func (d Decimal) Shift(places int) Decimal {
	d = initializeIfNeeded(d)
	if places == 0 {
		return d
	}

	targetPrec := int64(d.prec) - int64(places)
	if targetPrec < 0 {
		return Decimal{
			i:    new(big.Int).Mul(d.i, safeGetPrecisionMultiplier(int(-targetPrec))),
			prec: 0,
		}
	}
	if targetPrec > stdmath.MaxInt32 {
		panic("precision overflow")
	}
	return Decimal{
		i:    d.i,
		prec: int(targetPrec),
	}
}

// StripTrailingZeros returns a Decimal which is numerically equal to this one
// but with any trailing zeros removed from the representation.
func (d Decimal) StripTrailingZeros() Decimal {
	d = initializeIfNeeded(d)
	if d.prec == 0 {
		return d
	}

	// Fast path for values that fit in int64: use native arithmetic (zero allocs
	// when no trailing zeros, one alloc otherwise).
	if d.i.IsInt64() {
		v := d.i.Int64()
		if v%10 != 0 {
			return d
		}
		k := 0
		for k < d.prec && v%10 == 0 {
			v /= 10
			k++
		}
		return Decimal{i: big.NewInt(v), prec: d.prec - k}
	}

	// Large numbers: pre-allocate quo and mod, then binary search.
	quo := new(big.Int)
	mod := new(big.Int)

	// Quick check: if last digit is non-zero, no trailing zeros.
	quo.QuoRem(d.i, tenInt, mod)
	if mod.Sign() != 0 {
		return d
	}

	// Binary search for the number of trailing zeros k (known k >= 1).
	lo, hi := 1, d.prec
	for lo < hi {
		mid := (lo + hi + 1) / 2
		quo.QuoRem(d.i, safeGetPrecisionMultiplier(mid), mod)
		if mod.Sign() == 0 {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	value := new(big.Int).Quo(d.i, safeGetPrecisionMultiplier(lo))
	return Decimal{
		i:    value,
		prec: d.prec - lo,
	}
}

// SignificantFigures returns d rounded to figures significant figures.
// It may round within the fractional part or to tens/hundreds on the integer part.
// It panics if figures is not greater than 0.
func (d Decimal) SignificantFigures(figures int, roundingMode RoundingMode) Decimal {
	d = initializeIfNeeded(d)
	if figures <= 0 {
		panic("figures must be greater than 0")
	}
	if d.IsZero() {
		return d
	}

	absD := d.Abs()
	str := absD.String()
	intPart, fracPart, _ := strings.Cut(str, ".")

	targetPrec := 0
	if intPart != "0" {
		targetPrec = figures - len(intPart)
	} else {
		firstNonZero := -1
		for i := 0; i < len(fracPart); i++ {
			if fracPart[i] != '0' {
				firstNonZero = i
				break
			}
		}
		if firstNonZero == -1 {
			return d
		}
		targetPrec = firstNonZero + figures
	}

	if targetPrec >= d.prec {
		return d
	}
	if targetPrec >= 0 {
		return d.Rescale(targetPrec, roundingMode)
	}

	// Round to tens/hundreds/etc. when significant figures fall within the integer part.
	roundShift := d.prec - targetPrec
	rounded := Decimal{
		i:    new(big.Int).Set(d.i),
		prec: roundShift,
	}.round(roundingMode)
	return Decimal{
		i:    new(big.Int).Mul(rounded.i, safeGetPrecisionMultiplier(-targetPrec)),
		prec: 0,
	}
}

// MustNonNegative returns d and panics if d is negative.
func (d Decimal) MustNonNegative() Decimal {
	d = initializeIfNeeded(d)
	return d.requireNonNegative()
}

func (d Decimal) requireNonNegative() Decimal {
	if d.Sign() < 0 {
		panic("Negative value")
	}
	return d
}

// Cmp compares d and d2 and returns:
//
//	-1 if d < d2
//	 0 if d == d2
//	+1 if d > d2
func (d Decimal) Cmp(d2 Decimal) int {
	d1, d2, _ := rescalePair(d, d2)
	return d1.i.Cmp(d2.i)
}

// Equal returns true if d and d2 are equal.
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// NotEqual returns true if d and d2 are not equal.
func (d Decimal) NotEqual(d2 Decimal) bool {
	return d.Cmp(d2) != 0
}

// GT returns true if d is greater than d2.
func (d Decimal) GT(d2 Decimal) bool {
	return d.Cmp(d2) > 0
}

// GTE returns true if d is greater than or equal to d2.
func (d Decimal) GTE(d2 Decimal) bool {
	return d.Cmp(d2) >= 0
}

// LT returns true if d is less than d2.
func (d Decimal) LT(d2 Decimal) bool {
	return d.Cmp(d2) < 0
}

// LTE returns true if d is less than or equal to d2.
func (d Decimal) LTE(d2 Decimal) bool {
	return d.Cmp(d2) <= 0
}

// Sign returns:
//
//	-1 if d < 0
//	 0 if d == 0
//	+1 if d > 0
func (d Decimal) Sign() int {
	d = initializeIfNeeded(d)
	return d.i.Sign()
}

// IsNegative returns true if d is negative.
func (d Decimal) IsNegative() bool {
	return d.Sign() < 0
}

// IsNil returns true if d has no underlying value.
func (d Decimal) IsNil() bool {
	return d.i == nil
}

// IsZero returns true if d is zero or nil.
func (d Decimal) IsZero() bool {
	return d.IsNil() || d.Sign() == 0
}

// IsNotZero returns true if d is not zero.
func (d Decimal) IsNotZero() bool {
	return !d.IsZero()
}

// IsPositive returns true if d is positive.
func (d Decimal) IsPositive() bool {
	return d.Sign() > 0
}

// IsInteger returns true if d has no fractional part.
func (d Decimal) IsInteger() bool {
	_, fractionPart := d.Remainder()
	return fractionPart.Sign() == 0
}

// HasFraction returns true if d has a fractional part.
func (d Decimal) HasFraction() bool {
	return !d.IsInteger()
}

// Neg returns the negated decimal.
func (d Decimal) Neg() Decimal {
	d = initializeIfNeeded(d)
	return Decimal{new(big.Int).Neg(d.i), d.prec}
}

// Abs returns the absolute value of d.
func (d Decimal) Abs() Decimal {
	d = initializeIfNeeded(d)
	if d.IsNegative() {
		return d.Neg()
	}
	// We can return d directly, because there is no way to modify the value of d.i
	return d
}

// BigInt returns a copy of the underlying big.Int value.
func (d Decimal) BigInt() *big.Int {
	d = initializeIfNeeded(d)
	cp := new(big.Int)
	return cp.Set(d.i)
}

// BigRat returns d as an exact rational value.
func (d Decimal) BigRat() *big.Rat {
	d = initializeIfNeeded(d)
	return new(big.Rat).SetFrac(
		new(big.Int).Set(d.i),
		new(big.Int).Set(safeGetPrecisionMultiplier(d.prec)),
	)
}

// Float64 returns the nearest float64 value for d and whether it is exact.
func (d Decimal) Float64() (float64, bool) {
	d = initializeIfNeeded(d)
	rat := new(big.Rat).SetFrac(
		new(big.Int).Set(d.i),
		new(big.Int).Set(safeGetPrecisionMultiplier(d.prec)),
	)
	return rat.Float64()
}

// Float32 returns the nearest float32 value for d and whether it is exact.
func (d Decimal) Float32() (float32, bool) {
	d = initializeIfNeeded(d)
	rat := new(big.Rat).SetFrac(
		new(big.Int).Set(d.i),
		new(big.Int).Set(safeGetPrecisionMultiplier(d.prec)),
	)
	return rat.Float32()
}

// Int64 returns d as an int64 if it is an exact integer in range.
func (d Decimal) Int64() (int64, bool) {
	intPart, fractionPart := d.Remainder()
	if fractionPart.Sign() != 0 || !intPart.IsInt64() {
		return 0, false
	}
	return intPart.Int64(), true
}

// Uint64 returns d as a uint64 if it is a non-negative exact integer in range.
func (d Decimal) Uint64() (uint64, bool) {
	intPart, fractionPart := d.Remainder()
	if fractionPart.Sign() != 0 || intPart.Sign() < 0 || !intPart.IsUint64() {
		return 0, false
	}
	return intPart.Uint64(), true
}

// BitLen returns the bit length of d's underlying integer representation.
func (d Decimal) BitLen() int {
	d = initializeIfNeeded(d)
	return d.i.BitLen()
}

// Precision returns the number of decimal places in d.
func (d Decimal) Precision() int {
	return d.prec
}

// Max returns the larger of d and d2.
func (d Decimal) Max(d2 Decimal) Decimal {
	if d.GT(d2) {
		return d
	}
	return d2
}

// Min returns the smaller of d and d2.
func (d Decimal) Min(d2 Decimal) Decimal {
	if d.LT(d2) {
		return d
	}
	return d2
}

func rescalePair(d1, d2 Decimal) (rescaledD1, rescaledD2 Decimal, maxPrec int) {
	d1 = initializeIfNeeded(d1)
	d2 = initializeIfNeeded(d2)
	maxPrec = max(d1.prec, d2.prec)
	rescaledD1 = d1.RescaleDown(maxPrec)
	rescaledD2 = d2.RescaleDown(maxPrec)
	return
}

func initializeIfNeeded(value Decimal) Decimal {
	if value.i == nil {
		return Zero
	} else {
		return value
	}
}

func requireNonNegativePrecision(precision int) {
	if precision < 0 {
		panic("negative precision")
	}
}

func countFactor(value *big.Int, factor int64) (count int, remainder *big.Int) {
	remainder = new(big.Int).Set(value)
	divisor := big.NewInt(factor)
	mod := new(big.Int)

	for {
		mod.Mod(remainder, divisor)
		if mod.Sign() != 0 {
			return
		}
		remainder.Quo(remainder, divisor)
		count++
	}
}

// Clone returns a deep copy of d. The returned Decimal shares no mutable state
// with d, so mutating the underlying big.Int of either value will not affect
// the other.
func (d Decimal) Clone() Decimal {
	if d.i == nil {
		return Decimal{}
	}
	return Decimal{i: new(big.Int).Set(d.i), prec: d.prec}
}

// Format implements fmt.Formatter, providing support for the common numeric
// verbs along with width, precision, and flag handling.
//
// Supported verbs:
//
//	%v, %s  the canonical string form (same as String)
//	%q      the canonical string form wrapped in quotes
//	%d      the integer value; only valid for integer Decimals. Non-integer
//	        values produce the %!d(decimal.Decimal=<string>) error marker so
//	        callers do not silently truncate fractional digits.
//	%f      fixed-point notation; precision selects fractional digits
//	%e, %E  scientific notation (default precision 6)
//	%g, %G  the shorter of %e or %f for the given precision
//	%b      binary representation of the unscaled big.Int with a scale tag
//
// Width, precision, '-', '+', ' ', and '0' flags are honored where they make
// sense for the chosen verb.
func (d Decimal) Format(f fmt.State, verb rune) {
	dInit := initializeIfNeeded(d)

	// Build the core string representation for the verb.
	var core string
	switch verb {
	case 'v', 's':
		core = dInit.String()
	case 'q':
		core = strconv.Quote(dInit.String())
	case 'd':
		if !dInit.IsInteger() {
			fmt.Fprintf(f, "%%!d(decimal.Decimal=%s)", dInit.String())
			return
		}
		intPart, _ := dInit.Remainder()
		core = intPart.String()
		if prec, ok := f.Precision(); ok {
			neg := strings.HasPrefix(core, "-")
			digits := core
			if neg {
				digits = digits[1:]
			}
			if prec == 0 && digits == "0" {
				digits = ""
			} else if len(digits) < prec {
				digits = strings.Repeat("0", prec-len(digits)) + digits
			}
			if neg {
				core = "-" + digits
			} else {
				core = digits
			}
		}
	case 'f', 'F':
		if prec, ok := f.Precision(); ok {
			rescaled := dInit.Rescale(prec, RoundHalfEven)
			core = rescaled.StringWithTrailingZeros()
			if prec == 0 && f.Flag('#') {
				core += "."
			}
		} else {
			core = dInit.StringWithTrailingZeros()
			if f.Flag('#') && !strings.Contains(core, ".") {
				core += "."
			}
		}
	case 'e', 'E':
		prec := 6
		if p, ok := f.Precision(); ok {
			prec = p
		}
		core = formatScientific(dInit, prec, verb == 'E')
		if prec == 0 && f.Flag('#') {
			if eIdx := strings.IndexAny(core, "eE"); eIdx >= 0 {
				core = core[:eIdx] + "." + core[eIdx:]
			}
		}
	case 'g', 'G':
		prec := -1
		if p, ok := f.Precision(); ok {
			prec = p
		}
		eVerb := 'e'
		if verb == 'G' {
			eVerb = 'E'
		}
		core = formatG(dInit, prec, eVerb, f.Flag('#'))
	case 'b':
		core = dInit.i.Text(2) + "p-" + strconv.Itoa(dInit.prec)
	default:
		fmt.Fprintf(f, "%%!%c(decimal.Decimal=%s)", verb, dInit.String())
		return
	}

	// Apply sign flags for numeric verbs. %v, %s, %q do not receive a leading
	// '+' or ' ' for non-negatives.
	switch verb {
	case 'd', 'f', 'F', 'e', 'E', 'g', 'G', 'b':
		core = applySignFlags(core, f)
	}

	// Apply width/padding.
	writePadded(f, core, verb)
}

// applySignFlags prepends a '+' or ' ' for non-negative numeric values when the
// corresponding flag is set on f. It leaves already-signed strings untouched,
// and leaves empty strings empty (matching fmt's handling of %.0d on zero).
func applySignFlags(s string, f fmt.State) string {
	if s == "" {
		return s
	}
	if s[0] == '-' {
		return s
	}
	if f.Flag('+') {
		return "+" + s
	}
	if f.Flag(' ') {
		return " " + s
	}
	return s
}

// writePadded writes s to f honoring the width, '-' (left-align), and '0'
// (zero-pad) flags. Zero padding is only applied for numeric verbs when the
// '-' flag is not set.
func writePadded(f fmt.State, s string, verb rune) {
	width, hasWidth := f.Width()
	if !hasWidth || len(s) >= width {
		fmt.Fprint(f, s)
		return
	}

	pad := width - len(s)
	leftAlign := f.Flag('-')
	zeroPad := f.Flag('0') && !leftAlign && isNumericVerb(verb)
	// Match fmt: for %d, an explicit precision suppresses the '0' flag
	// (precision already performs digit-level zero padding).
	if zeroPad && verb == 'd' {
		if _, hasPrec := f.Precision(); hasPrec {
			zeroPad = false
		}
	}

	if leftAlign {
		fmt.Fprint(f, s)
		fmt.Fprint(f, strings.Repeat(" ", pad))
		return
	}

	if zeroPad {
		// Keep leading sign (if present) at the front when zero padding.
		if len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == ' ') {
			fmt.Fprint(f, string(s[0]))
			fmt.Fprint(f, strings.Repeat("0", pad))
			fmt.Fprint(f, s[1:])
			return
		}
		fmt.Fprint(f, strings.Repeat("0", pad))
		fmt.Fprint(f, s)
		return
	}

	fmt.Fprint(f, strings.Repeat(" ", pad))
	fmt.Fprint(f, s)
}

func isNumericVerb(verb rune) bool {
	switch verb {
	case 'd', 'f', 'F', 'e', 'E', 'g', 'G', 'b':
		return true
	}
	return false
}

// formatScientific formats d in scientific notation with the given number of
// fractional digits. The exponent is always emitted with at least two digits
// and an explicit sign, matching fmt's %e style.
func formatScientific(d Decimal, prec int, upper bool) string {
	if prec < 0 {
		prec = 0
	}
	if d.IsZero() {
		mantissa := "0"
		if prec > 0 {
			mantissa += "." + strings.Repeat("0", prec)
		}
		return mantissa + expString(0, upper)
	}

	neg := d.IsNegative()
	abs := d.Abs()

	// Determine the exponent: the position of the most significant digit
	// relative to the decimal point. Work from the canonical string which
	// already strips trailing zeros but preserves leading structure.
	digits, exp := significandAndExp(abs)

	// Round digits to the desired precision (prec digits after the first).
	digits, exp = roundSignificand(digits, exp, prec)

	// Build mantissa "d.dddd".
	var mantissa string
	if prec == 0 {
		mantissa = string(digits[0])
	} else {
		frac := digits[1:]
		if len(frac) < prec {
			frac += strings.Repeat("0", prec-len(frac))
		}
		mantissa = string(digits[0]) + "." + frac
	}

	sign := ""
	if neg {
		sign = "-"
	}
	return sign + mantissa + expString(exp, upper)
}

// significandAndExp returns the decimal digit string (without a decimal point
// or sign) of abs and the exponent such that the value equals 0.d... * 10^(exp+1).
// For convenience the returned digits start with the most significant digit
// and exp is the base-10 exponent of that digit.
func significandAndExp(abs Decimal) (digits string, exp int) {
	// Use canonical form (trailing zeros removed) for computing the
	// significand; this avoids spurious trailing zeros in the mantissa.
	s := abs.StripTrailingZeros().String()

	intPart, fracPart, hasDot := strings.Cut(s, ".")
	if !hasDot {
		// Pure integer.
		digits = strings.TrimLeft(intPart, "0")
		if digits == "" {
			digits = "0"
			exp = 0
			return
		}
		exp = len(digits) - 1
		return
	}

	if intPart != "0" {
		digits = intPart + fracPart
		// Strip leading zeros (shouldn't usually occur unless input was odd).
		trimmed := strings.TrimLeft(digits, "0")
		leadingZeros := len(digits) - len(trimmed)
		if trimmed == "" {
			digits = "0"
			exp = 0
			return
		}
		digits = trimmed
		exp = len(intPart) - 1 - leadingZeros
		return
	}

	// Integer part is "0"; count leading zeros in the fractional part.
	trimmed := strings.TrimLeft(fracPart, "0")
	if trimmed == "" {
		digits = "0"
		exp = 0
		return
	}
	leadingZeros := len(fracPart) - len(trimmed)
	digits = trimmed
	exp = -(leadingZeros + 1)
	return
}

// roundSignificand rounds digits to keep (prec+1) digits (one before the
// decimal point, prec after) using RoundHalfEven. It returns the possibly
// adjusted digit string and exponent.
func roundSignificand(digits string, exp, prec int) (string, int) {
	keep := prec + 1
	if len(digits) <= keep {
		return digits, exp
	}

	head := digits[:keep]
	tail := digits[keep:]

	// Determine rounding direction.
	roundUp := false
	first := tail[0]
	switch {
	case first < '5':
		roundUp = false
	case first > '5':
		roundUp = true
	default:
		// first == '5'; check remaining tail for non-zero.
		nonZero := false
		for i := 1; i < len(tail); i++ {
			if tail[i] != '0' {
				nonZero = true
				break
			}
		}
		if nonZero {
			roundUp = true
		} else {
			// Exactly halfway: round to even.
			lastKept := head[len(head)-1]
			if (lastKept-'0')%2 == 1 {
				roundUp = true
			}
		}
	}

	if !roundUp {
		return head, exp
	}

	// Propagate rounding carry through head.
	bs := []byte(head)
	for i := len(bs) - 1; i >= 0; i-- {
		if bs[i] < '9' {
			bs[i]++
			return string(bs), exp
		}
		bs[i] = '0'
	}
	// Carry out past the most significant digit.
	return "1" + string(bs[:len(bs)-1]) + strings.Repeat("0", 0), exp + 1
}

// expString returns the exponent suffix in the form "e±DD" (or "E±DD").
func expString(exp int, upper bool) string {
	marker := "e"
	if upper {
		marker = "E"
	}
	sign := "+"
	if exp < 0 {
		sign = "-"
		exp = -exp
	}
	if exp < 10 {
		return marker + sign + "0" + strconv.Itoa(exp)
	}
	return marker + sign + strconv.Itoa(exp)
}

// formatG renders d following Go's %g / %G conventions.
//
// For prec < 0 (default), the branch decision uses fmt's fixed eprec=6
// threshold regardless of the value's digit count, and the fixed-point
// branch emits the canonical (trailing-zero-stripped) string. The
// scientific branch uses the value's minimal significand so integers like
// 1000000 render as "1e+06" while 12345678 renders as "1.2345678e+07".
//
// For prec >= 0, precision is interpreted as the number of significant
// digits. Scientific form is used when the decimal exponent is below -4 or
// not less than the significant-digit count.
func formatG(d Decimal, prec int, eVerb rune, hashFlag bool) string {
	upper := eVerb == 'E'

	if d.IsZero() {
		if hashFlag {
			n := prec
			if n < 0 {
				n = 6
			}
			if n == 0 {
				n = 1
			}
			if n == 1 {
				return "0."
			}
			return "0." + strings.Repeat("0", n-1)
		}
		return "0"
	}

	_, exp := significandAndExp(d.Abs())

	// Default prec, no # flag: use the shortest-unique representation rules
	// described on the function comment (eprec=6 for the branch decision).
	if prec < 0 && !hashFlag {
		const defaultEprec = 6
		if exp < -4 || exp >= defaultEprec {
			minSig := minimalSignificantDigits(d)
			return stripGScientific(formatScientific(d, minSig-1, upper))
		}
		return d.String()
	}

	// Either precision was specified, or the # flag is set — use N
	// significant digits, keeping trailing zeros when # is present.
	var n int
	switch {
	case prec < 0:
		n = 6
	case prec == 0:
		n = 1
	default:
		n = prec
	}

	if exp < -4 || exp >= n {
		s := formatScientific(d, n-1, upper)
		if hashFlag {
			if n == 1 {
				if eIdx := strings.IndexAny(s, "eE"); eIdx >= 0 && !strings.Contains(s[:eIdx], ".") {
					s = s[:eIdx] + "." + s[eIdx:]
				}
			}
			return s
		}
		return stripGScientific(s)
	}

	fractional := n - 1 - exp
	if fractional < 0 {
		fractional = 0
	}
	rescaled := d.Rescale(fractional, RoundHalfEven)
	if hashFlag {
		core := rescaled.StringWithTrailingZeros()
		if fractional == 0 && !strings.Contains(core, ".") {
			core += "."
		}
		return core
	}
	return rescaled.String()
}

// minimalSignificantDigits returns the number of leading digits in |d|
// required to uniquely represent its value (trailing zeros in both the
// integer and fractional parts are removed). For 1000000 the result is 1;
// for 12345 it is 5; for 1.50 it is 2.
func minimalSignificantDigits(d Decimal) int {
	if d.IsZero() {
		return 1
	}
	s := new(big.Int).Abs(d.i).Text(10)
	s = strings.TrimRight(s, "0")
	if s == "" {
		return 1
	}
	return len(s)
}

// stripGScientific removes trailing zeros from the mantissa of a scientific
// string produced by formatScientific so that %g's scientific branch follows
// fmt's "no redundant zeros" rule (e.g. "1.200e+04" -> "1.2e+04").
func stripGScientific(s string) string {
	eIdx := strings.IndexAny(s, "eE")
	if eIdx < 0 {
		return s
	}
	mantissa := s[:eIdx]
	suffix := s[eIdx:]
	if !strings.Contains(mantissa, ".") {
		return s
	}
	mantissa = strings.TrimRight(mantissa, "0")
	mantissa = strings.TrimRight(mantissa, ".")
	return mantissa + suffix
}

// FormatWithSeparators returns d formatted with the supplied thousands and
// decimal separators. A thousands value of 0 disables grouping. A decimal
// value of 0 is treated as a zero rune and falls back to '.' so the output
// is always a syntactically valid decimal literal. The fractional part keeps
// its full precision as produced by StringWithTrailingZeros.
func (d Decimal) FormatWithSeparators(thousands, decimal rune) string {
	if decimal == 0 {
		decimal = '.'
	}

	base := d.StringWithTrailingZeros()

	neg := false
	if len(base) > 0 && base[0] == '-' {
		neg = true
		base = base[1:]
	}

	intPart, fracPart, hasDot := strings.Cut(base, ".")

	var grouped string
	if thousands == 0 || len(intPart) <= 3 {
		grouped = intPart
	} else {
		var b strings.Builder
		// Leading chunk may be 1, 2, or 3 digits.
		first := len(intPart) % 3
		if first == 0 {
			first = 3
		}
		b.WriteString(intPart[:first])
		for i := first; i < len(intPart); i += 3 {
			b.WriteRune(thousands)
			b.WriteString(intPart[i : i+3])
		}
		grouped = b.String()
	}

	var out strings.Builder
	if neg {
		out.WriteByte('-')
	}
	out.WriteString(grouped)
	if hasDot {
		out.WriteRune(decimal)
		out.WriteString(fracPart)
	}
	return out.String()
}
