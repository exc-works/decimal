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
