package decimal

import "errors"

// Sentinel errors returned by the decimal package.
//
// These errors are wrapped by the package's functions so that callers may use
// errors.Is to identify a specific failure category without pattern matching
// on the error message.
var (
	// ErrOverflow indicates a conversion overflowed the target numeric type,
	// such as int64, uint64, or float.
	ErrOverflow = errors.New("decimal: overflow")

	// ErrDivideByZero indicates an attempted division by zero in Quo,
	// QuoRem, or Mod.
	ErrDivideByZero = errors.New("decimal: division by zero")

	// ErrInvalidPrecision indicates a negative precision was supplied where
	// a non-negative value is required.
	ErrInvalidPrecision = errors.New("decimal: invalid precision")

	// ErrInvalidFormat indicates NewFromString failed to parse the input.
	ErrInvalidFormat = errors.New("decimal: invalid format")

	// ErrNegativeRoot indicates an attempt to take an even root of a
	// negative value in Sqrt or ApproxRoot.
	ErrNegativeRoot = errors.New("decimal: negative value for even root")

	// ErrInvalidRoot indicates a root value that is not strictly positive
	// was supplied to ApproxRoot.
	ErrInvalidRoot = errors.New("decimal: invalid root")

	// ErrInvalidLog indicates an attempt to take the logarithm of a
	// non-positive value.
	ErrInvalidLog = errors.New("decimal: log of non-positive value")

	// ErrRoundUnnecessary indicates rounding was required but
	// RoundUnnecessary was specified.
	ErrRoundUnnecessary = errors.New("decimal: rounding is necessary but RoundUnnecessary specified")

	// ErrUnmarshal indicates an unmarshal operation failed to parse input
	// into a Decimal.
	ErrUnmarshal = errors.New("decimal: unmarshal failed")

	// ErrInvalidArgument indicates an invalid argument was supplied to a
	// package-level function, typically during setup (for example a nil
	// validator or a missing required translation message).
	ErrInvalidArgument = errors.New("decimal: invalid argument")
)
