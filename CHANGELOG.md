# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/exc-works/decimal/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/exc-works/decimal/releases/tag/v0.3.0
[0.2.0]: https://github.com/exc-works/decimal/releases/tag/v0.2.0
[0.1.0]: https://github.com/exc-works/decimal/releases/tag/v0.1.0
