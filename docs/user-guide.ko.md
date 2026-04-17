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

## 8. 직렬화와 DB

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()`는 decimal 값을 JSON 문자열로 직렬화합니다.

### SQL

`Decimal`은 다음 둘 다를 구현합니다:

- `driver.Valuer`
- `sql.Scanner`

따라서 일반적인 데이터베이스 드라이버와 바로 동작합니다.

## 9. 흔한 함정

1. `MustFromString`은 panic을 발생시키므로, 신뢰할 수 없는 입력에 사용하지 마세요.
2. 음수 정밀도(`precision`)는 panic을 발생시킵니다.
3. `RoundUnnecessary`는 부정확한 연산에서 panic을 발생시킵니다.
4. `Log2()`는 0 이하 값에서 panic을 발생시킵니다.
5. `MarshalBinary()`는 후행 0을 정규화합니다.

## 10. 권장 패턴

1. 외부 입력은 `NewFromString`으로 파싱하고 오류를 처리하세요.
2. 사용자에게 표시되는 나눗셈 결과에는 `QuoWithPrec`를 사용하세요.
3. 고정 표시 스케일이 필요할 때만 `StringWithTrailingZeros`를 사용하세요.
4. 비즈니스 규칙에서 반올림 모드를 명시적으로 유지하세요.
