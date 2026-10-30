package decimal

import "errors"

// Sentinel errors used by returned errors and numeric panics.
//
// Error-returning functions wrap these values. Arithmetic functions documented
// to panic retain that behavior, with error-valued payloads for the categories
// below. After recovery, assert the payload to error before using errors.Is.
// Conversion methods returning (value, bool) retain their boolean contracts.
var (
	// ErrOverflow identifies a numeric panic when an exact arithmetic
	// intermediate exceeds the supported scale or coefficient-size bound.
	// Int64, Uint64, Float32, and Float64 report conversion status via bool.
	ErrOverflow = errors.New("decimal: overflow")

	// ErrDivideByZero identifies a panic on division by zero, including
	// QuoWithPrec, Quo, QuoRem, Mod, and negative powers of zero.
	ErrDivideByZero = errors.New("decimal: division by zero")

	// ErrInvalidPrecision identifies a negative-precision panic, or a returned
	// MarshalBinary error for precision outside the supported encoding range.
	ErrInvalidPrecision = errors.New("decimal: invalid precision")

	// ErrInvalidFormat indicates NewFromString failed to parse the input.
	ErrInvalidFormat = errors.New("decimal: invalid format")

	// ErrNegativeRoot indicates an attempt to take an even root of a
	// negative value in Sqrt or ApproxRoot.
	ErrNegativeRoot = errors.New("decimal: negative value for even root")

	// ErrInvalidRoot indicates a root value that is not strictly positive
	// was supplied to ApproxRoot.
	ErrInvalidRoot = errors.New("decimal: invalid root")

	// ErrNoConvergence indicates a root or transcendental calculation could
	// not certify its result within the iteration or precision budget.
	ErrNoConvergence = errors.New("decimal: iteration did not converge")

	// ErrInvalidLog indicates an attempt to take the logarithm of a
	// non-positive value.
	ErrInvalidLog = errors.New("decimal: log of non-positive value")

	// ErrRoundUnnecessary identifies a panic when rounding was required but
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
