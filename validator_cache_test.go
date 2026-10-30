package decimal

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidatorParameterReuseAndBounds(t *testing.T) {
	p := new(decimalValidationParams)
	first := p.decimal("0.1000000000000000000000000000000001")
	second := p.decimal("0.1000000000000000000000000000000001")
	if first.i != second.i {
		t.Fatal("parameter was reparsed")
	}
	lo, hi := p.rangeBounds("0.1~9.99")
	lo2, hi2 := p.rangeBounds("0.1~9.99")
	if lo.i != lo2.i || hi.i != hi2.i {
		t.Fatal("range was reparsed")
	}
	for i := 0; i < validatorParamCacheLimit+20; i++ {
		if !p.decimal(fmt.Sprint(i)).Equal(New(int64(i))) {
			t.Fatal("cache overflow changed comparison parameter")
		}
		lower, upper := p.rangeBounds(fmt.Sprintf("%d~%d", i, i+1))
		if !lower.Equal(New(int64(i))) || !upper.Equal(New(int64(i+1))) {
			t.Fatal("cache overflow changed range")
		}
	}
	if len(p.decimals)+len(p.ranges) > validatorParamCacheLimit {
		t.Fatal("unbounded parameter cache")
	}
	large := new(decimalValidationParams)
	text := "1." + strings.Repeat("0", validatorParamMaxBytes)
	if !large.decimal(text).Equal(One) {
		t.Fatal("large scalar rejected")
	}
	if lower, upper := large.rangeBounds(text + "~2"); !lower.Equal(One) || !upper.Equal(New(2)) {
		t.Fatal("large range rejected")
	}
	if len(large.decimals) != 0 || len(large.ranges) != 0 {
		t.Fatal("oversized key retained")
	}
	other := new(decimalValidationParams)
	if other.decimal("0.1000000000000000000000000000000001").i == first.i {
		t.Fatal("registrations share parameter storage")
	}
}

func TestValidatorMalformedParameterDoesNotPoisonCache(t *testing.T) {
	p := new(decimalValidationParams)
	for _, call := range []func(){func() { p.decimal("bad") }, func() { p.rangeBounds("2~1") }, func() { p.rangeBounds("bad") }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("malformed parameter did not panic")
				}
			}()
			call()
		}()
	}
	if !p.decimal("1").Equal(One) {
		t.Fatal("cache unusable after parser panic")
	}
	if len(p.decimals) != 1 || len(p.ranges) != 0 {
		t.Fatal("cached an invalid parameter")
	}
}

func TestValidatorConcurrentParameterCache(t *testing.T) {
	v := validator.New()
	if err := RegisterGoPlaygroundValidator(v); err != nil {
		t.Fatal(err)
	}
	type request struct {
		Amount Decimal `validate:"decimal_gt=0.1000000000000000000000000000000001,decimal_between=0.1~1"`
	}
	input := request{MustFromString("0.2")}
	// Warm validator's own schema metadata; parameter cache misses/hits may
	// then race across goroutines without changing shared Decimal values.
	if err := v.Struct(input); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 100 {
				if err := v.Struct(input); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	// A new type introduces concurrent misses after the cache is already live.
	type secondRequest struct {
		Amount Decimal `validate:"decimal_gte=0.200,decimal_lte=1.000,decimal_between=0.2~1.0"`
	}
	for range 16 {
		wg.Go(func() {
			if err := v.Struct(secondRequest{input.Amount}); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
}

func BenchmarkValidatorParameterReuse(b *testing.B) {
	type request struct {
		Amount Decimal `validate:"decimal_gt=0.1000000000000000000000000000000001"`
	}
	v := validator.New()
	if err := RegisterGoPlaygroundValidator(v); err != nil {
		b.Fatal(err)
	}
	input := request{MustFromString("0.2")}
	if err := v.Struct(input); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := v.Struct(input); err != nil {
			b.Fatal(err)
		}
	}
}
