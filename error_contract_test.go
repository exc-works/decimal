package decimal

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestNumericPanicSentinels(t *testing.T) {
	for _, tc := range []struct {
		name string
		want error
		call func()
	}{
		{"Quo", ErrDivideByZero, func() { One.Quo(Zero, RoundDown) }},
		{"QuoWithPrec", ErrDivideByZero, func() { One.QuoWithPrec(Zero, 2, RoundDown) }},
		{"QuoRem", ErrDivideByZero, func() { One.QuoRem(Zero) }},
		{"Mod", ErrDivideByZero, func() { One.Mod(Zero) }},
		{"negative power of zero", ErrDivideByZero, func() { Zero.Power(-1) }},
		{"constructor precision", ErrInvalidPrecision, func() { NewWithPrec(1, -1) }},
		{"rescale precision", ErrInvalidPrecision, func() { One.Rescale(-1, RoundDown) }},
		{"root precision", ErrInvalidPrecision, func() { One.SqrtWithPrec(-1) }},
		{"log precision", ErrInvalidPrecision, func() { One.LnWithPrec(-1) }},
		{"round unnecessary rescale", ErrRoundUnnecessary, func() { MustFromString("1.25").Rescale(1, RoundUnnecessary) }},
		{"round unnecessary division", ErrRoundUnnecessary, func() { One.QuoWithPrec(New(3), 2, RoundUnnecessary) }},
		{"round unnecessary rational", ErrRoundUnnecessary, func() { NewFromBigRatWithPrec(big.NewRat(1, 3), 2, RoundUnnecessary) }},
		{"round unnecessary multiply", ErrRoundUnnecessary, func() { MustFromString("0.1").Mul(MustFromString("0.1"), RoundUnnecessary) }},
		{"precision overflow", ErrOverflow, func() { NewWithPrec(1, math.MaxInt32).MulExact(NewWithPrec(1, 1)) }},
		{"native precision overflow", ErrOverflow, func() { NewWithPrec(1, math.MaxInt).MulExact(NewWithPrec(1, math.MaxInt)) }},
		{"power overflow", ErrOverflow, func() { New(2).Power(math.MinInt64) }},
		{"shift overflow", ErrOverflow, func() { One.Shift(math.MinInt) }},
		{"log domain", ErrInvalidLog, func() { Zero.Log2() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				value := recover()
				err, ok := value.(error)
				if !ok || !errors.Is(err, tc.want) {
					t.Fatalf("panic=%v (%T), want error wrapping %v", value, value, tc.want)
				}
			}()
			tc.call()
		})
	}
}

func TestTranslationRegistrationNilArguments(t *testing.T) {
	for _, err := range []error{
		RegisterGoPlaygroundValidatorTranslations(nil, nil),
		RegisterGoPlaygroundValidatorTranslations(validator.New(), nil),
		RegisterGoPlaygroundValidatorTranslations(nil, mustGetENTranslator(t)),
		RegisterGoPlaygroundValidatorTranslationsWithMessages(nil, nil, nil),
		RegisterGoPlaygroundValidatorTranslationsWithMessages(validator.New(), nil, nil),
	} {
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("got %v, want ErrInvalidArgument", err)
		}
	}
}
