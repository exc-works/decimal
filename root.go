package decimal

import (
	"fmt"
	"math/big"
	"math/bits"
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

// integerRoot returns floor(x^(1/n)). Newton iterates descend from an exact
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
	exponent, degree := big.NewInt(n-1), big.NewInt(n)
	guess := rootUpperBound(x, n, degree)
	next := new(big.Int)
	var power, weighted big.Int
	for range maxIterations {
		power.Exp(guess, exponent, nil)
		next.Quo(x, &power)
		weighted.Mul(guess, exponent)
		next.Add(next, &weighted).Quo(next, degree)
		if next.Cmp(guess) >= 0 {
			return guess, true
		}
		guess, next = next, guess
	}
	return nil, false
}

// rootUpperBound resolves a short prefix of the root with exact integer
// comparisons before Newton refinement. A plain power-of-two upper bound can
// be nearly twice the root; at high degrees Newton then shrinks it by only
// about 1/n per step. Resolving bitlen(n)+1 prefix bits instead puts its
// relative error below 1/n (or within one unit for a small integer root).
func rootUpperBound(x *big.Int, n int64, degree *big.Int) *big.Int {
	rootBits := (x.BitLen()-1)/int(n) + 1
	prefixBits := min(rootBits, bits.Len64(uint64(n))+1)
	shift := rootBits - prefixBits
	// shift*n <= x.BitLen()-1. Public root budgets also bound n, so the
	// prefix has at most 19 bits and fits native arithmetic.
	prefix := new(big.Int).Rsh(x, uint(shift*int(n)))
	lo, hi := uint64(1)<<uint(prefixBits-1), uint64(1)<<uint(prefixBits)
	var candidate, power big.Int
	for hi-lo > 1 {
		mid := lo + (hi-lo)/2
		candidate.SetUint64(mid)
		power.Exp(&candidate, degree, nil)
		if power.Cmp(prefix) <= 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	// prefix < hi^n and x < (prefix+1)*2^(shift*n), hence this is
	// strictly above the exact root. No floating-point estimate is used.
	guess := new(big.Int).SetUint64(hi)
	return guess.Lsh(guess, uint(shift))
}
