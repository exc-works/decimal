package decimal

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
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
	// Stack-allocated 48-byte scratch covers typical financial decimals
	// without an intermediate heap slice grow,leaving only the final
	// string conversion as the unavoidable allocation.
	var buf [48]byte
	return string(d.appendString(buf[:0], false))
}

// String removes trailing zeros from the decimal representation.
func (d Decimal) String() string {
	var buf [48]byte
	return string(d.appendString(buf[:0], true))
}

// Append appends the canonical decimal text (with trailing zeros stripped) to
// dst and returns the extended slice. Zero-allocation when dst has enough
// capacity, making this the preferred form for hot paths that build their own
// byte buffer (proto wire field setters, custom JSON, log builders, etc.).
//
// For "preserve trailing zeros" semantics use AppendWithTrailingZeros.
func (d Decimal) Append(dst []byte) []byte {
	return d.appendString(dst, true)
}

// AppendWithTrailingZeros is the trailing-zero-preserving counterpart of Append.
func (d Decimal) AppendWithTrailingZeros(dst []byte) []byte {
	return d.appendString(dst, false)
}

// appendString is the workhorse used by String, StringWithTrailingZeros, Append
// and MarshalJSON. It avoids the prior intermediate `bzStr := make(...) →
// string(bzStr)` round-trip and the explicit `new(big.Int).Neg` for negatives.
//
// Allocation profile (typical financial values fitting in ~47 digits):
//   - dst with enough capacity: 0 allocs in zero / prec==0 paths
//   - prec > 0 path: 0 alloc into sufficiently sized dst, plus 1 unavoidable
//     alloc when stripTrailingZeros triggers StripTrailingZeros (creates a
//     new big.Int internally)
//   - very large mantissa (> intBuf cap): 1 extra alloc from big.Int.Append's
//     own backing growth
func (d Decimal) appendString(dst []byte, stripTrailingZeros bool) []byte {
	d = initializeIfNeeded(d)

	// Fast path for zero preserves explicit scale only when requested.
	if d.IsZero() {
		if !stripTrailingZeros && d.prec > 0 {
			dst = append(dst, '0', '.')
			for i := 0; i < d.prec; i++ {
				dst = append(dst, '0')
			}
			return dst
		}
		return append(dst, '0')
	}

	if stripTrailingZeros {
		d = d.StripTrailingZeros()
	}
	if d.prec == 0 {
		return d.i.Append(dst, 10)
	}

	// Build the magnitude into a stack scratch buffer to avoid the
	// new(big.Int).Neg + intermediate string allocations the prior code
	// needed for negatives.
	//
	// 48 bytes is a fast-path threshold, not a hard cap: Decimal accepts
	// any precision up to maxParsedPrecision (1<<17), and the underlying
	// big.Int can hold mantissas of arbitrary length. When the result
	// of big.Int.Append doesn't fit, big.Int.Append itself allocates a
	// fresh heap-backed slice and returns that — correctness is unchanged,
	// only the zero-extra-alloc fast path is forfeited. 48 bytes covers
	// the entire space of canonical financial values (< 47 digits + sign)
	// without forcing the array to escape — verified via
	// `go build -gcflags=-m=2`: "marshal.go:.. append does not escape".
	var intBuf [48]byte
	intStr := d.i.Append(intBuf[:0], 10)
	isNeg := false
	if len(intStr) > 0 && intStr[0] == '-' {
		isNeg = true
		intStr = intStr[1:]
	}
	inputSize := len(intStr)

	if isNeg {
		dst = append(dst, '-')
	}

	if inputSize <= d.prec {
		// pure fraction: "0." + leading zeros + magnitude
		dst = append(dst, '0', '.')
		for i := 0; i < d.prec-inputSize; i++ {
			dst = append(dst, '0')
		}
		dst = append(dst, intStr...)
	} else {
		// has integer part: split at decPointPlace
		decPointPlace := inputSize - d.prec
		dst = append(dst, intStr[:decPointPlace]...)
		dst = append(dst, '.')
		dst = append(dst, intStr[decPointPlace:]...)
	}
	return dst
}

// MarshalJSON implements json.Marshaler.
// It encodes a decimal as a JSON string and encodes an uninitialized value as null.
//
// Implementation appends "..." directly via appendString instead of going
// through String() + json.Marshal(string), collapsing 4-6 allocs (typical)
// to 1.
func (d Decimal) MarshalJSON() ([]byte, error) {
	if d.i == nil {
		return []byte(`null`), nil
	}
	// 64 covers max ~62 digits + opening/closing quote; short enough to
	// keep the allocator's small-bucket fast path, large enough to avoid
	// grow on any realistic financial decimal.
	dst := make([]byte, 0, 64)
	dst = append(dst, '"')
	dst = d.appendString(dst, true)
	dst = append(dst, '"')
	return dst, nil
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
