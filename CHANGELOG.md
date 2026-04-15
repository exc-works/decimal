# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/exc-works/decimal/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/exc-works/decimal/releases/tag/v0.2.0
[0.1.0]: https://github.com/exc-works/decimal/releases/tag/v0.1.0
