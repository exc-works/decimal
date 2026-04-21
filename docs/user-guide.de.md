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

## 8. Erweiterte Mathematik

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

Alle drei geben bei ungültiger Eingabe einen Fehler zurück (`Log10`/`Ln`
erfordern einen positiven Empfänger). Verwenden Sie die Varianten
`*WithPrec(prec)`, um die Ausgabepräzision zu steuern.

## 9. Formatierung und Anzeige

`Decimal` implementiert `fmt.Formatter`, sodass die Standard-Verben funktionieren:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

Lokalisierte Anzeige:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. Serialisierung und DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` schreibt Dezimalwerte als JSON-Strings.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML` (einschließlich Varianten für Attribute) werden
bereitgestellt. Nicht initialisierte Werte werden als leeres Element bzw.
leeres Attribut kodiert.

### BSON (MongoDB)

`Decimal` implementiert `bson.ValueMarshaler` / `bson.ValueUnmarshaler` für
`go.mongodb.org/mongo-driver/v2/bson`. Werte werden als BSON-Strings kodiert;
beim Dekodieren werden String/Double/Int32/Int64/Decimal128/Null akzeptiert.

### SQL

`Decimal` implementiert beides:

- `driver.Valuer`
- `sql.Scanner`

Dadurch funktioniert es direkt mit üblichen Datenbanktreibern.

### NullDecimal (Nullable-Spalten)

Verwenden Sie `NullDecimal` für Spalten, die `NULL` sein können:

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

`NullDecimal` unterstützt SQL, JSON, YAML, Text, BSON sowie Gin-Binding. Die
Eingabe `null` oder eine leere Eingabe setzt `Valid=false`.

## 11. Validator-Integration

Registrieren Sie Decimal-Tags mit `go-playground/validator`:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

Verfügbare Tags:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (durch Tilde getrennte Grenzen, z. B. 1~100)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (ohne Parameter)
- `decimal_max_precision=N` (Dezimalstellen ≤ N)

Integrierte Übersetzungen: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. Fehlerbehandlung

Das Paket stellt Sentinel-Fehler für den Abgleich mit `errors.Is` bereit:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

Verfügbar: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`,
`ErrUnmarshal`.

## 13. Nebenläufigkeit

`Decimal`-Werte sind für gleichzeitigen Lesezugriff sicher, solange keine
Goroutine die Variable neu zuweist. Methoden mit Wert-Empfänger (`Add`, `Cmp`,
`String`, ...) verändern den Empfänger niemals. Methoden mit Pointer-Empfänger
(`Scan`, `UnmarshalJSON`, ...) verändern ihn hingegen und erfordern eine
externe Synchronisation, wenn derselbe `*Decimal` von mehreren Goroutinen
gemeinsam genutzt wird.

## 14. Häufige Fallstricke

1. `MustFromString` löst einen Panic aus; verwenden Sie es nicht für nicht vertrauenswürdige Eingaben.
2. Negative Präzision löst einen Panic aus.
3. `RoundUnnecessary` löst bei ungenauen Operationen einen Panic aus.
4. `Log2()` löst bei nicht-positiven Werten einen Panic aus; `Log10`/`Ln` geben stattdessen einen Fehler zurück.
5. `MarshalBinary()` normalisiert nachgestellte Nullen.

## 15. Empfohlene Muster

1. Parsen Sie externe Eingaben mit `NewFromString` und behandeln Sie Fehler.
2. Verwenden Sie `QuoWithPrec` für jede benutzersichtbare Divisionsausgabe.
3. Verwenden Sie `StringWithTrailingZeros` nur, wenn eine feste Anzeigeskala erforderlich ist.
4. Halten Sie den Rundungsmodus in Geschäftsregeln explizit.
5. Verwenden Sie `NullDecimal` für Nullable-SQL-Spalten anstelle von Pointer-Behelfslösungen.
6. Gleichen Sie Fehler mit `errors.Is(err, decimal.ErrXxx)` ab, um eine portable Fehlerbehandlung zu ermöglichen.
