# decimal

`decimal` is an immutable arbitrary-precision decimal type built on top of `math/big.Int`.
It keeps both unscaled integer digits and decimal precision, making it suitable for financial
and accounting workloads that require deterministic base-10 behavior.

Import path:

```go
import "github.com/exc-works/decimal"
```

## User Guides

- English User Guide: [docs/user-guide.en.md](docs/user-guide.en.md)
- Chinese User Guide: [docs/user-guide.zh.md](docs/user-guide.zh.md)

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/exc-works/decimal"
)

func main() {
	price := decimal.MustFromString("12.5000")
	fee := decimal.NewWithPrec(75, 2) // 0.75

	total := price.Add(fee)
	rounded := total.Rescale(2, decimal.RoundHalfEven)

	fmt.Println(price.String())                  // 12.5
	fmt.Println(price.StringWithTrailingZeros()) // 12.5000
	fmt.Println(total.String())                  // 13.25
	fmt.Println(rounded.String())                // 13.25
}
```

Common constants:

- `decimal.Zero`
- `decimal.One`
- `decimal.Ten`
- `decimal.Hundred`

## Core Design

`Decimal` uses immutable value semantics:

- methods like `Add`, `Sub`, `Mul`, `Quo`, and `Rescale` return new values
- pointer receiver methods (`Unmarshal*`, `Scan`) update the receiver
- `BigInt()` returns a copy, so internal state is not exposed for mutation

Example:

```go
a := decimal.MustFromString("1.20")
b := a.Add(decimal.MustFromString("0.30"))

fmt.Println(a.String()) // 1.2  (a is unchanged)
fmt.Println(b.String()) // 1.5
```

## Constructors

- `decimal.New(int64)`
- `decimal.NewFromInt(int)`
- `decimal.NewWithPrec(int64, prec)`
- `decimal.NewFromFloat64(float64)`
- `decimal.NewFromFloat32(float32)`
- `decimal.NewWithAppendPrec(int64, prec)`
- `decimal.NewFromUintWithAppendPrec(uint64, prec)`
- `decimal.NewFromBigInt(*big.Int)`
- `decimal.NewFromBigRat(*big.Rat)`
- `decimal.NewFromBigRatWithPrec(*big.Rat, prec, decimal.RoundingMode)`
- `decimal.NewFromBigIntWithPrec(*big.Int, prec)`
- `decimal.NewFromInt64(int64, precision)`
- `decimal.NewFromUint64(uint64, precision)`
- `decimal.NewFromString(string)`
- `decimal.MustFromString(string)`

### String Parsing

`NewFromString` supports:

- plain decimals: `123`, `-123.45`
- scientific notation: `1.234e3`, `123456E-3`

It trims leading/trailing spaces and rejects malformed formats such as:

- empty string
- `1.`
- `.1`
- multiple decimal points
- missing or invalid exponent

If parsing results in zero, precision is normalized to `0`.

## Arithmetic and Rounding

### Arithmetic

- `Add(Decimal)` / `SafeAdd(Decimal)` / `AddRaw(int64)`
- `Sub(Decimal)` / `SafeSub(Decimal)` / `SubRaw(int64)`
- `Mul(Decimal, decimal.RoundingMode)`
- `MulDown(Decimal)`
- `MulExact(Decimal)` (exact multiplication, no rounding, precision = `d.prec + d2.prec`)
- `Mul2(Decimal)` (deprecated alias of `MulExact`)
- `QuoWithPrec(Decimal, prec, decimal.RoundingMode)`
- `Quo(Decimal, decimal.RoundingMode)`
- `QuoDown(Decimal)`
- `QuoRem(Decimal)`
- `Mod(Decimal)`
- `Power(int64)`
- `Sqrt() (Decimal, error)`
- `ApproxRoot(int64) (Decimal, error)`
- `Log2() Decimal`

### Precision Utilities

- `RescaleDown(prec)`
- `Rescale(prec, decimal.RoundingMode)`
- `Shift(places)`
- `TruncateWithPrec(prec)` / `RoundWithPrec(prec)`
- `FloorWithPrec(prec)` / `CeilWithPrec(prec)`
- `Truncate()` / `Round()` / `Floor()` / `Ceil()`
- `StripTrailingZeros()`
- `SignificantFigures(figures, decimal.RoundingMode)`

### Comparison

- `Cmp(Decimal)`
- `Equal(Decimal)` / `NotEqual(Decimal)`
- `GT(Decimal)` / `GTE(Decimal)`
- `LT(Decimal)` / `LTE(Decimal)`
- `Max(Decimal)` / `Min(Decimal)`
- package-level helpers: `decimal.Max`, `decimal.Min`, `decimal.Between`

### Other Methods

- `IntPart()`
- `Remainder()`
- `Sign()` / `IsNegative()` / `IsZero()` / `IsNotZero()` / `IsPositive()`
- `IsInteger()` / `HasFraction()`
- `Neg()` / `Abs()`
- `BigInt()` / `BigRat()`
- `Float32() (float32, bool)` / `Float64() (float64, bool)`
- `Int64() (int64, bool)` / `Uint64() (uint64, bool)`
- `BitLen()`
- `Precision()`
- `MustNonNegative()`

### Rounding Modes

- `decimal.RoundDown` (toward zero)
- `decimal.RoundUp` (away from zero)
- `decimal.RoundCeiling` (toward +infinity)
- `decimal.RoundHalfUp`
- `decimal.RoundHalfDown`
- `decimal.RoundHalfEven` (banker's rounding)
- `decimal.RoundUnnecessary` (panics if rounding is required)

## Serialization

### String

- `String()` strips trailing zeros
- `StringWithTrailingZeros()` keeps trailing zeros

### JSON

- `MarshalJSON()` encodes as a JSON string
- `UnmarshalJSON()` accepts JSON string and (in some paths) raw JSON number text
- uninitialized value marshals as `null`

### YAML

- `MarshalYAML()` returns string form
- `UnmarshalYAML()` parses scalar string/number values

### Text

- `MarshalText()`
- `UnmarshalText()`
- `UnmarshalParam(string)` (for gin `BindUnmarshaler`)

### Gin

- `ShouldBindQuery` / `ShouldBind` / `ShouldBindUri` use `UnmarshalParam(string)`
- `ShouldBindJSON` uses `UnmarshalJSON()`

Example:

```go
type Req struct {
	Amount decimal.Decimal `form:"amount" uri:"amount" json:"amount"`
}

var req Req
if err := c.ShouldBindQuery(&req); err != nil {
	// handle error
}
```

### Binary / protobuf

- `MarshalBinary()` / `UnmarshalBinary()`
- `Marshal()` / `Unmarshal()`
- `MarshalTo([]byte)`
- `Size()`

Binary format:

- first 4 bytes: big-endian `uint32` precision
- remaining bytes: gob-encoded `big.Int`
- trailing zeros are stripped before serialization
- `decimal.PrecisionFixedSize == 4`

### Database

- `Value()` implements `driver.Valuer`
- `Scan(any)` implements `sql.Scanner`

`Scan` supports: `nil`, `float32`, `float64`, `int64`, `string`, `[]byte`, and quoted/unquoted decimal text.

## Notes and Pitfalls

- negative precision panics in constructors/rescaling
- `NewFromString` returns error; `MustFromString` panics
- `MustNonNegative` panics for negative values
- `Log2()` panics unless value `> 0`
- `ApproxRoot(root)` requires `root > 0`
- even root of negative values returns error
- `Quo` has a special integer-division path when both precisions are `0`
- binary encoding normalizes trailing zeros (`7.50` and `7.5000` can encode identically)

## Migration Notes

If you are migrating from older internal variants of this library:

- rely only on APIs present in this repository
- update formatting-sensitive code if it depended on fixed-scale output (`StringWithTrailingZeros`)
- validate binary compatibility if old code expected trailing-zero preservation

## Release and Versioning

- Semantic Versioning: <https://semver.org/>
- Changelog format: <https://keepachangelog.com/>
- changelog file: [`CHANGELOG.md`](CHANGELOG.md)
- GitHub release trigger: tags matching `v*.*.*`

Example:

```bash
git tag -a v0.1.0 -m "release v0.1.0"
git push origin v0.1.0
```
