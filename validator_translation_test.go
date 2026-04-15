package decimal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func TestRegisterGoPlaygroundValidatorTranslations(t *testing.T) {
	v := validator.New()
	if err := RegisterGoPlaygroundValidator(v); err != nil {
		t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
	}

	trans := mustGetENTranslator(t)
	if err := RegisterGoPlaygroundValidatorTranslations(v, trans); err != nil {
		t.Fatalf("RegisterGoPlaygroundValidatorTranslations() returned error: %v", err)
	}

	t.Run("required", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_required"`
		}

		err := v.Struct(req{})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount is required"}) {
			t.Fatalf("messages = %v, want [Amount is required]", messages)
		}
	})

	t.Run("eq", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_eq=1.23"`
		}

		err := v.Struct(req{Amount: MustFromString("1.24")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must be equal to 1.23"}) {
			t.Fatalf("messages = %v, want [Amount must be equal to 1.23]", messages)
		}
	})

	t.Run("gte", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_gte=10"`
		}

		err := v.Struct(req{Amount: MustFromString("9.99")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must be greater than or equal to 10"}) {
			t.Fatalf("messages = %v, want [Amount must be greater than or equal to 10]", messages)
		}
	})
}

func TestTranslateGoPlaygroundValidationErrors(t *testing.T) {
	trans := mustGetENTranslator(t)

	t.Run("nil error", func(t *testing.T) {
		messages := TranslateGoPlaygroundValidationErrors(nil, trans)
		if messages != nil {
			t.Fatalf("messages = %v, want nil", messages)
		}
	})

	t.Run("non validation error", func(t *testing.T) {
		messages := TranslateGoPlaygroundValidationErrors(errors.New("boom"), trans)
		if !reflect.DeepEqual(messages, []string{"boom"}) {
			t.Fatalf("messages = %v, want [boom]", messages)
		}
	})
}

func TestRegisterGoPlaygroundValidatorTranslations_MoreLocales(t *testing.T) {
	v := validator.New()
	if err := RegisterGoPlaygroundValidator(v); err != nil {
		t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
	}

	trans := mustGetZHTranslator(t)
	if err := RegisterGoPlaygroundValidatorTranslations(v, trans); err != nil {
		t.Fatalf("RegisterGoPlaygroundValidatorTranslations() returned error: %v", err)
	}

	type req struct {
		Amount Decimal `validate:"decimal_required,decimal_gt=1.23"`
	}

	err := v.Struct(req{})
	messages := TranslateGoPlaygroundValidationErrors(err, trans)
	if !reflect.DeepEqual(messages, []string{"Amount为必填项"}) {
		t.Fatalf("messages = %v, want [Amount为必填项]", messages)
	}
}

func TestRegisterGoPlaygroundValidatorTranslationsWithMessages(t *testing.T) {
	v := validator.New()
	if err := RegisterGoPlaygroundValidator(v); err != nil {
		t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
	}

	trans := mustGetENTranslator(t)
	custom := map[string]string{
		"decimal_required": "{0} cannot be empty",
	}
	if err := RegisterGoPlaygroundValidatorTranslationsWithMessages(v, trans, custom); err != nil {
		t.Fatalf("RegisterGoPlaygroundValidatorTranslationsWithMessages() returned error: %v", err)
	}

	type req struct {
		Amount Decimal `validate:"decimal_required"`
	}

	err := v.Struct(req{})
	messages := TranslateGoPlaygroundValidationErrors(err, trans)
	if !reflect.DeepEqual(messages, []string{"Amount cannot be empty"}) {
		t.Fatalf("messages = %v, want [Amount cannot be empty]", messages)
	}
}

func TestDefaultGoPlaygroundValidatorTranslationMessages(t *testing.T) {
	zhMsgs := DefaultGoPlaygroundValidatorTranslationMessages("zh-CN")
	if got := zhMsgs["decimal_required"]; got != "{0}为必填项" {
		t.Fatalf("zh decimal_required = %q, want %q", got, "{0}为必填项")
	}

	enMsgs := DefaultGoPlaygroundValidatorTranslationMessages("xx-YY")
	if got := enMsgs["decimal_required"]; got != "{0} is required" {
		t.Fatalf("fallback decimal_required = %q, want %q", got, "{0} is required")
	}
}

func mustGetENTranslator(t *testing.T) ut.Translator {
	t.Helper()

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale, zh.New())
	trans, ok := uni.GetTranslator("en")
	if !ok {
		t.Fatal("failed to get en translator")
	}
	return trans
}

func mustGetZHTranslator(t *testing.T) ut.Translator {
	t.Helper()

	enLocale := en.New()
	zhLocale := zh.New()
	uni := ut.New(enLocale, enLocale, zhLocale)
	trans, ok := uni.GetTranslator("zh")
	if !ok {
		t.Fatal("failed to get zh translator")
	}
	return trans
}
