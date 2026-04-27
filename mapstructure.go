package decimal

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// decimalType and nullDecimalType are reflect.Type sentinels used by DecodeHook
// to identify supported destination targets. They are computed once at package
// init to avoid repeated reflect.TypeOf calls per decode.
var (
	decimalType     = reflect.TypeOf(Decimal{})
	nullDecimalType = reflect.TypeOf(NullDecimal{})
)

// DecodeHook returns a mapstructure DecodeHookFuncType (also accepted by
// viper.DecodeHook) that converts numeric, []byte, json.Number, and nil source
// values into Decimal / NullDecimal targets.
//
// It is COMPLEMENTARY to mapstructure.TextUnmarshallerHookFunc(): the text
// hook already handles the string -> Decimal/NullDecimal path via
// Decimal.UnmarshalText. This hook adds every non-string path that the text
// hook misses (int / uint / float / json.Number / []byte / nil), and also
// handles strings so it remains correct when used standalone.
//
// nil/empty-string semantics:
//   - Decimal     + nil           -> error (cannot represent SQL NULL)
//   - Decimal     + ""            -> error (matches NewFromString)
//   - NullDecimal + nil           -> zero value, Valid=false
//   - NullDecimal + ""            -> zero value, Valid=false
//   - Decimal/NullDecimal + bool  -> error (rejected on purpose to avoid
//     silently mapping false/true to 0/1)
//
// Usage with viper:
//
//	viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
//	    mapstructure.TextUnmarshallerHookFunc(),
//	    decimal.DecodeHook(),
//	)))
//
// The returned function is compatible with mapstructure v2's
// DecodeHookFuncType. It is goroutine-safe and stateless.
func DecodeHook() func(from, to reflect.Type, data any) (any, error) {
	return func(_ reflect.Type, to reflect.Type, data any) (any, error) {
		switch to {
		case decimalType:
			return decodeToDecimal(data)
		case nullDecimalType:
			return decodeToNullDecimal(data)
		default:
			return data, nil
		}
	}
}

// decodeToDecimal converts a mapstructure source value to a Decimal.
func decodeToDecimal(data any) (any, error) {
	if data == nil {
		return nil, fmt.Errorf("could not convert nil mapstructure value to Decimal: %w", ErrUnmarshal)
	}
	switch v := data.(type) {
	case Decimal:
		return v, nil
	case *Decimal:
		// Reachable via ComposeDecodeHookFunc + TextUnmarshallerHookFunc:
		// the text hook allocates a *Decimal (UnmarshalText has pointer
		// receiver) and forwards it to the next hook in the chain.
		if v == nil {
			return nil, fmt.Errorf("could not convert nil *Decimal mapstructure value to Decimal: %w", ErrUnmarshal)
		}
		return *v, nil
	case NullDecimal:
		if !v.Valid {
			return nil, fmt.Errorf("could not convert invalid NullDecimal mapstructure value to Decimal: %w", ErrUnmarshal)
		}
		return v.Decimal, nil
	case string:
		return NewFromString(v)
	case []byte:
		return NewFromString(string(v))
	case json.Number:
		return NewFromString(v.String())
	case int:
		return NewFromInt(v), nil
	case int8:
		return New(int64(v)), nil
	case int16:
		return New(int64(v)), nil
	case int32:
		return New(int64(v)), nil
	case int64:
		return New(v), nil
	case uint:
		return NewFromUint64(uint64(v), 0), nil
	case uint8:
		return NewFromUint64(uint64(v), 0), nil
	case uint16:
		return NewFromUint64(uint64(v), 0), nil
	case uint32:
		return NewFromUint64(uint64(v), 0), nil
	case uint64:
		return NewFromUint64(v, 0), nil
	case float32:
		return decimalFromFloat32(v)
	case float64:
		return decimalFromFloat64(v)
	default:
		return nil, fmt.Errorf("could not convert mapstructure value of type '%T' to Decimal: %w", data, ErrUnmarshal)
	}
}

// decodeToNullDecimal converts a mapstructure source value to a NullDecimal.
// nil and empty string/bytes yield a zero NullDecimal (Valid=false).
func decodeToNullDecimal(data any) (any, error) {
	if data == nil {
		return NullDecimal{}, nil
	}
	switch v := data.(type) {
	case NullDecimal:
		return v, nil
	case *NullDecimal:
		// Reachable via ComposeDecodeHookFunc + TextUnmarshallerHookFunc.
		if v == nil {
			return NullDecimal{}, nil
		}
		return *v, nil
	case Decimal:
		return NewNullDecimal(v), nil
	case *Decimal:
		// Reachable via ComposeDecodeHookFunc + TextUnmarshallerHookFunc.
		if v == nil {
			return NullDecimal{}, nil
		}
		return NewNullDecimal(*v), nil
	case string:
		if v == "" {
			return NullDecimal{}, nil
		}
		d, err := NewFromString(v)
		if err != nil {
			return nil, err
		}
		return NewNullDecimal(d), nil
	case []byte:
		if len(v) == 0 {
			return NullDecimal{}, nil
		}
		d, err := NewFromString(string(v))
		if err != nil {
			return nil, err
		}
		return NewNullDecimal(d), nil
	case json.Number:
		d, err := NewFromString(v.String())
		if err != nil {
			return nil, err
		}
		return NewNullDecimal(d), nil
	case int:
		return NewNullDecimal(NewFromInt(v)), nil
	case int8:
		return NewNullDecimal(New(int64(v))), nil
	case int16:
		return NewNullDecimal(New(int64(v))), nil
	case int32:
		return NewNullDecimal(New(int64(v))), nil
	case int64:
		return NewNullDecimal(New(v)), nil
	case uint:
		return NewNullDecimal(NewFromUint64(uint64(v), 0)), nil
	case uint8:
		return NewNullDecimal(NewFromUint64(uint64(v), 0)), nil
	case uint16:
		return NewNullDecimal(NewFromUint64(uint64(v), 0)), nil
	case uint32:
		return NewNullDecimal(NewFromUint64(uint64(v), 0)), nil
	case uint64:
		return NewNullDecimal(NewFromUint64(v, 0)), nil
	case float32:
		d, err := decimalFromFloat32(v)
		if err != nil {
			return nil, err
		}
		return NewNullDecimal(d), nil
	case float64:
		d, err := decimalFromFloat64(v)
		if err != nil {
			return nil, err
		}
		return NewNullDecimal(d), nil
	default:
		return nil, fmt.Errorf("could not convert mapstructure value of type '%T' to NullDecimal: %w", data, ErrUnmarshal)
	}
}
