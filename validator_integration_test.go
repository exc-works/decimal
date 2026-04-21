package decimal

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestGoPlaygroundValidator(t *testing.T) {
	t.Run("required and gt", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_required,decimal_gt=0"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		err := validate.Struct(req{})
		assertValidationTag(t, err, "Amount", "decimal_required")

		err = validate.Struct(req{Amount: MustFromString("0")})
		assertValidationTag(t, err, "Amount", "decimal_gt")

		err = validate.Struct(req{Amount: MustFromString("-0.01")})
		assertValidationTag(t, err, "Amount", "decimal_gt")

		err = validate.Struct(req{Amount: MustFromString("1.23")})
		if err != nil {
			t.Fatalf("validate required,gt for valid decimal returned error: %v", err)
		}
	})

	t.Run("eq", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_required,decimal_eq=1.23"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		err := validate.Struct(req{})
		assertValidationTag(t, err, "Amount", "decimal_required")

		// Different scale but same numeric value should pass.
		err = validate.Struct(req{Amount: MustFromString("1.2300")})
		if err != nil {
			t.Fatalf("validate decimal_eq with same numeric value returned error: %v", err)
		}

		err = validate.Struct(req{Amount: MustFromString("1.24")})
		assertValidationTag(t, err, "Amount", "decimal_eq")
	})

	t.Run("omitempty with gte,lt,lte", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"omitempty,decimal_gte=1.5,decimal_lt=10,decimal_lte=9.5"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		// zero value struct (uninitialized Decimal) is considered empty and skips numeric checks.
		err := validate.Struct(req{})
		if err != nil {
			t.Fatalf("validate omitempty on zero decimal returned error: %v", err)
		}

		err = validate.Struct(req{Amount: MustFromString("0")})
		assertValidationTag(t, err, "Amount", "decimal_gte")

		err = validate.Struct(req{Amount: MustFromString("1.49")})
		assertValidationTag(t, err, "Amount", "decimal_gte")

		err = validate.Struct(req{Amount: MustFromString("10")})
		assertValidationTag(t, err, "Amount", "decimal_lt")

		err = validate.Struct(req{Amount: MustFromString("9.6")})
		assertValidationTag(t, err, "Amount", "decimal_lte")

		err = validate.Struct(req{Amount: MustFromString("9.5")})
		if err != nil {
			t.Fatalf("validate omitempty,gte,lt,lte for valid decimal returned error: %v", err)
		}
	})

	t.Run("pointer field", func(t *testing.T) {
		type req struct {
			Amount *Decimal `validate:"decimal_required,decimal_gt=0"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		err := validate.Struct(req{})
		assertValidationTag(t, err, "Amount", "decimal_required")

		negative := MustFromString("-1")
		err = validate.Struct(req{Amount: &negative})
		assertValidationTag(t, err, "Amount", "decimal_gt")

		positive := MustFromString("1")
		err = validate.Struct(req{Amount: &positive})
		if err != nil {
			t.Fatalf("validate required,gt for *Decimal returned error: %v", err)
		}
	})

	t.Run("exact precision comparison without float conversion", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_gte=0.1000000000000000000000000000000001"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		// 0.1 is strictly smaller than the configured lower bound and must fail.
		err := validate.Struct(req{Amount: MustFromString("0.1")})
		assertValidationTag(t, err, "Amount", "decimal_gte")

		err = validate.Struct(req{Amount: MustFromString("0.1000000000000000000000000000000001")})
		if err != nil {
			t.Fatalf("validate gte for exact equal decimal returned error: %v", err)
		}
	})

	t.Run("ne", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_required,decimal_ne=0"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		err := validate.Struct(req{Amount: MustFromString("0")})
		assertValidationTag(t, err, "Amount", "decimal_ne")

		err = validate.Struct(req{Amount: MustFromString("0.0000")})
		assertValidationTag(t, err, "Amount", "decimal_ne")

		err = validate.Struct(req{Amount: MustFromString("0.01")})
		if err != nil {
			t.Fatalf("validate decimal_ne for distinct value returned error: %v", err)
		}
	})

	t.Run("between tilde delimiter", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_between=1~100"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		// Bounds are inclusive.
		for _, inside := range []string{"1", "50", "100", "99.9999"} {
			if err := validate.Struct(req{Amount: MustFromString(inside)}); err != nil {
				t.Fatalf("validate decimal_between for value %s returned error: %v", inside, err)
			}
		}

		for _, outside := range []string{"0.99", "100.01", "-5"} {
			err := validate.Struct(req{Amount: MustFromString(outside)})
			assertValidationTag(t, err, "Amount", "decimal_between")
		}
	})

	t.Run("positive", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_positive"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		if err := validate.Struct(req{Amount: MustFromString("0.01")}); err != nil {
			t.Fatalf("validate decimal_positive for 0.01 returned error: %v", err)
		}

		err := validate.Struct(req{Amount: MustFromString("0")})
		assertValidationTag(t, err, "Amount", "decimal_positive")

		err = validate.Struct(req{Amount: MustFromString("-0.01")})
		assertValidationTag(t, err, "Amount", "decimal_positive")

		// Nil (zero struct) is invalid for decimal_positive.
		err = validate.Struct(req{})
		assertValidationTag(t, err, "Amount", "decimal_positive")
	})

	t.Run("negative", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_negative"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		if err := validate.Struct(req{Amount: MustFromString("-0.01")}); err != nil {
			t.Fatalf("validate decimal_negative for -0.01 returned error: %v", err)
		}

		err := validate.Struct(req{Amount: MustFromString("0")})
		assertValidationTag(t, err, "Amount", "decimal_negative")

		err = validate.Struct(req{Amount: MustFromString("0.01")})
		assertValidationTag(t, err, "Amount", "decimal_negative")

		err = validate.Struct(req{})
		assertValidationTag(t, err, "Amount", "decimal_negative")
	})

	t.Run("nonzero", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_nonzero"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		for _, v := range []string{"-0.01", "0.01", "-100", "100"} {
			if err := validate.Struct(req{Amount: MustFromString(v)}); err != nil {
				t.Fatalf("validate decimal_nonzero for %s returned error: %v", v, err)
			}
		}

		err := validate.Struct(req{Amount: MustFromString("0")})
		assertValidationTag(t, err, "Amount", "decimal_nonzero")

		err = validate.Struct(req{Amount: MustFromString("0.0000")})
		assertValidationTag(t, err, "Amount", "decimal_nonzero")

		err = validate.Struct(req{})
		assertValidationTag(t, err, "Amount", "decimal_nonzero")
	})

	t.Run("max_precision", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_max_precision=2"`
		}

		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		for _, v := range []string{"1", "1.2", "1.23", "100.00"} {
			if err := validate.Struct(req{Amount: MustFromString(v)}); err != nil {
				t.Fatalf("validate decimal_max_precision for %s returned error: %v", v, err)
			}
		}

		err := validate.Struct(req{Amount: MustFromString("1.234")})
		assertValidationTag(t, err, "Amount", "decimal_max_precision")

		err = validate.Struct(req{Amount: MustFromString("0.001")})
		assertValidationTag(t, err, "Amount", "decimal_max_precision")
	})

	t.Run("max_precision rejects negative parameter", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_max_precision=-1"`
		}
		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic for negative decimal_max_precision parameter")
			}
		}()
		_ = validate.Struct(req{Amount: MustFromString("1")})
	})

	t.Run("max_precision rejects non-integer parameter", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_max_precision=abc"`
		}
		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic for non-integer decimal_max_precision parameter")
			}
		}()
		_ = validate.Struct(req{Amount: MustFromString("1")})
	})

	t.Run("between rejects inverted min and max", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_between=100~1"`
		}
		validate := validator.New()
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}

		defer func() {
			r := recover()
			if r == nil {
				t.Fatalf("expected panic for inverted decimal_between parameter")
			}
			msg, ok := r.(string)
			if !ok {
				t.Fatalf("expected panic value to be string, got %T: %v", r, r)
			}
			if !containsAll(msg, "decimal_between", "min must be <= max") {
				t.Fatalf("panic message %q does not mention decimal_between min<=max", msg)
			}
		}()
		_ = validate.Struct(req{Amount: MustFromString("50")})
	})
}

func TestRegisterDecimalValidatorsIdempotent(t *testing.T) {
	type req struct {
		Amount Decimal `validate:"decimal_required,decimal_gt=0,decimal_max_precision=2"`
	}

	validate := validator.New()
	for i := 0; i < 3; i++ {
		if err := RegisterGoPlaygroundValidator(validate); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() call %d returned error: %v", i+1, err)
		}
	}

	// Valid input still passes after repeated registration.
	if err := validate.Struct(req{Amount: MustFromString("1.23")}); err != nil {
		t.Fatalf("validate after repeated registration returned error: %v", err)
	}

	// Invalid input still triggers the expected tag.
	err := validate.Struct(req{Amount: MustFromString("0")})
	assertValidationTag(t, err, "Amount", "decimal_gt")

	err = validate.Struct(req{Amount: MustFromString("1.234")})
	assertValidationTag(t, err, "Amount", "decimal_max_precision")
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func assertValidationTag(t *testing.T, err error, fieldName, tag string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected validation error for field %s with tag %s, got nil", fieldName, tag)
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T: %v", err, err)
	}

	for _, fieldErr := range validationErrs {
		if fieldErr.Field() == fieldName && fieldErr.Tag() == tag {
			return
		}
	}

	t.Fatalf("expected validation error tag %s on field %s, got %v", tag, fieldName, validationErrs)
}
