package decimal

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type yamlDecimalScalar struct {
	value Decimal
	valid bool
	kind  yamlScalarKind
}

type yamlScalarKind uint8

const (
	yamlStringScalar yamlScalarKind = iota
	yamlIntegerScalar
	yamlFloatScalar
)

func decodeYAMLScalar(unmarshal func(any) error) (yamlDecimalScalar, error) {
	var scalar yamlDecimalScalar
	// The callback API exposes text without its YAML tag or quoting style.
	// Resolve only the scalar kind first, so 012 and "012" keep their distinct
	// YAML-number and decimal-string meanings. Discard every resolved numeric
	// value: the original token below is the sole source of numeric digits.
	var resolved any
	if err := unmarshal(&resolved); err != nil {
		return scalar, fmt.Errorf("cannot resolve YAML scalar kind: %w: %w", err, ErrUnmarshal)
	}
	switch resolved.(type) {
	case nil:
		return scalar, nil
	case string, []byte:
		scalar.kind = yamlStringScalar
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		scalar.kind = yamlIntegerScalar
	case float32, float64:
		scalar.kind = yamlFloatScalar
	default:
		return scalar, fmt.Errorf("expected a numeric YAML scalar, got %T: %w", resolved, ErrUnmarshal)
	}
	if err := unmarshal(&scalar); err != nil {
		return scalar, fmt.Errorf("cannot decode YAML decimal: %w: %w", err, ErrUnmarshal)
	}
	return scalar, nil
}

// yaml.v2/v3 deliver the original scalar value to TextUnmarshaler. Interpret
// numeric tokens according to their resolved YAML kind, while string scalars
// retain NewFromString semantics. Decoder-owned floats never supply digits.
func (s *yamlDecimalScalar) UnmarshalText(text []byte) error {
	literal := string(text)
	if s.kind == yamlIntegerScalar || s.kind == yamlFloatScalar {
		// Base zero recognizes YAML's octal, hexadecimal and binary prefixes.
		// Underscore placement has already been accepted by the YAML resolver.
		integer, ok := new(big.Int).SetString(strings.ReplaceAll(literal, "_", ""), 0)
		if ok && (s.kind == yamlIntegerScalar || integer.IsInt64() || integer.IsUint64()) {
			// yaml.v2/v3 also resolve integer syntax before coercing an
			// explicit !!float tag: !!float 012 still denotes ten. Their
			// integer resolver falls back to decimal floating syntax beyond
			// 64 bits, so only use this path for those same integer ranges.
			s.value, s.valid = NewFromBigInt(integer), true
			return nil
		}
		if s.kind == yamlIntegerScalar {
			return fmt.Errorf("invalid YAML integer %q: %w", literal, ErrInvalidFormat)
		}
	}
	if s.kind == yamlFloatScalar {
		literal = normalizeYAMLFloat(literal)
	}
	d, err := NewFromString(literal)
	if err != nil {
		return err
	}
	s.value, s.valid = d, true
	return nil
}

func normalizeYAMLFloat(literal string) string {
	literal = strings.ReplaceAll(literal, "_", "")
	mantissa, exponent := literal, ""
	if index := strings.IndexAny(literal, "eE"); index >= 0 {
		mantissa, exponent = literal[:index], literal[index:]
	}
	sign := ""
	if len(mantissa) > 0 && (mantissa[0] == '+' || mantissa[0] == '-') {
		sign, mantissa = mantissa[:1], mantissa[1:]
	}
	if strings.HasPrefix(mantissa, ".") {
		mantissa = "0" + mantissa
	}
	if strings.HasSuffix(mantissa, ".") {
		mantissa += "0"
	}
	return sign + mantissa + exponent
}

// goccy/go-yaml converts even string destinations through float64. Its
// byte unmarshaler instead exposes the source, which we parse as an AST and
// read from the scalar token, never from FloatNode.Value.
func (s *yamlDecimalScalar) UnmarshalYAML(data []byte) error {
	file, err := parser.ParseBytes(data, 0)
	if err != nil {
		return err
	}
	if len(file.Docs) != 1 {
		return fmt.Errorf("expected one YAML scalar: %w", ErrInvalidFormat)
	}
	node := file.Docs[0].Body
	for {
		switch n := node.(type) {
		case *ast.TagNode:
			node = n.Value
		case *ast.AnchorNode:
			node = n.Value
		default:
			goto scalar
		}
	}
scalar:
	switch n := node.(type) {
	case nil, *ast.NullNode:
		return nil
	case *ast.StringNode:
		return s.UnmarshalText([]byte(n.Value))
	case *ast.IntegerNode:
		return s.UnmarshalText([]byte(n.Token.Value))
	case *ast.FloatNode:
		return s.UnmarshalText([]byte(n.Token.Value))
	default:
		return fmt.Errorf("expected a finite numeric YAML scalar: %w", ErrInvalidFormat)
	}
}
