package decimal

import (
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
