package decimal

import (
	"encoding/xml"
	"strings"
)

// MarshalXML implements xml.Marshaler.
// It encodes the decimal as character data using the canonical string form.
// An uninitialized Decimal is encoded as an empty element.
func (d Decimal) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d.i == nil {
		return e.EncodeElement("", start)
	}
	return e.EncodeElement(d.String(), start)
}

// UnmarshalXML implements xml.Unmarshaler.
// It parses character data into a Decimal via NewFromString.
// Empty content leaves d as the uninitialized zero Decimal.
func (d *Decimal) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var text string
	if err := dec.DecodeElement(&text, &start); err != nil {
		return err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		*d = Decimal{}
		return nil
	}

	parsed, err := NewFromString(text)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr.
// An uninitialized Decimal yields an attribute with an empty value.
func (d Decimal) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if d.i == nil {
		return xml.Attr{Name: name, Value: ""}, nil
	}
	return xml.Attr{Name: name, Value: d.String()}, nil
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr.
// An empty attribute value leaves d as the uninitialized zero Decimal.
func (d *Decimal) UnmarshalXMLAttr(attr xml.Attr) error {
	value := strings.TrimSpace(attr.Value)
	if value == "" {
		*d = Decimal{}
		return nil
	}

	parsed, err := NewFromString(value)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}
