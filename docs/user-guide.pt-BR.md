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

## 8. Serialização e Banco de Dados

### JSON

```go
type Item struct {
	Amount decimal.Decimal `json:"amount"`
}
```

`MarshalJSON()` grava valores decimais como strings JSON.

### SQL

`Decimal` implementa ambos:

- `driver.Valuer`
- `sql.Scanner`

Portanto, ele funciona diretamente com drivers de banco de dados comuns.

## 9. Armadilhas Comuns

1. `MustFromString` gera panic; não o use com entrada não confiável.
2. Precisão negativa gera panic.
3. `RoundUnnecessary` gera panic em operações inexatas.
4. `Log2()` gera panic para valores não positivos.
5. `MarshalBinary()` normaliza zeros à direita.

## 10. Padrões Recomendados

1. Analise entradas externas com `NewFromString` e trate erros.
2. Use `QuoWithPrec` para qualquer saída de divisão visível ao usuário.
3. Use `StringWithTrailingZeros` somente quando uma escala fixa de exibição for necessária.
4. Mantenha o modo de arredondamento explícito nas regras de negócio.
