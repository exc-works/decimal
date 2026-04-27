# उपयोगकर्ता मार्गदर्शिका (हिंदी)

यह मार्गदर्शिका `github.com/exc-works/decimal` के व्यावहारिक उपयोग पैटर्न दिखाती है।

## 1. स्थापना

```bash
go get github.com/exc-works/decimal
```

## 2. दशमलव सुरक्षित रूप से बनाएं

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

जब इनपुट अमान्य हो सकता है, तब `NewFromString` का उपयोग करें:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. परिशुद्धता और फॉर्मेटिंग

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` मानक निरूपण देता है और अनुगामी शून्यों को हटा देता है।
- `StringWithTrailingZeros()` स्केल (scale)-आधारित फॉर्मेटिंग को बनाए रखता है।

## 4. मूल अंकगणित

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

गुणा/भाग के लिए आपको राउंडिंग मोड चुनना आवश्यक है:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

राउंडिंग के बिना सटीक गुणा के लिए:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` को `MulExact` के अप्रचलित संगतता उपनाम के रूप में रखा गया है।

## 5. राउंडिंग और रीस्केलिंग

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

भाग के आउटपुट में स्पष्ट परिशुद्धता के लिए `QuoWithPrec` का उपयोग करें।

## 6. मान सिमैंटिक्स

`Decimal` अंकगणित/तुलना ऑपरेशनों के लिए अपरिवर्तनीय है:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. तुलना

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

तुलना संख्यात्मक है (स्केल (scale) से अप्रभावित)।

## 8. उन्नत गणित

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

तीनों अमान्य इनपुट के लिए त्रुटि लौटाते हैं (`Log10`/`Ln` के लिए रिसीवर धनात्मक होना आवश्यक है)। आउटपुट परिशुद्धता को नियंत्रित करने के लिए `*WithPrec(prec)` संस्करणों का उपयोग करें।

## 9. फ़ॉर्मेटिंग और प्रदर्शन

`Decimal`, `fmt.Formatter` को लागू करता है, इसलिए मानक verbs काम करते हैं:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

लोकेल-अनुरूप प्रदर्शन:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. सीरियलाइज़ेशन और DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` दशमलव मानों को JSON स्ट्रिंग के रूप में लिखता है।

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML` (और एट्रिब्यूट संस्करण) उपलब्ध हैं। अप्रारंभित मान खाली एलिमेंट/एट्रिब्यूट के रूप में एनकोड होते हैं।

### BSON (MongoDB)

`Decimal`, `go.mongodb.org/mongo-driver/v2/bson` के लिए `bson.ValueMarshaler` / `bson.ValueUnmarshaler` को लागू करता है। मान BSON स्ट्रिंग के रूप में एनकोड होते हैं, और डिकोड पर String/Double/Int32/Int64/Decimal128/Null स्वीकार किए जाते हैं।

### SQL

`Decimal` दोनों को लागू करता है:

- `driver.Valuer`
- `sql.Scanner`

इसलिए यह सामान्य डेटाबेस ड्राइवरों के साथ सीधे काम करता है।

### NullDecimal (nullable कॉलम)

उन कॉलम के लिए जो `NULL` हो सकते हैं, `NullDecimal` का उपयोग करें:

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

`NullDecimal`, SQL, JSON, YAML, Text, BSON, और gin binding को सपोर्ट करता है। `null` या खाली इनपुट `Valid=false` सेट करता है।

## 11. वैलिडेटर एकीकरण

`go-playground/validator` के साथ Decimal tags रजिस्टर करें:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

उपलब्ध tags:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (`~` से अलग की गई सीमाएँ, जैसे `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (कोई पैरामीटर नहीं)
- `decimal_max_precision=N` (दशमलव स्थान ≤ N)

अंतर्निहित अनुवाद: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`, `pt`, `pt_BR`, `ru`, `ar`, `hi`।

## 12. कॉन्फ़िग डिकोडिंग (viper / mapstructure)

`decimal.DecodeHook()` एक mapstructure-संगत हुक लौटाता है जो कॉन्फ़िग
मानों (`string`, `int`, `uint`, `float`, `json.Number`, `[]byte`, `nil`)
को `Decimal` और `NullDecimal` में डिकोड करता है। इसे
`mapstructure.TextUnmarshallerHookFunc()` के साथ कंपोज़ करने के लिए
डिज़ाइन किया गया है, जो पहले से ही `UnmarshalText` के माध्यम से स्ट्रिंग
पथ को संभालता है। क्रम सख्त नहीं है ── `decimal.DecodeHook()` स्ट्रिंग्स
को स्वतंत्र रूप से भी संभाल लेता है ── लेकिन नीचे दिखाया गया मानक क्रम
README उदाहरण से मेल खाता है:

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

व्यवहार:

- `Decimal` + `nil` / खाली स्ट्रिंग / खाली `[]byte` → `ErrUnmarshal` को रैप करने वाली त्रुटि (`Decimal` SQL NULL का प्रतिनिधित्व नहीं कर सकता)।
- `NullDecimal` + `nil` / खाली स्ट्रिंग / खाली `[]byte` → शून्य मान (`Valid: false`)।
- दोनों लक्ष्य प्रकारों के लिए `bool` स्रोत **अस्वीकार** किया जाता है, ताकि `false`/`true` को मौन रूप से `0`/`1` पर मैप होने से बचाया जा सके।
- `NaN` और `±Inf` फ़्लोट `ErrUnmarshal` को रैप करने वाली त्रुटि देते हैं।

यह हुक मुख्य मॉड्यूल में मौजूद है और viper या mapstructure की कोई भी
निर्भरता नहीं खींचता। यह किसी भी अन्य mapstructure-आधारित डिकोडर
(koanf, confita, cleanenv) के साथ भी काम करता है।

## 13. त्रुटि प्रबंधन

यह पैकेज `errors.Is` मिलान के लिए सेंटिनल त्रुटियाँ प्रदान करता है:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

उपलब्ध: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`, `ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`, `ErrUnmarshal`।

## 14. समवर्तीता

`Decimal` मान समवर्ती पढ़ने के लिए सुरक्षित हैं, बशर्ते कोई goroutine वेरिएबल को पुनः असाइन न करे। वैल्यू-रिसीवर विधियाँ (`Add`, `Cmp`, `String`, ...) कभी रिसीवर को परिवर्तित नहीं करतीं। पॉइंटर-रिसीवर विधियाँ (`Scan`, `UnmarshalJSON`, ...) परिवर्तन करती हैं और यदि एक ही `*Decimal` कई goroutines के बीच साझा किया गया है तो बाहरी सिंक्रोनाइज़ेशन आवश्यक है।

## 15. सामान्य त्रुटियाँ

1. `MustFromString` पैनिक करता है; इसे अविश्वसनीय इनपुट पर उपयोग न करें।
2. ऋणात्मक परिशुद्धता पैनिक कराती है।
3. `RoundUnnecessary` असटीक ऑपरेशनों पर पैनिक करता है।
4. `Log2()` गैर-धनात्मक मानों के लिए पैनिक करता है; `Log10`/`Ln` त्रुटि लौटाते हैं।
5. `MarshalBinary()` अनुगामी शून्यों को सामान्यीकृत करता है।

## 16. अनुशंसित पैटर्न

1. बाहरी इनपुट को `NewFromString` से पार्स करें और त्रुटियों को संभालें।
2. उपयोगकर्ता को दिखने वाले किसी भी भागफल आउटपुट के लिए `QuoWithPrec` का उपयोग करें।
3. `StringWithTrailingZeros` का उपयोग केवल तब करें जब निश्चित प्रदर्शन स्केल (display scale) आवश्यक हो।
4. व्यावसायिक नियमों में राउंडिंग मोड को स्पष्ट रखें।
5. nullable SQL कॉलम के लिए पॉइंटर वर्कअराउंड के बजाय `NullDecimal` का उपयोग करें।
6. पोर्टेबल हैंडलिंग के लिए `errors.Is(err, decimal.ErrXxx)` के माध्यम से त्रुटियों का मिलान करें।
