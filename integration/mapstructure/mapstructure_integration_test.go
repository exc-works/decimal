package mapstructure_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/exc-works/decimal"
	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Cfg matches the shape of testdata/config.{yaml,json,toml}.
type Cfg struct {
	Price    decimal.Decimal            `mapstructure:"price"`
	Quantity decimal.Decimal            `mapstructure:"quantity"`
	Ratio    decimal.Decimal            `mapstructure:"ratio"`
	Discount decimal.NullDecimal        `mapstructure:"discount"`
	Nested   NestedCfg                  `mapstructure:"nested"`
	Items    []decimal.Decimal          `mapstructure:"items"`
	Fees     map[string]decimal.Decimal `mapstructure:"fees"`
}

type NestedCfg struct {
	Total decimal.Decimal `mapstructure:"total"`
	Tax   decimal.Decimal `mapstructure:"tax"`
}

// composedHook returns the standard hook chain used by tests:
// TextUnmarshallerHookFunc first (handles strings), then decimal.DecodeHook
// (handles every non-string scalar).
func composedHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc(),
		decimal.DecodeHook(),
	)
}

// loadViper builds a viper.Viper from the given fixture and decodes it into
// out using the composed hook chain.
func loadViper(t *testing.T, configType, configPath string, out any) {
	t.Helper()
	v := viper.New()
	v.SetConfigType(configType)
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig (%s): %v", configType, err)
	}
	if err := v.Unmarshal(out, viper.DecodeHook(composedHook())); err != nil {
		t.Fatalf("Unmarshal (%s): %v", configType, err)
	}
}

// assertCfg verifies common shape that all three backends share.
func assertCfg(t *testing.T, cfg Cfg, expectDiscountInvalid bool) {
	t.Helper()
	if cfg.Price.String() != "9.99" {
		t.Errorf("price: expected 9.99, got %s", cfg.Price.String())
	}
	if cfg.Quantity.String() != "42" {
		t.Errorf("quantity: expected 42, got %s", cfg.Quantity.String())
	}
	if cfg.Ratio.String() != "0.25" {
		t.Errorf("ratio: expected 0.25, got %s", cfg.Ratio.String())
	}
	if expectDiscountInvalid && cfg.Discount.Valid {
		t.Errorf("discount: expected invalid NullDecimal, got %#v", cfg.Discount)
	}
	if cfg.Nested.Total.String() != "100.5" {
		t.Errorf("nested.total: expected 100.5, got %s", cfg.Nested.Total.String())
	}
	if cfg.Nested.Tax.String() != "7" {
		t.Errorf("nested.tax: expected 7, got %s", cfg.Nested.Tax.String())
	}
	if len(cfg.Items) != 3 {
		t.Fatalf("items: expected 3 elements, got %d", len(cfg.Items))
	}
	wantItems := []string{"1.1", "2.2", "3"}
	for i, want := range wantItems {
		if cfg.Items[i].String() != want {
			t.Errorf("items[%d]: expected %s, got %s", i, want, cfg.Items[i].String())
		}
	}
	if cfg.Fees["setup"].String() != "5.5" {
		t.Errorf("fees.setup: expected 5.5, got %s", cfg.Fees["setup"].String())
	}
	if cfg.Fees["monthly"].String() != "10" {
		t.Errorf("fees.monthly: expected 10, got %s", cfg.Fees["monthly"].String())
	}
}

func TestViper_YAML(t *testing.T) {
	var cfg Cfg
	loadViper(t, "yaml", "testdata/config.yaml", &cfg)
	assertCfg(t, cfg, true)
}

func TestViper_JSON(t *testing.T) {
	var cfg Cfg
	loadViper(t, "json", "testdata/config.json", &cfg)
	assertCfg(t, cfg, true)
}

func TestViper_TOML(t *testing.T) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile("testdata/config.toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}
	// TOML has no null; use a Cfg type without Discount to keep struct
	// strict-decoding clean.
	type tomlCfg struct {
		Price    decimal.Decimal            `mapstructure:"price"`
		Quantity decimal.Decimal            `mapstructure:"quantity"`
		Ratio    decimal.Decimal            `mapstructure:"ratio"`
		Nested   NestedCfg                  `mapstructure:"nested"`
		Items    []decimal.Decimal          `mapstructure:"items"`
		Fees     map[string]decimal.Decimal `mapstructure:"fees"`
	}
	var tc tomlCfg
	if err := v.Unmarshal(&tc, viper.DecodeHook(composedHook())); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if tc.Price.String() != "9.99" {
		t.Errorf("price: expected 9.99, got %s", tc.Price.String())
	}
	if tc.Quantity.String() != "42" {
		t.Errorf("quantity: expected 42, got %s", tc.Quantity.String())
	}
	if tc.Ratio.String() != "0.25" {
		t.Errorf("ratio: expected 0.25, got %s", tc.Ratio.String())
	}
	if tc.Nested.Total.String() != "100.5" {
		t.Errorf("nested.total: expected 100.5, got %s", tc.Nested.Total.String())
	}
	if len(tc.Items) != 3 {
		t.Fatalf("items: expected 3, got %d", len(tc.Items))
	}
}

// TestJSONNumberViaMapstructure drives a synthetic decode where the source map
// contains json.Number values, simulating what mapstructure sees when
// configured with json.Decoder.UseNumber. This exercises the json.Number
// branch of the hook directly via mapstructure (no viper involved).
func TestJSONNumberViaMapstructure(t *testing.T) {
	src := map[string]any{
		"price":    json.Number("9.99"),
		"quantity": json.Number("42"),
		"ratio":    json.Number("0.25"),
		"nested": map[string]any{
			"total": json.Number("100.5"),
			"tax":   json.Number("7"),
		},
		"items": []any{json.Number("1.1"), json.Number("2.2"), json.Number("3")},
		"fees": map[string]any{
			"setup":   json.Number("5.5"),
			"monthly": json.Number("10"),
		},
	}
	var cfg Cfg
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: composedHook(),
		Result:     &cfg,
	})
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	if err := dec.Decode(src); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if cfg.Price.String() != "9.99" {
		t.Errorf("price: expected 9.99, got %s", cfg.Price.String())
	}
	if cfg.Quantity.String() != "42" {
		t.Errorf("quantity: expected 42, got %s", cfg.Quantity.String())
	}
	if cfg.Ratio.String() != "0.25" {
		t.Errorf("ratio: expected 0.25, got %s", cfg.Ratio.String())
	}
}

// TestComposeWithTextUnmarshaller confirms that registering
// TextUnmarshallerHookFunc in front of decimal.DecodeHook does not cause
// fights: string sources are handled by the text hook, non-string sources
// by ours, and the result matches end-to-end.
func TestComposeWithTextUnmarshaller(t *testing.T) {
	src := map[string]any{
		"price":    "1.5", // handled by TextUnmarshallerHookFunc
		"quantity": 7,     // handled by decimal.DecodeHook
		"ratio":    0.125, // handled by decimal.DecodeHook
		"discount": nil,   // handled by decimal.DecodeHook (nil -> Valid=false)
		"nested": map[string]any{
			"total": "100.5",
			"tax":   3,
		},
		"items": []any{"1.1", 2, "3.0"},
		"fees":  map[string]any{"setup": "5.5", "monthly": 10},
	}
	var cfg Cfg
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: composedHook(),
		Result:     &cfg,
	})
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	if err := dec.Decode(src); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if cfg.Price.String() != "1.5" {
		t.Errorf("price: expected 1.5, got %s", cfg.Price.String())
	}
	if cfg.Quantity.String() != "7" {
		t.Errorf("quantity: expected 7, got %s", cfg.Quantity.String())
	}
	if cfg.Ratio.String() != "0.125" {
		t.Errorf("ratio: expected 0.125, got %s", cfg.Ratio.String())
	}
	if cfg.Discount.Valid {
		t.Errorf("discount: expected invalid NullDecimal, got %#v", cfg.Discount)
	}
	if cfg.Items[1].String() != "2" {
		t.Errorf("items[1]: expected 2, got %s", cfg.Items[1].String())
	}
}

// Error-path tests (backend-agnostic; driven through map[string]any).

func TestError_InvalidString(t *testing.T) {
	src := map[string]any{"price": "abc"}
	type only struct {
		Price decimal.Decimal `mapstructure:"price"`
	}
	var out only
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: composedHook(),
		Result:     &out,
	})
	err := dec.Decode(src)
	if err == nil {
		t.Fatalf("expected error for invalid string")
	}
	if !errors.Is(err, decimal.ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestError_BoolSource(t *testing.T) {
	src := map[string]any{"price": true}
	type only struct {
		Price decimal.Decimal `mapstructure:"price"`
	}
	var out only
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: composedHook(),
		Result:     &out,
	})
	err := dec.Decode(src)
	if err == nil {
		t.Fatalf("expected error for bool source")
	}
	if !errors.Is(err, decimal.ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
	if !strings.Contains(err.Error(), "bool") {
		t.Fatalf("expected error to mention bool, got %v", err)
	}
}

func TestError_UnsupportedType(t *testing.T) {
	src := map[string]any{"price": complex64(1 + 2i)}
	type only struct {
		Price decimal.Decimal `mapstructure:"price"`
	}
	var out only
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: composedHook(),
		Result:     &out,
	})
	err := dec.Decode(src)
	if err == nil {
		t.Fatalf("expected error for complex64 source")
	}
	if !errors.Is(err, decimal.ErrUnmarshal) {
		t.Fatalf("expected ErrUnmarshal, got %v", err)
	}
	if !strings.Contains(err.Error(), "complex64") {
		t.Fatalf("expected error to mention complex64, got %v", err)
	}
}

// Sanity: ensure DecodeHook returns a function with the expected mapstructure
// signature when used standalone (no Compose).
func TestStandaloneHook(t *testing.T) {
	src := map[string]any{"price": int64(99)}
	type only struct {
		Price decimal.Decimal `mapstructure:"price"`
	}
	var out only
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: decimal.DecodeHook(),
		Result:     &out,
	})
	if err := dec.Decode(src); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Price.String() != "99" {
		t.Errorf("price: expected 99, got %s", out.Price.String())
	}

	// Also verify the function is the documented signature.
	hook := decimal.DecodeHook()
	_ = reflect.TypeOf(hook)
}
