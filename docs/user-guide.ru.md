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

## 8. Сериализация и БД

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` записывает десятичные значения как строки JSON.

### SQL

`Decimal` реализует оба интерфейса:

- `driver.Valuer`
- `sql.Scanner`

Поэтому он напрямую работает с типичными драйверами баз данных.

## 9. Частые ошибки

1. `MustFromString` вызывает panic; не используйте его для недоверенного ввода.
2. Отрицательная точность вызывает panic.
3. `RoundUnnecessary` вызывает panic при неточных операциях.
4. `Log2()` вызывает panic для неположительных значений.
5. `MarshalBinary()` нормализует конечные нули.

## 10. Рекомендуемые практики

1. Разбирайте внешний ввод через `NewFromString` и обрабатывайте ошибки.
2. Используйте `QuoWithPrec` для любого пользовательского вывода результатов деления.
3. Используйте `StringWithTrailingZeros` только когда требуется фиксированный масштаб отображения.
4. Явно фиксируйте режим округления в бизнес-правилах.
