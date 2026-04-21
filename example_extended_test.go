package decimal_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/exc-works/decimal"
)

func ExampleNullDecimal() {
	n := decimal.NewNullDecimal(decimal.MustFromString("9.99"))
	fmt.Println(n.Valid, n.Decimal)

	var empty decimal.NullDecimal
	fmt.Println(empty.Valid, empty.String())
	// Output:
	// true 9.99
	// false null
}

func ExampleNullDecimal_MarshalJSON() {
	valid := decimal.NewNullDecimal(decimal.MustFromString("1.5"))
	validBytes, _ := json.Marshal(valid)

	var invalid decimal.NullDecimal
	invalidBytes, _ := json.Marshal(invalid)

	fmt.Println(string(validBytes))
	fmt.Println(string(invalidBytes))
	// Output:
	// "1.5"
	// null
}

func ExampleNullDecimal_UnmarshalJSON() {
	var valid decimal.NullDecimal
	_ = json.Unmarshal([]byte(`"2.5"`), &valid)
	fmt.Println(valid.Valid, valid.Decimal)

	var null decimal.NullDecimal
	_ = json.Unmarshal([]byte(`null`), &null)
	fmt.Println(null.Valid)
	// Output:
	// true 2.5
	// false
}

func ExampleDecimal_MarshalXML() {
	type Item struct {
		XMLName xml.Name        `xml:"item"`
		Amount  decimal.Decimal `xml:"amount"`
	}

	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	_ = enc.Encode(Item{Amount: decimal.MustFromString("42.5")})
	fmt.Println(buf.String())
	// Output:
	// <item><amount>42.5</amount></item>
}

// ErrInvalidFormat is returned (wrapped) by NewFromString on malformed input,
// so callers can match it using errors.Is.
func ExampleErrInvalidFormat() {
	_, err := decimal.NewFromString("not a number")
	fmt.Println(errors.Is(err, decimal.ErrInvalidFormat))
	// Output:
	// true
}

// ErrNegativeRoot is wrapped when asking for an even root of a negative value.
func ExampleErrNegativeRoot() {
	_, err := decimal.MustFromString("-4").Sqrt()
	fmt.Println(errors.Is(err, decimal.ErrNegativeRoot))
	// Output:
	// true
}

// ErrInvalidLog is wrapped by Log10/Ln when the input is not strictly positive.
func ExampleErrInvalidLog() {
	_, err := decimal.Zero.Log10()
	fmt.Println(errors.Is(err, decimal.ErrInvalidLog))
	// Output:
	// true
}

func ExampleRegisterGoPlaygroundValidator() {
	v := validator.New()
	_ = decimal.RegisterGoPlaygroundValidator(v)

	type Req struct {
		Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
		Rate  decimal.Decimal `validate:"decimal_between=0~1"`
	}

	good := Req{
		Price: decimal.MustFromString("9.99"),
		Rate:  decimal.MustFromString("0.25"),
	}
	fmt.Println(v.Struct(good))

	bad := Req{
		Price: decimal.MustFromString("9.999"), // too many decimal places
		Rate:  decimal.MustFromString("0.25"),
	}
	err := v.Struct(bad)
	if verr, ok := err.(validator.ValidationErrors); ok {
		fmt.Println(verr[0].Tag())
	}
	// Output:
	// <nil>
	// decimal_max_precision
}
