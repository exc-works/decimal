package decimal

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// decimalFromFloat64 is the non-panicking counterpart of NewFromFloat64.
// It is used by unmarshal paths that must reject NaN/±Inf with a wrapped
// ErrUnmarshal instead of the panic MustFromString would produce.
func decimalFromFloat64(f float64) (Decimal, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return Decimal{}, fmt.Errorf("decimal: non-finite float %v: %w", f, ErrUnmarshal)
	}
	return NewFromFloat64(f), nil
}

// decimalFromFloat32 is the non-panicking counterpart of NewFromFloat32.
func decimalFromFloat32(f float32) (Decimal, error) {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return Decimal{}, fmt.Errorf("decimal: non-finite float %v: %w", f, ErrUnmarshal)
	}
	return NewFromFloat32(f), nil
}

const (
	PrecisionFixedSize = 4
)

// StringWithTrailingZeros returns the string with trailing zeros in the decimal representation.
func (d Decimal) StringWithTrailingZeros() string {
	return d.string(false)
}

// String removes trailing zeros from the decimal representation.
func (d Decimal) String() string {
	return d.string(true)
}

func (d Decimal) string(stripTrailingZeros bool) string {
	d = initializeIfNeeded(d)

	// Fast path for zero preserves explicit scale only when requested.
	if d.IsZero() {
		if !stripTrailingZeros && d.prec > 0 {
			return "0." + strings.Repeat("0", d.prec)
		}
		return "0"
	}

	if stripTrailingZeros {
		d = d.StripTrailingZeros()
	}
	if d.prec == 0 {
		return d.i.String()
	}

	isNeg := d.IsNegative()

	if isNeg {
		d = Decimal{
			i:    new(big.Int).Neg(d.i),
			prec: d.prec,
		}
	}

	intStr := d.i.String()
	inputSize := len(intStr)

	var bzStr []byte
	// case 1, purely decimal
	if inputSize <= d.prec {
		bzStr = make([]byte, 0, d.prec+2+1) // +2 for "0." and +1 for "-" if negative

		// add "-" if needed
		if isNeg {
			bzStr = append(bzStr, byte('-'))
		}
		// add "0."
		bzStr = append(bzStr, byte('0'), byte('.'))

		// add "0"s to the left of the decimal point
		for i := 0; i < d.prec-inputSize; i++ {
			bzStr = append(bzStr, byte('0'))
		}
		bzStr = append(bzStr, intStr...)
	} else {
		bzStr = make([]byte, 0, inputSize+1+1) // +1 for "." and +1 for "-" if negative

		// add "-" if needed
		if isNeg {
			bzStr = append(bzStr, byte('-'))
		}
		// add integer part
		decPointPlace := inputSize - d.prec
		bzStr = append(bzStr, intStr[:decPointPlace]...)
		// add "."
		bzStr = append(bzStr, byte('.'))
		// add fractional part
		bzStr = append(bzStr, intStr[decPointPlace:]...)
	}
	return string(bzStr)
}

// MarshalJSON implements json.Marshaler.
// It encodes a decimal as a JSON string and encodes an uninitialized value as null.
func (d Decimal) MarshalJSON() ([]byte, error) {
	if d.i == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler.
// It accepts JSON strings and JSON numbers, treats null as a no-op, and updates d in place.
func (d *Decimal) UnmarshalJSON(bz []byte) error {
	if len(bz) == len("null") && string(bz) == "null" {
		return nil
	}

	if d.i == nil {
		d.i = new(big.Int)
	}

	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			dTemp, err := NewFromString(string(bz))
			if err == nil {
				*d = dTemp
				return nil
			}
		}
		return err
	}

	newDec, err := NewFromString(text)
	if err != nil {
		return err
	}

	*d = newDec
	return nil
}

// MarshalYAML implements yaml.Marshaler by returning the decimal string form.
func (d Decimal) MarshalYAML() (any, error) {
	return d.String(), nil
}

// UnmarshalYAML implements yaml unmarshaling by decoding scalar values into Decimal.
func (d *Decimal) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case string:
		parsed, err := NewFromString(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case []byte:
		parsed, err := NewFromString(string(v))
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case int:
		*d = NewFromInt(v)
		return nil
	case int8:
		*d = New(int64(v))
		return nil
	case int16:
		*d = New(int64(v))
		return nil
	case int32:
		*d = New(int64(v))
		return nil
	case int64:
		*d = New(v)
		return nil
	case uint:
		*d = NewFromUint64(uint64(v), 0)
		return nil
	case uint8:
		*d = NewFromUint64(uint64(v), 0)
		return nil
	case uint16:
		*d = NewFromUint64(uint64(v), 0)
		return nil
	case uint32:
		*d = NewFromUint64(uint64(v), 0)
		return nil
	case uint64:
		*d = NewFromUint64(v, 0)
		return nil
	case float32:
		parsed, err := decimalFromFloat32(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case float64:
		parsed, err := decimalFromFloat64(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	default:
		return fmt.Errorf("could not convert YAML value of type '%T' to Decimal: %w", raw, ErrUnmarshal)
	}
}

// MarshalText implements encoding.TextMarshaler by returning the decimal string form.
func (d Decimal) MarshalText() ([]byte, error) {
	d = initializeIfNeeded(d)
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler by parsing decimal text.
func (d *Decimal) UnmarshalText(text []byte) error {
	parsed, err := NewFromString(string(text))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// UnmarshalParam implements gin's BindUnmarshaler by parsing decimal text.
func (d *Decimal) UnmarshalParam(param string) error {
	return d.UnmarshalText([]byte(param))
}

// MarshalBinary implements encoding.BinaryMarshaler.
// The binary layout is 4 bytes of big-endian precision followed by big.Int Gob bytes.
// It strips trailing zeros before encoding and returns nil for an uninitialized value.
func (d Decimal) MarshalBinary() (data []byte, err error) {
	if d.i == nil {
		return nil, nil
	}
	d = d.StripTrailingZeros()
	precBytes := make([]byte, PrecisionFixedSize)
	binary.BigEndian.PutUint32(precBytes, uint32(d.prec))

	var intBytes []byte
	if intBytes, err = d.i.GobEncode(); err != nil {
		return nil, err
	}

	data = append(precBytes, intBytes...)
	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
// It expects the MarshalBinary layout, resets d to zero for empty input,
// and returns an error when data is shorter than the fixed precision prefix
// or carries a precision that exceeds maxParsedPrecision.
func (d *Decimal) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		d.i = &big.Int{}
		return nil
	}

	if len(data) < PrecisionFixedSize {
		return fmt.Errorf("error decoding binary %v: expected at least %d bytes, got %v: %w",
			data, PrecisionFixedSize, len(data), ErrUnmarshal)
	}

	// Read the precision as fixed-width bytes. Reject values that would lazily
	// blow up later when 10^prec is materialized for arithmetic.
	prec := int(binary.BigEndian.Uint32(data[:PrecisionFixedSize]))
	if prec > maxParsedPrecision {
		return fmt.Errorf("error decoding binary: precision %d exceeds maximum %d: %w",
			prec, maxParsedPrecision, ErrUnmarshal)
	}
	d.prec = prec

	// Read the big.Int.
	d.i = new(big.Int)
	return d.i.GobDecode(data[PrecisionFixedSize:])
}

// Value implements driver.Valuer by returning the canonical decimal string form.
func (d Decimal) Value() (driver.Value, error) {
	d = initializeIfNeeded(d)
	return d.String(), nil
}

// Scan implements sql.Scanner.
// It accepts nil, float32, float64, int64, string, and []byte inputs, and updates d in place.
// Nil input resets d to its uninitialized state.
func (d *Decimal) Scan(value any) error {
	if value == nil {
		d.i = nil
		d.prec = 0
		return nil
	}
	// first try to see if the data is stored in database as a Numeric datatype
	switch v := value.(type) {

	case float32:
		parsed, err := decimalFromFloat32(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil

	case float64:
		// numeric in sqlite3 sends us float64
		parsed, err := decimalFromFloat64(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil

	case int64:
		// at least in sqlite3 when the value is 0 in db, the data is sent
		// to us as an int64 instead of a float64 ...
		*d = New(v)
		return nil

	default:
		// default is trying to interpret value stored as string
		text, err := unquoteIfQuoted(v)
		if err != nil {
			return err
		}
		bTemp, err := NewFromString(text)
		if err != nil {
			return err
		}
		*d = bTemp
		return nil
	}
}

// Marshal implements gogo-protobuf custom type marshaling via MarshalBinary.
func (d Decimal) Marshal() ([]byte, error) {
	return d.MarshalBinary()
}

// MarshalTo implements gogo-protobuf custom type marshaling into data.
func (d Decimal) MarshalTo(data []byte) (n int, err error) {
	bz, err := d.MarshalBinary()
	if err != nil {
		return
	}
	n = copy(data, bz)
	return
}

// Unmarshal implements gogo-protobuf custom type unmarshaling via UnmarshalBinary.
func (d *Decimal) Unmarshal(data []byte) error {
	return d.UnmarshalBinary(data)
}

// Size implements gogo-protobuf custom type sizing based on Marshal output.
func (d Decimal) Size() int {
	bz, _ := d.Marshal()
	return len(bz)
}
