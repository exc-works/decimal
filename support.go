package decimal

import (
	"fmt"
	"math/big"
)

type RoundingMode int

const (
	// RoundDown rounds towards zero.
	RoundDown RoundingMode = iota
	// RoundUp rounds away from zero.
	RoundUp
	// RoundCeiling rounds towards positive infinity.
	RoundCeiling
	// RoundHalfUp rounds to nearest; ties round up.
	RoundHalfUp
	// RoundHalfDown rounds to nearest; ties round down.
	RoundHalfDown
	// RoundHalfEven rounds to nearest; ties to even.
	RoundHalfEven
	// RoundUnnecessary asserts no rounding is required.
	RoundUnnecessary
)

// MostSignificantBit returns the index of the most significant set bit in x.
// It returns 0 for x == 0 and panics if x < 0. The argument is read-only.
func MostSignificantBit(x *big.Int) uint {
	if x.Sign() < 0 {
		panic("MostSignificantBit of not positive number")
	}
	if x.Sign() == 0 {
		return 0
	}
	return uint(x.BitLen() - 1)
}

func unquoteIfQuoted(value any) (string, error) {
	var bytes []byte

	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return "", fmt.Errorf("could not convert value '%+v' to byte array of type '%T': %w", value, value, ErrUnmarshal)
	}

	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
