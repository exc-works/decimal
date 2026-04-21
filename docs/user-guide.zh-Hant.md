# 使用者指南（繁體中文）

本指南展示 `github.com/exc-works/decimal` 的實務使用模式。

## 1. 安裝

```bash
go get github.com/exc-works/decimal
```

## 2. 安全地建立 Decimal

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

當輸入可能無效時，使用 `NewFromString`：

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. 精度與格式化

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` 是標準表示，會去除尾隨零。
- `StringWithTrailingZeros()` 會保留小數位尺度（scale）風格的格式。

## 4. 基本算術

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

對於乘法/除法，你必須選擇捨入行為：

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

若要進行不捨入的精確乘法：

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` 保留為 `MulExact` 的已棄用相容別名。

## 5. 捨入與重定標

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

在除法中若要明確指定輸出精度，請使用 `QuoWithPrec`。

## 6. 值語義

對於算術/比較運算，`Decimal` 是不可變的：

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

比較基於數值（不受小數位尺度（scale）影響）。

## 8. 進階數學運算

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

以上三個函式在輸入無效時都會回傳錯誤（`Log10`／`Ln` 要求接收者為正值）。
使用 `*WithPrec(prec)` 變體以控制輸出精度。

## 9. 格式化與顯示

`Decimal` 實作了 `fmt.Formatter`，因此可使用標準的格式化動詞：

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

地區感知的顯示方式：

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. 序列化與資料庫

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` 會將十進位值寫成 JSON 字串。

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

套件提供 `MarshalXML`／`UnmarshalXML`（以及屬性版本）。
未初始化的值會編碼為空元素或空屬性。

### BSON（MongoDB）

`Decimal` 針對 `go.mongodb.org/mongo-driver/v2/bson` 實作了
`bson.ValueMarshaler`／`bson.ValueUnmarshaler`。值會編碼為 BSON 字串，
解碼時則可接受 String／Double／Int32／Int64／Decimal128／Null。

### SQL

`Decimal` 同時實作：

- `driver.Valuer`
- `sql.Scanner`

因此可直接搭配常見資料庫驅動程式使用。

### NullDecimal（可為 NULL 的欄位）

對於可能為 `NULL` 的欄位，請使用 `NullDecimal`：

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

`NullDecimal` 支援 SQL、JSON、YAML、Text、BSON 以及 gin 繫結。
當輸入為 `null` 或空字串時會將 `Valid` 設為 `false`。

## 11. 驗證器整合

將 Decimal 相關標籤註冊至 `go-playground/validator`：

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

可用的標籤：

- `decimal_required`、`decimal_eq`、`decimal_ne`、`decimal_gt`、`decimal_gte`、
  `decimal_lt`、`decimal_lte`、`decimal_between`（以 `~` 分隔上下界，例如 `1~100`）
- `decimal_positive`、`decimal_negative`、`decimal_nonzero`（無參數）
- `decimal_max_precision=N`（小數位數 ≤ N）

內建翻譯語系：`en`、`zh`、`zh_Hant`、`ja`、`ko`、`fr`、`es`、`de`、
`pt`、`pt_BR`、`ru`、`ar`、`hi`。

## 12. 錯誤處理

本套件對外公開了可供 `errors.Is` 比對的哨兵錯誤：

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

可用的錯誤：`ErrInvalidFormat`、`ErrInvalidPrecision`、`ErrOverflow`、
`ErrDivideByZero`、`ErrNegativeRoot`、`ErrInvalidLog`、`ErrRoundUnnecessary`、
`ErrUnmarshal`。

## 13. 並行處理

只要沒有任何 goroutine 重新指派該變數，`Decimal` 的值即可安全地被並行讀取。
值接收者方法（`Add`、`Cmp`、`String`、……）永遠不會改動接收者。
指標接收者方法（`Scan`、`UnmarshalJSON`、……）會修改接收者，若同一個
`*Decimal` 被多個 goroutine 共用，則需要外部同步機制。

## 14. 常見陷阱

1. `MustFromString` 會 panic；不要用在不可信輸入上。
2. 負的精度會 panic。
3. 在非精確運算上，`RoundUnnecessary` 會 panic。
4. 對非正值呼叫 `Log2()` 會 panic；`Log10`／`Ln` 則會回傳錯誤。
5. `MarshalBinary()` 會正規化尾隨零。

## 15. 建議做法

1. 使用 `NewFromString` 解析外部輸入並處理錯誤。
2. 對任何使用者可見的除法輸出使用 `QuoWithPrec`。
3. 僅在需要固定顯示小數位尺度（scale）時使用 `StringWithTrailingZeros`。
4. 在業務規則中明確指定捨入模式。
5. 對於可為 NULL 的 SQL 欄位，改用 `NullDecimal`，取代指標式的變通寫法。
6. 透過 `errors.Is(err, decimal.ErrXxx)` 比對錯誤，以獲得可移植的處理方式。
