//go:build bson
// +build bson

package decimal

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// MarshalBSONValue implements bson.ValueMarshaler.
// It encodes the decimal as a BSON string using its canonical string form.
// An uninitialized Decimal is encoded as BSON null.
func (d Decimal) MarshalBSONValue() (byte, []byte, error) {
	if d.i == nil {
		return byte(bson.TypeNull), nil, nil
	}
	t, data, err := bson.MarshalValue(d.String())
	return byte(t), data, err
}

// UnmarshalBSONValue implements bson.ValueUnmarshaler.
//
// Supported BSON types:
//   - Null       → uninitialized Decimal
//   - String     → parsed via NewFromString
//   - Double     → parsed via NewFromFloat64 (may lose precision)
//   - Int32      → constructed via New
//   - Int64      → constructed via New
//   - Decimal128 → parsed via NewFromString on the Decimal128 string form
//
// Any other BSON type returns an error wrapping ErrUnmarshal.
func (d *Decimal) UnmarshalBSONValue(t byte, data []byte) error {
	rv := bson.RawValue{Type: bson.Type(t), Value: data}

	switch bson.Type(t) {
	case bson.TypeNull:
		*d = Decimal{}
		return nil
	case bson.TypeString:
		s, ok := rv.StringValueOK()
		if !ok {
			return fmt.Errorf("decimal: cannot decode BSON string: %w", ErrUnmarshal)
		}
		parsed, err := NewFromString(s)
		if err != nil {
			return fmt.Errorf("decimal: cannot decode BSON string %q: %w: %w", s, err, ErrUnmarshal)
		}
		*d = parsed
		return nil
	case bson.TypeDouble:
		f, ok := rv.DoubleOK()
		if !ok {
			return fmt.Errorf("decimal: cannot decode BSON double: %w", ErrUnmarshal)
		}
		parsed, err := decimalFromFloat64(f)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case bson.TypeInt32:
		v, ok := rv.Int32OK()
		if !ok {
			return fmt.Errorf("decimal: cannot decode BSON int32: %w", ErrUnmarshal)
		}
		*d = New(int64(v))
		return nil
	case bson.TypeInt64:
		v, ok := rv.Int64OK()
		if !ok {
			return fmt.Errorf("decimal: cannot decode BSON int64: %w", ErrUnmarshal)
		}
		*d = New(v)
		return nil
	case bson.TypeDecimal128:
		dec128, ok := rv.Decimal128OK()
		if !ok {
			return fmt.Errorf("decimal: cannot decode BSON decimal128: %w", ErrUnmarshal)
		}
		s := dec128.String()
		parsed, err := NewFromString(s)
		if err != nil {
			return fmt.Errorf("decimal: cannot decode BSON decimal128 %q: %w: %w", s, err, ErrUnmarshal)
		}
		*d = parsed
		return nil
	default:
		return fmt.Errorf("decimal: unsupported BSON type %v: %w", bson.Type(t), ErrUnmarshal)
	}
}

// MarshalBSONValue implements bson.ValueMarshaler for NullDecimal.
// An invalid NullDecimal encodes as BSON null; otherwise it delegates to
// Decimal.MarshalBSONValue.
func (n NullDecimal) MarshalBSONValue() (byte, []byte, error) {
	if !n.Valid {
		return byte(bson.TypeNull), nil, nil
	}
	return n.Decimal.MarshalBSONValue()
}

// UnmarshalBSONValue implements bson.ValueUnmarshaler for NullDecimal.
// BSON null resets n to the invalid zero state; any other value is delegated
// to Decimal.UnmarshalBSONValue and marks n valid on success.
func (n *NullDecimal) UnmarshalBSONValue(t byte, data []byte) error {
	if bson.Type(t) == bson.TypeNull {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	if err := n.Decimal.UnmarshalBSONValue(t, data); err != nil {
		return err
	}
	n.Valid = true
	return nil
}
