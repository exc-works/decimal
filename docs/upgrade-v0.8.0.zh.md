# v0.7.0 → v0.8.0 升级说明

[English](numeric-contract-migration.md) · [数值契约](numeric-contract.md) ·
[发布说明](releases/v0.8.0.md)

v0.8.0 包含会改变计算结果和编码输出的修复。模块路径仍为
`github.com/exc-works/decimal`，最低 Go 版本仍为 1.26.0；没有移除 v0.7.0
的导出 API。编译通过后，仍需核对业务中的精度、舍入和持久化数据。

## 1. 默认输出精度统一为至少 20 位小数

这里的精度是小数位数，不是有效数字位数。

| 方法 | v0.7.0 默认输出 scale | v0.8.0 默认输出 scale |
| --- | --- | --- |
| Sqrt、ApproxRoot、Ln、Log10、Exp | `max(输入 scale,30)` | `max(输入 scale,20)` |
| Log2 | 输入 scale | `max(输入 scale,20)` |

需要原来的 30 位默认 scale 时，可显式调用
`d.SqrtWithPrec(max(d.Precision(), 30))`，其他方法同理。
需要 Log2 原来的 scale 时，使用 `d.Log2WithPrec(d.Precision())` 并处理错误。
这些调用只保留原输出位数；修正后的数值可能与旧版本不同。

根、对数和指数函数成功时返回经 HalfEven 正确舍入的结果；无法在计算预算内
确认结果时返回 `ErrNoConvergence`，不再以成功状态返回未经确认的近似值。
`Log2` 仍通过 panic 报告失败，新增 `Log2WithPrec` 返回错误。构造器接受一个
数值，不代表该数值上的所有运算都一定能完成；具体预算和必须成功的用例见
[数值契约](numeric-contract.md)及[正确性门槛](correctness-gates.md)。

## 2. 负整数幂改为 HalfEven

Power 在输入 scale 上对最终精确结果舍入，负指数不再使用 RoundUp。

| 表达式 | v0.7.0 | v0.8.0 |
| --- | --- | --- |
| `MustFromString("3.00").Power(-1)` | `0.34` | `0.33` |
| `New(2).Power(-1)` | `1` | `0`，因为在 scale 0 上将 0.5 舍入到偶数 |
| `MustFromString("2.00").Power(-1)` | `0.50` | `0.50`，scale 为 2 |

表中保留末尾零是为了展示 scale；`String()` 会去掉这些零。
需要更多小数位时，应在 Power **之前**构造或扩展底数的 scale。
先算出零、再 Rescale 无法恢复已丢弃的精度。`0^0=1` 保持不变；零的负次幂
会 panic `ErrDivideByZero`。精确中间结果超出支持范围时 panic `ErrOverflow`，
例如不可实际计算的 MinInt64 次幂；不再发生整数溢出或无限递归。

新增 `MulWithPrec` 可以指定乘法结果的 scale，并只舍入一次：

```go
a := decimal.MustFromString("1.23")
b := decimal.MustFromString("2.34")
p := a.MulWithPrec(b, 3, decimal.RoundHalfEven) // 2.878
```

Mul 仍使用两个输入的较大 scale，MulExact 使用两者之和。乘法舍入后再扩展
scale 无法找回精度。AddRaw/SubRaw 的行为也未改变：`1.23.AddRaw(1)` 是 `1.24`。

还需更新依赖旧错误结果的断言：`16.0 / 11` 在 scale 0 上 HalfEven 舍入应为
`1`；根的中点舍入、对数和指数的末位也有修正。显式精度的 `%g/%G` 改为
舍入后选择表示法，例如 `999.9` 的 `%.3g` 输出 `1e+03`。
实例 Max/Min 在数值相等时返回接收者本身，因此 scale 或 IsNil 可能改变。

## 3. 更新错误处理与 recover

除零、负精度、运算大小溢出、RoundUnnecessary 要求舍入等数值 panic，
现在使用可通过 `errors.Is` 判断的 error 值。Log2 的定义域错误和无法确认
舍入分别暴露 `ErrInvalidLog`、`ErrNoConvergence`。
已有 `recover().(string)` 或字符串比较应改成先断言 error，再检查对应错误。

解码和 Scan 返回的错误统一包含 `ErrUnmarshal`，并保留 `ErrInvalidFormat`
及 JSON/XML 底层错误类型。使用 `errors.Is/errors.As`，不要依赖最外层具体类型
或完整错误字符串。解码失败后，原 value、scale、IsNil、Valid 都保持不变。

非法舍入模式即使遇到零、精确结果或相同 scale 也会被拒绝；`--1`、`1e--2`
等重复符号输入现在会解析失败。validator/翻译注册的 nil 参数统一返回
`ErrInvalidArgument`，其他编码错误没有新增统一的 panic 分类。
Int64/Uint64/Float32/Float64 仍返回 `(value,bool)`；浮点转换的 bool 表示能否
精确表示原数值，并不表示一般意义上的转换成功。

## 4. 核对 nullable、YAML 和 Binary 数据

NullDecimal.Valid 是唯一的 null 标志。Valid=true、Decimal 未初始化时表示
有效零：`NewNullDecimal(Decimal{})` 现在编码成 JSON `"0"`、XML 文本 `0`、
BSON 字符串 `0`。需要 null 时使用 Valid=false。编码不会修改原对象状态。

YAML 数值从原始标量文本解析，避免 float64 中间值丢失高精度数字。
yaml.v2/v3 的未加引号数值仍遵循其 YAML 规则：`012` 是十，`"012"` 作为
十进制字符串是十二。请核对配置文件的引号和旧版本已经丢失精度的快照。

Binary 的线格式没有改变，仍会规范化小数末尾零，也不保证保留 scale 或
IsNil。不支持的编码 scale 和非法系数字节现在返回错误，解码失败不会改写原值。
应检查此前忽略编码/解码错误的调用。

各格式的 null 和空值含义不同，详见[解码表](numeric-contract.md#decoding-states-and-nullable-values)。
例如 JSON/SQL 空字符串是错误；空 Text/Param 清空 NullDecimal；空 Binary
解码为已初始化零。第三方框架可能在调用本库前自行处理 null。

## 5. 升级步骤

v0.8.0 tag 正式发布后执行：

```sh
go get github.com/exc-works/decimal@v0.8.0
go mod tidy
go test ./...
```

有 BSON 或独立集成模块时，也运行对应测试。业务回归至少覆盖：默认和显式
scale、负整数幂、舍入边界、ErrNoConvergence、recover、nullable 编码、精确
YAML 和 Binary 错误处理。更新断言时以数学要求和数值契约为准。

回退可重新固定 v0.7.0，但也会恢复旧数值错误；应同时处理应用断言、配置和
已写入数据的兼容性。性能优化及其测量范围见[benchmark 报告](allocation-benchmarks.md)，
它们不改变正确性要求。
