package mapstructure_test

import (
	"testing"

	"github.com/exc-works/decimal"
	yaml "go.yaml.in/yaml/v3"
)

func TestNativeYAMLPrecision(t *testing.T) {
	for _, input := range []string{"1.23456789012345678901234567890", "9007199254740993.0", "1e1000", "123456789012345678901234567890"} {
		var value struct {
			Amount   decimal.Decimal     `yaml:"amount"`
			Optional decimal.NullDecimal `yaml:"optional"`
		}
		if err := yaml.Unmarshal([]byte("amount: "+input+"\noptional: "+input), &value); err != nil {
			t.Fatal(err)
		}
		want := decimal.MustFromString(input)
		if !value.Amount.Equal(want) || !value.Optional.Valid || !value.Optional.Decimal.Equal(want) {
			t.Fatalf("input %s: got %s and %s", input, value.Amount, value.Optional)
		}
	}
}
