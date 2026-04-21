package decimal

import (
	"database/sql/driver"
	"encoding/xml"
	"strings"
)

// NullDecimal is a nullable Decimal, mirroring the pattern used by
// database/sql's NullString. The zero value represents SQL NULL.
type NullDecimal struct {
	Decimal Decimal
	Valid   bool // Valid is true if Decimal is not NULL
}

// NewNullDecimal returns a NullDecimal wrapping d with Valid set to true.
func NewNullDecimal(d Decimal) NullDecimal {
	return NullDecimal{Decimal: d, Valid: true}
}

// Scan implements sql.Scanner.
// A nil value resets n to an invalid zero state. Any other value is delegated
// to Decimal.Scan and marks n valid on success.
func (n *NullDecimal) Scan(value any) error {
	if value == nil {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.Scan(value); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// Value implements driver.Valuer by returning nil when n is not valid,
// otherwise the canonical string form from Decimal.Value.
func (n NullDecimal) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Decimal.Value()
}

// MarshalJSON implements json.Marshaler.
// An invalid NullDecimal is encoded as JSON null.
func (n NullDecimal) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Decimal.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
// JSON null sets Valid to false; any other value is parsed via Decimal.UnmarshalJSON.
func (n *NullDecimal) UnmarshalJSON(data []byte) error {
	if len(data) == len("null") && string(data) == "null" {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.UnmarshalJSON(data); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// MarshalYAML implements yaml.Marshaler.
// An invalid NullDecimal is encoded as the YAML nil value.
func (n NullDecimal) MarshalYAML() (any, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Decimal.MarshalYAML()
}

// UnmarshalYAML implements yaml unmarshaling. A nil scalar sets Valid to false;
// other scalars are delegated to Decimal.UnmarshalYAML.
func (n *NullDecimal) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}
	if raw == nil {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.UnmarshalYAML(func(target any) error {
		ptr, ok := target.(*any)
		if !ok {
			return unmarshal(target)
		}
		*ptr = raw
		return nil
	}); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// An invalid NullDecimal is encoded as an empty byte slice.
func (n NullDecimal) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte{}, nil
	}
	return n.Decimal.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
// Empty text sets Valid to false; any other value is parsed via Decimal.UnmarshalText.
func (n *NullDecimal) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.UnmarshalText(text); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// MarshalXML implements xml.Marshaler.
// An invalid NullDecimal is encoded as an empty element; otherwise the call
// is delegated to Decimal.MarshalXML.
func (n NullDecimal) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if !n.Valid {
		return e.EncodeElement("", start)
	}
	return n.Decimal.MarshalXML(e, start)
}

// UnmarshalXML implements xml.Unmarshaler.
// Empty character content sets Valid to false; any other value is delegated
// to Decimal.UnmarshalXML and marks n valid on success.
func (n *NullDecimal) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var text string
	if err := d.DecodeElement(&text, &start); err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	parsed, err := NewFromString(strings.TrimSpace(text))
	if err != nil {
		return err
	}
	n.Decimal = parsed
	n.Valid = true
	return nil
}

// MarshalXMLAttr implements xml.MarshalerAttr.
// An invalid NullDecimal returns an empty xml.Attr which encoding/xml omits.
// Otherwise the call is delegated to Decimal.MarshalXMLAttr.
func (n NullDecimal) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if !n.Valid {
		return xml.Attr{}, nil
	}
	return n.Decimal.MarshalXMLAttr(name)
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr.
// An empty attribute value sets Valid to false; any other value is delegated
// to Decimal.UnmarshalXMLAttr and marks n valid on success.
func (n *NullDecimal) UnmarshalXMLAttr(attr xml.Attr) error {
	if strings.TrimSpace(attr.Value) == "" {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.UnmarshalXMLAttr(attr); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// UnmarshalParam implements gin's BindUnmarshaler. An empty string marks the
// value invalid; any other value is delegated to Decimal.UnmarshalParam.
func (n *NullDecimal) UnmarshalParam(param string) error {
	if param == "" {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.UnmarshalParam(param); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// String returns "null" when n is not valid, otherwise Decimal.String.
func (n NullDecimal) String() string {
	if !n.Valid {
		return "null"
	}
	return n.Decimal.String()
}
