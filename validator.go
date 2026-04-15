package decimal

import (
	"fmt"
	"reflect"
	"strings"

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
//
// It selects built-in messages by trans.Locale(), with English fallback.
// Built-in locales include: en, zh, ja, fr, es, de, pt.
func RegisterGoPlaygroundValidatorTranslations(v *validator.Validate, trans ut.Translator) error {
	return RegisterGoPlaygroundValidatorTranslationsWithMessages(
		v,
		trans,
		DefaultGoPlaygroundValidatorTranslationMessages(trans.Locale()),
	)
}

// RegisterGoPlaygroundValidatorTranslationsWithMessages registers friendly
// Decimal validator messages using caller-provided templates.
//
// Any missing tag message falls back to English defaults.
func RegisterGoPlaygroundValidatorTranslationsWithMessages(v *validator.Validate, trans ut.Translator, messages map[string]string) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}
	if trans == nil {
		return fmt.Errorf("translator is nil")
	}

	merged := defaultGoPlaygroundValidatorTranslationMessages("en")
	for tag, msg := range messages {
		merged[tag] = msg
	}

	for _, tag := range []string{
		"decimal_required",
		"decimal_eq",
		"decimal_gt",
		"decimal_gte",
		"decimal_lt",
		"decimal_lte",
	} {
		message, ok := merged[tag]
		if !ok || message == "" {
			return fmt.Errorf("missing translation for tag %q", tag)
		}
		currentTag := tag
		currentMessage := message
		if err := v.RegisterTranslation(tag, trans,
			func(ut ut.Translator) error {
				return ut.Add(currentTag, currentMessage, true)
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

// DefaultGoPlaygroundValidatorTranslationMessages returns built-in translation
// templates for Decimal validation tags by locale (with English fallback).
func DefaultGoPlaygroundValidatorTranslationMessages(locale string) map[string]string {
	return defaultGoPlaygroundValidatorTranslationMessages(locale)
}

func defaultGoPlaygroundValidatorTranslationMessages(locale string) map[string]string {
	normalized := normalizeLocale(locale)
	switch {
	case normalized == "zh" || normalized == "zh_cn" || strings.HasPrefix(normalized, "zh_"):
		return map[string]string{
			"decimal_required": "{0}为必填项",
			"decimal_eq":       "{0}必须等于{1}",
			"decimal_gt":       "{0}必须大于{1}",
			"decimal_gte":      "{0}必须大于或等于{1}",
			"decimal_lt":       "{0}必须小于{1}",
			"decimal_lte":      "{0}必须小于或等于{1}",
		}
	case normalized == "ja" || strings.HasPrefix(normalized, "ja_"):
		return map[string]string{
			"decimal_required": "{0}は必須です",
			"decimal_eq":       "{0}は{1}と等しくなければなりません",
			"decimal_gt":       "{0}は{1}より大きくなければなりません",
			"decimal_gte":      "{0}は{1}以上でなければなりません",
			"decimal_lt":       "{0}は{1}より小さくなければなりません",
			"decimal_lte":      "{0}は{1}以下でなければなりません",
		}
	case normalized == "fr" || strings.HasPrefix(normalized, "fr_"):
		return map[string]string{
			"decimal_required": "{0} est obligatoire",
			"decimal_eq":       "{0} doit être égal à {1}",
			"decimal_gt":       "{0} doit être supérieur à {1}",
			"decimal_gte":      "{0} doit être supérieur ou égal à {1}",
			"decimal_lt":       "{0} doit être inférieur à {1}",
			"decimal_lte":      "{0} doit être inférieur ou égal à {1}",
		}
	case normalized == "es" || strings.HasPrefix(normalized, "es_"):
		return map[string]string{
			"decimal_required": "{0} es obligatorio",
			"decimal_eq":       "{0} debe ser igual a {1}",
			"decimal_gt":       "{0} debe ser mayor que {1}",
			"decimal_gte":      "{0} debe ser mayor o igual que {1}",
			"decimal_lt":       "{0} debe ser menor que {1}",
			"decimal_lte":      "{0} debe ser menor o igual que {1}",
		}
	case normalized == "de" || strings.HasPrefix(normalized, "de_"):
		return map[string]string{
			"decimal_required": "{0} ist erforderlich",
			"decimal_eq":       "{0} muss gleich {1} sein",
			"decimal_gt":       "{0} muss größer als {1} sein",
			"decimal_gte":      "{0} muss größer oder gleich {1} sein",
			"decimal_lt":       "{0} muss kleiner als {1} sein",
			"decimal_lte":      "{0} muss kleiner oder gleich {1} sein",
		}
	case normalized == "pt" || strings.HasPrefix(normalized, "pt_"):
		return map[string]string{
			"decimal_required": "{0} é obrigatório",
			"decimal_eq":       "{0} deve ser igual a {1}",
			"decimal_gt":       "{0} deve ser maior que {1}",
			"decimal_gte":      "{0} deve ser maior ou igual a {1}",
			"decimal_lt":       "{0} deve ser menor que {1}",
			"decimal_lte":      "{0} deve ser menor ou igual a {1}",
		}
	default:
		return map[string]string{
			"decimal_required": "{0} is required",
			"decimal_eq":       "{0} must be equal to {1}",
			"decimal_gt":       "{0} must be greater than {1}",
			"decimal_gte":      "{0} must be greater than or equal to {1}",
			"decimal_lt":       "{0} must be less than {1}",
			"decimal_lte":      "{0} must be less than or equal to {1}",
		}
	}
}

func normalizeLocale(locale string) string {
	return strings.ToLower(strings.ReplaceAll(locale, "-", "_"))
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
