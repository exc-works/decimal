# Guide d'utilisation (Français)

Ce guide présente des modèles d'utilisation pratiques pour `github.com/exc-works/decimal`.

## 1. Installation

```bash
go get github.com/exc-works/decimal
```

## 2. Créer des décimaux en toute sécurité

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

Utilisez `NewFromString` lorsque l'entrée peut être invalide :

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. Précision et formatage

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` est canonique et supprime les zéros de fin.
- `StringWithTrailingZeros()` conserve le format de type échelle.

## 4. Arithmétique de base

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

Pour la multiplication/division, vous devez choisir le comportement d'arrondi :

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

Pour une multiplication exacte sans arrondi :

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` est conservé comme alias de compatibilité déprécié de `MulExact`.

## 5. Arrondi et changement d'échelle

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

Utilisez `QuoWithPrec` pour une précision de sortie explicite lors d'une division.

## 6. Sémantique de valeur

`Decimal` est immuable pour les opérations arithmétiques/de comparaison :

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. Comparaison

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

La comparaison est numérique (insensible à l'échelle).

## 8. Mathématiques avancées

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

Les trois fonctions renvoient une erreur en cas d'entrée invalide (`Log10`/`Ln` exigent un récepteur strictement positif). Utilisez les variantes `*WithPrec(prec)` pour contrôler la précision de sortie.

## 9. Formatage et affichage

`Decimal` implémente `fmt.Formatter`, ce qui permet d'utiliser les verbes standard :

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

Affichage adapté à la locale :

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. Sérialisation et base de données

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` écrit les valeurs décimales sous forme de chaînes JSON.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML` (ainsi que leurs variantes pour les attributs) sont fournis. Les valeurs non initialisées sont encodées sous forme d'élément/d'attribut vide.

### BSON (MongoDB)

`Decimal` implémente `bson.ValueMarshaler` / `bson.ValueUnmarshaler` pour `go.mongodb.org/mongo-driver/v2/bson`. Les valeurs sont encodées sous forme de chaînes BSON, et les types String/Double/Int32/Int64/Decimal128/Null sont acceptés au décodage.

### SQL

`Decimal` implémente les deux interfaces suivantes :

- `driver.Valuer`
- `sql.Scanner`

Il fonctionne donc directement avec les pilotes de base de données classiques.

### NullDecimal (colonnes nullables)

Pour les colonnes pouvant être `NULL`, utilisez `NullDecimal` :

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

`NullDecimal` prend en charge SQL, JSON, YAML, Text, BSON et le binding gin. Une entrée `null` ou vide positionne `Valid=false`.

## 11. Intégration avec le validateur

Enregistrez les tags Decimal auprès de `go-playground/validator` :

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

Tags disponibles :

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`, `decimal_lt`, `decimal_lte`, `decimal_between` (bornes séparées par un tilde, p. ex. `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (sans paramètre)
- `decimal_max_precision=N` (nombre de décimales ≤ N)

Traductions intégrées : `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`, `pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. Décodage de configuration (viper / mapstructure)

`decimal.DecodeHook()` retourne un hook compatible mapstructure qui décode
les valeurs de configuration (`string`, `int`, `uint`, `float`,
`json.Number`, `[]byte`, `nil`) en `Decimal` et `NullDecimal`. Il est conçu
pour être composé avec `mapstructure.TextUnmarshallerHookFunc()`, qui gère
déjà le chemin chaîne via `UnmarshalText`. L'ordre n'est pas strict ──
`decimal.DecodeHook()` gère les chaînes seul ── mais l'ordre canonique
ci-dessous correspond à l'exemple du README :

```go
import (
	"github.com/exc-works/decimal"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Price    decimal.Decimal     `mapstructure:"price"`
	Discount decimal.NullDecimal `mapstructure:"discount"`
}

var cfg Config
err := viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
	mapstructure.TextUnmarshallerHookFunc(),
	decimal.DecodeHook(),
)))
```

Comportement :

- `Decimal` + `nil` / chaîne vide / `[]byte` vide → erreur enveloppant `ErrUnmarshal` (`Decimal` ne peut pas représenter un SQL NULL).
- `NullDecimal` + `nil` / chaîne vide / `[]byte` vide → valeur zéro (`Valid: false`).
- Une source `bool` est **rejetée** pour les deux cibles, afin d'éviter de mapper silencieusement `false`/`true` sur `0`/`1`.
- Les flottants `NaN` et `±Inf` produisent une erreur enveloppant `ErrUnmarshal`.

Le hook réside dans le module principal et n'introduit **aucune**
dépendance vers viper ou mapstructure. Il fonctionne aussi avec tout autre
décodeur basé sur mapstructure (koanf, confita, cleanenv).

## 13. Gestion des erreurs

Le package expose des erreurs sentinelles compatibles avec `errors.Is` :

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

Disponibles : `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`, `ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`, `ErrUnmarshal`.

## 14. Concurrence

Les valeurs `Decimal` sont sûres en accès concurrent en lecture tant qu'aucune goroutine ne réassigne la variable. Les méthodes à récepteur valeur (`Add`, `Cmp`, `String`, ...) ne modifient jamais le récepteur. Les méthodes à récepteur pointeur (`Scan`, `UnmarshalJSON`, ...) modifient le récepteur et exigent une synchronisation externe si le même `*Decimal` est partagé entre plusieurs goroutines.

## 15. Pièges courants

1. `MustFromString` panique ; ne l'utilisez pas sur des entrées non fiables.
2. Une précision négative provoque une panique.
3. `RoundUnnecessary` panique sur les opérations inexactes.
4. `Log2()` panique pour les valeurs non positives ; `Log10`/`Ln` renvoient une erreur.
5. `MarshalBinary()` normalise les zéros de fin.

## 16. Modèles recommandés

1. Analysez les entrées externes avec `NewFromString` et gérez les erreurs.
2. Utilisez `QuoWithPrec` pour toute sortie de division visible par l'utilisateur.
3. Utilisez `StringWithTrailingZeros` uniquement lorsqu'une échelle d'affichage fixe est requise.
4. Gardez le mode d'arrondi explicite dans les règles métier.
5. Utilisez `NullDecimal` pour les colonnes SQL nullables plutôt que des contournements à base de pointeurs.
6. Faites correspondre les erreurs via `errors.Is(err, decimal.ErrXxx)` pour une gestion portable.
