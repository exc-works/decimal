# Correctness gates

The [numerical contract](numeric-contract.md) determines correct behavior.
Performance results cannot relax its rounding or state requirements.

## Three complementary checks

1. **Independent numerical judgement.** Arithmetic uses exact math/big.Rat
   comparisons with neighboring grid points. Roots are checked against their
   exact integer midpoint boundaries, including odd/even ties. Logarithms and
   exponentials are checked against independently computed exact rational
   intervals; both endpoints must round to the same coefficient. The oracle
   calls none of the production fixed-point interval or series helpers.
2. **Properties and failure states.** Bounded fuzzing compares arithmetic with
   independent rationals and checks root rounding cells. Deterministic cases
   check binary floating ties/subnormals, scale, input immutability, nullable
   validity, decoder failure atomicity, and error identities/types.
3. **Algorithm review and regression commitments.** Review the derivations of
   bounds, rounding decisions, and termination, and retain confirmed failures
   as regressions. The versioned must-succeed corpus checks successful completion
   as well as values, preventing new errors or panics from hiding wrong results.
   Review and CI provide complementary evidence; passing samples is not a proof
   for every possible input.

## Must-succeed corpus v1 (v0.8.0)

[must_succeed.json](../testdata/must_succeed.json) contains 45 reviewable cases
for v0.8.0. Its membership is part of the maintainer-approved release contract.

The corpus covers all six default mathematical methods, higher input scale,
60-place requests, exact positive and negative midpoint ties, large and tiny
roots, both sides of logarithm/exponential rounding transitions, a 401st-degree
root regression, the irrational 500th root of two at 50 places, a large exact
dyadic logarithm, and exact/tiny-result paths at
high precision. Precision-boundary fast paths do not promise that every other
input at that same precision will finish.

Every listed case must:

* return successfully without an unexpected error or panic;
* return the specified coefficient at exactly the specified scale;
* preserve its input state;
* agree with an independent exact midpoint or interval certificate.

The oracle itself fails if its finite refinement budget cannot certify the
result; it never falls back to an epsilon or accepts an uncertified reference.
The exact-rational oracle deliberately has a smaller, curated workload than
the production library. It is test infrastructure, not a new production cap.

The JSON is reproducible with Python's standard-library Decimal at 500
significant digits:

```sh
python3 testdata/generate_must_succeed.py
git diff --exit-code -- testdata/must_succeed.json
```

Its expected values are also certified independently by Go tests. A mistaken
fixture edit cannot establish correctness merely by matching an implementation.
The older 1500-digit Python fixtures provide additional large-precision
cross-checks; they are not alone a strict interval certificate.

The CI regeneration check detects fixture/generator drift. Nonempty coverage,
unique IDs, all mathematical operations, and key historical regression IDs are
checked separately. Removing promised cases or shrinking a public support range
requires an explicit compatibility decision, not an unexplained fixture update.

## Independent bounds

For a reduced logarithm argument `1<=m<=2`, the oracle uses
`z=(m-1)/(m+1)` and exact rational terms of the atanh identity. After `N`
terms its positive tail is bounded by
`2*z^(2N+1)/((2N+1)*(1-z²))`. Powers of two are accounted for using independently
bounded ln(2), with interval directions reversed for negative multipliers.
Base changes use all four ratio corners against positive base-log intervals.

For Exp, reduce the positive argument to `r<=1/4`, sum its exact rational
Taylor terms, and bound the omitted tail by the first omitted term divided by
`1-r/(N+2)`. Square positive bounds during reconstruction and invert for
negative inputs. Very negative arguments use the independent inequality `e>2`
to certify zero cheaply. Exact rational arithmetic avoids the production
fixed-point roundoff bookkeeping; the mathematical identities and inequalities
still require review.

## CI and local commands

The existing Test job runs the complete unit, race, BSON and integration suites.
It includes the deterministic strict tests. Dedicated Correctness jobs run the
strict suite and reference regeneration on amd64 and 386; amd64 additionally
runs each numerical fuzzer for 20 seconds with two workers. The 386 run covers
fixed-width binary precision values before conversion to native int.

```sh
go test -count=1 -run 'Test(MustSucceed|Contract|MulWithPrec|DirectArithmetic|Independent)' ./...
go test -run '^$' -fuzz '^FuzzArithmeticCorrectRounding$' -fuzztime=20s -parallel=2
go test -run '^$' -fuzz '^FuzzRootCorrectRounding$' -fuzztime=20s -parallel=2
```

Fuzz inputs bound coefficient size, scale and degree for predictable CI work.
Any discovered failure must be retained as a replayable regression before the
implementation is repaired. Workflow checks are added by this change; repository
branch-protection settings are not modified.
