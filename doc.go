// Package decimal provides an immutable arbitrary-precision decimal type
// built on top of math/big.Int.
//
// Decimal values keep both unscaled integer digits and decimal precision,
// which makes the package suitable for financial and accounting scenarios that
// require deterministic base-10 behavior.
package decimal
