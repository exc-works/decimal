# Benutzerhandbuch (Deutsch)

Dieser Leitfaden zeigt praktische Verwendungsmuster für `github.com/exc-works/decimal`.

## 1. Installation

```bash
go get github.com/exc-works/decimal
```

## 2. Dezimalwerte sicher erstellen

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

Verwenden Sie `NewFromString`, wenn die Eingabe ungültig sein kann:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. Präzision und Formatierung

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` ist kanonisch und entfernt nachgestellte Nullen.
- `StringWithTrailingZeros()` bewahrt eine an der Skala orientierte Formatierung.

## 4. Grundlegende Arithmetik

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

Für Multiplikation/Division müssen Sie ein Rundungsverhalten wählen:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

Für exakte Multiplikation ohne Rundung:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` bleibt als veralteter Kompatibilitätsalias von `MulExact` erhalten.

## 5. Rundung und Reskalierung

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

Verwenden Sie `QuoWithPrec` für eine explizite Ausgabepräzision bei Divisionen.

## 6. Wertsemantik

`Decimal` ist für arithmetische/Vergleichsoperationen unveränderlich:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. Vergleich

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

Der Vergleich ist numerisch (skalenunabhängig).

## 8. Serialisierung und DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` schreibt Dezimalwerte als JSON-Strings.

### SQL

`Decimal` implementiert beides:

- `driver.Valuer`
- `sql.Scanner`

Dadurch funktioniert es direkt mit üblichen Datenbanktreibern.

## 9. Häufige Stolperfallen

1. `MustFromString` löst einen Panic aus; verwenden Sie es nicht für nicht vertrauenswürdige Eingaben.
2. Negative Präzision löst einen Panic aus.
3. `RoundUnnecessary` löst bei ungenauen Operationen einen Panic aus.
4. `Log2()` löst bei nicht-positiven Werten einen Panic aus.
5. `MarshalBinary()` normalisiert nachgestellte Nullen.

## 10. Empfohlene Muster

1. Parsen Sie externe Eingaben mit `NewFromString` und behandeln Sie Fehler.
2. Verwenden Sie `QuoWithPrec` für jede benutzersichtbare Divisionsausgabe.
3. Verwenden Sie `StringWithTrailingZeros` nur, wenn eine feste Anzeigeskala erforderlich ist.
4. Halten Sie den Rundungsmodus in Geschäftsregeln explizit.
