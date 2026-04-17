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

## 8. 序列化與資料庫

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` 會將十進位值寫成 JSON 字串。

### SQL

`Decimal` 同時實作：

- `driver.Valuer`
- `sql.Scanner`

因此可直接搭配常見資料庫驅動使用。

## 9. 常見陷阱

1. `MustFromString` 會 panic；不要用在不可信輸入上。
2. 負的精度會 panic。
3. 在非精確運算上，`RoundUnnecessary` 會 panic。
4. 對非正值呼叫 `Log2()` 會 panic。
5. `MarshalBinary()` 會正規化尾隨零。

## 10. 建議做法

1. 使用 `NewFromString` 解析外部輸入並處理錯誤。
2. 對任何使用者可見的除法輸出使用 `QuoWithPrec`。
3. 僅在需要固定顯示小數位尺度（scale）時使用 `StringWithTrailingZeros`。
4. 在業務規則中明確指定捨入模式。
