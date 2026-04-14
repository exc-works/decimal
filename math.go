package decimal

// Max returns the greater of a and b.
func Max(a, b Decimal) Decimal {
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

// Min returns the smaller of a and b.
func Min(a, b Decimal) Decimal {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}

// Between reports whether v is within the inclusive range [lower, upper].
func Between(v, lower, upper Decimal) bool {
	return v.Cmp(lower) >= 0 && v.Cmp(upper) <= 0
}
