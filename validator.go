package decimal

import (
	"fmt"
	"reflect"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterGoPlaygroundValidator registers Decimal-specific validation tags
// for go-playground/validator.
//
// Registered tags:
// - decimal_required
// - decimal_eq
// - decimal_gt
// - decimal_gte
// - decimal_lt
// - decimal_lte
//
// Use built-in omitempty as usual.
func RegisterGoPlaygroundValidator(v *validator.Validate) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}

	if err := v.RegisterValidation("decimal_required", validateDecimalRequired); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_eq", validateDecimalEQ); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_gt", validateDecimalGT); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_gte", validateDecimalGTE); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_lt", validateDecimalLT); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_lte", validateDecimalLTE); err != nil {
		return err
	}

	return nil
}

// RegisterGoPlaygroundValidatorTranslations registers friendly error messages
// for Decimal validator tags on the provided translator.
func RegisterGoPlaygroundValidatorTranslations(v *validator.Validate, trans ut.Translator) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}
	if trans == nil {
		return fmt.Errorf("translator is nil")
	}

	translations := []struct {
		tag     string
		message string
	}{
		{tag: "decimal_required", message: "{0} is required"},
		{tag: "decimal_eq", message: "{0} must be equal to {1}"},
		{tag: "decimal_gt", message: "{0} must be greater than {1}"},
		{tag: "decimal_gte", message: "{0} must be greater than or equal to {1}"},
		{tag: "decimal_lt", message: "{0} must be less than {1}"},
		{tag: "decimal_lte", message: "{0} must be less than or equal to {1}"},
	}

	for _, item := range translations {
		tag := item.tag
		message := item.message
		if err := v.RegisterTranslation(tag, trans,
			func(ut ut.Translator) error {
				return ut.Add(tag, message, true)
			},
			func(ut ut.Translator, fe validator.FieldError) string {
				translated, err := ut.T(fe.Tag(), fe.Field(), fe.Param())
				if err != nil {
					return fe.Error()
				}
				return translated
			},
		); err != nil {
			return err
		}
	}

	return nil
}

// TranslateGoPlaygroundValidationErrors converts validator errors into friendly messages.
func TranslateGoPlaygroundValidationErrors(err error, trans ut.Translator) []string {
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}
	}

	messages := make([]string, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		if trans == nil {
			messages = append(messages, fieldErr.Error())
			continue
		}
		messages = append(messages, fieldErr.Translate(trans))
	}
	return messages
}

func validateDecimalRequired(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	return !field.IsNil()
}

func validateDecimalEQ(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	param := mustParseDecimalParam(fl.Param())
	return field.Equal(param)
}

func validateDecimalGT(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	param := mustParseDecimalParam(fl.Param())
	return field.GT(param)
}

func validateDecimalGTE(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	param := mustParseDecimalParam(fl.Param())
	return field.GTE(param)
}

func validateDecimalLT(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	param := mustParseDecimalParam(fl.Param())
	return field.LT(param)
}

func validateDecimalLTE(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	param := mustParseDecimalParam(fl.Param())
	return field.LTE(param)
}

func decimalFromField(field reflect.Value) (Decimal, bool) {
	if !field.IsValid() || !field.CanInterface() {
		return Decimal{}, false
	}
	dec, ok := field.Interface().(Decimal)
	if !ok {
		return Decimal{}, false
	}
	return dec, true
}

func mustParseDecimalParam(param string) Decimal {
	parsed, err := NewFromString(param)
	if err != nil {
		panic(fmt.Sprintf("invalid decimal validator parameter %q", param))
	}
	return parsed
}
