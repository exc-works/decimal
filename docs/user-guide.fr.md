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

## 8. Sérialisation et base de données

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` écrit les valeurs décimales sous forme de chaînes JSON.

### SQL

`Decimal` implémente les deux interfaces suivantes :

- `driver.Valuer`
- `sql.Scanner`

Il fonctionne donc directement avec les pilotes de base de données classiques.

## 9. Pièges courants

1. `MustFromString` panique ; ne l'utilisez pas sur des entrées non fiables.
2. Une précision négative provoque une panique.
3. `RoundUnnecessary` panique sur les opérations inexactes.
4. `Log2()` panique pour les valeurs non positives.
5. `MarshalBinary()` normalise les zéros de fin.

## 10. Modèles recommandés

1. Analysez les entrées externes avec `NewFromString` et gérez les erreurs.
2. Utilisez `QuoWithPrec` pour toute sortie de division visible par l'utilisateur.
3. Utilisez `StringWithTrailingZeros` uniquement lorsqu'une échelle d'affichage fixe est requise.
4. Gardez le mode d'arrondi explicite dans les règles métier.
