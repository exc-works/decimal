# دليل المستخدم (العربية)

يوضح هذا الدليل أنماط استخدام عملية لـ `github.com/exc-works/decimal`.

## 1. التثبيت

```bash
go get github.com/exc-works/decimal
```

## 2. إنشاء القيم العشرية بأمان

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

استخدم `NewFromString` عندما قد يكون الإدخال غير صالح:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. الدقة والتنسيق

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` هو تمثيل قياسي ويزيل الأصفار اللاحقة.
- `StringWithTrailingZeros()` يحافظ على تنسيق يعتمد على المقياس.

## 4. العمليات الحسابية الأساسية

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

في الضرب/القسمة يجب أن تختار سلوك التقريب:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

للحصول على ضرب دقيق من دون تقريب:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` مُحتفَظ به كاسم مستعار توافقي مُهمَل لـ `MulExact`.

## 5. التقريب وإعادة ضبط المقياس

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

استخدم `QuoWithPrec` لتحديد دقة ناتج القسمة بشكل صريح.

## 6. دلالات القيمة

`Decimal` غير قابل للتغيير في عمليات الحساب/المقارنة:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. المقارنة

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

المقارنة رقمية (غير حساسة للمقياس).

## 8. التسلسل وقاعدة البيانات

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` يخرّج القيم العشرية كسلاسل نصية في JSON.

### SQL

`Decimal` يطبّق كلتا الواجهتين:

- `driver.Valuer`
- `sql.Scanner`

لذلك يعمل مباشرةً مع مشغلات قواعد البيانات الشائعة.

## 9. الأخطاء الشائعة

1. `MustFromString` يؤدي إلى panic؛ لا تستخدمه مع مدخلات غير موثوقة.
2. الدقة السالبة تؤدي إلى panic.
3. `RoundUnnecessary` يؤدي إلى panic عند العمليات غير الدقيقة.
4. `Log2()` يؤدي إلى panic للقيم غير الموجبة.
5. `MarshalBinary()` يطبّع الأصفار اللاحقة.

## 10. الأنماط الموصى بها

1. حلّل المدخلات الخارجية باستخدام `NewFromString` وتعامل مع الأخطاء.
2. استخدم `QuoWithPrec` لأي مخرجات قسمة مرئية للمستخدم.
3. استخدم `StringWithTrailingZeros` فقط عندما تكون هناك حاجة إلى مقياس عرض ثابت.
4. اجعل وضع التقريب صريحًا في قواعد العمل.
