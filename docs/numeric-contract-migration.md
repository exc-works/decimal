# Numerical contract changes (unreleased)

These changes apply to the forthcoming release containing the contract PR
stack. A release version has not yet been assigned. The
[numerical contract](numeric-contract.md) specifies the target behavior.

## Arithmetic and default precision

* Sqrt, ApproxRoot, Ln, Log10, and Exp now default to `max(input scale,20)`
  decimal places, previously `max(input scale,30)`.
* Log2 now uses that same default, previously the input scale alone. Use
  `Log2WithPrec(d.Precision())` for the previous output scale and an explicit
  error result. Use any WithPrec variant to request a specific scale such as 30.
* Negative integer powers now use HalfEven instead of RoundUp. The output scale
  remains the input scale. For example, `3.00^-1` changes from `0.34` to `0.33`;
  `New(2).Power(-1)` changes from `1` to `0`, since the exact midpoint `0.5` is
  rounded at scale 0. Supply a base with the needed scale to keep more places.
* Equal-valued instance Max/Min return the receiver, matching package Max/Min.
  The numerical value is unchanged, but the selected scale or IsNil can change.
* MulWithPrec is new. It takes an explicit output scale and rounds the exact
  product once; Mul retains max-input-scale behavior and MulExact remains exact.

Raw arithmetic behavior is unchanged; its comments now describe coefficient
units. Formatting behavior is unchanged; `%g/%G` comments now describe the
actual exponent thresholds. No new general parameter-error classification,
uniform precision cap, or nil big.Int check is introduced.

## Decoding and nullable values

* Failed decodes now preserve all receiver state, including IsNil and Valid.
  In particular, invalid JSON no longer initializes an uninitialized Decimal.
* NullDecimal.Valid alone determines whether encoding produces an empty value.
  Valid=true with an uninitialized Decimal now encodes as valid zero in
  JSON/XML/BSON, rather than null or empty content. Existing valid-zero SQL,
  Text, and YAML behavior is retained.
* Returned decoder and Scan errors consistently wrap ErrUnmarshal while retaining
  underlying categories such as ErrInvalidFormat and JSON/XML syntax error types.
  Code should use errors.Is/errors.As rather than error-string equality or direct
  concrete-type assertions against the outer wrapper.

The entry-point-specific null and empty-input rules remain as specified in the
[contract's decoding table](numeric-contract.md#decoding-states-and-nullable-values).
In particular, JSON/SQL empty strings are errors, while empty Text/Param clears
NullDecimal. Binary empty input continues to produce initialized zero.
