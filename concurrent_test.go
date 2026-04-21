package decimal_test

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/exc-works/decimal"
)

// TestConcurrentReadsImmutable spins up many goroutines that all share a
// single Decimal value and exercise read-only / value-receiver methods.
// Under the -race detector this will fail if any of these methods mutate
// shared state.
func TestConcurrentReadsImmutable(t *testing.T) {
	d := decimal.MustFromString("123.456789")
	const N = 64

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_ = d.String()
				_ = d.StringWithTrailingZeros()
				_ = d.IntPart()
				_ = d.Precision()
				_ = d.IsZero()
				_ = d.Cmp(decimal.Zero)
				_ = d.Add(decimal.One)
				_ = d.Sub(decimal.One)
				_ = d.Mul(decimal.Ten, decimal.RoundHalfEven)
				_ = d.BigInt()
			}
		}()
	}
	wg.Wait()
}

// TestConcurrentOperationsOnSeparateValues verifies that arithmetic on
// independently-constructed Decimal values produces the expected results
// when executed concurrently.
func TestConcurrentOperationsOnSeparateValues(t *testing.T) {
	const N = 64
	const iters = 500

	var wg sync.WaitGroup
	errCh := make(chan error, N*iters)

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			base := decimal.NewFromInt(id + 1)
			for j := 0; j < iters; j++ {
				step := decimal.NewFromInt(j + 1)

				sum := base.Add(step)
				expectedSum := decimal.NewFromInt(id + 1 + j + 1)
				if sum.Cmp(expectedSum) != 0 {
					errCh <- fmt.Errorf("goroutine %d iter %d: Add = %s, want %s",
						id, j, sum.String(), expectedSum.String())
					return
				}

				diff := sum.Sub(step)
				if diff.Cmp(base) != 0 {
					errCh <- fmt.Errorf("goroutine %d iter %d: Sub = %s, want %s",
						id, j, diff.String(), base.String())
					return
				}

				product := base.Mul(decimal.Ten, decimal.RoundHalfEven)
				expectedProduct := decimal.NewFromInt((id + 1) * 10)
				if product.Cmp(expectedProduct) != 0 {
					errCh <- fmt.Errorf("goroutine %d iter %d: Mul = %s, want %s",
						id, j, product.String(), expectedProduct.String())
					return
				}

				// Make sure the base receiver was not mutated by the
				// value-receiver arithmetic above.
				expectedBase := decimal.NewFromInt(id + 1)
				if base.Cmp(expectedBase) != 0 {
					errCh <- fmt.Errorf("goroutine %d iter %d: base mutated to %s, want %s",
						id, j, base.String(), expectedBase.String())
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestBigIntReturnsCopy guards against regressions in BigInt()'s defensive copy.
// Mutating the returned *big.Int must not be observable through the receiver's
// String(), because the caller should only hold an independent copy.
func TestBigIntReturnsCopy(t *testing.T) {
	d := decimal.MustFromString("1234.56")
	before := d.String()

	got := d.BigInt()
	got.Add(got, big.NewInt(999))

	if after := d.String(); after != before {
		t.Fatalf("BigInt() returned a shared pointer: String() went from %q to %q after mutating the returned big.Int",
			before, after)
	}
}

// TestBigRatReturnsCopy guards against regressions in BigRat()'s defensive
// copy. Mutating the returned *big.Rat must not be observable through the
// receiver's String().
func TestBigRatReturnsCopy(t *testing.T) {
	d := decimal.MustFromString("1234.56")
	before := d.String()

	got := d.BigRat()
	// Mutate numerator and denominator in place.
	got.Num().Add(got.Num(), big.NewInt(999))
	got.Denom().Add(got.Denom(), big.NewInt(7))

	if after := d.String(); after != before {
		t.Fatalf("BigRat() returned a shared pointer: String() went from %q to %q after mutating the returned big.Rat",
			before, after)
	}
}

// TestConcurrentJSONMarshal exercises the read-only MarshalJSON path on a
// single shared Decimal from many goroutines at once. All marshaled results
// must be identical, and none of them may race against the others.
func TestConcurrentJSONMarshal(t *testing.T) {
	d := decimal.MustFromString("987654321.0123456789")
	want, err := d.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() baseline failed: %v", err)
	}

	const N = 64
	const iters = 500

	var wg sync.WaitGroup
	errCh := make(chan error, N)

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				got, err := d.MarshalJSON()
				if err != nil {
					errCh <- fmt.Errorf("goroutine %d iter %d: MarshalJSON error: %v",
						id, j, err)
					return
				}
				if string(got) != string(want) {
					errCh <- fmt.Errorf("goroutine %d iter %d: MarshalJSON = %s, want %s",
						id, j, string(got), string(want))
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}
