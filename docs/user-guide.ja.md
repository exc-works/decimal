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

## 8. 高度な数学関数

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

これら 3 つの関数は、不正な入力に対してエラーを返します（`Log10` および `Ln` はレシーバが正の値であることを要求します）。出力精度を制御するには、`*WithPrec(prec)` バリアントを使用してください。

## 9. フォーマットと表示

`Decimal` は `fmt.Formatter` を実装しているため、標準の動詞をそのまま使用できます。

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

ロケールを考慮した表示:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (ヨーロッパ式)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. シリアル化と DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` は Decimal の値を JSON 文字列として書き出します。

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML`（および属性用のバリアント）が提供されています。未初期化の値は、空の要素または属性としてエンコードされます。

### BSON (MongoDB)

`Decimal` は `go.mongodb.org/mongo-driver/v2/bson` 向けに `bson.ValueMarshaler` / `bson.ValueUnmarshaler` を実装しています。値は BSON 文字列としてエンコードされ、デコード時には String/Double/Int32/Int64/Decimal128/Null が受け付けられます。

### SQL

`Decimal` は次の両方を実装しています。

- `driver.Valuer`
- `sql.Scanner`

そのため、一般的なデータベースドライバで直接動作します。

### NullDecimal（NULL 許容カラム）

`NULL` の可能性があるカラムには、`NullDecimal` を使用します。

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

`NullDecimal` は SQL、JSON、YAML、Text、BSON、および gin バインディングをサポートします。`null` または空の入力の場合、`Valid=false` に設定されます。

## 11. バリデータの統合

`go-playground/validator` に Decimal 用のタグを登録します。

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

利用可能なタグ:

- `decimal_required`、`decimal_eq`、`decimal_ne`、`decimal_gt`、`decimal_gte`、
  `decimal_lt`、`decimal_lte`、`decimal_between`（チルダ区切りの上下限、例: `1~100`）
- `decimal_positive`、`decimal_negative`、`decimal_nonzero`（パラメータなし）
- `decimal_max_precision=N`（小数点以下の桁数が N 以下）

組み込みの翻訳: `en`、`zh`、`zh_Hant`、`ja`、`ko`、`fr`、`es`、`de`、
`pt`、`pt_BR`、`ru`、`ar`、`hi`。

## 12. エラーハンドリング

本パッケージは `errors.Is` によるマッチング用にセンチネルエラーを公開しています。

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// 処理する
}
```

利用可能なもの: `ErrInvalidFormat`、`ErrInvalidPrecision`、`ErrOverflow`、
`ErrDivideByZero`、`ErrNegativeRoot`、`ErrInvalidLog`、`ErrRoundUnnecessary`、
`ErrUnmarshal`。

## 13. 並行性

`Decimal` の値は、どのゴルーチンもその変数を再代入しない限り、並行的な読み取りアクセスに対して安全です。値レシーバのメソッド（`Add`、`Cmp`、`String`、...）はレシーバを変更することはありません。ポインタレシーバのメソッド（`Scan`、`UnmarshalJSON`、...）は値を変更するため、同じ `*Decimal` を複数のゴルーチンで共有する場合は、外部での同期が必要です。

## 14. よくある落とし穴

1. `MustFromString` は panic するため、信頼できない入力には使用しないでください。
2. 負の精度は panic します。
3. `RoundUnnecessary` は不正確な演算で panic します。
4. `Log2()` は 0 以下の値で panic します。`Log10` / `Ln` はエラーを返します。
5. `MarshalBinary()` は末尾のゼロを正規化します。

## 15. 推奨パターン

1. 外部入力は `NewFromString` で解析し、エラーを処理する。
2. ユーザーに表示する除算結果には `QuoWithPrec` を使用する。
3. 固定表示スケールが必要な場合にのみ `StringWithTrailingZeros` を使用する。
4. ビジネスルールでは丸めモードを明示する。
5. NULL 許容の SQL カラムには、ポインタによる回避策ではなく `NullDecimal` を使用する。
6. 移植性のあるエラーハンドリングには、`errors.Is(err, decimal.ErrXxx)` でエラーをマッチングする。
