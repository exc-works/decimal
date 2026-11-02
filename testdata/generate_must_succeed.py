"""Regenerate proposed corpus v1 using Python's independent Decimal arithmetic.

Go tests also certify roots by integer midpoint comparisons and selected
transcendentals by exact rational intervals; these fixtures alone are not proofs.
Only the Python standard library is needed. Run from any working directory.
"""

import json
from decimal import Decimal, ROUND_HALF_EVEN, localcontext
from pathlib import Path


cases = []


def add(name, op, text, prec, degree=None, default=False, exact=None):
    with localcontext() as ctx:
        ctx.prec = 500
        x = Decimal(text)
        if exact is not None:
            want = exact
        else:
            if op == "sqrt":
                y = x.sqrt()
            elif op == "root":
                y = abs(x) ** (Decimal(1) / degree)
                if x < 0:
                    y = -y
            elif op == "ln":
                y = x.ln()
            elif op == "log2":
                y = x.ln() / Decimal(2).ln()
            elif op == "log10":
                y = x.log10()
            elif op == "exp":
                y = x.exp()
            else:
                raise ValueError(op)
            want = "0" if not y else format(
                y.quantize(Decimal(1).scaleb(-prec), rounding=ROUND_HALF_EVEN), "f"
            )
    case = dict(id=name, op=op, input=text, prec=prec, want=want)
    if degree is not None:
        case["degree"] = degree
    if default:
        case["default"] = True
    cases.append(case)


for op, text in [("sqrt", "2"), ("root", "2"), ("ln", "2"),
                 ("log2", "3"), ("log10", "2"), ("exp", "1")]:
    add("default-" + op, op, text, 20, degree=3 if op == "root" else None,
        default=True)

add("default-high-scale-sqrt", "sqrt", "2." + "0" * 25, 25, default=True)
for name, text, prec in [
    ("sqrt-zero-midpoint", "0.25", 0),
    ("sqrt-integer-midpoint", "2.25", 0),
    ("sqrt-lower-even-midpoint", "1.5625", 1),
    ("sqrt-upper-even-midpoint", "1.8225", 1),
    ("sqrt-large", "4e200", 2),
    ("sqrt-small", "1e-200", 105),
    ("sqrt-sixty-places", "2", 60),
]:
    add(name, "sqrt", text, prec)

for name, text, degree, prec in [
    ("odd-negative-lower-midpoint", "-1.953125", 3, 1),
    ("odd-negative-upper-midpoint", "-2.460375", 3, 1),
    ("odd-negative-large", "-1e21", 3, 2),
    ("root-high-degree", "1e401", 401, 2),
    ("root-high-degree-negative", "-1e401", 401, 2),
    ("root-high-degree-irrational", "2", 500, 50),
    ("cube-root-sixty-places", "2", 3, 60),
]:
    add(name, "root", text, prec, degree)

for op, text, prec in [
    ("ln", "0.5", 20), ("ln", "10", 60),
    ("ln", "1.0000000000000000000000000000000000000001", 50),
    ("ln", "0.9999999999999999999999999999999999999999", 50),
    ("log2", "0.3", 20), ("log2", "3", 60),
    ("log10", "0.999", 20), ("log10", "2", 60),
    ("exp", "-1", 60), ("exp", "0.1", 20), ("exp", "10", 20),
]:
    add(f"{op}-{text}-{prec}", op, text, prec)

# Inputs on both sides of p=0 rounding transitions. Keep the source decimals
# modest in size so exact-rational certification stays cheap enough for CI.
with localcontext() as ctx:
    ctx.prec = 120
    for op, boundary in [("ln", Decimal("0.5").exp()),
                         ("log2", Decimal(2).sqrt()),
                         ("log10", Decimal(10).sqrt()),
                         ("exp", Decimal("1.5").ln())]:
        for sign in [-1, 1]:
            text = format(boundary + sign * Decimal("1e-35"), ".40f")
            add(f"{op}-rounding-transition-{sign}", op, text, 0)

add("log2-large-dyadic", "log2", str(2 ** 1000), 0, exact="1000")
add("exp-negative-huge", "exp", "-1e1000", 1000, exact="0")
add("ln-one-precision-boundary", "ln", "1", 131072, exact="0")
add("root-zero-precision-boundary", "root", "0", 131072, degree=3, exact="0")
add("exp-zero-precision-boundary", "exp", "0", 131072, exact="1")

document = dict(version=1, target_release="unreleased", cases=cases)
Path(__file__).with_name("must_succeed.json").write_text(
    json.dumps(document, indent=2) + "\n", encoding="utf-8"
)
