package decimal

import (
	"errors"
	"math/big"
	"testing"
)

func TestZivRefinement(t *testing.T) {
	t.Run("refines internal precision and preserves output scale", func(t *testing.T) {
		attempts, previousBits := 0, 0
		result, err := certifyDecimal(4, 0, func(scale *big.Int, bits int) (fixedInterval, error) {
			attempts++
			if bits <= previousBits {
				t.Fatal("working precision did not increase")
			}
			previousBits = bits
			if attempts < 3 {
				// Both intervals enclose 1.23456; the first two are too wide
				// to determine its rounding to four decimal places.
				return fixedInterval{new(big.Int).Set(scale), new(big.Int).Lsh(scale, 1)}, nil
			}
			return fractionBounds(big.NewInt(123456), big.NewInt(100000), scale), nil
		})
		if err != nil || attempts != 3 || result.Precision() != 4 || !result.Equal(MustFromString("1.2346")) {
			t.Fatalf("got %s, scale %d, attempts %d, error %v", result, result.Precision(), attempts, err)
		}
	})
	t.Run("ambiguous enclosures exhaust the bounded budget", func(t *testing.T) {
		attempts := 0
		result, err := certifyDecimal(0, 0, func(scale *big.Int, _ int) (fixedInterval, error) {
			attempts++
			// The center is always exactly 1/2. Agreement of successive
			// approximations would be misleading: every enclosure still
			// contains values rounding to either zero or one.
			half := new(big.Int).Rsh(scale, 1)
			return fixedInterval{new(big.Int).Sub(half, oneInt), new(big.Int).Add(half, oneInt)}, nil
		})
		if !errors.Is(err, ErrNoConvergence) || attempts != 8 || !result.IsNil() {
			t.Fatalf("result %s, attempts %d, error %v", result, attempts, err)
		}
	})
	t.Run("evaluation errors stop immediately", func(t *testing.T) {
		attempts := 0
		wantErr := errors.New("enclosure calculation failed")
		_, err := certifyDecimal(4, 0, func(*big.Int, int) (fixedInterval, error) { attempts++; return fixedInterval{}, wantErr })
		if !errors.Is(err, wantErr) || attempts != 1 {
			t.Fatalf("attempts %d, error %v", attempts, err)
		}
	})
}

func TestRoundBoundsUsesEntireEnclosure(t *testing.T) {
	for _, tc := range []struct {
		lo, hi, scale int64
		prec          int
		want          string
		ok            bool
	}{
		{123454, 123456, 100000, 4, "", false},
		{1234551, 1234552, 1000000, 4, "1.2346", true},
		{-123456, -123454, 100000, 4, "", false},
		{-1234552, -1234551, 1000000, 4, "-1.2346", true},
		{125, 125, 100, 1, "1.2", true},
		{135, 135, 100, 1, "1.4", true},
	} {
		result, ok := roundBounds(fixedInterval{big.NewInt(tc.lo), big.NewInt(tc.hi)}, big.NewInt(tc.scale), tc.prec)
		if ok != tc.ok || (ok && !result.Equal(MustFromString(tc.want))) {
			t.Fatalf("[%d,%d]/%d: result %s, accepted %v", tc.lo, tc.hi, tc.scale, result, ok)
		}
	}
}
