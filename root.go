package decimal

import (
	"fmt"
	"math/big"
)

// roundedIntegerRoot obtains the coefficient below the exact root, then
// compares its exact half-unit boundary to the original input. No rounded
// Newton approximation participates in the rounding decision.
func roundedIntegerRoot(d Decimal, root int64, prec int) (Decimal, error) {
	// Bound exact intermediate scales and powers before allocating them.
	if d.prec > maxParsedPrecision || root > maxParsedPrecision || prec > maxParsedPrecision || root > int64(maxParsedPrecision)/int64(max(1, prec)) {
		return Decimal{}, fmt.Errorf("root calculation exceeds precision budget: %w", ErrNoConvergence)
	}
	shift := int(root)*prec - d.prec
	num, den := new(big.Int).Set(d.i), big.NewInt(1)
	if shift >= 0 {
		num.Mul(num, safeGetPrecisionMultiplier(shift))
	} else {
		den = safeGetPrecisionMultiplier(-shift)
	}
	floor, ok := integerRoot(new(big.Int).Quo(num, den), root)
	if !ok {
		return Decimal{}, fmt.Errorf("integer root iteration limit reached: %w", ErrNoConvergence)
	}
	midpoint := new(big.Int).Add(new(big.Int).Lsh(floor, 1), oneInt)
	right := new(big.Int).Exp(midpoint, big.NewInt(root), nil)
	right.Mul(right, den)
	left := new(big.Int).Lsh(num, uint(root))
	cmp := left.Cmp(right)
	if cmp > 0 || (cmp == 0 && floor.Bit(0) != 0) {
		floor.Add(floor, oneInt)
	}
	return Decimal{i: floor, prec: prec}, nil
}

// integerRoot returns floor(x^(1/n)). Newton iterates descend from a binary
// upper bound. A stationary or increasing next iterate certifies the floor.
func integerRoot(x *big.Int, n int64) (*big.Int, bool) {
	if x.Sign() == 0 {
		return new(big.Int), true
	}
	if n == 2 {
		return new(big.Int).Sqrt(x), true
	}
	if n >= int64(x.BitLen()) {
		return big.NewInt(1), true
	}
	guess := new(big.Int).Lsh(big.NewInt(1), uint((int64(x.BitLen())+n-1)/n))
	exponent, degree := big.NewInt(n-1), big.NewInt(n)
	for range maxIterations {
		power := new(big.Int).Exp(guess, exponent, nil)
		next := new(big.Int).Quo(x, power)
		next.Add(next, new(big.Int).Mul(guess, exponent)).Quo(next, degree)
		if next.Cmp(guess) >= 0 {
			return guess, true
		}
		guess = next
	}
	return nil, false
}
