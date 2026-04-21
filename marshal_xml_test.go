package decimal_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/exc-works/decimal"
)

type xmlElementPayload struct {
	XMLName xml.Name        `xml:"Payment"`
	Amount  decimal.Decimal `xml:"Amount"`
}

type xmlAttrPayload struct {
	XMLName xml.Name        `xml:"Payment"`
	Amount  decimal.Decimal `xml:"amount,attr"`
}

func TestDecimal_MarshalXML_Roundtrip(t *testing.T) {
	in := xmlElementPayload{Amount: decimal.MustFromString("123.45")}
	data, err := xml.Marshal(in)
	if err != nil {
		t.Fatalf("xml.Marshal: %v", err)
	}
	if !strings.Contains(string(data), "<Amount>123.45</Amount>") {
		t.Fatalf("unexpected XML: %s", data)
	}

	var out xmlElementPayload
	if err := xml.Unmarshal(data, &out); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}
	if !out.Amount.Equal(in.Amount) {
		t.Fatalf("roundtrip mismatch: got %s want %s", out.Amount.String(), in.Amount.String())
	}
}

func TestDecimal_MarshalXML_UninitializedEmptyElement(t *testing.T) {
	in := xmlElementPayload{} // Amount uninitialized
	data, err := xml.Marshal(in)
	if err != nil {
		t.Fatalf("xml.Marshal: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "<Amount></Amount>") && !strings.Contains(s, "<Amount/>") {
		t.Fatalf("expected empty element, got: %s", s)
	}

	var out xmlElementPayload
	if err := xml.Unmarshal(data, &out); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}
	// Uninitialized should still be the uninitialized zero Decimal.
	if !out.Amount.IsZero() {
		t.Fatalf("expected zero decimal, got %s", out.Amount.String())
	}
}

func TestDecimal_UnmarshalXML_Malformed(t *testing.T) {
	xmlData := []byte(`<Payment><Amount>not-a-number</Amount></Payment>`)
	var out xmlElementPayload
	if err := xml.Unmarshal(xmlData, &out); err == nil {
		t.Fatalf("expected error for malformed number, got nil")
	}
}

func TestDecimal_MarshalXMLAttr_Roundtrip(t *testing.T) {
	in := xmlAttrPayload{Amount: decimal.MustFromString("-7.25")}
	data, err := xml.Marshal(in)
	if err != nil {
		t.Fatalf("xml.Marshal: %v", err)
	}
	if !strings.Contains(string(data), `amount="-7.25"`) {
		t.Fatalf("unexpected XML attribute: %s", data)
	}

	var out xmlAttrPayload
	if err := xml.Unmarshal(data, &out); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}
	if !out.Amount.Equal(in.Amount) {
		t.Fatalf("roundtrip mismatch: got %s want %s", out.Amount.String(), in.Amount.String())
	}
}

func TestDecimal_MarshalXMLAttr_Uninitialized(t *testing.T) {
	in := xmlAttrPayload{} // uninitialized
	data, err := xml.Marshal(in)
	if err != nil {
		t.Fatalf("xml.Marshal: %v", err)
	}
	if !strings.Contains(string(data), `amount=""`) {
		t.Fatalf("expected empty attr, got: %s", data)
	}

	var out xmlAttrPayload
	if err := xml.Unmarshal(data, &out); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}
	if !out.Amount.IsZero() {
		t.Fatalf("expected zero decimal, got %s", out.Amount.String())
	}
}

func TestDecimal_UnmarshalXMLAttr_Malformed(t *testing.T) {
	xmlData := []byte(`<Payment amount="bad-number"/>`)
	var out xmlAttrPayload
	if err := xml.Unmarshal(xmlData, &out); err == nil {
		t.Fatalf("expected error for malformed attr, got nil")
	}
}
