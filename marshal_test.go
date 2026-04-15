package decimal

import (
	"encoding/json"
	"math/big"
	"testing"
)

type marshalTest struct {
	Value Decimal
}

func TestStringFormatting(t *testing.T) {
	got := Decimal{}
	if got.String() != "0" {
		t.Fatalf("zero value String() = %s, want 0", got.String())
	}

	got = New(0)
	if got.String() != "0" {
		t.Fatalf("New(0).String() = %s, want 0", got.String())
	}

	got = NewFromBigIntWithPrec(big.NewInt(1000), 18)
	if got.String() != "0.000000000000001" {
		t.Fatalf("String() = %s, want 0.000000000000001", got.String())
	}

	got = MustFromString("1.234000000")
	if got.String() != "1.234" {
		t.Fatalf("String() = %s, want 1.234", got.String())
	}

	got = NewFromBigIntWithPrec(big.NewInt(-1000), 18)
	if got.String() != "-0.000000000000001" {
		t.Fatalf("String() = %s, want -0.000000000000001", got.String())
	}

	got = MustFromString("-1.234000000")
	if got.String() != "-1.234" {
		t.Fatalf("String() = %s, want -1.234", got.String())
	}
}

func TestStringWithTrailingZeros(t *testing.T) {
	got := Decimal{}
	if got.StringWithTrailingZeros() != "0" {
		t.Fatalf("zero value StringWithTrailingZeros() = %s, want 0", got.StringWithTrailingZeros())
	}

	got = NewWithPrec(0, 18)
	if got.StringWithTrailingZeros() != "0.000000000000000000" {
		t.Fatalf("StringWithTrailingZeros() = %s, want 0.000000000000000000", got.StringWithTrailingZeros())
	}

	got = MustFromString("1.234000000")
	if got.StringWithTrailingZeros() != "1.234000000" {
		t.Fatalf("StringWithTrailingZeros() = %s, want 1.234000000", got.StringWithTrailingZeros())
	}

	got = MustFromString("-1.234000000")
	if got.StringWithTrailingZeros() != "-1.234000000" {
		t.Fatalf("StringWithTrailingZeros() = %s, want -1.234000000", got.StringWithTrailingZeros())
	}
}

func TestMarshalBinary(t *testing.T) {
	var zero Decimal
	bz, err := zero.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() returned error: %v", err)
	}
	if bz != nil {
		t.Fatalf("MarshalBinary() = %v, want nil", bz)
	}

	var decoded Decimal
	if err := decoded.UnmarshalBinary(nil); err != nil {
		t.Fatalf("UnmarshalBinary(nil) returned error: %v", err)
	}
	if !decoded.IsZero() {
		t.Fatalf("UnmarshalBinary(nil) produced %s, want zero", decoded.String())
	}

	original := MustFromString("1.234000000")
	bz, err = original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() returned error: %v", err)
	}

	var roundTripped Decimal
	if err := roundTripped.UnmarshalBinary(bz); err != nil {
		t.Fatalf("UnmarshalBinary() returned error: %v", err)
	}
	if roundTripped.String() != "1.234" {
		t.Fatalf("round-trip String() = %s, want 1.234", roundTripped.String())
	}
	if roundTripped.Precision() != 3 {
		t.Fatalf("round-trip Precision() = %d, want 3", roundTripped.Precision())
	}
}

func TestJSON(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		val := marshalTest{}
		bz, err := json.Marshal(val)
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}
		if string(bz) != `{"Value":null}` {
			t.Fatalf("json.Marshal = %s, want {\"Value\":null}", string(bz))
		}

		var decoded marshalTest
		if err := json.Unmarshal(bz, &decoded); err != nil {
			t.Fatalf("json.Unmarshal returned error: %v", err)
		}
		if !decoded.Value.IsZero() {
			t.Fatalf("json.Unmarshal result = %s, want zero", decoded.Value.String())
		}
	})

	t.Run("non-zero value", func(t *testing.T) {
		val := marshalTest{Value: NewWithPrec(10001, 4)}
		bz, err := json.Marshal(val)
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}
		if string(bz) != `{"Value":"1.0001"}` {
			t.Fatalf("json.Marshal = %s, want {\"Value\":\"1.0001\"}", string(bz))
		}

		var decoded marshalTest
		if err := json.Unmarshal(bz, &decoded); err != nil {
			t.Fatalf("json.Unmarshal returned error: %v", err)
		}
		if !decoded.Value.Equal(val.Value) {
			t.Fatalf("json.Unmarshal result = %s, want %s", decoded.Value.String(), val.Value.String())
		}
	})
}

func TestDecimalUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "null", input: "null", want: "0"},
		{name: "quoted number", input: `"0.123456789"`, want: "0.123456789"},
		{name: "raw number", input: `0.123456789`, want: "0.123456789"},
		{name: "negative quoted number", input: `"-0.123456789"`, want: "-0.123456789"},
		{name: "negative raw number", input: `-0.123456789`, want: "-0.123456789"},
		{name: "large raw integer", input: `999999999999999999999999999999999999999999`, want: "999999999999999999999999999999999999999999"},
		{name: "large quoted integer", input: `"999999999999999999999999999999999999999999"`, want: "999999999999999999999999999999999999999999"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got Decimal
			if err := json.Unmarshal([]byte(tc.input), &got); err != nil {
				t.Fatalf("json.Unmarshal(%s) returned error: %v", tc.input, err)
			}
			if got.String() != tc.want {
				t.Fatalf("json.Unmarshal(%s) = %s, want %s", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestValueAndMarshalYAML(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var d Decimal
		v, err := d.Value()
		if err != nil {
			t.Fatalf("Value() returned error: %v", err)
		}
		if v != "0" {
			t.Fatalf("Value() = %v, want 0", v)
		}
	})

	t.Run("non-zero value", func(t *testing.T) {
		d := MustFromString("7.5000")
		v, err := d.Value()
		if err != nil {
			t.Fatalf("Value() returned error: %v", err)
		}
		if v != "7.5" {
			t.Fatalf("Value() = %v, want 7.5", v)
		}
	})

	t.Run("marshal yaml", func(t *testing.T) {
		d := MustFromString("1.2300")
		v, err := d.MarshalYAML()
		if err != nil {
			t.Fatalf("MarshalYAML() returned error: %v", err)
		}
		if v != "1.23" {
			t.Fatalf("MarshalYAML() = %v, want 1.23", v)
		}
	})
}

func TestDecimalUnmarshalYAML(t *testing.T) {
	decode := func(value any, err error) func(any) error {
		return func(target any) error {
			if err != nil {
				return err
			}
			ptr, ok := target.(*any)
			if !ok {
				t.Fatalf("unexpected target type %T", target)
			}
			*ptr = value
			return nil
		}
	}

	t.Run("string", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalYAML(decode("1.2300", nil)); err != nil {
			t.Fatalf("UnmarshalYAML(string) returned error: %v", err)
		}
		if d.StringWithTrailingZeros() != "1.2300" {
			t.Fatalf("UnmarshalYAML(string) = %s, want 1.2300", d.StringWithTrailingZeros())
		}
	})

	t.Run("float64", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalYAML(decode(float64(1.25), nil)); err != nil {
			t.Fatalf("UnmarshalYAML(float64) returned error: %v", err)
		}
		if d.String() != "1.25" {
			t.Fatalf("UnmarshalYAML(float64) = %s, want 1.25", d.String())
		}
	})

	t.Run("int64", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalYAML(decode(int64(42), nil)); err != nil {
			t.Fatalf("UnmarshalYAML(int64) returned error: %v", err)
		}
		if d.String() != "42" {
			t.Fatalf("UnmarshalYAML(int64) = %s, want 42", d.String())
		}
	})

	t.Run("null no-op", func(t *testing.T) {
		d := MustFromString("9.99")
		if err := d.UnmarshalYAML(decode(nil, nil)); err != nil {
			t.Fatalf("UnmarshalYAML(nil) returned error: %v", err)
		}
		if d.String() != "9.99" {
			t.Fatalf("UnmarshalYAML(nil) = %s, want 9.99", d.String())
		}
	})

	t.Run("invalid string", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalYAML(decode("bad", nil)); err == nil {
			t.Fatal("UnmarshalYAML(invalid string) expected error, got nil")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalYAML(decode(true, nil)); err == nil {
			t.Fatal("UnmarshalYAML(unsupported type) expected error, got nil")
		}
	})
}

func TestMarshalTextAndUnmarshalText(t *testing.T) {
	t.Run("marshal text", func(t *testing.T) {
		bz, err := MustFromString("1.2300").MarshalText()
		if err != nil {
			t.Fatalf("MarshalText() returned error: %v", err)
		}
		if string(bz) != "1.23" {
			t.Fatalf("MarshalText() = %s, want 1.23", string(bz))
		}
	})

	t.Run("marshal text zero value", func(t *testing.T) {
		var d Decimal
		bz, err := d.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText() returned error: %v", err)
		}
		if string(bz) != "0" {
			t.Fatalf("MarshalText() = %s, want 0", string(bz))
		}
	})

	t.Run("unmarshal text", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalText([]byte("1.2300")); err != nil {
			t.Fatalf("UnmarshalText() returned error: %v", err)
		}
		if d.StringWithTrailingZeros() != "1.2300" {
			t.Fatalf("UnmarshalText() = %s, want 1.2300", d.StringWithTrailingZeros())
		}
	})

	t.Run("unmarshal text invalid", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalText([]byte("")); err == nil {
			t.Fatal("UnmarshalText() expected error, got nil")
		}
	})

	t.Run("unmarshal param", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalParam("1.2300"); err != nil {
			t.Fatalf("UnmarshalParam() returned error: %v", err)
		}
		if d.StringWithTrailingZeros() != "1.2300" {
			t.Fatalf("UnmarshalParam() = %s, want 1.2300", d.StringWithTrailingZeros())
		}
	})

	t.Run("unmarshal param invalid", func(t *testing.T) {
		var d Decimal
		if err := d.UnmarshalParam(""); err == nil {
			t.Fatal("UnmarshalParam() expected error, got nil")
		}
	})
}

func TestScan(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		d := MustFromString("1")
		if err := d.Scan(nil); err != nil {
			t.Fatalf("Scan(nil) returned error: %v", err)
		}
		if !d.IsNil() {
			t.Fatal("Scan(nil) should produce nil decimal")
		}
	})

	t.Run("float64", func(t *testing.T) {
		var d Decimal
		if err := d.Scan(float64(1.25)); err != nil {
			t.Fatalf("Scan(float64) returned error: %v", err)
		}
		if d.String() != "1.25" {
			t.Fatalf("Scan(float64) = %s, want 1.25", d.String())
		}
	})

	t.Run("float32", func(t *testing.T) {
		var d Decimal
		if err := d.Scan(float32(1.25)); err != nil {
			t.Fatalf("Scan(float32) returned error: %v", err)
		}
		if d.String() != "1.25" {
			t.Fatalf("Scan(float32) = %s, want 1.25", d.String())
		}
	})

	t.Run("int64", func(t *testing.T) {
		var d Decimal
		if err := d.Scan(int64(42)); err != nil {
			t.Fatalf("Scan(int64) returned error: %v", err)
		}
		if d.String() != "42" {
			t.Fatalf("Scan(int64) = %s, want 42", d.String())
		}
	})

	t.Run("quoted bytes", func(t *testing.T) {
		var d Decimal
		if err := d.Scan([]byte(`"3.1400"`)); err != nil {
			t.Fatalf("Scan([]byte quoted) returned error: %v", err)
		}
		if d.StringWithTrailingZeros() != "3.1400" {
			t.Fatalf("Scan([]byte quoted) = %s, want 3.1400", d.StringWithTrailingZeros())
		}
	})

	t.Run("raw string", func(t *testing.T) {
		var d Decimal
		if err := d.Scan("0.125"); err != nil {
			t.Fatalf("Scan(string) returned error: %v", err)
		}
		if d.String() != "0.125" {
			t.Fatalf("Scan(string) = %s, want 0.125", d.String())
		}
	})
}

func TestProtoCompatMethods(t *testing.T) {
	original := MustFromString("9.8760")

	bz, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() returned error: %v", err)
	}
	if original.Size() != len(bz) {
		t.Fatalf("Size() = %d, want %d", original.Size(), len(bz))
	}

	buf := make([]byte, len(bz))
	n, err := original.MarshalTo(buf)
	if err != nil {
		t.Fatalf("MarshalTo() returned error: %v", err)
	}
	if n != len(bz) {
		t.Fatalf("MarshalTo() copied = %d, want %d", n, len(bz))
	}

	var decoded Decimal
	if err := decoded.Unmarshal(buf[:n]); err != nil {
		t.Fatalf("Unmarshal() returned error: %v", err)
	}
	if decoded.String() != "9.876" {
		t.Fatalf("Unmarshal() = %s, want 9.876", decoded.String())
	}
}
