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

## 8. الرياضيات المتقدمة

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

تُعيد الدوال الثلاث خطأً عند المدخلات غير الصالحة (`Log10`/`Ln` تتطلبان
مُستقبِلًا موجبًا). استخدم متغيرات `*WithPrec(prec)` للتحكم في دقة الناتج.

## 9. التنسيق والعرض

يُطبّق `Decimal` الواجهة `fmt.Formatter`، لذا تعمل رموز التنسيق القياسية:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

عرض مُراعٍ للّغة المحلية:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. التسلسل وقاعدة البيانات

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` يخرّج القيم العشرية كسلاسل نصية في JSON.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

تتوفر الدوال `MarshalXML` / `UnmarshalXML` (بالإضافة إلى متغيرات السمات).
القيم غير المُهيَّأة تُرمَّز كعنصر/سمة فارغة.

### BSON (MongoDB)

يُطبّق `Decimal` الواجهتين `bson.ValueMarshaler` / `bson.ValueUnmarshaler`
لحزمة `go.mongodb.org/mongo-driver/v2/bson`. تُرمَّز القيم كسلاسل BSON نصية،
ويُقبَل عند فكّ الترميز كلٌّ من String و Double و Int32 و Int64 و Decimal128
و Null.

### SQL

`Decimal` يطبّق كلتا الواجهتين:

- `driver.Valuer`
- `sql.Scanner`

لذلك يعمل مباشرةً مع مشغلات قواعد البيانات الشائعة.

### NullDecimal (الأعمدة القابلة للإبطال)

للأعمدة التي قد تحمل القيمة `NULL`، استخدم `NullDecimal`:

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

يدعم `NullDecimal` كلًّا من SQL و JSON و YAML و Text و BSON وربط gin.
المدخل `null` أو الفارغ يضبط `Valid=false`.

## 11. التكامل مع المدقق

سجّل وسوم Decimal مع `go-playground/validator`:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

الوسوم المتاحة:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (الحدود مفصولة بعلامة `~`، مثل `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (بدون وسيط)
- `decimal_max_precision=N` (المنازل العشرية ≤ N)

الترجمات المدمجة: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. فك ترميز الإعدادات (viper / mapstructure)

تُرجِع `decimal.DecodeHook()` خطّافًا متوافقًا مع mapstructure يقوم بفك
ترميز قيم الإعدادات (`string`, `int`, `uint`, `float`, `json.Number`,
`[]byte`, `nil`) إلى `Decimal` و `NullDecimal`. صُمّم ليُستخدم مع
`mapstructure.TextUnmarshallerHookFunc()` الذي يعالج بالفعل مسار السلسلة
عبر `UnmarshalText`. الترتيب ليس صارمًا ── إذ يعالج `decimal.DecodeHook()`
السلاسل بمفرده ── لكن الترتيب القياسي الموضّح أدناه يطابق مثال README:

```go
import (
	"github.com/exc-works/decimal"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Price    decimal.Decimal     `mapstructure:"price"`
	Discount decimal.NullDecimal `mapstructure:"discount"`
}

var cfg Config
err := viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
	mapstructure.TextUnmarshallerHookFunc(),
	decimal.DecodeHook(),
)))
```

السلوك:

- `Decimal` + `nil` / سلسلة فارغة / `[]byte` فارغ → خطأ يلفّ `ErrUnmarshal` (لا يمكن لـ `Decimal` تمثيل SQL NULL).
- `NullDecimal` + `nil` / سلسلة فارغة / `[]byte` فارغ → قيمة صفرية (`Valid: false`).
- يُعدّ مصدر `bool` **مرفوضًا** لكلا الهدفين، تجنّبًا لربط `false`/`true` ضمنيًا بـ `0`/`1`.
- تُنتج أعداد الفاصلة العائمة `NaN` و `±Inf` خطأً يلفّ `ErrUnmarshal`.

يقع الخطّاف في الوحدة الرئيسية ولا يجلب **أي** اعتماديات على viper أو
mapstructure. يعمل أيضًا مع أيّ مفكّك ترميز آخر مبني على mapstructure
(مثل koanf و confita و cleanenv).

## 13. معالجة الأخطاء

تكشف الحزمة عن أخطاء حارسة للمطابقة عبر `errors.Is`:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

المتاح: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`,
`ErrUnmarshal`.

## 14. التزامن

قيم `Decimal` آمنة للقراءة المتزامنة ما دام لا توجد غوروتين تُعيد تعيين
المتغيّر. الدوال ذات المُستقبِل بالقيمة (`Add`, `Cmp`, `String`, ...) لا تُعدّل
المُستقبِل أبدًا. أما الدوال ذات المُستقبِل بالمؤشر (`Scan`, `UnmarshalJSON`,
...) فتُعدّله، وتتطلب مزامنة خارجية إذا تمّت مشاركة نفس `*Decimal` بين عدة
غوروتينات.

## 15. المزالق الشائعة

1. `MustFromString` يؤدي إلى panic؛ لا تستخدمه مع مدخلات غير موثوقة.
2. الدقة السالبة تؤدي إلى panic.
3. `RoundUnnecessary` يؤدي إلى panic عند العمليات غير الدقيقة.
4. `Log2()` يؤدي إلى panic للقيم غير الموجبة؛ بينما `Log10`/`Ln` تُعيدان خطأً.
5. `MarshalBinary()` يطبّع الأصفار اللاحقة.

## 16. الأنماط الموصى بها

1. حلّل المدخلات الخارجية باستخدام `NewFromString` وتعامل مع الأخطاء.
2. استخدم `QuoWithPrec` لأي مخرجات قسمة مرئية للمستخدم.
3. استخدم `StringWithTrailingZeros` فقط عندما تكون هناك حاجة إلى مقياس عرض ثابت.
4. اجعل وضع التقريب صريحًا في قواعد العمل.
5. استخدم `NullDecimal` للأعمدة القابلة للإبطال في SQL بدلًا من الحلول البديلة بالمؤشرات.
6. طابق الأخطاء عبر `errors.Is(err, decimal.ErrXxx)` لمعالجة قابلة للنقل.
