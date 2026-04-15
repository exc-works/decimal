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

## 8. Serialization and DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` writes decimal values as JSON strings.

### SQL

`Decimal` implements both:

- `driver.Valuer`
- `sql.Scanner`

So it works with typical database drivers directly.

## 9. Common Pitfalls

1. `MustFromString` panics; do not use it on untrusted input.
2. Negative precision panics.
3. `RoundUnnecessary` panics on inexact operations.
4. `Log2()` panics for non-positive values.
5. `MarshalBinary()` normalizes trailing zeros.

## 10. Recommended Patterns

1. Parse external input with `NewFromString` and handle errors.
2. Use `QuoWithPrec` for any user-visible division output.
3. Use `StringWithTrailingZeros` only when fixed display scale is required.
4. Keep rounding mode explicit in business rules.
