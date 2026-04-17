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

## 8. Serialización y BD

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` escribe valores decimales como cadenas JSON.

### SQL

`Decimal` implementa ambos:

- `driver.Valuer`
- `sql.Scanner`

Por lo tanto, funciona directamente con controladores de base de datos típicos.

## 9. Errores comunes

1. `MustFromString` provoca pánico; no lo use con entradas no confiables.
2. La precisión negativa provoca pánico.
3. `RoundUnnecessary` provoca pánico en operaciones inexactas.
4. `Log2()` provoca pánico para valores no positivos.
5. `MarshalBinary()` normaliza los ceros finales.

## 10. Patrones recomendados

1. Analice la entrada externa con `NewFromString` y gestione los errores.
2. Use `QuoWithPrec` para cualquier salida de división visible para el usuario.
3. Use `StringWithTrailingZeros` solo cuando se requiera una escala de visualización fija.
4. Mantenga explícito el modo de redondeo en las reglas de negocio.
