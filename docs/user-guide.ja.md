# ユーザーガイド（日本語）

このガイドでは、`github.com/exc-works/decimal` の実践的な使用パターンを示します。

## 1. インストール

```bash
go get github.com/exc-works/decimal
```

## 2. Decimal を安全に作成する

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

入力が不正な可能性がある場合は、`NewFromString` を使用します。

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. 精度とフォーマット

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` は標準表現で、末尾のゼロを取り除きます。
- `StringWithTrailingZeros()` はスケールを保った表示形式を維持します。

## 4. 基本的な算術

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

乗算・除算では、丸め動作を選択する必要があります。

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

丸めなしで正確な乗算を行う場合:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` は `MulExact` の非推奨互換エイリアスとして維持されています。

## 5. 丸めとリスケーリング

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

除算で出力精度を明示するには、`QuoWithPrec` を使用します。

## 6. 値セマンティクス

`Decimal` は算術演算および比較演算に対して不変です。

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. 比較

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

比較は数値ベースです（スケール非依存）。

## 8. シリアライズと DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` は Decimal の値を JSON 文字列として書き出します。

### SQL

`Decimal` は次の両方を実装しています。

- `driver.Valuer`
- `sql.Scanner`

そのため、一般的なデータベースドライバで直接動作します。

## 9. よくある落とし穴

1. `MustFromString` は panic するため、信頼できない入力には使用しないでください。
2. 負の精度は panic します。
3. `RoundUnnecessary` は不正確な演算で panic します。
4. `Log2()` は 0 以下の値で panic します。
5. `MarshalBinary()` は末尾のゼロを正規化します。

## 10. 推奨パターン

1. 外部入力は `NewFromString` で解析し、エラーを処理する。
2. ユーザーに表示する除算結果には `QuoWithPrec` を使用する。
3. 固定表示スケールが必要な場合にのみ `StringWithTrailingZeros` を使用する。
4. ビジネスルールでは丸めモードを明示する。
