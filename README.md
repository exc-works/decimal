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
- Chinese (Simplified) User Guide: [docs/user-guide.zh.md](docs/user-guide.zh.md)
- Chinese (Traditional) User Guide: [docs/user-guide.zh-Hant.md](docs/user-guide.zh-Hant.md)
- Japanese User Guide: [docs/user-guide.ja.md](docs/user-guide.ja.md)
- Korean User Guide: [docs/user-guide.ko.md](docs/user-guide.ko.md)
- Spanish User Guide: [docs/user-guide.es.md](docs/user-guide.es.md)
- French User Guide: [docs/user-guide.fr.md](docs/user-guide.fr.md)
- German User Guide: [docs/user-guide.de.md](docs/user-guide.de.md)
- Portuguese (Brazil) User Guide: [docs/user-guide.pt-BR.md](docs/user-guide.pt-BR.md)
- Russian User Guide: [docs/user-guide.ru.md](docs/user-guide.ru.md)
- Arabic User Guide: [docs/user-guide.ar.md](docs/user-guide.ar.md)
- Hindi User Guide: [docs/user-guide.hi.md](docs/user-guide.hi.md)

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
- `decimal.NewFromDecimal(Decimal)` (deep copy)
- `d.Clone()` (deep copy; useful after `NewFromBigInt` with an externally mutable `*big.Int`)

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
- `Sqrt() (Decimal, error)` / `SqrtWithPrec(prec)`
- `ApproxRoot(int64) (Decimal, error)` / `ApproxRootWithPrec(root, prec)`
- `Log2() Decimal`
- `Log10() (Decimal, error)` / `Log10WithPrec(prec)` (input must be > 0)
- `Ln() (Decimal, error)` / `LnWithPrec(prec)` (input must be > 0)
- `Exp() (Decimal, error)` / `ExpWithPrec(prec)` (Taylor series with argument reduction)

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
- `FormatWithSeparators(thousands, decimal rune)` for locale-aware display
  (e.g., `12345.67` → `"12,345.67"` or European `"12.345,67"`)
- `Format(fmt.State, verb rune)` implements `fmt.Formatter`, supporting
  `%v`, `%s`, `%q`, `%d`, `%f`, `%e`, `%g`, `%b` with width/precision/flags

### JSON

- `MarshalJSON()` encodes as a JSON string
- `UnmarshalJSON()` accepts JSON string and (in some paths) raw JSON number text
- uninitialized value marshals as `null`

### XML

- `MarshalXML()` / `UnmarshalXML()`
- `MarshalXMLAttr()` / `UnmarshalXMLAttr()` for use in XML attributes
- Uninitialized values encode as empty element/attribute

### BSON

- `MarshalBSONValue()` / `UnmarshalBSONValue()` via `go.mongodb.org/mongo-driver/v2/bson`
- Encodes as BSON string; uninitialized encodes as BSON null
- Decodes from String, Double, Int32, Int64, Decimal128, Null
- `NullDecimal` also implements BSON value marshaling

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

### Validator

- Use `decimal_required` to require Decimal field presence
- Built-in `omitempty` can be used as usual
- Decimal numeric comparison tags:
  `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`, `decimal_lt`, `decimal_lte`,
  `decimal_between` (tilde-separated bounds, e.g. `decimal_between=1~100`;
  `min` must be `<=` `max`)
- Sign/zero tags (no param): `decimal_positive`, `decimal_negative`, `decimal_nonzero`
- Precision tag: `decimal_max_precision=N` — max number of **decimal places
  (scale)**, i.e. digits after the decimal point; **not** total significant
  digits. `123.45` has scale `2` and passes `decimal_max_precision=2`.
- Uses exact `Decimal` comparison (`Cmp`), without `Float64` conversion
- Supports friendly error messages via translation helpers:
  `RegisterGoPlaygroundValidatorTranslations`,
  `RegisterGoPlaygroundValidatorTranslationsWithMessages`,
  and `TranslateGoPlaygroundValidationErrors`
- Built-in translation locales (13): `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`,
  `de`, `pt`, `pt_BR`, `ru`, `ar`, `hi`
- Register once before any validation; calling `RegisterGoPlaygroundValidator`
  multiple times on the same `*validator.Validate` is idempotent — later calls
  simply overwrite the previously registered handlers.

> **Safety note.** Validator tag parameters must be compile-time constants.
> Passing malformed parameters (non-numeric limits, unparseable decimal values,
> `min > max` for `decimal_between`, negative `decimal_max_precision`) causes
> panics at validation time — do not splice untrusted input into struct tags.

Example:

```go
import (
	"github.com/exc-works/decimal"
	"github.com/go-playground/validator/v10"
)

type Req struct {
	Amount decimal.Decimal `validate:"decimal_required,decimal_eq=12.34"`
}

v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)
err := v.Struct(Req{Amount: decimal.MustFromString("12.34")})
```

Friendly messages example:

```go
import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
)

enLocale := en.New()
uni := ut.New(enLocale, enLocale)
trans, _ := uni.GetTranslator("en")

_ = decimal.RegisterGoPlaygroundValidatorTranslations(v, trans)
messages := decimal.TranslateGoPlaygroundValidationErrors(err, trans)
```

Custom language template override example:

```go
_ = decimal.RegisterGoPlaygroundValidatorTranslationsWithMessages(v, trans, map[string]string{
	"decimal_required": "{0} cannot be empty",
})
```

For gin:

```go
import (
	"github.com/exc-works/decimal"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
	_ = decimal.RegisterGoPlaygroundValidator(v)
}
```

### viper / mapstructure

`decimal.DecodeHook()` returns a mapstructure-compatible hook that converts
non-string scalars (int / uint / float / `json.Number` / `[]byte` / `nil`)
into `Decimal` and `NullDecimal`. Compose it with
`mapstructure.TextUnmarshallerHookFunc()` to also cover the string path:

```go
import (
	"github.com/exc-works/decimal"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
	mapstructure.TextUnmarshallerHookFunc(),
	decimal.DecodeHook(),
)))
```

`nil` source maps to `Valid=false` for `NullDecimal` and to an error wrapping
`ErrUnmarshal` for `Decimal`. The hook lives in the main module and pulls in
no viper / mapstructure dependencies.

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

### NullDecimal

For nullable SQL columns, use `NullDecimal`:

```go
type Row struct {
    Amount decimal.NullDecimal
}

var r Row
_ = db.QueryRow("SELECT amount FROM t").Scan(&r.Amount)
if r.Amount.Valid {
    fmt.Println(r.Amount.Decimal.String())
}
```

`NullDecimal` implements `sql.Scanner`, `driver.Valuer`, JSON/YAML/Text/BSON
marshaling, and gin `UnmarshalParam`. `null`/empty input sets `Valid=false`.

## Error Handling

The package exposes sentinel errors so callers can switch on error category
via `errors.Is`:

- `ErrInvalidFormat` — malformed decimal string in `NewFromString`, `UnmarshalJSON`, etc.
- `ErrInvalidPrecision` — negative precision
- `ErrOverflow` — int64/uint64/float conversion overflow
- `ErrDivideByZero` — division by zero
- `ErrNegativeRoot` — even root of a negative value (`Sqrt`, `ApproxRoot`)
- `ErrInvalidRoot` — non-positive `root` passed to `ApproxRoot`
- `ErrInvalidLog` — logarithm of a non-positive value
- `ErrRoundUnnecessary` — rounding required under `RoundUnnecessary` mode
- `ErrUnmarshal` — binary/YAML/BSON/SQL unmarshal failures
- `ErrInvalidArgument` — invalid setup argument (e.g. nil validator/translator)

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
    // handle
}
```

## Concurrency

- `Decimal` values are safe for concurrent read access by multiple goroutines
  as long as no goroutine reassigns the variable.
- Value-receiver methods (`Add`, `Sub`, `Mul`, `Cmp`, `String`, etc.) never
  mutate the receiver and are safe to call concurrently.
- Pointer-receiver methods (`Scan`, `UnmarshalJSON`, `UnmarshalYAML`,
  `UnmarshalText`, `UnmarshalBinary`) mutate the receiver; external
  synchronization is required when the same `*Decimal` may be accessed
  concurrently.
- Accessors like `BigInt()` / `BigRat()` return defensive copies.
- Package-level constants (`Zero`, `One`, `Ten`, `Hundred`) are read-only.

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

### BSON support (optional)

BSON support is compiled out by default so that downstream projects are not
forced to pull in `go.mongodb.org/mongo-driver/v2`. To enable it, build with
the `bson` build tag:

```bash
go build -tags bson ./...
go test  -tags bson ./...
```

When the tag is set, `Decimal` and `NullDecimal` implement
`bson.ValueMarshaler` / `bson.ValueUnmarshaler` (see `marshal_bson.go`).
Without the tag, no BSON code is compiled and the MongoDB driver is not
linked into the resulting binary, keeping the core library dependency-free.
