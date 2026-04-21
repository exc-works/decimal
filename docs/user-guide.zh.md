# 用户指南（中文）

本指南展示 `github.com/exc-works/decimal` 的实用用法。

## 1. 安装

```bash
go get github.com/exc-works/decimal
```

## 2. 安全地创建 Decimal

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

对外部输入建议使用 `NewFromString`：

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. 精度与格式

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` 用于规范输出，会去掉尾随零。
- `StringWithTrailingZeros()` 用于保留显示精度。

## 4. 基础运算

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

乘除法请明确舍入策略：

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

如果需要“不舍入的精确乘法”：

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` 保留为 `MulExact` 的 deprecated 兼容别名。

## 5. 舍入与重设精度

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  向零
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  远离零
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  向正无穷
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

做除法时建议优先使用 `QuoWithPrec` 控制输出小数位。

## 6. 不可变语义

算术和比较方法不会修改原值：

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. 比较行为

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

比较是“数值相等”，不按显示精度区分。

## 8. 高级数学运算

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ≈ 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

三者在输入非法时都会返回错误（`Log10` 与 `Ln` 要求接收者为正数）。
可以使用 `*WithPrec(prec)` 变体控制输出精度。

## 9. 格式化与显示

`Decimal` 实现了 `fmt.Formatter`，标准格式化动词都可直接使用：

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 （按 RoundHalfEven 舍入）
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

本地化显示：

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" （欧洲风格）
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. 序列化与数据库

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` 会把 decimal 编码为 JSON 字符串。

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

同时提供 `MarshalXML` / `UnmarshalXML` 以及属性形式的变体。
未初始化的值会编码为空元素或空属性。

### BSON（MongoDB）

针对 `go.mongodb.org/mongo-driver/v2/bson`，`Decimal` 实现了
`bson.ValueMarshaler` / `bson.ValueUnmarshaler`。序列化时以 BSON 字符串编码，
反序列化可接受 String、Double、Int32、Int64、Decimal128 以及 Null。

### SQL

`Decimal` 同时实现了：

- `driver.Valuer`
- `sql.Scanner`

可直接用于常见数据库驱动。

### NullDecimal（可空列）

对于可能为 `NULL` 的列，使用 `NullDecimal`：

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

`NullDecimal` 支持 SQL、JSON、YAML、Text、BSON 以及 gin 绑定。
`null` 或空字符串输入会令 `Valid=false`。

## 11. Validator 集成

向 `go-playground/validator` 注册 Decimal 相关标签：

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

可用标签：

- `decimal_required`、`decimal_eq`、`decimal_ne`、`decimal_gt`、`decimal_gte`、
  `decimal_lt`、`decimal_lte`、`decimal_between`（边界以 `~` 分隔，如 `1~100`；
  `min` 必须 `<=` `max`）
- `decimal_positive`、`decimal_negative`、`decimal_nonzero`（无参数）
- `decimal_max_precision=N` —— 限制**小数位数（scale）**，即小数点**后**的
  位数，**不是**有效数字位数。`123.45` 的 scale 为 `2`，可通过
  `decimal_max_precision=2` 的校验。

内置翻译语言：`en`、`zh`、`zh_Hant`、`ja`、`ko`、`fr`、`es`、`de`、
`pt`、`pt_BR`、`ru`、`ar`、`hi`。

验证器标签参数必须是编译期常量。若传入非法参数（无法解析为整数的限制、
无法解析为 Decimal 的数值、`decimal_between` 的 `min > max`、
`decimal_max_precision` 为负数等），会在校验时触发 panic —— 切勿将不受
信任的输入拼接进 struct tag。`RegisterGoPlaygroundValidator` 是幂等的：
对同一个 `*validator.Validate` 重复调用只会覆盖已注册的处理函数，不会报错。

## 12. 错误处理

包内导出了若干哨兵错误（sentinel error），可用 `errors.Is` 匹配：

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// 处理非法输入
}
```

可用错误：`ErrInvalidFormat`、`ErrInvalidPrecision`、`ErrOverflow`、
`ErrDivideByZero`、`ErrNegativeRoot`、`ErrInvalidLog`、`ErrRoundUnnecessary`、
`ErrUnmarshal`。

## 13. 并发

只要没有任何 goroutine 重新赋值，`Decimal` 值可以被多个 goroutine 安全地并发读取。
值接收者方法（`Add`、`Cmp`、`String` 等）不会修改接收者；
而指针接收者方法（`Scan`、`UnmarshalJSON` 等）会修改接收者本身，
如果多个 goroutine 共享同一个 `*Decimal`，需要由调用方自行同步。

## 14. 常见坑点

1. `MustFromString` 解析失败会 panic，不要用于不可信输入。
2. 负精度会 panic。
3. `RoundUnnecessary` 遇到非精确结果会 panic。
4. `Log2()` 对非正数会 panic；而 `Log10` / `Ln` 会返回错误。
5. `MarshalBinary()` 会规范化尾随零。

## 15. 推荐实践

1. 外部输入统一使用 `NewFromString` 并处理错误。
2. 用户可见的除法结果统一使用 `QuoWithPrec`。
3. 仅在确实需要固定显示位数时使用 `StringWithTrailingZeros`。
4. 在业务规则中明确指定 rounding mode，避免隐式默认。
5. 对于可空 SQL 列，使用 `NullDecimal`，不要用指针去绕。
6. 使用 `errors.Is(err, decimal.ErrXxx)` 匹配错误，便于跨层级处理。
