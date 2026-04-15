package decimal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/locales/en"
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

func mustGetENTranslator(t *testing.T) ut.Translator {
	t.Helper()

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, ok := uni.GetTranslator("en")
	if !ok {
		t.Fatal("failed to get en translator")
	}
	return trans
}
