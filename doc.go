// Package decimal provides an immutable arbitrary-precision decimal type
// built on top of math/big.Int.
//
// Decimal values keep both unscaled integer digits and decimal precision,
// which makes the package suitable for financial and accounting scenarios that
// require deterministic base-10 behavior.
//
// # Concurrency
//
// A Decimal value is safe for concurrent read access by multiple goroutines
// provided that no goroutine reassigns the variable holding it. The type is
// designed around immutable semantics: arithmetic and inspection methods have
// value receivers (for example Add, Sub, Mul, Quo, Cmp, Sign, String,
// StringWithTrailingZeros, IntPart, Precision, IsZero, MarshalJSON,
// MarshalText, MarshalBinary) and return new Decimal values without mutating
// the receiver, so invoking them concurrently on the same Decimal is safe.
//
// Methods with pointer receivers mutate the receiver and therefore require
// external synchronization whenever the same *Decimal may be accessed from
// more than one goroutine. These include the decoding entry points used by
// the standard library and popular frameworks: Scan, UnmarshalJSON,
// UnmarshalYAML, UnmarshalText, UnmarshalBinary, and UnmarshalParam.
//
// Accessors that expose the underlying math/big types (for example BigInt
// and BigRat) return defensive copies rather than the internal state, so the
// returned *big.Int or *big.Rat may be read or mutated by the caller without
// affecting other goroutines that share the original Decimal.
//
// The package-level constants Zero, One, Ten, and Hundred are intended to be
// treated as read-only singletons; do not pass them to APIs that would
// mutate a Decimal in place.
package decimal
