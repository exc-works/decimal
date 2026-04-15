# decimal

`decimal` 包（仓库根目录）提供一个基于 `math/big.Int` 的小数类型 `Decimal`。它同时保留整数部分和精度信息，适合需要精确十进制运算、序列化和数据库存储的场景。

导入路径：

```go
import "github.com/exc-works/decimal"
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/exc-works/decimal"
)

func main() {
	price := decimal.MustFromString("12.5000")
	fee := decimal.NewWithPrec(75, 2) // 0.75

	total := price.Add(fee)
	rounded := total.Rescale(2, decimal.RoundHalfEven)

	fmt.Println(price.String())
	fmt.Println(price.StringWithTrailingZeros())
	fmt.Println(total.String())
	fmt.Println(rounded.String())
}
```

常用全局值：

- `decimal.Zero`
- `decimal.One`
- `decimal.Ten`
- `decimal.Hundred`

## 构造方法

`Decimal` 采用**不可变值语义**：构造函数与算术/比较方法都会返回新值，不会修改原值。  
只有显式的指针接收者方法（如 `Unmarshal*`、`Scan`）会写入接收者本身。

### 不可变语义

- `a.Add(b)`、`a.Sub(b)`、`a.Rescale(...)` 等调用不会改变 `a`
- 需要“更新值”时，请显式接收返回值（例如 `a = a.Add(b)`）
- `BigInt()` 返回底层整数的副本，不会暴露可写内部状态

示例：

```go
a := decimal.MustFromString("1.20")
b := a.Add(decimal.MustFromString("0.30"))

fmt.Println(a.String()) // 1.2（a 不变）
fmt.Println(b.String()) // 1.5
```

- `decimal.New(int64)`：按整数构造，精度为 `0`
- `decimal.NewFromInt(int)`：按整数构造，精度为 `0`
- `decimal.NewWithPrec(int64, prec)`：按整数值构造，并指定精度
- `decimal.NewFromFloat64(float64)`：通过十进制字符串转换构造
- `decimal.NewFromFloat32(float32)`：通过十进制字符串转换构造
- `decimal.NewWithAppendPrec(int64, prec)`：按整数值构造，并补足 `prec` 位小数零
- `decimal.NewFromUintWithAppendPrec(uint64, prec)`：与上面类似，但输入为无符号整数
- `decimal.NewFromBigInt(*big.Int)`：从 `big.Int` 构造，精度为 `0`
- `decimal.NewFromBigRat(*big.Rat)`：从 `big.Rat` 构造，无法精确表示为有限小数时返回错误
- `decimal.NewFromBigRatWithPrec(*big.Rat, prec, decimal.RoundingMode)`：从 `big.Rat` 构造，并按指定精度和舍入方式转换
- `decimal.NewFromBigIntWithPrec(*big.Int, prec)`：从 `big.Int` 构造，并指定精度
- `decimal.NewFromInt64(int64, precision)`：从 `int64` 构造，并指定精度
- `decimal.NewFromUint64(uint64, precision)`：从 `uint64` 构造，并指定精度
- `decimal.NewFromString(string)`：解析十进制字符串，返回 `(Decimal, error)`
- `decimal.MustFromString(string)`：与 `NewFromString` 相同，但解析失败会 panic

### 字符串输入规则

`NewFromString` 支持：

- 普通十进制：`123`、`-123.45`
- 科学计数法：`1.234e3`、`123456E-3`

它会忽略首尾空白，但会拒绝以下格式：

- 空字符串
- `1.`
- `.1`
- 多个小数点
- 指数部分缺失或非法

如果解析后结果为 `0`，内部精度会被规整为 `0`。

## 运算与舍入

`Decimal` 提供常见算术、比较、取整和根运算。

### 算术

- `Add(Decimal)` / `SafeAdd(Decimal)` / `AddRaw(int64)`
- `Sub(Decimal)` / `SafeSub(Decimal)` / `SubRaw(int64)`
- `Mul(Decimal, decimal.RoundingMode)`
- `MulDown(Decimal)`
- `Mul2(Decimal)`：结果精度为两个操作数精度之和
- `QuoWithPrec(Decimal, prec, decimal.RoundingMode)`
- `Quo(Decimal, decimal.RoundingMode)`
- `QuoDown(Decimal)`
- `QuoRem(Decimal)`：返回截断商和余数
- `Mod(Decimal)`：返回与 `QuoRem` 一致的余数
- `Power(int64)`
- `Sqrt() (Decimal, error)`
- `ApproxRoot(int64) (Decimal, error)`
- `Log2() Decimal`

### 精度调整

- `RescaleDown(prec)`：按 `RoundDown` 调整到目标精度
- `Rescale(prec, decimal.RoundingMode)`：按指定舍入模式调整到目标精度
- `TruncateWithPrec(prec)` / `RoundWithPrec(prec)`：向零截断 / half-even 到目标精度
- `FloorWithPrec(prec)` / `CeilWithPrec(prec)`：向负无穷 / 向正无穷到目标精度
- `Truncate()` / `Round()` / `Floor()` / `Ceil()`：以上能力在 `prec=0` 下的便捷形式
- `StripTrailingZeros()`：移除尾随零
- `SignificantFigures(figures, decimal.RoundingMode)`：保留指定有效数字

### 比较

- `Cmp(Decimal)`
- `Equal(Decimal)` / `NotEqual(Decimal)`
- `GT(Decimal)` / `GTE(Decimal)`
- `LT(Decimal)` / `LTE(Decimal)`
- `Max(Decimal)` / `Min(Decimal)`
- 包级函数 `decimal.Max(a, b)`、`decimal.Min(a, b)`、`decimal.Between(v, lower, upper)`

### 其他常用方法

- `IntPart()`：返回整数部分 `*big.Int`
- `Remainder()`：返回整数部分和小数部分
- `Sign()`、`IsNegative()`、`IsZero()`、`IsNotZero()`、`IsPositive()`
- `IsInteger()`、`HasFraction()`
- `Neg()`、`Abs()`
- `BigInt()`：返回底层 `*big.Int` 的副本
- `Float32() (float32, bool)`：转换为 `float32`，并返回是否精确
- `Float64() (float64, bool)`：转换为 `float64`，并返回是否精确
- `Int64() (int64, bool)` / `Uint64() (uint64, bool)`：仅在“可精确表示”时返回 `ok=true`
- `BitLen()`：返回底层整数的 bit 长度
- `Precision()`：返回当前精度
- `MustNonNegative()`：负数会 panic

### 舍入模式

`Rescale`、`Mul`、`Quo`、`QuoWithPrec`、`SignificantFigures` 使用 `RoundingMode`：

- `decimal.RoundDown`
- `decimal.RoundUp`
- `decimal.RoundCeiling`
- `decimal.RoundHalfUp`
- `decimal.RoundHalfDown`
- `decimal.RoundHalfEven`
- `decimal.RoundUnnecessary`

说明：

- `RoundDown` 表示向零舍入
- `RoundUp` 表示远离零舍入
- `RoundCeiling` 表示向正无穷舍入
- `RoundHalfEven` 也叫 bankers rounding
- `RoundUnnecessary` 要求结果必须精确，否则会 panic

## 序列化

### 字符串表示

- `String()`：去掉尾随零后输出
- `StringWithTrailingZeros()`：保留尾随零输出

示例：

```go
d := decimal.MustFromString("7.5000")
fmt.Println(d.String())               // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
```

### JSON

- `MarshalJSON()`：输出 JSON 字符串
- `UnmarshalJSON()`：接受 JSON 字符串；在某些场景下也可兼容未加引号的数值文本
- `nil` 值会被编码为 `null`

### YAML

- `MarshalYAML()`：返回字符串值
- `UnmarshalYAML()`：解析 YAML 标量（字符串/数字）

### Text

- `MarshalText()`：输出十进制文本
- `UnmarshalText()`：解析十进制文本

### Binary / protobuf

- `MarshalBinary()` / `UnmarshalBinary()`
- `Marshal()` / `Unmarshal()`
- `MarshalTo([]byte)`
- `Size()`

二进制格式特点：

- 前 4 字节是大端序 `uint32` 精度
- 后续内容是 `big.Int` 的 gob 编码
- 序列化前会先移除尾随零
- `decimal.PrecisionFixedSize` 的值为 `4`

### 数据库

- `Value()`：实现 `driver.Valuer`
- `Scan(any)`：实现 `sql.Scanner`

`Scan` 支持：

- `nil`
- `float32`
- `float64`
- `int64`
- `string`
- `[]byte`
- 带引号或不带引号的字符串文本

## 错误 / 注意事项

- 精度参数不能为负数，相关构造和 `Rescale` 会 panic
- `NewFromString` 会返回错误，不会 panic；`MustFromString` 会 panic
- `MustNonNegative()` 遇到负数会 panic
- `Log2()` 要求输入严格大于 `0`，否则会 panic
- `ApproxRoot(root)` 要求 `root > 0`
- 偶数次方根不能作用于负数
- `Quo` 在两个操作数精度都为 `0` 时会走特殊路径，以保持整数除法的一致性
- `MarshalBinary()` 会先去掉尾随零，因此 `7.50` 和 `7.5000` 序列化后可能得到相同的规范化结果
- `String()` 会去掉尾随零，如果你需要保留格式，请使用 `StringWithTrailingZeros()`

## 迁移说明

如果你是从旧版 decimal API 迁移过来，请注意：

- 当前实现只保留现在仓库中实际存在的 API
- 迁移时不要依赖历史兼容接口，即使它们曾在其他代码库里存在过
- 如果旧代码依赖精度固定格式输出，检查是否需要把 `String()` 改成 `StringWithTrailingZeros()`
- 如果旧代码依赖序列化保留尾随零，请注意二进制序列化会先规范化尾随零

## 参考用法

```go
x := decimal.MustFromString("1.20")
y := decimal.MustFromString("2.34")

sum := x.Add(y)
product := x.Mul(y, decimal.RoundHalfEven)
quotient := y.Quo(x, decimal.RoundDown)

fmt.Println(sum.String())      // 3.54
fmt.Println(product.String())   // 2.81
fmt.Println(quotient.String())  // 1.95
```
