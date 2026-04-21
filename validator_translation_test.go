package decimal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/ar"
	"github.com/go-playground/locales/de"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/es"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/locales/hi"
	"github.com/go-playground/locales/ja"
	"github.com/go-playground/locales/ko"
	"github.com/go-playground/locales/pt"
	"github.com/go-playground/locales/pt_BR"
	"github.com/go-playground/locales/ru"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant"
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

	t.Run("ne", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_ne=0"`
		}

		err := v.Struct(req{Amount: MustFromString("0")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must not be equal to 0"}) {
			t.Fatalf("messages = %v, want [Amount must not be equal to 0]", messages)
		}
	})

	t.Run("between", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_between=1~100"`
		}

		err := v.Struct(req{Amount: MustFromString("200")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must be between 1~100"}) {
			t.Fatalf("messages = %v, want [Amount must be between 1~100]", messages)
		}
	})

	t.Run("positive", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_positive"`
		}

		err := v.Struct(req{Amount: MustFromString("0")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must be positive"}) {
			t.Fatalf("messages = %v, want [Amount must be positive]", messages)
		}
	})

	t.Run("negative", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_negative"`
		}

		err := v.Struct(req{Amount: MustFromString("0")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must be negative"}) {
			t.Fatalf("messages = %v, want [Amount must be negative]", messages)
		}
	})

	t.Run("nonzero", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_nonzero"`
		}

		err := v.Struct(req{Amount: MustFromString("0")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must not be zero"}) {
			t.Fatalf("messages = %v, want [Amount must not be zero]", messages)
		}
	})

	t.Run("max_precision", func(t *testing.T) {
		type req struct {
			Amount Decimal `validate:"decimal_max_precision=2"`
		}

		err := v.Struct(req{Amount: MustFromString("1.234")})
		messages := TranslateGoPlaygroundValidationErrors(err, trans)
		if !reflect.DeepEqual(messages, []string{"Amount must have at most 2 decimal places"}) {
			t.Fatalf("messages = %v, want [Amount must have at most 2 decimal places]", messages)
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

// TestRegisterGoPlaygroundValidatorTranslations_AllBuiltInLocales verifies a
// sample translated message for every built-in locale so that all 12 locales
// are wired up end-to-end.
func TestRegisterGoPlaygroundValidatorTranslations_AllBuiltInLocales(t *testing.T) {
	cases := []struct {
		name    string
		locale  locales.Translator
		tag     string
		invalid Decimal
		want    string
	}{
		{
			name:    "en required",
			locale:  en.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount is required",
		},
		{
			name:    "zh required",
			locale:  zh.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount为必填项",
		},
		{
			name:    "zh_Hant required",
			locale:  zh_Hant.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount為必填項",
		},
		{
			name:    "ja required",
			locale:  ja.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amountは必須です",
		},
		{
			name:    "ko required",
			locale:  ko.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount은(는) 필수 항목입니다",
		},
		{
			name:    "fr required",
			locale:  fr.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount est obligatoire",
		},
		{
			name:    "es required",
			locale:  es.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount es obligatorio",
		},
		{
			name:    "de required",
			locale:  de.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount ist erforderlich",
		},
		{
			name:    "pt required",
			locale:  pt.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount é obrigatório",
		},
		{
			name:    "pt_BR required",
			locale:  pt_BR.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount é obrigatório",
		},
		{
			name:    "ru required",
			locale:  ru.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount обязательно для заполнения",
		},
		{
			name:    "ar required",
			locale:  ar.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount مطلوب",
		},
		{
			name:    "hi required",
			locale:  hi.New(),
			tag:     "decimal_required",
			invalid: Decimal{},
			want:    "Amount आवश्यक है",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trans := mustGetTranslator(t, tc.locale)
			v := validator.New()
			if err := RegisterGoPlaygroundValidator(v); err != nil {
				t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
			}
			if err := RegisterGoPlaygroundValidatorTranslations(v, trans); err != nil {
				t.Fatalf("RegisterGoPlaygroundValidatorTranslations() returned error: %v", err)
			}

			type req struct {
				Amount Decimal `validate:"decimal_required"`
			}

			err := v.Struct(req{Amount: tc.invalid})
			messages := TranslateGoPlaygroundValidationErrors(err, trans)
			if len(messages) != 1 || messages[0] != tc.want {
				t.Fatalf("messages = %v, want [%s]", messages, tc.want)
			}
		})
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

	// All new tags should have English defaults.
	for _, tag := range []string{
		"decimal_ne",
		"decimal_between",
		"decimal_positive",
		"decimal_negative",
		"decimal_nonzero",
		"decimal_max_precision",
	} {
		if enMsgs[tag] == "" {
			t.Fatalf("missing English default for tag %q", tag)
		}
	}

	// zh_Hant should route to Traditional Chinese (distinct from Simplified).
	zhHantMsgs := DefaultGoPlaygroundValidatorTranslationMessages("zh-Hant")
	if got := zhHantMsgs["decimal_required"]; got != "{0}為必填項" {
		t.Fatalf("zh_Hant decimal_required = %q, want %q", got, "{0}為必填項")
	}

	// zh_TW should also route to Traditional.
	zhTWMsgs := DefaultGoPlaygroundValidatorTranslationMessages("zh_TW")
	if got := zhTWMsgs["decimal_required"]; got != "{0}為必填項" {
		t.Fatalf("zh_TW decimal_required = %q, want %q", got, "{0}為必填項")
	}

	// zh_SG should still route to Simplified via the zh_* fallback.
	zhSGMsgs := DefaultGoPlaygroundValidatorTranslationMessages("zh_SG")
	if got := zhSGMsgs["decimal_required"]; got != "{0}为必填项" {
		t.Fatalf("zh_SG decimal_required = %q, want %q", got, "{0}为必填项")
	}

	// pt_BR should resolve to its own branch (still Portuguese text, but
	// selected via the pt_br branch before the generic pt branch).
	ptBRMsgs := DefaultGoPlaygroundValidatorTranslationMessages("pt_BR")
	if got := ptBRMsgs["decimal_required"]; got != "{0} é obrigatório" {
		t.Fatalf("pt_BR decimal_required = %q, want %q", got, "{0} é obrigatório")
	}

	// Spot-check a couple of other new locales.
	koMsgs := DefaultGoPlaygroundValidatorTranslationMessages("ko")
	if got := koMsgs["decimal_positive"]; got != "{0}은(는) 양수여야 합니다" {
		t.Fatalf("ko decimal_positive = %q", got)
	}
	ruMsgs := DefaultGoPlaygroundValidatorTranslationMessages("ru")
	if got := ruMsgs["decimal_max_precision"]; got == "" {
		t.Fatalf("ru decimal_max_precision missing")
	}
	arMsgs := DefaultGoPlaygroundValidatorTranslationMessages("ar")
	if got := arMsgs["decimal_nonzero"]; got == "" {
		t.Fatalf("ar decimal_nonzero missing")
	}
	hiMsgs := DefaultGoPlaygroundValidatorTranslationMessages("hi")
	if got := hiMsgs["decimal_between"]; got == "" {
		t.Fatalf("hi decimal_between missing")
	}
}

func mustGetENTranslator(t *testing.T) ut.Translator {
	t.Helper()
	return mustGetTranslator(t, en.New())
}

func mustGetZHTranslator(t *testing.T) ut.Translator {
	t.Helper()
	return mustGetTranslator(t, zh.New())
}

func mustGetTranslator(t *testing.T, loc locales.Translator) ut.Translator {
	t.Helper()

	uni := ut.New(en.New(), loc)
	trans, ok := uni.GetTranslator(loc.Locale())
	if !ok {
		t.Fatalf("failed to get %s translator", loc.Locale())
	}
	return trans
}
