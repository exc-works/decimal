# Upgrading from v0.7.0 to v0.8.0

[中文升级说明](upgrade-v0.8.0.zh.md) · [Numerical contract](numeric-contract.md) ·
[Release notes](releases/v0.8.0.md)

v0.8.0 changes observable behavior as well as fixing incorrect results. No
exported API is removed relative to v0.7.0. The module path remains
`github.com/exc-works/decimal`, and the Go requirement remains 1.26.0 or newer.
The version can be installed after the v0.8.0 tag is published:

```sh
go get github.com/exc-works/decimal@v0.8.0
go mod tidy
go test ./...
```

Run your separate integration modules and BSON-tagged tests where applicable.
Review stored expected outputs and serialized data; successful compilation alone
does not detect the following behavior changes.

## Default scale and explicit precision

`Precision()` counts decimal places, not significant digits.

| Operation | v0.7.0 default output scale | v0.8.0 default output scale |
| --- | --- | --- |
| Sqrt, ApproxRoot, Ln, Log10, Exp | `max(input scale,30)` | `max(input scale,20)` |
| Log2 | Input scale | `max(input scale,20)` |

Use a WithPrec method to make your output requirement explicit. For example,
`d.SqrtWithPrec(max(d.Precision(), 30))` retains the old default **scale**;
`d.Log2WithPrec(d.Precision())` retains Log2's old scale and returns an error.
Corrected calculations may still differ from old numerical values at that scale.
Explicit precision methods return exactly their requested scale on success.

Do not ignore root/logarithm/exponential errors. Calculations now certify
HalfEven rounding using exact root midpoints or bounded intervals. They return
`ErrNoConvergence` if the result cannot be certified within the calculation
budget, instead of returning an uncertified approximation as success. `Log2`
retains its panic API; prefer `Log2WithPrec` when failure should be handled as an
error. A successful constructor does not imply every calculation on that value
will fit its budget. See the [contract](numeric-contract.md#errors-resources-and-ownership)
and [must-succeed corpus](correctness-gates.md) for the guarantees.

## Powers, multiplication, and corrected rounding

Negative integer powers now use HalfEven instead of RoundUp, rounding the exact
reciprocal only once at the input scale. Positive powers also use HalfEven.

| Expression | v0.7.0 | v0.8.0 |
| --- | --- | --- |
| `MustFromString("3.00").Power(-1)` | `0.34` | `0.33` |
| `New(2).Power(-1)` | `1` | `0` (HalfEven at scale 0) |
| `MustFromString("2.00").Power(-1)` | `0.50` | `0.50` (scale 2) |

Displayed trailing zeros above denote output scale; `String()` removes them.
To keep fractional places, construct or rescale the base **before** Power.
`New(2).Power(-1).Rescale(2, RoundHalfEven)` stays zero and cannot recover digits.
`0^0=1` is unchanged; zero to a negative power panics with `ErrDivideByZero`.
Unsupported exact power intermediates panic with `ErrOverflow`, including
impractical MinInt64 powers, rather than overflowing or recursing indefinitely.

`MulWithPrec` is new. It rounds the exact product once at the requested scale:

```go
a := decimal.MustFromString("1.23")
b := decimal.MustFromString("2.34")
p := a.MulWithPrec(b, 3, decimal.RoundHalfEven) // 2.878
```

`Mul` still uses the larger input scale, and `MulExact` uses their sum. Rescaling
a rounded product afterward cannot recover its discarded digits. `AddRaw` and
`SubRaw` still operate on coefficient units: `1.23.AddRaw(1)` is `1.24`.

Regression expectations must reflect corrected mathematics: for example,
`16.0 / 11` via `QuoWithPrec(...,0,RoundHalfEven)` is now `1`, not `2`.
Root ties and logarithm/exponential last digits have also been corrected.
Explicit `%g/%G` formatting now chooses notation after rounding; `%.3g` of
`999.9` is `1e+03`. Do not preserve old buggy values as compatibility targets.

Equal-valued instance Max/Min now return the receiver itself, matching package
Max/Min's first-operand rule. Numeric equality is unchanged, but output scale or
IsNil may change when equal operands have different representations.

## Error handling and invalid input

Numeric panic payloads for divide-by-zero, negative precision, arithmetic size
overflow, and required rounding under RoundUnnecessary are now errors exposing
their sentinel categories. Log2's domain/certification panics also expose
`ErrInvalidLog`/`ErrNoConvergence`. Update code that asserts a string after
`recover()`; first assert `error`, then use `errors.Is`.

Returned decoding/Scan errors wrap `ErrUnmarshal` while retaining underlying
identities, including `ErrInvalidFormat` and JSON/XML syntax/type errors. Use
`errors.Is`/`errors.As`, not outer concrete-type assertions or error-string
equality. Failed decodes preserve value, scale, IsNil, and NullDecimal.Valid.

Invalid rounding modes are rejected even on zero, exact, and unchanged-scale
fast paths. Repeated signs such as `--1` and `1e--2` are rejected by parsing.
Nil validator/translation setup arguments now consistently return
`ErrInvalidArgument`. There is no blanket new classification for other
programmer-error panics. Native conversions retain their `(value,bool)` APIs;
Float32/Float64's bool reports exactness, not general conversion success.

## Decoding, nullable values, and persisted formats

NullDecimal.Valid is the sole nullness discriminator. Valid=true with an
uninitialized Decimal is valid zero. For example, `NewNullDecimal(Decimal{})`
now encodes as JSON `"0"`, XML text `0`, and BSON string `0`, rather than an
empty/null representation. Use Valid=false for null. Encoding does not change
the original initialization state.

YAML numeric decoding uses the original scalar text instead of a binary float
intermediate, preserving high-precision digits. yaml.v2/v3 numeric syntax still
applies to unquoted scalars: unquoted `012` denotes ten, while quoted `"012"`
is parsed as decimal twelve. Check fixtures and persisted configuration values
that previously depended on floating-point loss or scalar quoting.

There is no new binary wire format. Encoding still normalizes fractional
trailing zeros and does not promise to preserve scale or IsNil. Unsupported
binary scales and malformed coefficients now return errors, and failed binary
decodes preserve the receiver. Audit callers that ignored these errors.

Null/empty rules remain format-specific; refer to the
[decoding table](numeric-contract.md#decoding-states-and-nullable-values).
In particular, JSON/SQL empty strings are errors, empty Text/Param clears a
NullDecimal, and empty Binary input produces initialized zero. Third-party
frameworks may intercept null before invoking these methods.

## Application upgrade checks

1. Set explicit WithPrec scales where application output requires 30 places or
   Log2's former scale; check both numerical value and Precision().
2. Recheck negative powers and rounding boundaries against the mathematical
   requirement, including scale-zero reciprocals and multiplication precision.
3. Handle ErrNoConvergence, numeric error-valued panics, and wrapped decode
   errors; verify failures leave existing receiver state unchanged.
4. Recheck JSON/XML/BSON nullable outputs, precise YAML inputs, and binary error
   handling against representative application data.
5. Run application and integration tests, including `-tags bson` if used, before
   deploying the dependency update. Rollback means pinning v0.7.0 and restoring
   application expectations/data as needed; it also restores the old bugs.
