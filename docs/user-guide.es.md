# Guía del usuario (Español)

Esta guía muestra patrones prácticos de uso para `github.com/exc-works/decimal`.

## 1. Instalación

```bash
go get github.com/exc-works/decimal
```

## 2. Crear decimales de forma segura

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

Use `NewFromString` cuando la entrada pueda ser inválida:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. Precisión y formato

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` devuelve una representación canónica y elimina los ceros finales.
- `StringWithTrailingZeros()` conserva el formato según la escala.

## 4. Aritmética básica

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

Para la multiplicación/división debe elegir el comportamiento de redondeo:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

Para la multiplicación exacta sin redondeo:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` se mantiene como un alias de compatibilidad obsoleto de `MulExact`.

## 5. Redondeo y reescalado

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

Use `QuoWithPrec` para establecer una precisión de salida explícita en la división.

## 6. Semántica de valor

`Decimal` es inmutable para operaciones aritméticas/de comparación:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. Comparación

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

La comparación es numérica (insensible a la escala).

## 8. Matemáticas avanzadas

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

Las tres funciones retornan un error para entradas inválidas (`Log10`/`Ln`
requieren un receptor positivo). Use las variantes `*WithPrec(prec)` para
controlar la precisión de salida.

## 9. Formato y visualización

`Decimal` implementa `fmt.Formatter`, por lo que los verbos estándar funcionan:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

Visualización según la configuración regional:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. Serialización y BD

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` escribe valores decimales como cadenas JSON.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

Se proporcionan `MarshalXML` / `UnmarshalXML` (además de variantes para
atributos). Los valores no inicializados se codifican como elemento/atributo
vacío.

### BSON (MongoDB)

`Decimal` implementa `bson.ValueMarshaler` / `bson.ValueUnmarshaler` para
`go.mongodb.org/mongo-driver/v2/bson`. Los valores se codifican como cadenas
BSON, y al decodificar se aceptan String/Double/Int32/Int64/Decimal128/Null.

### SQL

`Decimal` implementa ambos:

- `driver.Valuer`
- `sql.Scanner`

Por lo tanto, funciona directamente con controladores de base de datos típicos.

### NullDecimal (columnas anulables)

Para columnas que pueden ser `NULL`, utilice `NullDecimal`:

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

`NullDecimal` admite SQL, JSON, YAML, Text, BSON y binding de gin. Una entrada
`null` o vacía establece `Valid=false`.

## 11. Integración con validator

Registre las etiquetas de Decimal con `go-playground/validator`:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

Etiquetas disponibles:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (límites separados por tildes, p. ej. `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (sin parámetro)
- `decimal_max_precision=N` (decimales ≤ N)

Traducciones integradas: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. Manejo de errores

El paquete expone errores centinela para usar con `errors.Is`:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

Disponibles: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`,
`ErrUnmarshal`.

## 13. Concurrencia

Los valores `Decimal` son seguros para el acceso concurrente de lectura siempre
que ninguna goroutine reasigne la variable. Los métodos con receptor por valor
(`Add`, `Cmp`, `String`, ...) nunca mutan el receptor. Los métodos con receptor
por puntero (`Scan`, `UnmarshalJSON`, ...) sí mutan y requieren sincronización
externa si el mismo `*Decimal` se comparte entre goroutines.

## 14. Errores comunes

1. `MustFromString` provoca pánico; no lo use con entradas no confiables.
2. La precisión negativa provoca pánico.
3. `RoundUnnecessary` provoca pánico en operaciones inexactas.
4. `Log2()` provoca pánico para valores no positivos; `Log10`/`Ln` retornan un error.
5. `MarshalBinary()` normaliza los ceros finales.

## 15. Patrones recomendados

1. Analice la entrada externa con `NewFromString` y gestione los errores.
2. Use `QuoWithPrec` para cualquier salida de división visible para el usuario.
3. Use `StringWithTrailingZeros` solo cuando se requiera una escala de visualización fija.
4. Mantenga explícito el modo de redondeo en las reglas de negocio.
5. Use `NullDecimal` para columnas SQL anulables en lugar de soluciones con punteros.
6. Haga coincidir errores mediante `errors.Is(err, decimal.ErrXxx)` para un manejo portable.
