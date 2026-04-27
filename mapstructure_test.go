package decimal

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

// callHook invokes the DecodeHook with the supplied source data and a
// destination reflect.Type, returning whatever the hook returns.
func callHook(to reflect.Type, data any) (any, error) {
	hook := DecodeHook()
	var from reflect.Type
	if data != nil {
		from = reflect.TypeOf(data)
	}
	return hook(from, to, data)
}

func TestDecodeHook_Decimal_Strings(t *testing.T) {
	got, err := callHook(decimalType, "1.5")
	if err != nil {
		t.Fatalf("non-empty string -> Decimal: unexpected error: %v", err)
	}
	d, ok := got.(Decimal)
	if !ok {
		t.Fatalf("expected Decimal, got %T", got)
	}
	if d.String() != "1.5" {
		t.Fatalf("expected 1.5, got %s", d.String())
	}

	_, err = callHook(decimalType, "")
	if err == nil {
		t.Fatalf("empty string -> Decimal: expected error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestDecodeHook_Decimal_Bytes(t *testing.T) {
	got, err := callHook(decimalType, []byte("2.25"))
	if err != nil {
		t.Fatalf("[]byte -> Decimal: %v", err)
	}
	if got.(Decimal).String() != "2.25" {
		t.Fatalf("expected 2.25, got %s", got.(Decimal).String())
	}

	_, err = callHook(decimalType, []byte{})
	if err == nil {
		t.Fatalf("empty []byte -> Decimal: expected error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}

	_, err = callHook(decimalType, []byte(nil))
	if err == nil {
		t.Fatalf("nil []byte -> Decimal: expected error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestDecodeHook_Decimal_JSONNumber(t *testing.T) {
	got, err := callHook(decimalType, json.Number("3.14"))
	if err != nil {
		t.Fatalf("json.Number -> Decimal: %v", err)
	}
	if got.(Decimal).String() != "3.14" {
		t.Fatalf("expected 3.14, got %s", got.(Decimal).String())
	}

	_, err = callHook(decimalType, json.Number("not-a-number"))
	if err == nil {
		t.Fatalf("invalid json.Number -> Decimal: expected error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestDecodeHook_Decimal_Ints(t *testing.T) {
	cases := []struct {
		name string
		data any
		want string
	}{
		{"int", int(7), "7"},
		{"int8", int8(-8), "-8"},
		{"int16", int16(16), "16"},
		{"int32", int32(-32), "-32"},
		{"int64", int64(64), "64"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := callHook(decimalType, tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.(Decimal).String() != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got.(Decimal).String())
			}
		})
	}
}

func TestDecodeHook_Decimal_Uints(t *testing.T) {
	cases := []struct {
		name string
		data any
		want string
	}{
		{"uint", uint(7), "7"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"uint64-max", uint64(math.MaxUint64), "18446744073709551615"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := callHook(decimalType, tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.(Decimal).String() != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got.(Decimal).String())
			}
		})
	}
}

func TestDecodeHook_Decimal_Floats(t *testing.T) {
	got, err := callHook(decimalType, float32(1.5))
	if err != nil {
		t.Fatalf("float32 -> Decimal: %v", err)
	}
	if got.(Decimal).String() != "1.5" {
		t.Fatalf("float32 expected 1.5, got %s", got.(Decimal).String())
	}

	got, err = callHook(decimalType, float64(2.5))
	if err != nil {
		t.Fatalf("float64 -> Decimal: %v", err)
	}
	if got.(Decimal).String() != "2.5" {
		t.Fatalf("float64 expected 2.5, got %s", got.(Decimal).String())
	}

	for _, bad := range []any{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		float32(math.NaN()),
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
	} {
		_, err := callHook(decimalType, bad)
		if err == nil {
			t.Fatalf("non-finite float %v -> Decimal: expected error", bad)
		}
		if !errors.Is(err, ErrUnmarshal) {
			t.Fatalf("non-finite float: expected ErrUnmarshal, got %v", err)
		}
	}
}

func TestDecodeHook_Decimal_Nil(t *testing.T) {
	_, err := callHook(decimalType, nil)
	if err == nil {
		t.Fatalf("nil -> Decimal: expected error")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
	// nil error message must mention "nil" (not "<nil>") for readability.
	if msg := err.Error(); !strings.Contains(msg, "nil mapstructure value") {
		t.Fatalf("expected error to mention 'nil mapstructure value', got %q", msg)
	}

	// Value-form passthrough cases for Decimal-typed source data — covers the
	// ComposeDecodeHookFunc + TextUnmarshallerHookFunc handoff where upstream
	// hooks return value (not pointer) types.
	got, err := callHook(decimalType, Decimal{})
	if err != nil {
		t.Fatalf("Decimal{} -> Decimal: unexpected error: %v", err)
	}
	if d, ok := got.(Decimal); !ok || d.String() != "0" {
		t.Fatalf("Decimal{} -> Decimal: expected zero Decimal, got %#v", got)
	}

	got, err = callHook(decimalType, NewNullDecimal(NewFromInt(7)))
	if err != nil {
		t.Fatalf("NullDecimal(valid) -> Decimal: unexpected error: %v", err)
	}
	if d, ok := got.(Decimal); !ok || d.String() != "7" {
		t.Fatalf("NullDecimal(valid 7) -> Decimal: expected 7, got %#v", got)
	}

	_, err = callHook(decimalType, NullDecimal{})
	if err == nil {
		t.Fatalf("NullDecimal(invalid) -> Decimal: expected error")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("NullDecimal(invalid) -> Decimal: expected ErrUnmarshal, got %v", err)
	}

	// Typed nil pointer -> Decimal: §5 says nil cannot become Decimal.
	if _, err := callHook(decimalType, (*Decimal)(nil)); err == nil || !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("(*Decimal)(nil) -> Decimal: expected ErrUnmarshal, got %v", err)
	}
}

func TestDecodeHook_Decimal_BoolRejected(t *testing.T) {
	_, err := callHook(decimalType, true)
	if err == nil {
		t.Fatalf("bool -> Decimal: expected error")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
	// Error message must mention bool's %T form.
	if msg := err.Error(); !strings.Contains(msg, "bool") {
		t.Fatalf("expected error to mention bool, got %q", msg)
	}
}

func TestDecodeHook_Decimal_StructRejected(t *testing.T) {
	_, err := callHook(decimalType, struct{}{})
	if err == nil {
		t.Fatalf("struct -> Decimal: expected error")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
}

func TestDecodeHook_NullDecimal_Strings(t *testing.T) {
	got, err := callHook(nullDecimalType, "9.99")
	if err != nil {
		t.Fatalf("non-empty string -> NullDecimal: %v", err)
	}
	nd := got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "9.99" {
		t.Fatalf("expected valid 9.99, got %#v", nd)
	}

	got, err = callHook(nullDecimalType, "")
	if err != nil {
		t.Fatalf("empty string -> NullDecimal: %v", err)
	}
	nd = got.(NullDecimal)
	if nd.Valid {
		t.Fatalf("empty string should give invalid NullDecimal, got %#v", nd)
	}
}

func TestDecodeHook_NullDecimal_Bytes(t *testing.T) {
	got, err := callHook(nullDecimalType, []byte("4.5"))
	if err != nil {
		t.Fatalf("[]byte -> NullDecimal: %v", err)
	}
	nd := got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "4.5" {
		t.Fatalf("expected valid 4.5, got %#v", nd)
	}

	got, err = callHook(nullDecimalType, []byte{})
	if err != nil {
		t.Fatalf("empty []byte -> NullDecimal: %v", err)
	}
	if got.(NullDecimal).Valid {
		t.Fatalf("empty []byte should give invalid NullDecimal")
	}

	got, err = callHook(nullDecimalType, []byte(nil))
	if err != nil {
		t.Fatalf("nil []byte -> NullDecimal: %v", err)
	}
	if got.(NullDecimal).Valid {
		t.Fatalf("nil []byte should give invalid NullDecimal")
	}
}

func TestDecodeHook_NullDecimal_JSONNumber(t *testing.T) {
	got, err := callHook(nullDecimalType, json.Number("0.5"))
	if err != nil {
		t.Fatalf("json.Number -> NullDecimal: %v", err)
	}
	nd := got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "0.5" {
		t.Fatalf("expected valid 0.5, got %#v", nd)
	}

	_, err = callHook(nullDecimalType, json.Number("oops"))
	if err == nil {
		t.Fatalf("invalid json.Number -> NullDecimal: expected error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestDecodeHook_NullDecimal_Ints(t *testing.T) {
	cases := []struct {
		name string
		data any
		want string
	}{
		{"int", int(7), "7"},
		{"int8", int8(-1), "-1"},
		{"int16", int16(2), "2"},
		{"int32", int32(-3), "-3"},
		{"int64", int64(4), "4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := callHook(nullDecimalType, tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			nd := got.(NullDecimal)
			if !nd.Valid || nd.Decimal.String() != tc.want {
				t.Fatalf("expected valid %s, got %#v", tc.want, nd)
			}
		})
	}
}

func TestDecodeHook_NullDecimal_Uints(t *testing.T) {
	cases := []struct {
		name string
		data any
		want string
	}{
		{"uint", uint(7), "7"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"uint64-max", uint64(math.MaxUint64), "18446744073709551615"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := callHook(nullDecimalType, tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			nd := got.(NullDecimal)
			if !nd.Valid || nd.Decimal.String() != tc.want {
				t.Fatalf("expected valid %s, got %#v", tc.want, nd)
			}
		})
	}
}

func TestDecodeHook_NullDecimal_Floats(t *testing.T) {
	got, err := callHook(nullDecimalType, float32(1.5))
	if err != nil {
		t.Fatalf("float32 -> NullDecimal: %v", err)
	}
	nd := got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "1.5" {
		t.Fatalf("expected valid 1.5, got %#v", nd)
	}

	got, err = callHook(nullDecimalType, float64(2.5))
	if err != nil {
		t.Fatalf("float64 -> NullDecimal: %v", err)
	}
	nd = got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "2.5" {
		t.Fatalf("expected valid 2.5, got %#v", nd)
	}

	for _, bad := range []any{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		float32(math.NaN()),
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
	} {
		_, err := callHook(nullDecimalType, bad)
		if err == nil {
			t.Fatalf("non-finite float %v -> NullDecimal: expected error", bad)
		}
		if !errors.Is(err, ErrUnmarshal) {
			t.Fatalf("non-finite float: expected ErrUnmarshal, got %v", err)
		}
	}
}

func TestDecodeHook_NullDecimal_Nil(t *testing.T) {
	got, err := callHook(nullDecimalType, nil)
	if err != nil {
		t.Fatalf("nil -> NullDecimal: %v", err)
	}
	nd := got.(NullDecimal)
	if nd.Valid {
		t.Fatalf("nil should give invalid NullDecimal, got %#v", nd)
	}

	// Value-form passthrough cases for Decimal/NullDecimal-typed source data.
	got, err = callHook(nullDecimalType, Decimal{})
	if err != nil {
		t.Fatalf("Decimal{} -> NullDecimal: unexpected error: %v", err)
	}
	nd = got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "0" {
		t.Fatalf("Decimal{} -> NullDecimal: expected valid 0, got %#v", nd)
	}

	got, err = callHook(nullDecimalType, NewNullDecimal(NewFromInt(3)))
	if err != nil {
		t.Fatalf("NullDecimal(valid) -> NullDecimal: unexpected error: %v", err)
	}
	nd = got.(NullDecimal)
	if !nd.Valid || nd.Decimal.String() != "3" {
		t.Fatalf("NullDecimal(valid 3) -> NullDecimal: expected valid 3, got %#v", nd)
	}

	got, err = callHook(nullDecimalType, NullDecimal{})
	if err != nil {
		t.Fatalf("NullDecimal{} -> NullDecimal: unexpected error: %v", err)
	}
	nd = got.(NullDecimal)
	if nd.Valid {
		t.Fatalf("NullDecimal{} -> NullDecimal: expected invalid, got %#v", nd)
	}

	// Typed nil pointers -> NullDecimal: align with untyped nil (Valid=false).
	for _, src := range []any{(*Decimal)(nil), (*NullDecimal)(nil)} {
		got, err := callHook(nullDecimalType, src)
		if err != nil {
			t.Fatalf("%T(nil) -> NullDecimal: unexpected error: %v", src, err)
		}
		if nd := got.(NullDecimal); nd.Valid {
			t.Fatalf("%T(nil) -> NullDecimal: expected invalid, got %#v", src, nd)
		}
	}
}

func TestDecodeHook_NullDecimal_BoolRejected(t *testing.T) {
	_, err := callHook(nullDecimalType, true)
	if err == nil {
		t.Fatalf("bool -> NullDecimal: expected error")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
	if msg := err.Error(); !strings.Contains(msg, "bool") {
		t.Fatalf("expected error to mention bool, got %q", msg)
	}
}

func TestDecodeHook_NullDecimal_StructRejected(t *testing.T) {
	_, err := callHook(nullDecimalType, struct{}{})
	if err == nil {
		t.Fatalf("struct -> NullDecimal: expected error")
	}
	if !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
}

func TestDecodeHook_PassThrough(t *testing.T) {
	// Unrelated target type — hook should return data unchanged with no error.
	hook := DecodeHook()
	type Other struct{}
	otherType := reflect.TypeOf(Other{})
	for _, data := range []any{"hello", 42, nil, []byte("x"), true} {
		var from reflect.Type
		if data != nil {
			from = reflect.TypeOf(data)
		}
		got, err := hook(from, otherType, data)
		if err != nil {
			t.Fatalf("passthrough for %T: unexpected error: %v", data, err)
		}
		if !reflect.DeepEqual(got, data) {
			t.Fatalf("passthrough for %T: expected unchanged, got %#v", data, got)
		}
	}

	// Pointer-to-Decimal target also falls through (mapstructure auto-derefs
	// and re-invokes the hook with the value-type target).
	ptrType := reflect.TypeOf((*Decimal)(nil))
	got, err := hook(reflect.TypeOf("1"), ptrType, "1")
	if err != nil {
		t.Fatalf("passthrough for *Decimal: unexpected error: %v", err)
	}
	if got != "1" {
		t.Fatalf("passthrough for *Decimal: expected unchanged 1, got %#v", got)
	}
}

func TestDecodeHook_ErrorMessageType(t *testing.T) {
	// Verify %T string appears in the unsupported-type error.
	type weird struct{ X int }
	_, err := callHook(decimalType, weird{})
	if err == nil {
		t.Fatalf("weird -> Decimal: expected error")
	}
	if !strings.Contains(err.Error(), "decimal.weird") {
		t.Fatalf("expected error to include type name decimal.weird, got %q", err.Error())
	}
}
