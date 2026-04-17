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

## 8. सीरियलाइज़ेशन और DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` दशमलव मानों को JSON स्ट्रिंग के रूप में लिखता है।

### SQL

`Decimal` दोनों को लागू करता है:

- `driver.Valuer`
- `sql.Scanner`

इसलिए यह सामान्य डेटाबेस ड्राइवरों के साथ सीधे काम करता है।

## 9. आम त्रुटियाँ

1. `MustFromString` पैनिक करता है; इसे अविश्वसनीय इनपुट पर उपयोग न करें।
2. ऋणात्मक परिशुद्धता पैनिक कराती है।
3. `RoundUnnecessary` असटीक ऑपरेशनों पर पैनिक करता है।
4. `Log2()` गैर-धनात्मक मानों के लिए पैनिक करता है।
5. `MarshalBinary()` अनुगामी शून्यों को सामान्यीकृत करता है।

## 10. अनुशंसित पैटर्न

1. बाहरी इनपुट को `NewFromString` से पार्स करें और त्रुटियों को संभालें।
2. उपयोगकर्ता को दिखने वाले किसी भी भागफल आउटपुट के लिए `QuoWithPrec` का उपयोग करें।
3. `StringWithTrailingZeros` का उपयोग केवल तब करें जब निश्चित प्रदर्शन स्केल (display scale) आवश्यक हो।
4. व्यावसायिक नियमों में राउंडिंग मोड को स्पष्ट रखें।
