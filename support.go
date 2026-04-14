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

// MostSignificantBit returns the index of the most significant set bit.
func MostSignificantBit(x *big.Int) uint {
	if x.Sign() < 0 {
		panic("MostSignificantBit of not positive number")
	}
	if x.Sign() == 0 {
		return 0
	}

	var msb uint
	for bitLen := x.BitLen(); bitLen > 0; {
		bitLenHalf := bitLen >> 1
		if bitLenHalf<<1 != bitLen {
			bitLenHalf++
		}
		mask := new(big.Int).Lsh(big.NewInt(1), uint(bitLenHalf))
		mask = mask.Sub(mask, big.NewInt(1))
		if x.Cmp(mask) >= 0 {
			msb += uint(bitLenHalf)
			bitLen -= bitLenHalf
			x = x.Rsh(x, uint(bitLenHalf))
		} else {
			bitLen = bitLenHalf
		}
	}
	return msb - 1
}

func unquoteIfQuoted(value any) (string, error) {
	var bytes []byte

	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return "", fmt.Errorf("could not convert value '%+v' to byte array of type '%T'", value, value)
	}

	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
