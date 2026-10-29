package mapstructure_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/exc-works/decimal"
	yaml "go.yaml.in/yaml/v3"
	yaml2 "gopkg.in/yaml.v2"
	yaml3 "gopkg.in/yaml.v3"
)

func TestNativeYAMLScalarSyntax(t *testing.T) {
	data, err := os.ReadFile("../../testdata/yaml_scalars.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input, Want string
		Error       bool
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, decoder := range []struct {
		name      string
		unmarshal func([]byte, any) error
	}{
		{"yaml.v2", yaml2.Unmarshal}, {"yaml.v3", yaml3.Unmarshal}, {"go.yaml.in/yaml/v3", yaml.Unmarshal},
	} {
		t.Run(decoder.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.Input, func(t *testing.T) {
					var value decimal.Decimal
					var nullable decimal.NullDecimal
					err := decoder.unmarshal([]byte(tc.Input), &value)
					nullErr := decoder.unmarshal([]byte(tc.Input), &nullable)
					if tc.Error {
						if err == nil || nullErr == nil {
							t.Fatalf("expected errors, got Decimal=%s (%v), NullDecimal=%s (%v)", value, err, nullable, nullErr)
						}
						return
					}
					if err != nil || nullErr != nil {
						t.Fatalf("Decimal error %v, NullDecimal error %v", err, nullErr)
					}
					want := decimal.MustFromString(tc.Want)
					if !value.Equal(want) || !nullable.Valid || !nullable.Decimal.Equal(want) {
						t.Fatalf("got %s and %s, want %s", value, nullable, want)
					}
				})
			}
		})
	}
}

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
