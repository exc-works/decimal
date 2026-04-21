# 사용자 가이드 (한국어)

이 가이드는 `github.com/exc-works/decimal`의 실용적인 사용 패턴을 보여줍니다.

## 1. 설치

```bash
go get github.com/exc-works/decimal
```

## 2. Decimal 안전하게 생성하기

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

입력이 유효하지 않을 수 있을 때는 `NewFromString`을 사용하세요:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. 정밀도와 포맷팅

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()`은 표준 표현을 반환하며 후행 0을 제거합니다.
- `StringWithTrailingZeros()`는 스케일 기반 포맷을 유지합니다.

## 4. 기본 산술

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

곱셈/나눗셈에서는 반올림 동작을 선택해야 합니다:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

반올림 없이 정확한 곱셈이 필요하면:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2`는 `MulExact`의 하위 호환을 위한 사용 중단 예정(deprecated) 별칭(alias)으로 유지됩니다.

## 5. 반올림과 리스케일링

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

나눗셈에서 출력 정밀도를 명시하려면 `QuoWithPrec`를 사용하세요.

## 6. 값 의미론

`Decimal`은 산술/비교 연산에서 불변입니다:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. 비교

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

비교는 수치 기준(스케일 비민감)으로 수행됩니다.

## 8. 고급 수학 함수

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

세 함수 모두 유효하지 않은 입력에 대해 오류를 반환합니다(`Log10`/`Ln`은 수신자가 양수여야 합니다). 출력 정밀도를 제어하려면 `*WithPrec(prec)` 변형을 사용하세요.

## 9. 포맷팅과 표시

`Decimal`은 `fmt.Formatter`를 구현하므로 표준 verb를 사용할 수 있습니다:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

로케일을 고려한 표시:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (유럽식)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. 직렬화와 DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()`는 decimal 값을 JSON 문자열로 직렬화합니다.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML`(및 속성용 변형)이 제공됩니다. 초기화되지 않은 값은 빈 요소/속성으로 인코딩됩니다.

### BSON (MongoDB)

`Decimal`은 `go.mongodb.org/mongo-driver/v2/bson`용으로 `bson.ValueMarshaler` / `bson.ValueUnmarshaler`를 구현합니다. 값은 BSON 문자열로 인코딩되며, 디코드 시에는 String/Double/Int32/Int64/Decimal128/Null이 허용됩니다.

### SQL

`Decimal`은 다음 둘 다를 구현합니다:

- `driver.Valuer`
- `sql.Scanner`

따라서 일반적인 데이터베이스 드라이버와 바로 동작합니다.

### NullDecimal (nullable 컬럼)

`NULL`이 될 수 있는 컬럼에는 `NullDecimal`을 사용하세요:

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

`NullDecimal`은 SQL, JSON, YAML, Text, BSON 및 gin 바인딩을 지원합니다. `null` 또는 빈 입력은 `Valid=false`로 설정됩니다.

## 11. Validator 통합

`go-playground/validator`에 Decimal 태그를 등록하세요:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

사용 가능한 태그:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (`~`로 구분된 경계값, 예: `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (파라미터 없음)
- `decimal_max_precision=N` (소수 자릿수 ≤ N)

기본 제공 번역: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. 오류 처리

이 패키지는 `errors.Is` 매칭을 위한 센티넬 오류를 공개합니다:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// 처리
}
```

사용 가능한 오류: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`,
`ErrUnmarshal`.

## 13. 동시성

`Decimal` 값은 어떤 고루틴도 변수를 재할당하지 않는 한 동시 읽기 접근에 안전합니다. 값 수신자 메서드(`Add`, `Cmp`, `String`, ...)는 수신자를 절대 변경하지 않습니다. 포인터 수신자 메서드(`Scan`, `UnmarshalJSON`, ...)는 수신자를 변경하므로 동일한 `*Decimal`이 여러 고루틴에서 공유될 경우 외부 동기화가 필요합니다.

## 14. 흔한 함정

1. `MustFromString`은 panic을 발생시키므로, 신뢰할 수 없는 입력에 사용하지 마세요.
2. 음수 정밀도(`precision`)는 panic을 발생시킵니다.
3. `RoundUnnecessary`는 부정확한 연산에서 panic을 발생시킵니다.
4. `Log2()`는 0 이하 값에서 panic을 발생시키며, `Log10`/`Ln`은 오류를 반환합니다.
5. `MarshalBinary()`는 후행 0을 정규화합니다.

## 15. 권장 패턴

1. 외부 입력은 `NewFromString`으로 파싱하고 오류를 처리하세요.
2. 사용자에게 표시되는 나눗셈 결과에는 `QuoWithPrec`를 사용하세요.
3. 고정 표시 스케일이 필요할 때만 `StringWithTrailingZeros`를 사용하세요.
4. 비즈니스 규칙에서 반올림 모드를 명시적으로 유지하세요.
5. nullable한 SQL 컬럼에는 포인터 우회 대신 `NullDecimal`을 사용하세요.
6. 이식 가능한 오류 처리를 위해 `errors.Is(err, decimal.ErrXxx)`로 오류를 매칭하세요.
