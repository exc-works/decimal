package decimal

import (
	"fmt"
	"math/big"
)

// This file contains only the shared precision driver. interval.go implements
// enclosure arithmetic and rounding; logarithm.go and exponential.go provide
// the function-specific enclosure algorithms.

// These guard limits are this library's resource policy, not requirements of
// Ziv's strategy. Doubling from 64 through 8192 permits at most eight attempts;
// the total working-precision budget may stop refinement sooner.
const (
	zivInitialGuardBits = 64
	zivMaxGuardBits     = 8192
)

// 3.322 is a rational upper bound for log2(10); no floating-point estimate
// enters the numeric calculation or its precision budget.
func decimalBits(prec int) int { return (prec*3322 + 999) / 1000 }

// certifyDecimal implements a bounded form of Ziv's strategy for correct
// rounding:
//
//  1. Evaluate at the target precision plus guard bits, obtaining a rigorous
//     interval that includes both arithmetic error and the series remainder.
//  2. Round both endpoints to prec decimal places using RoundHalfEven. If they
//     agree, every possible exact value in the interval has that same result.
//  3. Otherwise increase the guard precision and recompute the enclosure.
//
// Only internal working precision increases; the requested output scale stays
// fixed. Agreement between successive approximate values would not certify
// rounding. Success relies on enclose returning a proven enclosure.
//
// enclose receives a borrowed, read-only scale and must return independently
// owned bounds. Its input scale is reused by the next attempt, and bounds from
// different scales must not be mixed. An evaluation error stops immediately;
// exhausted guard/working budgets return ErrNoConvergence without a result.
//
// Background: MPFR's discussion of Ziv's strategy:
// https://www.vinc17.net/research/slides/sieste2010.pdf#page=35
func certifyDecimal(prec, extraDigits int, enclose func(*big.Int, int) (fixedInterval, error)) (Decimal, error) {
	if prec > maxParsedPrecision || extraDigits > maxParsedPrecision-prec {
		return Decimal{}, fmt.Errorf("transcendental precision budget exceeded: %w", ErrNoConvergence)
	}
	baseBits := decimalBits(prec + extraDigits)
	maxBits := decimalBits(maxParsedPrecision)
	multiplier := safeGetPrecisionMultiplier(prec)
	coefficient := new(big.Int)
	var scale big.Int
	var rounding roundingWorkspace
	for guard := zivInitialGuardBits; guard <= zivMaxGuardBits; guard *= 2 {
		bits := baseBits + guard
		if bits > maxBits {
			break
		}
		scale.SetInt64(1).Lsh(&scale, uint(bits))
		// Stage 1: enclose the exact value at this binary working scale.
		bounds, err := enclose(&scale, bits)
		if err != nil {
			return Decimal{}, err
		}
		// Stages 2 and 3: accept a unique decimal rounding, or refine.
		if rounding.roundBoundsTo(coefficient, bounds, &scale, multiplier) {
			return Decimal{i: coefficient, prec: prec}, nil
		}
	}
	return Decimal{}, fmt.Errorf("cannot certify rounding within precision budget: %w", ErrNoConvergence)
}
