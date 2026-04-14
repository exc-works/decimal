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

	Zero    = New(0)
	One     = New(1)
	Ten     = New(10)
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

func New(value int64) Decimal {
	return NewFromBigInt(big.NewInt(value))
}

func NewFromInt(value int) Decimal {
	return NewFromBigInt(big.NewInt(int64(value)))
}

func NewWithPrec(value int64, prec int) Decimal {
	requireNonNegativePrecision(prec)
	return NewFromBigIntWithPrec(big.NewInt(value), prec)
}

func NewFromFloat64(value float64) Decimal {
	return MustFromString(strconv.FormatFloat(value, 'f', -1, 64))
}

// NewWithAppendPrec create a new Decimal from value, and append number of zeros to make it fit the required precision
// If `value` is 1, `prec` is 2, then return 1.00.
// If `value` is 1, `prec` is 18, then return 1.000000000000000000
func NewWithAppendPrec(value int64, prec int) Decimal {
	requireNonNegativePrecision(prec)
	return Decimal{
		i:    new(big.Int).Mul(big.NewInt(value), safeGetPrecisionMultiplier(prec)),
		prec: prec,
	}
}

func NewFromUintWithAppendPrec(value uint64, prec int) Decimal {
	requireNonNegativePrecision(prec)
	return Decimal{
		i:    new(big.Int).Mul(new(big.Int).SetUint64(value), safeGetPrecisionMultiplier(prec)),
		prec: prec,
	}
}

// NewFromBigInt create a new Decimal from big integer assuming whole numbers
func NewFromBigInt(value *big.Int) Decimal {
	return NewFromBigIntWithPrec(value, 0)
}

// NewFromBigIntWithPrec create a new Decimal from big integer assuming whole numbers
func NewFromBigIntWithPrec(value *big.Int, precision int) Decimal {
	requireNonNegativePrecision(precision)
	return Decimal{
		i:    new(big.Int).Set(value),
		prec: precision,
	}
}

func NewFromInt64(value int64, precision int) Decimal {
	requireNonNegativePrecision(precision)
	return Decimal{
		i:    new(big.Int).SetInt64(value),
		prec: precision,
	}
}

// NewFromUint64 create a new Decimal from uint64 value.
func NewFromUint64(value uint64, precision int) Decimal {
	requireNonNegativePrecision(precision)
	return Decimal{
		i:    new(big.Int).SetUint64(value),
		prec: precision,
	}
}

// NewFromString create a new Decimal from decimal string.
// valid must come in the form:
//
//	(-) whole integers (.) decimal integers
//
// examples of acceptable input include:
//
//	-123.456
//	456.7890
//	345
//	-456789
//	1.23456e3
//	1.23456E3
//	123456e-3
//	123456E-3
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
func NewFromString(str string) (d Decimal, err error) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return Decimal{}, errors.New("decimal string cannot be empty")
	}

	// Parse scientific notation first
	var expOffset int64 = 0
	eIndex := strings.IndexAny(str, "Ee")
	if eIndex != -1 {
		// Parse the exponent
		expStr := str[eIndex+1:]
		if len(expStr) == 0 {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: missing exponent", str)
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
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: missing exponent value", str)
		}

		expInt, err := strconv.ParseInt(expStr, 10, 32)
		if err != nil {
			var e *strconv.NumError
			if errors.As(err, &e) && errors.Is(e.Err, strconv.ErrRange) {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent too large", str)
			}
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", str)
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
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string", str)
	}

	// Parse the mantissa (number part without exponent)
	var precision int
	strs := strings.Split(str, ".")
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		precision = len(strs[1])
		// Maintain backward compatibility: reject formats like "1." and ".1"
		if len(combinedStr) == 0 || precision == 0 {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string", str)
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string", str)
	}

	if combinedStr == "" {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: invalid decimal string", str)
	}

	// Apply exponent offset to precision
	precision -= int(expOffset)

	// Parse the combined string as big.Int first
	combined, ok := new(big.Int).SetString(combinedStr, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("failed to set decimal string: %s", combinedStr)
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

func MustFromString(str string) Decimal {
	d, err := NewFromString(str)
	if err != nil {
		panic(err)
	}
	return d
}

func (d Decimal) Add(d2 Decimal) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Add(d1.i, d2.i),
		prec: maxPrec,
	}
}

func (d Decimal) SafeAdd(d2 Decimal) Decimal {
	return d.Add(d2).requireNonNegative()
}

func (d Decimal) AddRaw(i int64) Decimal {
	d = initializeIfNeeded(d)
	return Decimal{
		i:    new(big.Int).Add(d.i, big.NewInt(i)),
		prec: d.prec,
	}
}

func (d Decimal) Sub(d2 Decimal) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Sub(d1.i, d2.i),
		prec: maxPrec,
	}
}

func (d Decimal) SafeSub(d2 Decimal) Decimal {
	return d.Sub(d2).requireNonNegative()
}

func (d Decimal) SubRaw(i int64) Decimal {
	d = initializeIfNeeded(d)
	return Decimal{
		i:    new(big.Int).Sub(d.i, big.NewInt(i)),
		prec: d.prec,
	}
}

// Mul multiplies two decimals and returns the result.
//
// The precision of the result is the maximum of the precisions of the two decimals.
func (d Decimal) Mul(d2 Decimal, roundingMode RoundingMode) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Mul(d1.i, d2.i),
		prec: maxPrec,
	}.round(roundingMode)
}

func (d Decimal) MulDown(d2 Decimal) Decimal {
	return d.Mul(d2, RoundDown)
}

// Mul2 multiplies two decimals and returns the result.
//
// The precision of the result is the sum of the precisions of the two decimals.
func (d Decimal) Mul2(d2 Decimal) Decimal {
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

// QuoWithPrec divides two decimals and returns the result with the specified precision.
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

func (d Decimal) QuoDown(d2 Decimal) Decimal {
	return d.Quo(d2, RoundDown)
}

// IntPart returns integer part.
func (d Decimal) IntPart() *big.Int {
	intPart, _ := d.Remainder()
	return intPart
}

// Remainder returns integer part and fractional part.
func (d Decimal) Remainder() (intPart *big.Int, fractionPart *big.Int) {
	d = initializeIfNeeded(d)
	return new(big.Int).QuoRem(d.i, safeGetPrecisionMultiplier(d.prec), new(big.Int))
}

// Power returns a result of raising to integer power.
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
			tmp = tmp.Mul2(resultD)
		}
		i /= 2
		resultD = resultD.Mul2(resultD)
	}
	return resultD.Mul2(tmp).Rescale(d.prec, RoundHalfEven)
}

// Sqrt returns the square root using ApproxRoot(2).
// It returns an error for negative inputs.
func (d Decimal) Sqrt() (guess Decimal, err error) {
	return d.ApproxRoot(2)
}

func (d Decimal) ApproxRoot(root int64) (guess Decimal, err error) {
	if root <= 0 {
		return Decimal{}, fmt.Errorf("root must be greater than 0")
	}

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = errors.New("out of bounds")
			}
		}
	}()

	d = initializeIfNeeded(d)
	if d.IsNegative() {
		if root%2 == 0 {
			return Decimal{}, fmt.Errorf("cannot take even root of negative value")
		}
		absRoot, err := d.Neg().ApproxRoot(root)
		return absRoot.Neg(), err
	}

	if root == 1 || d.IsZero() || d.Equal(One) {
		return d, nil
	}

	rootInt := big.NewInt(0).SetInt64(root)
	guess = NewWithAppendPrec(1, d.prec)
	delta := guess

	for iter := 0; delta.Abs().i.Cmp(oneInt) > 0 && iter < maxIterations; iter++ {
		prev := guess.Power(root - 1)
		if prev.IsZero() {
			prev = One
		}
		delta = d.Quo(prev, RoundHalfEven)
		delta = delta.Sub(guess)

		quo := new(big.Int).Quo(delta.i, rootInt)
		delta = Decimal{i: quo, prec: d.prec}

		guess = guess.Add(delta)
	}
	return
}

// Log2 returns log2.
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

func (d Decimal) RescaleDown(prec int) Decimal {
	d = initializeIfNeeded(d)
	return d.Rescale(prec, RoundDown)
}

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

// SignificantFigures returns a Decimal with the specified number of significant figures
func (d Decimal) SignificantFigures(figures int, roundingMode RoundingMode) Decimal {
	d = initializeIfNeeded(d)
	if figures <= 0 {
		panic("figures must be greater than 0")
	}
	if d.prec == 0 || d.prec <= figures {
		return d
	}

	absD := d.Abs()
	str := absD.String()
	splits := strings.Split(str, ".")
	if splits[0] != "0" {
		figures -= len(splits[0])
		if figures < 0 {
			figures = 0
		}
		return d.Rescale(min(figures, d.prec), roundingMode)
	} else {
		for i := 0; i < len(splits[1]); i++ {
			if splits[1][i] != '0' {
				figures += i
				return d.Rescale(min(figures, d.prec), roundingMode)
			}
		}
	}
	return d
}

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

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (d Decimal) Cmp(d2 Decimal) int {
	d1, d2, _ := rescalePair(d, d2)
	return d1.i.Cmp(d2.i)
}

// Equal returns equal other value
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// NotEqual returns not equal other value
func (d Decimal) NotEqual(d2 Decimal) bool {
	return d.Cmp(d2) != 0
}

// GT greater than other value
func (d Decimal) GT(d2 Decimal) bool {
	return d.Cmp(d2) > 0
}

// GTE greater than or equal other value
func (d Decimal) GTE(d2 Decimal) bool {
	return d.Cmp(d2) >= 0
}

// LT less than other value
func (d Decimal) LT(d2 Decimal) bool {
	return d.Cmp(d2) < 0
}

// LTE less than or equal other value
func (d Decimal) LTE(d2 Decimal) bool {
	return d.Cmp(d2) <= 0
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (d Decimal) Sign() int {
	d = initializeIfNeeded(d)
	return d.i.Sign()
}

// IsNegative returns is negative value
func (d Decimal) IsNegative() bool {
	return d.Sign() < 0
}

// IsNil returns true if the decimal is nil
func (d Decimal) IsNil() bool {
	return d.i == nil
}

// IsZero returns true if the decimal is zero or nil
func (d Decimal) IsZero() bool {
	return d.IsNil() || d.Sign() == 0
}

// IsNotZero returns true if the decimal is not zero
func (d Decimal) IsNotZero() bool {
	return !d.IsZero()
}

// IsPositive returns is positive value
func (d Decimal) IsPositive() bool {
	return d.Sign() > 0
}

// Neg reverse the decimal sign
func (d Decimal) Neg() Decimal {
	d = initializeIfNeeded(d)
	return Decimal{new(big.Int).Neg(d.i), d.prec}
}

// Abs returns absolute value
func (d Decimal) Abs() Decimal {
	d = initializeIfNeeded(d)
	if d.IsNegative() {
		return d.Neg()
	}
	// We can return d directly, because there is no way to modify the value of d.i
	return d
}

// BigInt returns a copy of the underlying big.Int.
func (d Decimal) BigInt() *big.Int {
	d = initializeIfNeeded(d)
	cp := new(big.Int)
	return cp.Set(d.i)
}

func (d Decimal) BitLen() int {
	d = initializeIfNeeded(d)
	return d.i.BitLen()
}

func (d Decimal) Precision() int {
	return d.prec
}

func (d Decimal) Max(d2 Decimal) Decimal {
	if d.GT(d2) {
		return d
	}
	return d2
}

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
