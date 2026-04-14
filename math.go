package decimal

func Max(a, b Decimal) Decimal {
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

func Min(a, b Decimal) Decimal {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}

func Between(v, lower, upper Decimal) bool {
	return v.Cmp(lower) >= 0 && v.Cmp(upper) <= 0
}
