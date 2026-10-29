"""Regenerate independent decimal references with Python's standard library.

Run from the repository root: python3 testdata/generate_transcendental.py
CI consumes the checked-in JSON and does not require Python.
"""

import json
from decimal import Decimal, ROUND_CEILING, ROUND_FLOOR, localcontext
from pathlib import Path


def calculate(operation, value):
    if operation == "ln":
        return value.ln()
    if operation == "log2":
        return value.ln() / Decimal(2).ln()
    if operation == "log10":
        return value.log10()
    return value.exp()


cases = []


def add(operation, text, precision):
    value = calculate(operation, Decimal(text))
    rounded = value.quantize(Decimal(1).scaleb(-precision))
    cases.append(dict(op=operation, input=text, prec=precision, want=str(rounded)))


with localcontext() as context:
    context.prec = 1500
    for operation, inputs in [
        ("ln", ["2", "3", "0.2", "1e-50"]),
        ("log2", ["3", "1.9", "0.3", "33", "2.12345678"]),
        ("log10", ["2", "1e100", "0.1"]),
        ("exp", ["1", "-1", "100", "-100", "0.5"]),
    ]:
        for text in inputs:
            for precision in [0, 1, 18, 30, 100]:
                add(operation, text, precision)
    for operation, text, precision in [
        ("ln", "2", 1000),
        ("log2", "3", 1000),
        ("log10", "2", 1000),
        ("exp", "1", 1000),
        ("exp", "1000", 0),
    ]:
        add(operation, text, precision)
    for operation, threshold in [
        ("exp", Decimal("1.5").ln()),
        ("ln", Decimal("0.5").exp()),
        ("log2", Decimal(2).sqrt()),
        ("log10", Decimal(10).sqrt()),
    ]:
        for rounding in [ROUND_FLOOR, ROUND_CEILING]:
            text = str(threshold.quantize(Decimal("1e-100"), rounding=rounding))
            add(operation, text, 0)

Path(__file__).with_name("transcendental.json").write_text(
    json.dumps(cases, indent=2) + "\n", encoding="utf-8"
)
