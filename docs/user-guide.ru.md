# Руководство пользователя (Русский)

В этом руководстве показаны практические шаблоны использования `github.com/exc-works/decimal`.

## 1. Установка

```bash
go get github.com/exc-works/decimal
```

## 2. Безопасное создание Decimal-значений

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

Используйте `NewFromString`, когда входные данные могут быть некорректными:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. Точность и форматирование

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` выдает каноническое представление и удаляет конечные нули.
- `StringWithTrailingZeros()` сохраняет форматирование, зависящее от масштаба.

## 4. Базовая арифметика

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

Для умножения/деления необходимо явно выбрать режим округления:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

Для точного умножения без округления:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` сохранен как устаревший алиас совместимости для `MulExact`.

## 5. Округление и изменение масштаба

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

Используйте `QuoWithPrec`, чтобы явно задавать точность результата при делении.

## 6. Семантика значений

`Decimal` неизменяем для операций арифметики и сравнения:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. Сравнение

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

Сравнение выполняется по числовому значению (без учета масштаба).

## 8. Расширенная математика

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

Все три метода возвращают ошибку при некорректном входе (`Log10` и `Ln`
требуют положительное значение получателя). Используйте варианты
`*WithPrec(prec)` для управления точностью результата.

## 9. Форматирование и отображение

`Decimal` реализует интерфейс `fmt.Formatter`, поэтому стандартные
глаголы форматирования работают напрямую:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

Отображение с учетом локали:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (европейский стиль)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. Сериализация и БД

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` записывает десятичные значения как строки JSON.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

Предоставляются `MarshalXML` / `UnmarshalXML` (а также варианты для
атрибутов). Неинициализированные значения кодируются как пустой
элемент или атрибут.

### BSON (MongoDB)

`Decimal` реализует `bson.ValueMarshaler` / `bson.ValueUnmarshaler` для
`go.mongodb.org/mongo-driver/v2/bson`. Значения кодируются как строки
BSON, а при декодировании принимаются типы String, Double, Int32, Int64,
Decimal128 и Null.

### SQL

`Decimal` реализует оба интерфейса:

- `driver.Valuer`
- `sql.Scanner`

Поэтому он напрямую работает с типичными драйверами баз данных.

### NullDecimal (столбцы, допускающие NULL)

Для столбцов, которые могут содержать `NULL`, используйте `NullDecimal`:

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

`NullDecimal` поддерживает SQL, JSON, YAML, Text, BSON, а также
привязку в gin. Значение `null` или пустой ввод устанавливает
`Valid=false`.

## 11. Интеграция с валидатором

Зарегистрируйте теги Decimal в `go-playground/validator`:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

Доступные теги:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (границы через тильду, напр. `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (без параметра)
- `decimal_max_precision=N` (количество знаков после запятой ≤ N)

Встроенные переводы: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. Декодирование конфигурации (viper / mapstructure)

`decimal.DecodeHook()` возвращает совместимый с mapstructure хук, который
декодирует значения конфигурации (`string`, `int`, `uint`, `float`,
`json.Number`, `[]byte`, `nil`) в `Decimal` и `NullDecimal`. Он
спроектирован так, чтобы компоноваться с
`mapstructure.TextUnmarshallerHookFunc()`, который уже обрабатывает путь
строки через `UnmarshalText`. Порядок не строгий ──
`decimal.DecodeHook()` обрабатывает строки и самостоятельно ── но
канонический порядок, показанный ниже, соответствует примеру из README:

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

Поведение:

- `Decimal` + `nil` / пустая строка / пустой `[]byte` → ошибка, оборачивающая `ErrUnmarshal` (`Decimal` не может представлять SQL NULL).
- `NullDecimal` + `nil` / пустая строка / пустой `[]byte` → нулевое значение (`Valid: false`).
- Источник `bool` **отклоняется** для обеих целей, чтобы избежать молчаливого отображения `false`/`true` в `0`/`1`.
- Числа с плавающей точкой `NaN` и `±Inf` приводят к ошибке, оборачивающей `ErrUnmarshal`.

Хук размещён в основном модуле и **не** тянет за собой зависимости от
viper или mapstructure. Он также работает с любым другим декодером на
основе mapstructure (koanf, confita, cleanenv).

## 13. Обработка ошибок

Пакет предоставляет сигнальные ошибки для сопоставления через
`errors.Is`:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// обработка
}
```

Доступны: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`,
`ErrUnmarshal`.

## 14. Конкурентность

Значения `Decimal` безопасны для одновременного чтения, пока ни одна
горутина не переприсваивает переменную. Методы с получателем по
значению (`Add`, `Cmp`, `String`, ...) никогда не изменяют получателя.
Методы с получателем-указателем (`Scan`, `UnmarshalJSON`, ...)
изменяют его и требуют внешней синхронизации, если один и тот же
`*Decimal` используется несколькими горутинами.

## 15. Распространённые ошибки

1. `MustFromString` вызывает panic; не используйте его для недоверенного ввода.
2. Отрицательная точность вызывает panic.
3. `RoundUnnecessary` вызывает panic при неточных операциях.
4. `Log2()` вызывает panic для неположительных значений; `Log10` и `Ln`
   возвращают ошибку.
5. `MarshalBinary()` нормализует конечные нули.

## 16. Рекомендуемые шаблоны

1. Разбирайте внешний ввод через `NewFromString` и обрабатывайте ошибки.
2. Используйте `QuoWithPrec` для любого пользовательского вывода
   результатов деления.
3. Используйте `StringWithTrailingZeros` только когда требуется
   фиксированный масштаб отображения.
4. Явно фиксируйте режим округления в бизнес-правилах.
5. Используйте `NullDecimal` для SQL-столбцов, допускающих NULL, вместо
   обходных решений с указателями.
6. Сопоставляйте ошибки через `errors.Is(err, decimal.ErrXxx)` для
   переносимой обработки.
