# Numerical contract

These are the maintainer-approved rules for v0.8.0. They do not redefine earlier
releases. See the
[migration notes](numeric-contract-migration.md) for observable changes.

This document is the primary specification for the numerical operations,
conversions, decoding states, and formatting described below. Go comments, examples, and guides
must agree with it. An implementation, an existing expected value, or another
library's output is evidence to check against this contract, not its definition.

## Value, representation, and correctness

An initialized Decimal has an integer coefficient `c` and nonnegative scale `s`:

```text
value = c × 10^(-s)
```

`Precision()` is this scale: decimal places, not significant digits, binary
precision, or an estimate of how many digits are accurate. Results must satisfy
both their numerical value and the output scale specified by their API.
Comparisons use mathematical values, so `1.0` and `1.00` compare equal.

Exact operations must equal the exact mathematical result. Rounded operations
must return the correctly rounded result at the specified scale and mode,
including its last digit. This includes roots, logarithms, and exponentials.
An absolute-error tolerance alone does not establish that guarantee.

An operation that cannot finish or certify its rounding must use its documented
error or panic path. It must not return an uncertified approximation as success.
This success guarantee does not promise that every mathematically defined input
can be calculated within finite implementation budgets.

The uninitialized value `Decimal{}` acts numerically as zero in arithmetic, but
`IsNil()` distinguishes it from initialized zero. Decimal does not store a
negative-zero sign. A zero arithmetic result retains the operation's target
scale, even when compact string formatting displays it as `0`.

## Rounding

Rounding selects a multiple of `10^(-p)` for target scale `p`.

| Mode | Rule |
| --- | --- |
| `RoundDown` | Toward zero |
| `RoundUp` | Away from zero whenever inexact |
| `RoundCeiling` | Toward positive infinity |
| `RoundHalfUp` | Nearest; ties away from zero |
| `RoundHalfDown` | Nearest; ties toward zero |
| `RoundHalfEven` | Nearest; ties to an even coefficient on the target grid |
| `RoundUnnecessary` | Exact results only; panic with `ErrRoundUnnecessary` if rounding is required |

HalfUp and HalfDown have these same meanings for negative values. Rounding is
performed from the exact mathematical result at the final target scale, rather
than from a result already rounded at an intermediate scale.

## Arithmetic and output scale

For operand scales `s` and `t`:

| Operation | Value and output scale |
| --- | --- |
| `Add`, `Sub` | Exact sum or difference, scale `max(s,t)` |
| `AddRaw(k)`, `SubRaw(k)` | Add or subtract `k × 10^(-s)`, preserving `s` |
| `Mul(e,mode)` | Product rounded once at scale `max(s,t)` |
| `MulExact(e)` | Exact product, scale `s+t`, subject to its supported scale bound |
| `MulWithPrec(e,p,mode)` | Exact product rounded once at scale `p`; pad zeros if needed |
| `Quo(e,mode)` | Exact quotient rounded once at scale `max(s,t)` |
| `QuoWithPrec(e,p,mode)` | Exact quotient rounded once at scale `p` |
| `Power(n)` | Integer power rounded once using HalfEven at input scale `s` |

For example, `1.23.AddRaw(1)` is `1.24`; numerical addition of one is
`d.Add(New(1))`. For `1.23 × 2.34`, Mul at scale 2 yields `2.88`, MulExact yields
`2.8782` at scale 4, and MulWithPrec at scale 3 yields `2.878`. Increasing scale
after an already rounded multiplication cannot recover discarded digits.

MulWithPrec computes directly at the target scale. It does not require a
MulExact intermediate and does not inherit MulExact's intermediate scale cap.
Negative target precision and invalid rounding modes follow the existing
arithmetic parameter conventions; allocation cost still depends on the input
coefficients and requested representation.

`Power(0)` is defined as one for every base, including `0^0`, with input scale.
Zero to a negative power panics with `ErrDivideByZero`. Positive and negative
powers both use HalfEven. Power does not automatically increase a small input
scale: `New(2).Power(-1)` yields zero at scale 0; a base represented as `2.00`
yields `0.50` at scale 2.

`QuoRem` truncates its quotient toward zero at scale 0. Its remainder has scale
`max(s,t)` and satisfies `x = q*y+r`, `abs(r)<abs(y)`, and, if nonzero, has the
sign of `x`. `Mod` returns that same remainder.

`SignificantFigures(n,mode)` counts significant digits. It rounds when needed,
but does not add zeros to reach `n` digits when no rounding is needed.

Max and Min, including instance methods, return the first operand itself when
values compare equal. Thus `a.Max(b)` preserves `a`'s scale and initialization
state for equal values.

## Roots, logarithms, and exponential

The default output scale of Sqrt, ApproxRoot, Ln, Log2, Log10, and Exp is
`max(input scale,20)`. Their WithPrec variants use exactly the requested output
scale `p`. These scales count decimal places, not significant digits.
Successful results use HalfEven, including exact-value and zero-result paths.

Sqrt requires a nonnegative input. ApproxRoot requires positive integer degree;
even-degree roots require a nonnegative input, and odd-degree roots select the
real root, allowing negative inputs. Logarithms require a positive input.

Root implementations use exact integer midpoint comparisons. Logarithms and
Exp certify rounding with outward intervals and remainder bounds. Their bounded
Ziv strategy increases internal guard precision when an interval cannot yet
certify a single rounded result. The requested output scale stays fixed.
Internal guard bits and iteration counts are implementation budgets, not
additional accurate decimal digits promised to the caller.

Invalid root or logarithm inputs use their documented `ErrInvalidRoot`,
`ErrNegativeRoot`, or `ErrInvalidLog` paths. Budget or certification failures use
`ErrNoConvergence` where documented. Log2 retains its panic-based failure API;
Log2WithPrec returns an error. This contract does not change those API shapes.

## Construction, zero, and native conversions

NewFromString parses decimal text exactly within its supported input range.
Nonzero values preserve the resulting nonnegative scale; parsed zero is
normalized to scale 0. To preserve a chosen zero scale, use an explicit scale
constructor such as `NewWithPrec(0,p)`.

StripTrailingZeros is explicit normalization: remove fractional trailing zeros
without reducing scale below 0; numerical zero becomes initialized Zero at
scale 0. It does not mutate its input.

NewFromFloat32/64 format finite input using its own float bit width and the
shortest round-tripping decimal digits, then parse that decimal exactly. For
example, float64 `0.1` becomes decimal `0.1`, rather than the full binary
rational value stored by float64. Information already lost before conversion
cannot be recovered; use strings for exact decimal input.

NaN and infinities are rejected: native float constructors panic, and supported
float decoding paths return an error identifiable as ErrUnmarshal. Floating
negative zero becomes initialized decimal zero at scale 0, without its sign.

Float32/64 round the exact Decimal value to the nearest target floating value,
with ties to even in the binary significand. The returned bool means exact
mathematical equality, not general conversion success: decimal `0.5` is exact
in float64; decimal `0.1` is not.

Floating overflow returns same-sign infinity and false, without a numeric
overflow panic. Conversion supports subnormal values rather than flushing them
all to zero. Nonzero values that round to zero return same-sign floating zero
and false. A negative floating-zero output does not add negative-zero state to
Decimal itself. Exactly representable subnormals can return true.

## Decoding states and nullable values

When an error-returning decoder or Scan method fails, the receiver's value,
scale, IsNil, and, for NullDecimal, Valid must all remain as before the call.
Parsing into a temporary and committing only on success avoids partial updates.

Successful null or empty input has an entry-point-specific meaning:

| Input | Decimal | NullDecimal |
| --- | --- | --- |
| JSON literal `null` | Preserve the previous state | Clear to uninitialized Decimal, scale 0, Valid=false |
| YAML null passed to UnmarshalYAML | Preserve the previous state | Clear to uninitialized Decimal, scale 0, Valid=false |
| JSON string `""` | Parse error; preserve state | Parse error; preserve state |
| Empty Text / Param | Parse error; preserve state | Clear to uninitialized Decimal, scale 0, Valid=false |
| Empty or whitespace-only XML content / attribute value | Clear to uninitialized Decimal, scale 0 | Clear to uninitialized Decimal, scale 0, Valid=false |
| SQL Scan(nil) | Clear to uninitialized Decimal, scale 0 | Clear to uninitialized Decimal, scale 0, Valid=false |
| SQL empty string or byte slice | Parse error; preserve state | Parse error; preserve state |
| BSON Null | Clear to uninitialized Decimal, scale 0 | Clear to uninitialized Decimal, scale 0, Valid=false |
| DecodeHook nil | ErrUnmarshal | Uninitialized Decimal, scale 0, Valid=false |
| DecodeHook empty string or byte slice | ErrUnmarshal and ErrInvalidFormat | Uninitialized Decimal, scale 0, Valid=false |
| Binary empty byte input | Initialized zero at scale 0 | No corresponding Binary API |

The table describes calls to the package's methods or hook. A third-party
framework may handle null or missing fields before invoking custom methods;
its dispatch behavior is not established by this table. Quoted YAML `"null"`
is a string and fails numerical parsing. Zero-length input and whitespace-only
text are not generally interchangeable.

NullDecimal.Valid is its sole nullness discriminator. Valid=false encodes using
that format's empty-value rule. Valid=true encodes as a valid number even if its
Decimal field is uninitialized; that state means valid zero. For example,
NewNullDecimal(Decimal{}) encodes as JSON `"0"`, XML character data `0`, a BSON
string `0`, and SQL string `0`. Encoding does not mutate the source state.

Compact encodings generally preserve numerical value, not every representation
detail. Fractional trailing zeros can be normalized. In particular, an
uninitialized Decimal encodes as empty Binary data, which decodes to initialized
zero: the value is zero but IsNil changes from true to false. That behavior is
intentional. NullDecimal uses a named Decimal field and does not acquire its
Binary methods through promotion.

## Formatting

Formatting does not mutate the Decimal. String emits compact decimal text
without fractional trailing zeros; StringWithTrailingZeros preserves scale.
Neither preserves original spelling such as a leading plus sign or exponent.

`%f` without precision retains the receiver's scale and trailing zeros. With
precision `p`, it uses HalfEven at `p` decimal places and displays those places.
`%e/%E` precision counts places after the leading digit, defaulting to 6.

Ordinary `%g/%G` without precision or `#` preserves the complete numerical
value and removes redundant trailing zeros. It uses scientific notation for
decimal exponent `e<-4` or `e>=6`; the threshold 6 does not truncate the value
to six significant digits.

With explicit significant-digit count `n`, `%g/%G` first rounds using HalfEven,
then chooses scientific notation when the rounded exponent `e<-4` or `e>=n`;
otherwise it uses fixed form. Explicit precision 0 uses one significant digit.
The existing `#` behavior retains zeros and uses six significant digits when
precision is omitted. Notation is chosen by exponent thresholds, not by comparing
the lengths of two candidate strings.

## Errors, resources, and ownership

Documented error categories are stable through errors.Is; diagnostic wording
is not a fixed interface. Existing numeric panic categories retain their
published behavior. There is no blanket requirement to change programmer-error
panics into ErrInvalidArgument or to add nil-argument checks.

The package's error-returning decoding and Scan entry points wrap ErrUnmarshal
and preserve underlying identities. This includes JSON, YAML, XML, BSON,
Text/Param, Binary, SQL Scan, and supported DecodeHook targets. Invalid numeric
text also remains identifiable as ErrInvalidFormat. JSON/XML syntax error types
remain available through errors.As. NewFromString itself returns ErrInvalidFormat,
without adding the decoding category. Errors produced by external frameworks
outside these entry points are outside this wrapping guarantee.

Current entry points have different representation and calculation budgets.
The parsed fractional/exponent magnitude is limited to `1<<17`; that is not
a universal bound on coefficients or every arithmetic scale. A constructible
value does not guarantee that every subsequent calculation can finish.

Publicly promised support ranges and the versioned must-succeed corpus are
compatibility commitments. Reducing them requires an explicit compatibility
decision. Internal budgets may change while satisfying numerical correctness,
the promised successful cases, and documented failure behavior. Tests must
check unexpected errors and panics as well as values on successful cases.
The v0.8.0 corpus and independent verification methods are specified in
[correctness gates](correctness-gates.md).

Arithmetic does not mutate operands. Exposed big.Int and big.Rat values are
independent copies; internal sharing is allowed only for read-only data.
Concurrent reads require that no goroutine reassign the same Decimal variable.
Pointer-receiver decoding requires caller synchronization. Exported singleton
variables such as Zero and One must be treated as read-only.
