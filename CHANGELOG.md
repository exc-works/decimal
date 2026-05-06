# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `(Decimal).Append(dst []byte) []byte` and
  `(Decimal).AppendWithTrailingZeros(dst []byte) []byte`: append the canonical
  decimal text to `dst` without an intermediate `string` allocation. Intended
  for hot paths that build their own byte buffer (proto wire field setters,
  custom JSON, log builders).

### Changed
- `MarshalJSON` rewritten to append `"…"` directly via the new internal
  `appendString` helper, bypassing the prior `String() + json.Marshal(string)`
  round-trip. Microbench (Apple M5 Max, BENCHTIME=2s × 3):
    - integer / decimal-with-prec: 100-110 ns / 4 allocs → 60-65 ns / 2 allocs
    - negative: 127 ns / 5 allocs → 77 ns / 2 allocs
    - trailing-zeros: 118 ns / 6 allocs → 87 ns / 4 allocs
- `String` and `StringWithTrailingZeros` now share the same `appendString`
  workhorse and use a 48-byte stack scratch to skip one slice grow; behavior
  is identical, allocation count is the same or one lower (negative path).

## [0.5.0] - 2026-04-28

### Fixed
- `MostSignificantBit(*big.Int)` no longer mutates its argument. Previously
  the binary-search loop right-shifted into the input via `x.Rsh(x, ...)`,
  silently corrupting the caller's `*big.Int`. The implementation is now a
  one-liner backed by `(*big.Int).BitLen`.
- `Quo` now panics on an invalid `RoundingMode` even when the division is
  exact, matching its godoc. Previously the exact-division fast path
  returned without validating the mode.

### Changed
- **Behavior, edge cases only**: `Quo`'s halfway-rounding decision is now
  driven by the exact remainder versus the divisor, rather than the prior
  approach of scaling the numerator by `10^(2·maxPrec)` and rounding the
  oversized intermediate. Results are unchanged for the vast majority of
  inputs; for inputs whose true tail straddled the halfway point in a
  digit beyond what the prior scaling preserved, the rounded last digit
  may differ by one ulp (the new result is the mathematically correct one).

### Security
- `NewFromString` rejects parsed precisions whose magnitude exceeds
  `1<<17` (e.g. `"1e2000000000"`), preventing pathological inputs from
  triggering a `10^|precision|` `big.Int` allocation. The same cap is
  applied to `UnmarshalBinary`, which previously stored an attacker-supplied
  `uint32` precision verbatim and only failed lazily on first arithmetic.

### Performance
- `Quo` now performs a single `Mul` plus `QuoRem` (using the exact integer
  remainder for rounding) instead of two `Mul`s, a `Quo`, and a follow-up
  rounding division. Microbenchmarks show ~30–60% latency and ~30–70%
  allocation reductions across the `Quo` matrix; downstream `Log2` /
  `Ln` / `Log10` / `Exp` / `Sqrt` benefit transitively.
- `BigRat`, `Float64`, and `Float32` no longer wrap the cached precision
  multiplier in a redundant `new(big.Int).Set(...)`. `big.Rat.SetFrac`
  already copies its inputs internally, so the wrapper produced one extra
  heap allocation per call with no safety benefit.
- `Ln` / `LnWithPrec` parse the `ln2` constant once at package init
  (`ln2Decimal`) instead of re-parsing the literal on every call.

### Internal
- New shared helper `applyDivisionRounding(quo, rem, divisor, mode)` in
  `round.go` consolidates the integer-division rounding logic; both `Quo`
  and `NewFromBigRatWithPrec` now route through it. Rounding for `Quo`
  also correctly handles a negative divisor.
- New `validateRoundingMode(mode)` helper in `round.go` enforces the
  invalid-mode panic from public entry points whose fast paths bypass
  `applyDivisionRounding`.
- Minor cleanups: `big.NewInt(0).SetInt64(root)` → `big.NewInt(root)` in
  `ApproxRootWithPrec`; dead `strings.Repeat("0", 0)` removed from
  `roundSignificand`; unused `twoInt` package var removed.

## [0.4.0] - 2026-04-27

### Added
- `DecodeHook()` — mapstructure-compatible config decoder hook for viper,
  koanf, confita, cleanenv, and any other mapstructure-based loader.
  Decodes non-string scalar config values (`int` / `uint` / `float` /
  `json.Number` / `[]byte` / `nil`) into `Decimal` and `NullDecimal`,
  with `bool` and non-finite floats explicitly rejected via `ErrUnmarshal`.
  Designed to compose with `mapstructure.TextUnmarshallerHookFunc()`,
  which already covers the string path through `Decimal.UnmarshalText`.
  The hook lives in the main module and pulls in **no** viper or
  mapstructure dependencies; real end-to-end viper round-trip tests
  (YAML / JSON / TOML, plus the `ComposeDecodeHookFunc` interop case)
  live in an isolated sub-module at `integration/mapstructure/` so the
  main module's dependency surface is unchanged.
- New "Config Decoding (viper / mapstructure)" chapter in all 12
  multilingual user guides (en/zh/zh-Hant/ja/ko/fr/es/de/pt-BR/ru/ar/hi),
  with a byte-for-byte identical Go example across languages.

## [0.3.0] - 2026-04-21

### Added
- Sentinel errors for programmatic handling via `errors.Is`: `ErrOverflow`,
  `ErrDivideByZero`, `ErrInvalidPrecision`, `ErrInvalidFormat`, `ErrNegativeRoot`,
  `ErrInvalidRoot`, `ErrInvalidLog`, `ErrRoundUnnecessary`, `ErrUnmarshal`,
  `ErrInvalidArgument`.
- `SqrtWithPrec(prec)` and `ApproxRootWithPrec(root, prec)` for explicit
  precision control, mirroring `Log10WithPrec`/`LnWithPrec`/`ExpWithPrec`.
- `NullDecimal` — nullable Decimal wrapper implementing `sql.Scanner`,
  `driver.Valuer`, JSON/YAML/Text/XML/BSON marshalers, and gin `UnmarshalParam`.
- New arbitrary-precision math functions: `Log10`/`Log10WithPrec`,
  `Ln`/`LnWithPrec`, `Exp`/`ExpWithPrec`.
- `fmt.Formatter` implementation on `Decimal` supporting `%v`/`%s`/`%q`/`%d`/`%f`/`%e`/`%g`/`%b`
  with width, precision, and flag handling. `%d` honors the precision flag
  (`%.4d` pads with leading zeros, matching `fmt` stdlib); `%d` on a
  non-integer Decimal produces the `%!d(decimal.Decimal=<value>)` error
  marker instead of silently truncating; `%b` honors the `+` and ` ` sign
  flags consistent with `fmt`'s treatment of floats.
- `Decimal.Clone()` and package-level `NewFromDecimal(Decimal)` deep-copy constructor.
- `Decimal.FormatWithSeparators(thousands, decimal rune)` for locale-aware display.
- XML serialization on `Decimal` and `NullDecimal`: `MarshalXML` /
  `UnmarshalXML` / `MarshalXMLAttr` / `UnmarshalXMLAttr`.
- Optional BSON serialization (opt-in via `-tags bson`) through
  `go.mongodb.org/mongo-driver/v2/bson` on both `Decimal` and `NullDecimal`
  (supports String, Double, Int32, Int64, Decimal128, Null). Keeps the core
  module free of MongoDB dependency for users that do not need BSON.
- Six new validator tags: `decimal_ne`, `decimal_between`, `decimal_positive`,
  `decimal_negative`, `decimal_nonzero`, `decimal_max_precision`. The
  `decimal_max_precision` tag checks **decimal places (scale)**, not total
  significant digits; `decimal_between` now also validates that `min <= max`
  at registration-time.
- Validator translations expanded to all 13 locales matching user-guide coverage:
  en/zh/zh-Hant/ja/ko/fr/es/de/pt/pt-BR/ru/ar/hi.
- Thread-safety guarantees documented in `doc.go`; concurrent read tests
  verified under `-race`, including defensive-copy regression tests for
  `BigInt()` / `BigRat()`.
- `TestMathConstantsSanity` verifies the hard-coded 70-digit `ln2Literal`
  against independently computed cross-checks, preventing silent drift in
  `Log10` / `Ln` / `Exp`.

### Changed
- Existing `fmt.Errorf` returns in `decimal.go` and `marshal.go` (and the new
  `marshal_bson.go` / `marshal_xml.go`) now wrap the new sentinel errors
  for `errors.Is` compatibility on every failure path.
- **Breaking**: `Sqrt()` and `ApproxRoot()` now auto-bump the output precision
  to `max(d.Precision(), 30)`, consistent with `Log10`/`Ln`/`Exp`. Previously,
  integer receivers (e.g. `decimal.New(4).Sqrt()`) returned poor low-precision
  approximations such as `"1"`; they now return the correct value. Callers
  that need the old behavior can use `SqrtWithPrec(d.Precision())` or
  `ApproxRootWithPrec(root, d.Precision())`. Doc comments on all four
  functions flag the breaking change.
- `ApproxRoot` non-positive `root` errors now wrap `ErrInvalidRoot` instead of
  returning a plain string error.
- Validator tag parameters are now documented as compile-time constants:
  passing malformed parameters (non-numeric limits, unparseable decimal
  literals, or inverted `min > max` for `decimal_between`) panics at
  validation time. Do not splice untrusted input into struct tags.

## [0.2.0] - 2026-04-15

### Added
- Publishing baseline artifacts: `LICENSE`, CI/release workflows, package-level docs, and contributor/security guides.
- Added `MulExact` as the clear API name for exact multiplication and kept `Mul2` as a deprecated compatibility alias.
- Added `Decimal.UnmarshalParam(string)` to support gin `BindUnmarshaler` for query/form/uri binding.
- Added real gin integration tests covering `ShouldBindQuery`, `ShouldBindUri`, and `ShouldBindJSON`.
- Documented gin support and usage in `README.md`.
- Added `RegisterGoPlaygroundValidator(*validator.Validate)` to register Decimal comparison tags:
  `decimal_required`, `decimal_eq`, `decimal_gt`, `decimal_gte`, `decimal_lt`, `decimal_lte`.
- Added exact decimal comparison for custom decimal validator tags (uses `Decimal.Cmp`, not `Float64`).
- Added validator error-message helpers:
  `RegisterGoPlaygroundValidatorTranslations` and `TranslateGoPlaygroundValidationErrors`.
- Added multi-language validator translation support (en/zh/ja/fr/es/de/pt) with
  locale-aware defaults and `RegisterGoPlaygroundValidatorTranslationsWithMessages`
  for custom message overrides.
- Added real `go-playground/validator` and gin integration tests for `decimal_*` numeric tags on `Decimal`.

## [0.1.0] - 2026-04-15

### Added
- Initial public release of the `decimal` package.
- Arbitrary-precision decimal arithmetic based on `math/big.Int`.
- Rounding modes, precision rescaling, comparisons, and formatting helpers.
- JSON/YAML/Text/Binary/SQL serialization support.
- Extensive unit tests, examples, and benchmarks.

[Unreleased]: https://github.com/exc-works/decimal/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/exc-works/decimal/releases/tag/v0.5.0
[0.4.0]: https://github.com/exc-works/decimal/releases/tag/v0.4.0
[0.3.0]: https://github.com/exc-works/decimal/releases/tag/v0.3.0
[0.2.0]: https://github.com/exc-works/decimal/releases/tag/v0.2.0
[0.1.0]: https://github.com/exc-works/decimal/releases/tag/v0.1.0
