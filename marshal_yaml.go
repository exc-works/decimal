package decimal

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type yamlDecimalScalar struct {
	value Decimal
	valid bool
}

func decodeYAMLScalar(unmarshal func(any) error) (yamlDecimalScalar, error) {
	var scalar yamlDecimalScalar
	if err := unmarshal(&scalar); err != nil {
		return scalar, fmt.Errorf("cannot decode YAML decimal: %w: %w", err, ErrUnmarshal)
	}
	return scalar, nil
}

// yaml.v2/v3 deliver the original scalar value to TextUnmarshaler. Null
// scalars bypass it and leave valid=false. No decoder-owned float is used.
func (s *yamlDecimalScalar) UnmarshalText(text []byte) error {
	d, err := NewFromString(string(text))
	if err != nil {
		return err
	}
	s.value, s.valid = d, true
	return nil
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
		// These AST values are exact integers and preserve YAML base prefixes.
		switch v := n.Value.(type) {
		case int64:
			s.value, s.valid = New(v), true
			return nil
		case uint64:
			s.value, s.valid = NewFromUint64(v, 0), true
			return nil
		}
		return s.UnmarshalText([]byte(strings.ReplaceAll(n.Token.Value, "_", "")))
	case *ast.FloatNode:
		text := strings.ReplaceAll(n.Token.Value, "_", "")
		// YAML also allows .5 and 1.; normalize these without arithmetic.
		parts := strings.FieldsFunc(text, func(r rune) bool { return r == 'e' || r == 'E' })
		mantissa := parts[0]
		sign := ""
		if mantissa[0] == '+' || mantissa[0] == '-' {
			sign, mantissa = mantissa[:1], mantissa[1:]
		}
		if strings.HasPrefix(mantissa, ".") {
			mantissa = "0" + mantissa
		}
		if strings.HasSuffix(mantissa, ".") {
			mantissa += "0"
		}
		text = sign + mantissa
		if len(parts) == 2 {
			text += "e" + parts[1]
		}
		return s.UnmarshalText([]byte(text))
	default:
		return fmt.Errorf("expected a finite numeric YAML scalar: %w", ErrInvalidFormat)
	}
}
