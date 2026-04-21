package decimal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterGoPlaygroundValidator registers Decimal-specific validation tags
// for go-playground/validator.
//
// Registered tags:
//   - decimal_required
//   - decimal_eq
//   - decimal_ne
//   - decimal_gt
//   - decimal_gte
//   - decimal_lt
//   - decimal_lte
//   - decimal_between       (param: "min~max" — tilde-separated bounds, inclusive; min must be <= max)
//   - decimal_positive
//   - decimal_negative
//   - decimal_nonzero
//   - decimal_max_precision (param: non-negative integer, max number of decimal places (scale); see validateDecimalMaxPrecision)
//
// Use built-in omitempty as usual.
//
// Validator tag parameters must be compile-time constants. Passing malformed
// parameters (non-numeric limits, unparseable decimal values, min > max for
// decimal_between, negative decimal_max_precision) causes panics at validation
// time — do not splice untrusted input into struct tags.
//
// Calling this function multiple times on the same *validator.Validate is safe
// and idempotent: re-registration overwrites the previous handler for each tag.
func RegisterGoPlaygroundValidator(v *validator.Validate) error {
	if v == nil {
		return fmt.Errorf("validator is nil: %w", ErrInvalidArgument)
	}

	if err := v.RegisterValidation("decimal_required", validateDecimalRequired); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_eq", validateDecimalEQ); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_ne", validateDecimalNE); err != nil {
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
	if err := v.RegisterValidation("decimal_between", validateDecimalBetween); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_positive", validateDecimalPositive); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_negative", validateDecimalNegative); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_nonzero", validateDecimalNonzero); err != nil {
		return err
	}
	if err := v.RegisterValidation("decimal_max_precision", validateDecimalMaxPrecision); err != nil {
		return err
	}

	return nil
}

// RegisterGoPlaygroundValidatorTranslations registers friendly error messages
// for Decimal validator tags on the provided translator.
//
// It selects built-in messages by trans.Locale(), with English fallback.
// Built-in locales: en, zh (Simplified), zh_Hant (Traditional Chinese, also
// matches zh_TW), ja, ko, fr, es, de, pt, pt_BR, ru, ar, hi.
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
		return fmt.Errorf("validator is nil: %w", ErrInvalidArgument)
	}
	if trans == nil {
		return fmt.Errorf("translator is nil: %w", ErrInvalidArgument)
	}

	merged := defaultGoPlaygroundValidatorTranslationMessages("en")
	for tag, msg := range messages {
		merged[tag] = msg
	}

	for _, tag := range []string{
		"decimal_required",
		"decimal_eq",
		"decimal_ne",
		"decimal_gt",
		"decimal_gte",
		"decimal_lt",
		"decimal_lte",
		"decimal_between",
		"decimal_positive",
		"decimal_negative",
		"decimal_nonzero",
		"decimal_max_precision",
	} {
		message, ok := merged[tag]
		if !ok || message == "" {
			return fmt.Errorf("missing translation for tag %q: %w", tag, ErrInvalidArgument)
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
	case normalized == "zh_hant" || normalized == "zh_tw" || strings.HasPrefix(normalized, "zh_hant_") || strings.HasPrefix(normalized, "zh_tw_"):
		return map[string]string{
			"decimal_required":      "{0}為必填項",
			"decimal_eq":            "{0}必須等於{1}",
			"decimal_ne":            "{0}不得等於{1}",
			"decimal_gt":            "{0}必須大於{1}",
			"decimal_gte":           "{0}必須大於或等於{1}",
			"decimal_lt":            "{0}必須小於{1}",
			"decimal_lte":           "{0}必須小於或等於{1}",
			"decimal_between":       "{0}必須介於{1}之間",
			"decimal_positive":      "{0}必須為正數",
			"decimal_negative":      "{0}必須為負數",
			"decimal_nonzero":       "{0}不得為零",
			"decimal_max_precision": "{0}最多只能有{1}位小數",
		}
	case normalized == "zh" || normalized == "zh_cn" || strings.HasPrefix(normalized, "zh_"):
		return map[string]string{
			"decimal_required":      "{0}为必填项",
			"decimal_eq":            "{0}必须等于{1}",
			"decimal_ne":            "{0}不得等于{1}",
			"decimal_gt":            "{0}必须大于{1}",
			"decimal_gte":           "{0}必须大于或等于{1}",
			"decimal_lt":            "{0}必须小于{1}",
			"decimal_lte":           "{0}必须小于或等于{1}",
			"decimal_between":       "{0}必须介于{1}之间",
			"decimal_positive":      "{0}必须为正数",
			"decimal_negative":      "{0}必须为负数",
			"decimal_nonzero":       "{0}不得为零",
			"decimal_max_precision": "{0}最多只能有{1}位小数",
		}
	case normalized == "ja" || strings.HasPrefix(normalized, "ja_"):
		return map[string]string{
			"decimal_required":      "{0}は必須です",
			"decimal_eq":            "{0}は{1}と等しくなければなりません",
			"decimal_ne":            "{0}は{1}と等しくてはなりません",
			"decimal_gt":            "{0}は{1}より大きくなければなりません",
			"decimal_gte":           "{0}は{1}以上でなければなりません",
			"decimal_lt":            "{0}は{1}より小さくなければなりません",
			"decimal_lte":           "{0}は{1}以下でなければなりません",
			"decimal_between":       "{0}は{1}の範囲内でなければなりません",
			"decimal_positive":      "{0}は正の数でなければなりません",
			"decimal_negative":      "{0}は負の数でなければなりません",
			"decimal_nonzero":       "{0}はゼロであってはなりません",
			"decimal_max_precision": "{0}の小数点以下は最大{1}桁までです",
		}
	case normalized == "ko" || strings.HasPrefix(normalized, "ko_"):
		return map[string]string{
			"decimal_required":      "{0}은(는) 필수 항목입니다",
			"decimal_eq":            "{0}은(는) {1}와(과) 같아야 합니다",
			"decimal_ne":            "{0}은(는) {1}와(과) 같지 않아야 합니다",
			"decimal_gt":            "{0}은(는) {1}보다 커야 합니다",
			"decimal_gte":           "{0}은(는) {1} 이상이어야 합니다",
			"decimal_lt":            "{0}은(는) {1}보다 작아야 합니다",
			"decimal_lte":           "{0}은(는) {1} 이하여야 합니다",
			"decimal_between":       "{0}은(는) {1} 사이여야 합니다",
			"decimal_positive":      "{0}은(는) 양수여야 합니다",
			"decimal_negative":      "{0}은(는) 음수여야 합니다",
			"decimal_nonzero":       "{0}은(는) 0이 아니어야 합니다",
			"decimal_max_precision": "{0}의 소수점 이하 자릿수는 최대 {1}자리여야 합니다",
		}
	case normalized == "fr" || strings.HasPrefix(normalized, "fr_"):
		return map[string]string{
			"decimal_required":      "{0} est obligatoire",
			"decimal_eq":            "{0} doit être égal à {1}",
			"decimal_ne":            "{0} ne doit pas être égal à {1}",
			"decimal_gt":            "{0} doit être supérieur à {1}",
			"decimal_gte":           "{0} doit être supérieur ou égal à {1}",
			"decimal_lt":            "{0} doit être inférieur à {1}",
			"decimal_lte":           "{0} doit être inférieur ou égal à {1}",
			"decimal_between":       "{0} doit être compris entre {1}",
			"decimal_positive":      "{0} doit être positif",
			"decimal_negative":      "{0} doit être négatif",
			"decimal_nonzero":       "{0} ne doit pas être nul",
			"decimal_max_precision": "{0} doit avoir au plus {1} décimales",
		}
	case normalized == "es" || strings.HasPrefix(normalized, "es_"):
		return map[string]string{
			"decimal_required":      "{0} es obligatorio",
			"decimal_eq":            "{0} debe ser igual a {1}",
			"decimal_ne":            "{0} no debe ser igual a {1}",
			"decimal_gt":            "{0} debe ser mayor que {1}",
			"decimal_gte":           "{0} debe ser mayor o igual que {1}",
			"decimal_lt":            "{0} debe ser menor que {1}",
			"decimal_lte":           "{0} debe ser menor o igual que {1}",
			"decimal_between":       "{0} debe estar entre {1}",
			"decimal_positive":      "{0} debe ser positivo",
			"decimal_negative":      "{0} debe ser negativo",
			"decimal_nonzero":       "{0} no debe ser cero",
			"decimal_max_precision": "{0} debe tener como máximo {1} decimales",
		}
	case normalized == "de" || strings.HasPrefix(normalized, "de_"):
		return map[string]string{
			"decimal_required":      "{0} ist erforderlich",
			"decimal_eq":            "{0} muss gleich {1} sein",
			"decimal_ne":            "{0} darf nicht gleich {1} sein",
			"decimal_gt":            "{0} muss größer als {1} sein",
			"decimal_gte":           "{0} muss größer oder gleich {1} sein",
			"decimal_lt":            "{0} muss kleiner als {1} sein",
			"decimal_lte":           "{0} muss kleiner oder gleich {1} sein",
			"decimal_between":       "{0} muss zwischen {1} liegen",
			"decimal_positive":      "{0} muss positiv sein",
			"decimal_negative":      "{0} muss negativ sein",
			"decimal_nonzero":       "{0} darf nicht null sein",
			"decimal_max_precision": "{0} darf höchstens {1} Dezimalstellen haben",
		}
	case normalized == "pt_br" || strings.HasPrefix(normalized, "pt_br_"):
		return map[string]string{
			"decimal_required":      "{0} é obrigatório",
			"decimal_eq":            "{0} deve ser igual a {1}",
			"decimal_ne":            "{0} não deve ser igual a {1}",
			"decimal_gt":            "{0} deve ser maior que {1}",
			"decimal_gte":           "{0} deve ser maior ou igual a {1}",
			"decimal_lt":            "{0} deve ser menor que {1}",
			"decimal_lte":           "{0} deve ser menor ou igual a {1}",
			"decimal_between":       "{0} deve estar entre {1}",
			"decimal_positive":      "{0} deve ser positivo",
			"decimal_negative":      "{0} deve ser negativo",
			"decimal_nonzero":       "{0} não deve ser zero",
			"decimal_max_precision": "{0} deve ter no máximo {1} casas decimais",
		}
	case normalized == "pt" || strings.HasPrefix(normalized, "pt_"):
		return map[string]string{
			"decimal_required":      "{0} é obrigatório",
			"decimal_eq":            "{0} deve ser igual a {1}",
			"decimal_ne":            "{0} não deve ser igual a {1}",
			"decimal_gt":            "{0} deve ser maior que {1}",
			"decimal_gte":           "{0} deve ser maior ou igual a {1}",
			"decimal_lt":            "{0} deve ser menor que {1}",
			"decimal_lte":           "{0} deve ser menor ou igual a {1}",
			"decimal_between":       "{0} deve estar entre {1}",
			"decimal_positive":      "{0} deve ser positivo",
			"decimal_negative":      "{0} deve ser negativo",
			"decimal_nonzero":       "{0} não deve ser zero",
			"decimal_max_precision": "{0} deve ter no máximo {1} casas decimais",
		}
	case normalized == "ru" || strings.HasPrefix(normalized, "ru_"):
		return map[string]string{
			"decimal_required":      "{0} обязательно для заполнения",
			"decimal_eq":            "{0} должно быть равно {1}",
			"decimal_ne":            "{0} не должно быть равно {1}",
			"decimal_gt":            "{0} должно быть больше {1}",
			"decimal_gte":           "{0} должно быть больше или равно {1}",
			"decimal_lt":            "{0} должно быть меньше {1}",
			"decimal_lte":           "{0} должно быть меньше или равно {1}",
			"decimal_between":       "{0} должно быть в диапазоне {1}",
			"decimal_positive":      "{0} должно быть положительным",
			"decimal_negative":      "{0} должно быть отрицательным",
			"decimal_nonzero":       "{0} не должно быть равно нулю",
			"decimal_max_precision": "{0} должно иметь не более {1} знаков после запятой",
		}
	case normalized == "ar" || strings.HasPrefix(normalized, "ar_"):
		return map[string]string{
			"decimal_required":      "{0} مطلوب",
			"decimal_eq":            "يجب أن يكون {0} مساوياً لـ {1}",
			"decimal_ne":            "يجب ألا يكون {0} مساوياً لـ {1}",
			"decimal_gt":            "يجب أن يكون {0} أكبر من {1}",
			"decimal_gte":           "يجب أن يكون {0} أكبر من أو يساوي {1}",
			"decimal_lt":            "يجب أن يكون {0} أصغر من {1}",
			"decimal_lte":           "يجب أن يكون {0} أصغر من أو يساوي {1}",
			"decimal_between":       "يجب أن يكون {0} بين {1}",
			"decimal_positive":      "يجب أن يكون {0} موجباً",
			"decimal_negative":      "يجب أن يكون {0} سالباً",
			"decimal_nonzero":       "يجب ألا يكون {0} صفراً",
			"decimal_max_precision": "يجب ألا يحتوي {0} على أكثر من {1} منزلة عشرية",
		}
	case normalized == "hi" || strings.HasPrefix(normalized, "hi_"):
		return map[string]string{
			"decimal_required":      "{0} आवश्यक है",
			"decimal_eq":            "{0} {1} के बराबर होना चाहिए",
			"decimal_ne":            "{0} {1} के बराबर नहीं होना चाहिए",
			"decimal_gt":            "{0} {1} से अधिक होना चाहिए",
			"decimal_gte":           "{0} {1} से अधिक या बराबर होना चाहिए",
			"decimal_lt":            "{0} {1} से कम होना चाहिए",
			"decimal_lte":           "{0} {1} से कम या बराबर होना चाहिए",
			"decimal_between":       "{0} {1} के बीच होना चाहिए",
			"decimal_positive":      "{0} धनात्मक होना चाहिए",
			"decimal_negative":      "{0} ऋणात्मक होना चाहिए",
			"decimal_nonzero":       "{0} शून्य नहीं होना चाहिए",
			"decimal_max_precision": "{0} में अधिकतम {1} दशमलव स्थान होने चाहिए",
		}
	default:
		return map[string]string{
			"decimal_required":      "{0} is required",
			"decimal_eq":            "{0} must be equal to {1}",
			"decimal_ne":            "{0} must not be equal to {1}",
			"decimal_gt":            "{0} must be greater than {1}",
			"decimal_gte":           "{0} must be greater than or equal to {1}",
			"decimal_lt":            "{0} must be less than {1}",
			"decimal_lte":           "{0} must be less than or equal to {1}",
			"decimal_between":       "{0} must be between {1}",
			"decimal_positive":      "{0} must be positive",
			"decimal_negative":      "{0} must be negative",
			"decimal_nonzero":       "{0} must not be zero",
			"decimal_max_precision": "{0} must have at most {1} decimal places",
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

func validateDecimalNE(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	param := mustParseDecimalParam(fl.Param())
	return !field.Equal(param)
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

func validateDecimalBetween(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	if field.IsNil() {
		return false
	}
	min, max := mustParseDecimalRangeParam(fl.Param())
	return field.GTE(min) && field.LTE(max)
}

func validateDecimalPositive(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	if field.IsNil() {
		return false
	}
	return field.IsPositive()
}

func validateDecimalNegative(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	if field.IsNil() {
		return false
	}
	return field.IsNegative()
}

func validateDecimalNonzero(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	if field.IsNil() {
		return false
	}
	return !field.IsZero()
}

// validateDecimalMaxPrecision implements the decimal_max_precision tag.
//
// **This tag checks decimal places (scale), not total significant digits.**
// The tag name "precision" is retained for backward compatibility, but the
// underlying check is field.Precision() <= max, where Precision() returns the
// number of digits after the decimal point (the scale). For example,
// 123.45 has scale 2 and passes decimal_max_precision=2, even though it has
// 5 significant digits.
func validateDecimalMaxPrecision(fl validator.FieldLevel) bool {
	field, ok := decimalFromField(fl.Field())
	if !ok {
		return false
	}
	max, err := strconv.Atoi(strings.TrimSpace(fl.Param()))
	if err != nil {
		panic(fmt.Sprintf("invalid decimal_max_precision parameter %q", fl.Param()))
	}
	if max < 0 {
		panic(fmt.Sprintf("invalid decimal_max_precision parameter %q: must be non-negative", fl.Param()))
	}
	return field.Precision() <= max
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

// mustParseDecimalRangeParam parses a "min~max" parameter pair used by
// decimal_between. The tilde is picked because struct-tag values cannot
// contain spaces cleanly and "~" reads naturally as a range (e.g. "1~100").
//
// It panics if min > max: an empty or inverted range is almost certainly a
// typo in the struct tag (a compile-time constant), so failing loudly at
// validation time surfaces the bug sooner than silently rejecting every input.
func mustParseDecimalRangeParam(param string) (Decimal, Decimal) {
	parts := strings.Split(strings.TrimSpace(param), "~")
	if len(parts) != 2 {
		panic(fmt.Sprintf("invalid decimal_between parameter %q: expected \"min~max\"", param))
	}
	min := mustParseDecimalParam(strings.TrimSpace(parts[0]))
	max := mustParseDecimalParam(strings.TrimSpace(parts[1]))
	if min.GT(max) {
		panic(fmt.Sprintf("invalid decimal_between parameter %q: min must be <= max", param))
	}
	return min, max
}
