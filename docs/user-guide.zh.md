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

## 8. 序列化与数据库

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` 会把 decimal 编码为 JSON 字符串。

### SQL

`Decimal` 实现了：

- `driver.Valuer`
- `sql.Scanner`

可直接用于常见数据库驱动。

## 9. 常见坑点

1. `MustFromString` 解析失败会 panic，不要用于不可信输入。
2. 负精度会 panic。
3. `RoundUnnecessary` 遇到非精确结果会 panic。
4. `Log2()` 对非正数会 panic。
5. `MarshalBinary()` 会规范化尾随零。

## 10. 推荐实践

1. 外部输入统一使用 `NewFromString` 并处理错误。
2. 用户可见的除法结果统一使用 `QuoWithPrec`。
3. 仅在确实需要固定显示位数时使用 `StringWithTrailingZeros`。
4. 在业务规则中明确指定 rounding mode，避免隐式默认。
