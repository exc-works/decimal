# User Guide (English)

This guide shows practical usage patterns for `github.com/exc-works/decimal`.

## 1. Installation

```bash
go get github.com/exc-works/decimal
```

## 2. Create Decimals Safely

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

Use `NewFromString` when input can be invalid:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. Precision and Formatting

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` is canonical and strips trailing zeros.
- `StringWithTrailingZeros()` preserves scale-style formatting.

## 4. Basic Arithmetic

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

For multiplication/division you must choose rounding behavior:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

For exact multiplication without rounding:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` is kept as a deprecated compatibility alias of `MulExact`.

## 5. Rounding and Rescaling

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

Use `QuoWithPrec` for explicit output precision in division.

## 6. Value Semantics

`Decimal` is immutable for arithmetic/comparison operations:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. Comparison

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

Comparison is numeric (scale-insensitive).

## 8. Advanced Math

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e

sqrt, _ := decimal.New(2).Sqrt()                    // 1.414213...
root, _ := decimal.New(27).ApproxRoot(3)            // 3
```

`Log10`/`Ln` require a positive receiver; `Sqrt`/`ApproxRoot` reject even
roots of negative values. All expose `*WithPrec(prec)` variants for explicit
precision control. Without one, these functions auto-bump precision to
`max(d.prec, 30)` so integer receivers still produce meaningful results.

## 9. Formatting and Display

`Decimal` implements `fmt.Formatter`, so standard verbs work:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

Locale-aware display:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. Serialization and DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` writes decimal values as JSON strings.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML` (plus attribute variants) are provided.
Uninitialized values encode as empty element/attribute.

### BSON (MongoDB)

`Decimal` implements `bson.ValueMarshaler` / `bson.ValueUnmarshaler` for
`go.mongodb.org/mongo-driver/v2/bson`. Values encode as BSON strings, and
String/Double/Int32/Int64/Decimal128/Null are accepted on decode.

### SQL

`Decimal` implements both:

- `driver.Valuer`
- `sql.Scanner`

So it works with typical database drivers directly.

### NullDecimal (nullable columns)

For columns that may be `NULL`, use `NullDecimal`:

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

`NullDecimal` supports SQL, JSON, YAML, Text, BSON, and gin binding. `null`
or empty input sets `Valid=false`.

## 11. Validator Integration

Register Decimal tags with `go-playground/validator`:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

Available tags:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (tilde-separated bounds, e.g.
  `1~100`; `min` must be `<=` `max`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (no param)
- `decimal_max_precision=N` — max number of **decimal places (scale)**, i.e.
  digits after the decimal point; **not** total significant digits. `123.45`
  has scale `2` and passes `decimal_max_precision=2`.

Built-in translations: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

Validator tag parameters must be compile-time constants. Passing malformed
parameters (non-numeric limits, unparseable decimal values, `min > max` for
`decimal_between`, negative `decimal_max_precision`) causes panics at
validation time — do not splice untrusted input into struct tags.
`RegisterGoPlaygroundValidator` is idempotent: calling it more than once on
the same `*validator.Validate` overwrites the previously registered handlers
without error.

## 12. Error Handling

The package exposes sentinel errors for `errors.Is` matching:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

Available: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidRoot`, `ErrInvalidLog`,
`ErrRoundUnnecessary`, `ErrUnmarshal`, `ErrInvalidArgument`.

## 13. Concurrency

`Decimal` values are safe for concurrent read access as long as no goroutine
reassigns the variable. Value-receiver methods (`Add`, `Cmp`, `String`, ...)
never mutate the receiver. Pointer-receiver methods (`Scan`, `UnmarshalJSON`,
...) do mutate and require external synchronization if the same `*Decimal` is
shared across goroutines.

## 14. Common Pitfalls

1. `MustFromString` panics; do not use it on untrusted input.
2. Negative precision panics.
3. `RoundUnnecessary` panics on inexact operations.
4. `Log2()` panics for non-positive values; `Log10`/`Ln` return an error.
5. `MarshalBinary()` normalizes trailing zeros.

## 15. Recommended Patterns

1. Parse external input with `NewFromString` and handle errors.
2. Use `QuoWithPrec` for any user-visible division output.
3. Use `StringWithTrailingZeros` only when fixed display scale is required.
4. Keep rounding mode explicit in business rules.
5. Use `NullDecimal` for nullable SQL columns instead of pointer workarounds.
6. Match errors via `errors.Is(err, decimal.ErrXxx)` for portable handling.
