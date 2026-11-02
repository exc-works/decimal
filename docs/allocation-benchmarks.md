# Allocation benchmarks

Measured on Apple M5 Max, darwin/arm64, Go 1.26.6. Baseline: main at
`b9c3d5d4956189bf5becf0fcef7c1b2397a0e0ba`. Each entry is the median of three
150 ms samples. Timings are indicative; allocation counts are the main evidence.

```sh
go test -run '^$' -bench 'Benchmark(IntegerInspection|NativeFloatConversion|RationalConstruction|IntegerFormatting)$' -benchmem -benchtime=150ms -count=3
```

## Integer inspection

| Operation / input | ns/op before → after | B/op before → after | allocs/op before → after |
| --- | ---: | ---: | ---: |
| IsInteger / integer | 32.9 → 1.74 | 72 → 0 | 3 → 0 |
| Int64 / integer | 33.0 → 1.86 | 72 → 0 | 3 → 0 |
| IntPart / integer | 31.6 → 20.2 | 72 → 40 | 3 → 2 |
| Int64 / scaled integer | 34.8 → 14.7 | 72 → 8 | 3 → 1 |
| IsInteger / coefficient 1, scale 1000 | 566 → 2.10 | 1233 → 0 | 12 → 0 |
| `%d` / integer | 187 → 144 | 224 → 120 | 11 → 7 |
| `%d` / scaled integer | 187 → 176 | 224 → 128 | 11 → 8 |

Integer queries calculate only the needed quotient or remainder, with direct
inspection at scale zero and a bit-length bound for values below one. `IntPart`
still returns an independently mutable integer. Native integer conversions use
local scratch values for division and retain exactness and range checks.

Tests compare against exact rationals and cover extreme scales and ownership.
Benchmarks also include fractions, wide integers, tiny values, and scaled zero.
No cache or shared mutable scratch storage is introduced.

## Rational and floating conversions

| Operation / input | ns/op before → after | B/op before → after | allocs/op before → after |
| --- | ---: | ---: | ---: |
| Float64 / integer | 80.8 → 2.64 | 96 → 0 | 7 → 0 |
| Float32 / integer | 81.5 → 26.2 | 96 → 16 | 7 → 2 |
| Float64 / wide integer | 218 → 60.7 | 712 → 256 | 9 → 2 |
| Float32 / scaled zero | 592 → 1.66 | 1609 → 0 | 10 → 0 |
| NewFromBigRat / integer | 92.9 → 19.7 | 176 → 40 | 10 → 2 |
| NewFromBigRat / zero | 12.7 → 1.64 | 32 → 0 | 1 → 0 |
| NewFromBigRatWithPrec / terminating | 62.6 → 39.6 | 160 → 96 | 5 → 2 |
| NewFromBigRatWithPrec / repeating | 87.6 → 65.1 | 216 → 152 | 7 → 4 |

Exact integer rationals bypass denominator factorization; rounded rational
construction reuses its owned coefficient buffer and reads the input directly.
Tests check all rounding modes against an independent rational oracle and verify
input ownership. Integer floating conversions round directly to the target
binary format; Float32 deliberately avoids conversion through Float64.
Random wide integers, overflow, and an explicit double-rounding counterexample
are compared with exact rational conversion. Fractional floating conversions
keep their rational path (unchanged allocation counts); exact noninteger rational
construction also retains its factorization path.

## Rescale and significant figures

For this group, five 300 ms samples were run sequentially against baseline
`ffac032` and the optimized tree, with no other test process running. The
preceding stack commits leave these rounding implementations unchanged.

```sh
go test -run '^$' -bench 'Benchmark(RescaleRounding|SignificantFiguresRounding)$' -benchmem -benchtime=300ms -count=5
```

| Operation / input | ns/op before → after | B/op before → after | allocs/op before → after |
| --- | ---: | ---: | ---: |
| Rescale HalfEven / positive tie | 60.4 → 50.3 | 128 → 96 | 5 → 4 |
| Rescale HalfEven / negative tie | 99.1 → 54.8 | 208 → 96 | 9 → 4 |
| Rescale HalfEven / negative exact | 100 → 25.0 | 200 → 40 | 8 → 2 |
| SignificantFigures / positive | 130 → 134 | 192 → 152 | 8 → 6 |
| SignificantFigures / negative | 195 → 132 | 312 → 192 | 14 → 8 |

Rescale uses the same signed quotient/remainder rounding implementation as
arithmetic, removing separate sign-recursion and duplicated tie decisions.
Truncation still computes only a quotient. SignificantFigures borrows its input
coefficient read-only, then rescales the independently owned rounding result in
place. Tests cover all modes, ties, carries, rounding to integer multiples,
requested scale, panic identity, and input immutability against exact rationals.
The positive SignificantFigures timing is effectively unchanged; its measured
benefit is fewer allocations.

## Audit boundaries

The audit inspected arithmetic, roots, certified logarithms/exponentials,
interval helpers, formatting/parsing, codecs, and integration helpers. Returned
big integers and interval endpoints need independent ownership. Existing
alternating series buffers and per-call scratch storage preserve that property;
they should not be replaced with shared mutable integers to reduce allocation
counts. These benchmarks target the concrete paths changed in this stack and do
not establish an allocation limit for every API or arbitrary coefficient size.
