//go:build bson
// +build bson

package decimal_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/exc-works/decimal"
)

func ExampleDecimal_MarshalBSONValue() {
	type Row struct {
		Amount decimal.Decimal `bson:"amount"`
	}

	data, _ := bson.Marshal(Row{Amount: decimal.MustFromString("123.45")})

	var out Row
	_ = bson.Unmarshal(data, &out)
	fmt.Println(out.Amount)
	// Output:
	// 123.45
}

type bsonPayload struct {
	Amount decimal.Decimal `bson:"amount"`
}

type bsonNullPayload struct {
	Amount decimal.NullDecimal `bson:"amount"`
}

func TestDecimal_BSON_RoundtripString(t *testing.T) {
	in := bsonPayload{Amount: decimal.MustFromString("123.456")}
	data, err := bson.Marshal(in)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if !out.Amount.Equal(in.Amount) {
		t.Fatalf("roundtrip mismatch: got %s want %s", out.Amount.String(), in.Amount.String())
	}
}

func TestDecimal_BSON_Uninitialized_IsNull(t *testing.T) {
	in := bsonPayload{} // Amount uninitialized
	data, err := bson.Marshal(in)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	// Decode into a raw document and confirm the amount field is BSON null.
	var raw bson.Raw
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("bson.Unmarshal raw: %v", err)
	}
	amount, err := raw.LookupErr("amount")
	if err != nil {
		t.Fatalf("LookupErr: %v", err)
	}
	if amount.Type != bson.TypeNull {
		t.Fatalf("expected BSON null, got type %v", amount.Type)
	}

	// Decoding back into the typed struct should produce an uninitialized
	// Decimal (IsZero returns true for the zero value).
	var out bsonPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if !out.Amount.IsZero() {
		t.Fatalf("expected zero decimal, got %s", out.Amount.String())
	}
}

func TestDecimal_UnmarshalBSONValue_Double(t *testing.T) {
	type doublePayload struct {
		Amount float64 `bson:"amount"`
	}
	data, err := bson.Marshal(doublePayload{Amount: 12.5})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	want := decimal.NewFromFloat64(12.5)
	if !out.Amount.Equal(want) {
		t.Fatalf("got %s, want %s", out.Amount.String(), want.String())
	}
}

func TestDecimal_UnmarshalBSONValue_Int64(t *testing.T) {
	type int64Payload struct {
		Amount int64 `bson:"amount"`
	}
	data, err := bson.Marshal(int64Payload{Amount: 42})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if !out.Amount.Equal(decimal.New(42)) {
		t.Fatalf("got %s, want 42", out.Amount.String())
	}
}

func TestDecimal_UnmarshalBSONValue_Int32(t *testing.T) {
	type int32Payload struct {
		Amount int32 `bson:"amount"`
	}
	data, err := bson.Marshal(int32Payload{Amount: 7})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if !out.Amount.Equal(decimal.New(7)) {
		t.Fatalf("got %s, want 7", out.Amount.String())
	}
}

func TestDecimal_UnmarshalBSONValue_UnsupportedType(t *testing.T) {
	type boolPayload struct {
		Amount bool `bson:"amount"`
	}
	data, err := bson.Marshal(boolPayload{Amount: true})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	if err := bson.Unmarshal(data, &out); err == nil {
		t.Fatalf("expected error for unsupported BSON type")
	}
}

func TestNullDecimal_BSON_ValidRoundtrip(t *testing.T) {
	in := bsonNullPayload{Amount: decimal.NewNullDecimal(decimal.MustFromString("99.99"))}
	data, err := bson.Marshal(in)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonNullPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if !out.Amount.Valid {
		t.Fatalf("expected Valid=true")
	}
	if !out.Amount.Decimal.Equal(in.Amount.Decimal) {
		t.Fatalf("got %s, want %s", out.Amount.Decimal.String(), in.Amount.Decimal.String())
	}
}

func TestNullDecimal_BSON_InvalidEncodesAsNull(t *testing.T) {
	in := bsonNullPayload{} // Valid=false
	data, err := bson.Marshal(in)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var raw bson.Raw
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("bson.Unmarshal raw: %v", err)
	}
	amount, err := raw.LookupErr("amount")
	if err != nil {
		t.Fatalf("LookupErr: %v", err)
	}
	if amount.Type != bson.TypeNull {
		t.Fatalf("expected BSON null, got type %v", amount.Type)
	}

	var out bsonNullPayload
	if err := bson.Unmarshal(data, &out); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	if out.Amount.Valid {
		t.Fatalf("expected Valid=false after decoding BSON null")
	}
}

func TestDecimal_UnmarshalBSONValue_Double_NaN(t *testing.T) {
	type doublePayload struct {
		Amount float64 `bson:"amount"`
	}
	data, err := bson.Marshal(doublePayload{Amount: math.NaN()})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	err = bson.Unmarshal(data, &out)
	if err == nil {
		t.Fatalf("expected error for NaN BSON double, got nil")
	}
	if !errors.Is(err, decimal.ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal wrap, got %v", err)
	}
}

func TestDecimal_UnmarshalBSONValue_Double_Inf(t *testing.T) {
	type doublePayload struct {
		Amount float64 `bson:"amount"`
	}
	for _, f := range []float64{math.Inf(1), math.Inf(-1)} {
		data, err := bson.Marshal(doublePayload{Amount: f})
		if err != nil {
			t.Fatalf("bson.Marshal: %v", err)
		}

		var out bsonPayload
		err = bson.Unmarshal(data, &out)
		if err == nil {
			t.Fatalf("expected error for Inf BSON double (%v), got nil", f)
		}
		if !errors.Is(err, decimal.ErrUnmarshal) {
			t.Fatalf("expected ErrUnmarshal wrap for %v, got %v", f, err)
		}
	}
}

// TestUnmarshalBSONDecimal128MalformedWrapsErrUnmarshal ensures that when the
// Decimal128 branch produces a string that NewFromString cannot parse (e.g.
// NaN / Infinity strings produced by Decimal128.String()), the error wraps
// ErrUnmarshal so callers can match it with errors.Is.
func TestUnmarshalBSONDecimal128MalformedWrapsErrUnmarshal(t *testing.T) {
	// Build a Decimal128 representing NaN. Per MongoDB spec a Decimal128 NaN
	// has the top bits 0x7c00 in the high 64 bits; its String() returns "NaN",
	// which NewFromString rejects as an invalid format.
	// High bits: 0x7c00_0000_0000_0000 is NaN.
	nan := bson.NewDecimal128(0x7c00000000000000, 0)

	type dec128Payload struct {
		Amount bson.Decimal128 `bson:"amount"`
	}
	data, err := bson.Marshal(dec128Payload{Amount: nan})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	err = bson.Unmarshal(data, &out)
	if err == nil {
		t.Fatalf("expected error for malformed Decimal128, got nil")
	}
	if !errors.Is(err, decimal.ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal wrap, got %v", err)
	}
}

// TestUnmarshalBSONString_MalformedWrapsErrUnmarshal asserts the string branch
// wraps ErrUnmarshal when NewFromString rejects the payload.
func TestUnmarshalBSONString_MalformedWrapsErrUnmarshal(t *testing.T) {
	type stringPayload struct {
		Amount string `bson:"amount"`
	}
	data, err := bson.Marshal(stringPayload{Amount: "not-a-number"})
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}

	var out bsonPayload
	err = bson.Unmarshal(data, &out)
	if err == nil {
		t.Fatalf("expected error for malformed BSON string, got nil")
	}
	if !errors.Is(err, decimal.ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal wrap, got %v", err)
	}
}
