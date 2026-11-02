"""Generate independent high-degree root references; run from repository root."""

import json
from decimal import Decimal, localcontext
from pathlib import Path

cases = []
with localcontext() as context:
    context.prec = 350
    for text, degree, precision in [
        ("2", 500, 50), ("2", 500, 100), ("2", 2000, 30),
        ("1e401", 401, 2), ("-1e401", 401, 2),
        ("1.1", 1000, 50), ("0.5", 1000, 50),
        ("1e-100", 401, 30), ("2", 131072, 0),
    ]:
        value = Decimal(text)
        root = (value.copy_abs().ln() / degree).exp()
        if value < 0:
            root = -root
        rounded = root.quantize(Decimal(1).scaleb(-precision))
        cases.append(dict(input=text, degree=degree, prec=precision, want=str(rounded)))

Path(__file__).with_name("roots.json").write_text(json.dumps(cases, indent=2) + "\n")
