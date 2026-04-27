# Guia do Usuário (Português do Brasil)

Este guia mostra padrões práticos de uso para `github.com/exc-works/decimal`.

## 1. Instalação

```bash
go get github.com/exc-works/decimal
```

## 2. Criar Decimais com Segurança

```go
price := decimal.MustFromString("99.9900")
discount := decimal.NewWithPrec(125, 2) // 1.25
```

Use `NewFromString` quando a entrada puder ser inválida:

```go
v, err := decimal.NewFromString(input)
if err != nil {
	return err
}
```

## 3. Precisão e Formatação

```go
d := decimal.MustFromString("7.5000")

fmt.Println(d.String())                  // 7.5
fmt.Println(d.StringWithTrailingZeros()) // 7.5000
fmt.Println(d.Precision())               // 4
```

- `String()` retorna uma representação canônica e remove zeros à direita.
- `StringWithTrailingZeros()` preserva a formatação de acordo com a escala.

## 4. Aritmética Básica

```go
subtotal := decimal.MustFromString("12.50")
fee := decimal.MustFromString("0.75")
total := subtotal.Add(fee) // 13.25
```

Para multiplicação/divisão, você deve escolher o comportamento de arredondamento:

```go
a := decimal.MustFromString("2.555")
b := decimal.MustFromString("1.00")

mul := a.Mul(b, decimal.RoundHalfEven)
quo := a.QuoWithPrec(decimal.MustFromString("3"), 2, decimal.RoundHalfEven)
```

Para multiplicação exata sem arredondamento:

```go
exact := decimal.MustFromString("1.20").MulExact(decimal.MustFromString("2.30"))
fmt.Println(exact.StringWithTrailingZeros()) // 2.7600
```

`Mul2` é mantido como um alias de compatibilidade obsoleto de `MulExact`.

## 5. Arredondamento e Reescalonamento

```go
v := decimal.MustFromString("-1.23")

fmt.Println(v.Rescale(0, decimal.RoundDown))    // -1  (toward zero)
fmt.Println(v.Rescale(0, decimal.RoundUp))      // -2  (away from zero)
fmt.Println(v.Rescale(0, decimal.RoundCeiling)) // -1  (toward +infinity)
fmt.Println(v.Floor())                           // -2
fmt.Println(v.Ceil())                            // -1
```

Use `QuoWithPrec` para definir explicitamente a precisão de saída em divisões.

## 6. Semântica de Valor

`Decimal` é imutável em operações aritméticas e de comparação:

```go
x := decimal.MustFromString("1.20")
y := x.Add(decimal.MustFromString("0.30"))

fmt.Println(x.String()) // 1.2
fmt.Println(y.String()) // 1.5
```

## 7. Comparação

```go
a := decimal.MustFromString("1.0")
b := decimal.MustFromString("1.00")

fmt.Println(a.Equal(b)) // true
fmt.Println(a.Cmp(b))   // 0
```

A comparação é numérica (insensível à escala).

## 8. Matemática Avançada

```go
x := decimal.MustFromString("100")
log10, _ := x.Log10() // 2
ln, _   := decimal.MustFromString("2.71828").Ln()  // ~= 1
exp, _  := decimal.MustFromString("1").Exp()        // e
```

Os três retornam erro para entradas inválidas (`Log10`/`Ln` exigem um receptor
positivo). Use as variantes `*WithPrec(prec)` para controlar a precisão de saída.

## 9. Formatação e Exibição

`Decimal` implementa `fmt.Formatter`, portanto os verbos padrão funcionam:

```go
d := decimal.MustFromString("1234.5678")
fmt.Sprintf("%s", d)   // 1234.5678
fmt.Sprintf("%.2f", d) // 1234.57 (RoundHalfEven)
fmt.Sprintf("%e", d)   // 1.234568e+03
fmt.Sprintf("%+10.1f", d) // "   +1234.6"
```

Exibição sensível à localidade:

```go
d := decimal.MustFromString("12345.678")
d.FormatWithSeparators(',', '.') // "12,345.678"
d.FormatWithSeparators('.', ',') // "12.345,678" (European)
d.FormatWithSeparators(' ', '.') // "12 345.678"
```

## 10. Serialização e Banco de Dados

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` grava valores decimais como strings JSON.

### XML

```go
type Item struct {
	Amount decimal.Decimal `xml:"amount"`
}
```

`MarshalXML` / `UnmarshalXML` (além das variantes para atributos) são fornecidos.
Valores não inicializados são codificados como elemento/atributo vazio.

### BSON (MongoDB)

`Decimal` implementa `bson.ValueMarshaler` / `bson.ValueUnmarshaler` para
`go.mongodb.org/mongo-driver/v2/bson`. Os valores são codificados como strings
BSON, e String/Double/Int32/Int64/Decimal128/Null são aceitos na decodificação.

### SQL

`Decimal` implementa ambos:

- `driver.Valuer`
- `sql.Scanner`

Portanto, ele funciona diretamente com drivers de banco de dados comuns.

### NullDecimal (colunas anuláveis)

Para colunas que podem ser `NULL`, use `NullDecimal`:

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

`NullDecimal` suporta SQL, JSON, YAML, Text, BSON e binding do gin. Entrada
`null` ou vazia define `Valid=false`.

## 11. Integração com validator

Registre as tags do Decimal com `go-playground/validator`:

```go
v := validator.New()
_ = decimal.RegisterGoPlaygroundValidator(v)

type Req struct {
	Price decimal.Decimal `validate:"decimal_required,decimal_positive,decimal_max_precision=2"`
	Rate  decimal.Decimal `validate:"decimal_between=0~1"`
}
```

Tags disponíveis:

- `decimal_required`, `decimal_eq`, `decimal_ne`, `decimal_gt`, `decimal_gte`,
  `decimal_lt`, `decimal_lte`, `decimal_between` (limites separados por til, p. ex. `1~100`)
- `decimal_positive`, `decimal_negative`, `decimal_nonzero` (sem parâmetro)
- `decimal_max_precision=N` (casas decimais ≤ N)

Traduções integradas: `en`, `zh`, `zh_Hant`, `ja`, `ko`, `fr`, `es`, `de`,
`pt`, `pt_BR`, `ru`, `ar`, `hi`.

## 12. Decodificação de Configuração (viper / mapstructure)

`decimal.DecodeHook()` retorna um hook compatível com mapstructure que
decodifica valores de configuração (`string`, `int`, `uint`, `float`,
`json.Number`, `[]byte`, `nil`) em `Decimal` e `NullDecimal`. Foi
projetado para ser composto com `mapstructure.TextUnmarshallerHookFunc()`,
que já trata o caminho de string via `UnmarshalText`. A ordem não é
estrita ── `decimal.DecodeHook()` lida com strings de forma autônoma ──
mas a ordem canônica mostrada abaixo corresponde ao exemplo do README:

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

Comportamento:

- `Decimal` + `nil` / string vazia / `[]byte` vazio → erro envolvendo `ErrUnmarshal` (`Decimal` não pode representar SQL NULL).
- `NullDecimal` + `nil` / string vazia / `[]byte` vazio → valor zero (`Valid: false`).
- Uma fonte `bool` é **rejeitada** para ambos os destinos, para evitar mapear silenciosamente `false`/`true` para `0`/`1`.
- Floats `NaN` e `±Inf` produzem um erro envolvendo `ErrUnmarshal`.

O hook reside no módulo principal e **não** introduz dependências de viper
ou mapstructure. Funciona também com qualquer outro decodificador baseado
em mapstructure (koanf, confita, cleanenv).

## 13. Tratamento de Erros

O pacote expõe erros sentinela para correspondência com `errors.Is`:

```go
_, err := decimal.NewFromString("not a number")
if errors.Is(err, decimal.ErrInvalidFormat) {
	// handle
}
```

Disponíveis: `ErrInvalidFormat`, `ErrInvalidPrecision`, `ErrOverflow`,
`ErrDivideByZero`, `ErrNegativeRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`,
`ErrUnmarshal`.

## 14. Concorrência

Os valores de `Decimal` são seguros para acesso concorrente de leitura desde
que nenhuma goroutine reatribua a variável. Os métodos com receptor por valor
(`Add`, `Cmp`, `String`, ...) nunca alteram o receptor. Os métodos com receptor
por ponteiro (`Scan`, `UnmarshalJSON`, ...) alteram o valor e exigem
sincronização externa se o mesmo `*Decimal` for compartilhado entre goroutines.

## 15. Armadilhas Comuns

1. `MustFromString` gera panic; não o use com entrada não confiável.
2. Precisão negativa gera panic.
3. `RoundUnnecessary` gera panic em operações inexatas.
4. `Log2()` gera panic para valores não positivos; `Log10`/`Ln` retornam erro.
5. `MarshalBinary()` normaliza zeros à direita.

## 16. Padrões Recomendados

1. Analise entradas externas com `NewFromString` e trate erros.
2. Use `QuoWithPrec` para qualquer saída de divisão visível ao usuário.
3. Use `StringWithTrailingZeros` somente quando uma escala fixa de exibição for necessária.
4. Mantenha o modo de arredondamento explícito nas regras de negócio.
5. Use `NullDecimal` para colunas SQL anuláveis em vez de soluções com ponteiros.
6. Faça correspondência de erros via `errors.Is(err, decimal.ErrXxx)` para um tratamento portátil.
