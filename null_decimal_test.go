package decimal

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
)

func TestNewNullDecimal(t *testing.T) {
	d := MustFromString("1.25")
	n := NewNullDecimal(d)
	if !n.Valid {
		t.Fatal("NewNullDecimal should set Valid to true")
	}
	if !n.Decimal.Equal(d) {
		t.Fatalf("NewNullDecimal.Decimal = %s, want %s", n.Decimal.String(), d.String())
	}
}

func TestNullDecimalScan(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantValid bool
		want      string
		wantErr   bool
	}{
		{name: "nil", input: nil, wantValid: false, want: "0"},
		{name: "string", input: "1.23", wantValid: true, want: "1.23"},
		{name: "bytes", input: []byte("2.5"), wantValid: true, want: "2.5"},
		{name: "int64", input: int64(3), wantValid: true, want: "3"},
		{name: "float64", input: float64(1.25), wantValid: true, want: "1.25"},
		{name: "invalid string", input: "not-a-number", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var n NullDecimal
			err := n.Scan(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Scan(%v) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Scan(%v) returned error: %v", tc.input, err)
			}
			if n.Valid != tc.wantValid {
				t.Fatalf("Scan(%v).Valid = %v, want %v", tc.input, n.Valid, tc.wantValid)
			}
			if tc.wantValid && n.Decimal.String() != tc.want {
				t.Fatalf("Scan(%v).Decimal = %s, want %s", tc.input, n.Decimal.String(), tc.want)
			}
		})
	}
}

func TestNullDecimalValue(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		n := NullDecimal{}
		v, err := n.Value()
		if err != nil {
			t.Fatalf("Value() returned error: %v", err)
		}
		if v != nil {
			t.Fatalf("Value() = %v, want nil", v)
		}
	})

	t.Run("valid", func(t *testing.T) {
		n := NewNullDecimal(MustFromString("7.5"))
		v, err := n.Value()
		if err != nil {
			t.Fatalf("Value() returned error: %v", err)
		}
		if v != "7.5" {
			t.Fatalf("Value() = %v, want 7.5", v)
		}
	})
}

func TestNullDecimalMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   NullDecimal
		want string
	}{
		{name: "invalid", in: NullDecimal{}, want: "null"},
		{name: "valid", in: NewNullDecimal(MustFromString("1.23")), want: `"1.23"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := tc.in.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() returned error: %v", err)
			}
			if string(bz) != tc.want {
				t.Fatalf("MarshalJSON() = %s, want %s", string(bz), tc.want)
			}
		})
	}
}

func TestNullDecimalUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
		want      string
		wantErr   bool
	}{
		{name: "null", input: "null", wantValid: false, want: "0"},
		{name: "quoted number", input: `"1.23"`, wantValid: true, want: "1.23"},
		{name: "raw number", input: `1.23`, wantValid: true, want: "1.23"},
		{name: "invalid", input: `"not-a-number"`, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var n NullDecimal
			err := n.UnmarshalJSON([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalJSON(%s) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON(%s) returned error: %v", tc.input, err)
			}
			if n.Valid != tc.wantValid {
				t.Fatalf("UnmarshalJSON(%s).Valid = %v, want %v", tc.input, n.Valid, tc.wantValid)
			}
			if tc.wantValid && n.Decimal.String() != tc.want {
				t.Fatalf("UnmarshalJSON(%s).Decimal = %s, want %s", tc.input, n.Decimal.String(), tc.want)
			}
		})
	}
}

func TestNullDecimalJSONRoundTrip(t *testing.T) {
	type wrapper struct {
		Amount NullDecimal `json:"amount"`
	}

	t.Run("null", func(t *testing.T) {
		var w wrapper
		if err := json.Unmarshal([]byte(`{"amount":null}`), &w); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}
		if w.Amount.Valid {
			t.Fatal("expected Valid=false for null")
		}

		bz, err := json.Marshal(w)
		if err != nil {
			t.Fatalf("Marshal returned error: %v", err)
		}
		if string(bz) != `{"amount":null}` {
			t.Fatalf("Marshal = %s, want {\"amount\":null}", string(bz))
		}
	})

	t.Run("value", func(t *testing.T) {
		var w wrapper
		if err := json.Unmarshal([]byte(`{"amount":"1.23"}`), &w); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}
		if !w.Amount.Valid {
			t.Fatal("expected Valid=true for value")
		}
		if w.Amount.Decimal.String() != "1.23" {
			t.Fatalf("Decimal = %s, want 1.23", w.Amount.Decimal.String())
		}

		bz, err := json.Marshal(w)
		if err != nil {
			t.Fatalf("Marshal returned error: %v", err)
		}
		if string(bz) != `{"amount":"1.23"}` {
			t.Fatalf("Marshal = %s, want {\"amount\":\"1.23\"}", string(bz))
		}
	})
}

func TestNullDecimalMarshalYAML(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		n := NullDecimal{}
		v, err := n.MarshalYAML()
		if err != nil {
			t.Fatalf("MarshalYAML() returned error: %v", err)
		}
		if v != nil {
			t.Fatalf("MarshalYAML() = %v, want nil", v)
		}
	})

	t.Run("valid", func(t *testing.T) {
		n := NewNullDecimal(MustFromString("1.23"))
		v, err := n.MarshalYAML()
		if err != nil {
			t.Fatalf("MarshalYAML() returned error: %v", err)
		}
		if v != "1.23" {
			t.Fatalf("MarshalYAML() = %v, want 1.23", v)
		}
	})
}

func TestNullDecimalUnmarshalYAML(t *testing.T) {
	decode := func(value any) func(any) error {
		return func(target any) error {
			ptr, ok := target.(*any)
			if !ok {
				return nil
			}
			*ptr = value
			return nil
		}
	}

	tests := []struct {
		name      string
		value     any
		wantValid bool
		want      string
		wantErr   bool
	}{
		{name: "nil", value: nil, wantValid: false, want: "0"},
		{name: "string", value: "1.25", wantValid: true, want: "1.25"},
		{name: "int64", value: int64(42), wantValid: true, want: "42"},
		{name: "float64", value: float64(1.5), wantValid: true, want: "1.5"},
		{name: "invalid string", value: "bad", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var n NullDecimal
			err := n.UnmarshalYAML(decode(tc.value))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalYAML(%v) expected error, got nil", tc.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalYAML(%v) returned error: %v", tc.value, err)
			}
			if n.Valid != tc.wantValid {
				t.Fatalf("UnmarshalYAML(%v).Valid = %v, want %v", tc.value, n.Valid, tc.wantValid)
			}
			if tc.wantValid && n.Decimal.String() != tc.want {
				t.Fatalf("UnmarshalYAML(%v).Decimal = %s, want %s", tc.value, n.Decimal.String(), tc.want)
			}
		})
	}
}

func TestNullDecimalMarshalText(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		n := NullDecimal{}
		bz, err := n.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText() returned error: %v", err)
		}
		if string(bz) != "" {
			t.Fatalf("MarshalText() = %q, want empty", string(bz))
		}
	})

	t.Run("valid", func(t *testing.T) {
		n := NewNullDecimal(MustFromString("1.23"))
		bz, err := n.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText() returned error: %v", err)
		}
		if string(bz) != "1.23" {
			t.Fatalf("MarshalText() = %s, want 1.23", string(bz))
		}
	})
}

func TestNullDecimalUnmarshalText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
		want      string
		wantErr   bool
	}{
		{name: "empty", input: "", wantValid: false, want: "0"},
		{name: "value", input: "1.25", wantValid: true, want: "1.25"},
		{name: "invalid", input: "not-a-number", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var n NullDecimal
			err := n.UnmarshalText([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalText(%s) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalText(%s) returned error: %v", tc.input, err)
			}
			if n.Valid != tc.wantValid {
				t.Fatalf("UnmarshalText(%s).Valid = %v, want %v", tc.input, n.Valid, tc.wantValid)
			}
			if tc.wantValid && n.Decimal.String() != tc.want {
				t.Fatalf("UnmarshalText(%s).Decimal = %s, want %s", tc.input, n.Decimal.String(), tc.want)
			}
		})
	}
}

func TestNullDecimalUnmarshalParam(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
		want      string
		wantErr   bool
	}{
		{name: "empty", input: "", wantValid: false, want: "0"},
		{name: "value", input: "1.5", wantValid: true, want: "1.5"},
		{name: "invalid", input: "not-a-number", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var n NullDecimal
			err := n.UnmarshalParam(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalParam(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalParam(%q) returned error: %v", tc.input, err)
			}
			if n.Valid != tc.wantValid {
				t.Fatalf("UnmarshalParam(%q).Valid = %v, want %v", tc.input, n.Valid, tc.wantValid)
			}
			if tc.wantValid && n.Decimal.String() != tc.want {
				t.Fatalf("UnmarshalParam(%q).Decimal = %s, want %s", tc.input, n.Decimal.String(), tc.want)
			}
		})
	}
}

type nullXMLElementPayload struct {
	XMLName xml.Name    `xml:"Payment"`
	Amount  NullDecimal `xml:"Amount"`
}

type nullXMLAttrPayload struct {
	XMLName xml.Name    `xml:"Payment"`
	Amount  NullDecimal `xml:"amount,attr"`
}

func TestNullDecimalMarshalXML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := nullXMLElementPayload{Amount: NewNullDecimal(MustFromString("123.45"))}
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}
		if !strings.Contains(string(data), "<Amount>123.45</Amount>") {
			t.Fatalf("unexpected XML: %s", data)
		}

		var out nullXMLElementPayload
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if !out.Amount.Valid {
			t.Fatalf("expected Valid=true, got false")
		}
		if !out.Amount.Decimal.Equal(in.Amount.Decimal) {
			t.Fatalf("roundtrip mismatch: got %s want %s", out.Amount.Decimal.String(), in.Amount.Decimal.String())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		in := nullXMLElementPayload{} // Amount invalid
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}
		s := string(data)
		if !strings.Contains(s, "<Amount></Amount>") && !strings.Contains(s, "<Amount/>") {
			t.Fatalf("expected empty element, got: %s", s)
		}

		var out nullXMLElementPayload
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if out.Amount.Valid {
			t.Fatalf("expected Valid=false for empty element, got true")
		}
		if !out.Amount.Decimal.IsZero() {
			t.Fatalf("expected zero decimal, got %s", out.Amount.Decimal.String())
		}
	})

	t.Run("malformed", func(t *testing.T) {
		xmlData := []byte(`<Payment><Amount>not-a-number</Amount></Payment>`)
		var out nullXMLElementPayload
		if err := xml.Unmarshal(xmlData, &out); err == nil {
			t.Fatalf("expected error for malformed number, got nil")
		}
	})
}

func TestNullDecimalMarshalXMLAttr(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := nullXMLAttrPayload{Amount: NewNullDecimal(MustFromString("-7.25"))}
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}
		if !strings.Contains(string(data), `amount="-7.25"`) {
			t.Fatalf("unexpected XML attribute: %s", data)
		}

		var out nullXMLAttrPayload
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if !out.Amount.Valid {
			t.Fatalf("expected Valid=true, got false")
		}
		if !out.Amount.Decimal.Equal(in.Amount.Decimal) {
			t.Fatalf("roundtrip mismatch: got %s want %s", out.Amount.Decimal.String(), in.Amount.Decimal.String())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		in := nullXMLAttrPayload{} // Amount invalid
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}
		// Invalid should yield no amount attribute (encoding/xml omits empty xml.Attr).
		if strings.Contains(string(data), "amount=") {
			t.Fatalf("expected no amount attribute, got: %s", data)
		}

		var out nullXMLAttrPayload
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if out.Amount.Valid {
			t.Fatalf("expected Valid=false for missing attr, got true")
		}
		if !out.Amount.Decimal.IsZero() {
			t.Fatalf("expected zero decimal, got %s", out.Amount.Decimal.String())
		}
	})

	t.Run("empty attr", func(t *testing.T) {
		// Explicit empty attribute should also decode to Invalid.
		xmlData := []byte(`<Payment amount=""></Payment>`)
		var out nullXMLAttrPayload
		if err := xml.Unmarshal(xmlData, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if out.Amount.Valid {
			t.Fatalf("expected Valid=false for empty attr, got true")
		}
	})

	t.Run("malformed", func(t *testing.T) {
		xmlData := []byte(`<Payment amount="bad-number"/>`)
		var out nullXMLAttrPayload
		if err := xml.Unmarshal(xmlData, &out); err == nil {
			t.Fatalf("expected error for malformed attr, got nil")
		}
	})
}

func TestNullDecimalXMLStructRoundTrip(t *testing.T) {
	type wrapper struct {
		XMLName xml.Name    `xml:"Invoice"`
		Total   NullDecimal `xml:"Total"`
		Tax     NullDecimal `xml:"tax,attr"`
	}

	t.Run("both valid", func(t *testing.T) {
		in := wrapper{
			Total: NewNullDecimal(MustFromString("99.99")),
			Tax:   NewNullDecimal(MustFromString("0.07")),
		}
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}
		if !strings.Contains(string(data), "<Total>99.99</Total>") {
			t.Fatalf("expected Total element, got: %s", data)
		}
		if !strings.Contains(string(data), `tax="0.07"`) {
			t.Fatalf("expected tax attr, got: %s", data)
		}

		var out wrapper
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if !out.Total.Valid || !out.Tax.Valid {
			t.Fatalf("expected both Valid=true, got Total.Valid=%v Tax.Valid=%v", out.Total.Valid, out.Tax.Valid)
		}
		if !out.Total.Decimal.Equal(in.Total.Decimal) {
			t.Fatalf("Total roundtrip mismatch: got %s want %s", out.Total.Decimal.String(), in.Total.Decimal.String())
		}
		if !out.Tax.Decimal.Equal(in.Tax.Decimal) {
			t.Fatalf("Tax roundtrip mismatch: got %s want %s", out.Tax.Decimal.String(), in.Tax.Decimal.String())
		}
	})

	t.Run("both invalid", func(t *testing.T) {
		in := wrapper{} // both invalid
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}

		var out wrapper
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if out.Total.Valid {
			t.Fatalf("expected Total.Valid=false, got true")
		}
		if out.Tax.Valid {
			t.Fatalf("expected Tax.Valid=false, got true")
		}
	})

	t.Run("mixed", func(t *testing.T) {
		in := wrapper{
			Total: NewNullDecimal(MustFromString("42")),
			// Tax left invalid
		}
		data, err := xml.Marshal(in)
		if err != nil {
			t.Fatalf("xml.Marshal returned error: %v", err)
		}
		if strings.Contains(string(data), "tax=") {
			t.Fatalf("expected no tax attribute for invalid NullDecimal, got: %s", data)
		}

		var out wrapper
		if err := xml.Unmarshal(data, &out); err != nil {
			t.Fatalf("xml.Unmarshal returned error: %v", err)
		}
		if !out.Total.Valid {
			t.Fatalf("expected Total.Valid=true, got false")
		}
		if out.Tax.Valid {
			t.Fatalf("expected Tax.Valid=false, got true")
		}
		if !out.Total.Decimal.Equal(in.Total.Decimal) {
			t.Fatalf("Total roundtrip mismatch: got %s want %s", out.Total.Decimal.String(), in.Total.Decimal.String())
		}
	})
}

func TestNullDecimalString(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		n := NullDecimal{}
		if n.String() != "null" {
			t.Fatalf("String() = %s, want null", n.String())
		}
	})

	t.Run("valid", func(t *testing.T) {
		n := NewNullDecimal(MustFromString("1.23"))
		if n.String() != "1.23" {
			t.Fatalf("String() = %s, want 1.23", n.String())
		}
	})
}
