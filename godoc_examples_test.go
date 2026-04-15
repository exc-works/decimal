package decimal

import (
	"fmt"
	"math/big"
)

func ExampleNew() {
	fmt.Println(New(42))
	// Output:
	// 42
}

func ExampleNewFromInt() {
	fmt.Println(NewFromInt(42))
	// Output:
	// 42
}

func ExampleNewWithPrec() {
	fmt.Println(NewWithPrec(1234, 2))
	// Output:
	// 12.34
}

func ExampleNewFromFloat64() {
	fmt.Println(NewFromFloat64(12.34))
	// Output:
	// 12.34
}

func ExampleNewFromFloat32() {
	fmt.Println(NewFromFloat32(12.34))
	// Output:
	// 12.34
}

func ExampleNewWithAppendPrec() {
	fmt.Println(NewWithAppendPrec(12, 3).StringWithTrailingZeros())
	// Output:
	// 12.000
}

func ExampleNewFromUintWithAppendPrec() {
	fmt.Println(NewFromUintWithAppendPrec(12, 3).StringWithTrailingZeros())
	// Output:
	// 12.000
}

func ExampleNewFromBigInt() {
	fmt.Println(NewFromBigInt(big.NewInt(123)))
	// Output:
	// 123
}

func ExampleNewFromBigRat() {
	d, err := NewFromBigRat(big.NewRat(7, 4))
	fmt.Println(d, err == nil)
	// Output:
	// 1.75 true
}

func ExampleNewFromBigRatWithPrec() {
	d, err := NewFromBigRatWithPrec(big.NewRat(1, 3), 2, RoundHalfEven)
	fmt.Println(d, err == nil)
	// Output:
	// 0.33 true
}

func ExampleNewFromBigIntWithPrec() {
	fmt.Println(NewFromBigIntWithPrec(big.NewInt(12345), 2))
	// Output:
	// 123.45
}

func ExampleNewFromInt64() {
	fmt.Println(NewFromInt64(12345, 2))
	// Output:
	// 123.45
}

func ExampleNewFromUint64() {
	fmt.Println(NewFromUint64(12345, 2))
	// Output:
	// 123.45
}

func ExampleMustFromString() {
	fmt.Println(MustFromString("12.34"))
	// Output:
	// 12.34
}

func ExampleMax() {
	fmt.Println(Max(New(1), New(2)))
	// Output:
	// 2
}

func ExampleMin() {
	fmt.Println(Min(New(1), New(2)))
	// Output:
	// 1
}

func ExampleBetween() {
	fmt.Println(Between(New(5), New(1), New(10)))
	// Output:
	// true
}

func ExampleMostSignificantBit() {
	fmt.Println(MostSignificantBit(big.NewInt(16)))
	// Output:
	// 4
}

func ExampleDecimal_SafeAdd() {
	fmt.Println(New(2).SafeAdd(New(3)))
	// Output:
	// 5
}

func ExampleDecimal_AddRaw() {
	fmt.Println(MustFromString("1.23").AddRaw(1))
	// Output:
	// 1.24
}

func ExampleDecimal_Add_differentPrecision() {
	r := MustFromString("1.2").Add(MustFromString("0.030"))
	fmt.Println(r.String(), r.Precision())
	fmt.Println(r.StringWithTrailingZeros())
	// Output:
	// 1.23 3
	// 1.230
}

func ExampleDecimal_Sub() {
	fmt.Println(MustFromString("5.5").Sub(MustFromString("2.2")))
	// Output:
	// 3.3
}

func ExampleDecimal_SafeSub() {
	fmt.Println(New(5).SafeSub(New(2)))
	// Output:
	// 3
}

func ExampleDecimal_SubRaw() {
	fmt.Println(MustFromString("5.5").SubRaw(1))
	// Output:
	// 5.4
}

func ExampleDecimal_Sub_differentPrecision() {
	r := MustFromString("5.00").Sub(MustFromString("0.125"))
	fmt.Println(r.String(), r.Precision())
	// Output:
	// 4.875 3
}

func ExampleDecimal_Mul() {
	fmt.Println(MustFromString("1.25").Mul(MustFromString("2.00"), RoundHalfEven))
	// Output:
	// 2.5
}

func ExampleDecimal_MulDown() {
	fmt.Println(MustFromString("1.25").MulDown(MustFromString("2.00")))
	// Output:
	// 2.5
}

func ExampleDecimal_Mul2() {
	fmt.Println(MustFromString("1.20").Mul2(MustFromString("2.30")).StringWithTrailingZeros())
	// Output:
	// 2.7600
}

func ExampleDecimal_Mul_differentPrecision() {
	r := MustFromString("1.234").Mul(MustFromString("2.5"), RoundHalfEven)
	fmt.Println(r.String(), r.Precision())
	// Output:
	// 3.085 3
}

func ExampleDecimal_QuoWithPrec() {
	fmt.Println(New(1).QuoWithPrec(New(3), 6, RoundHalfEven))
	// Output:
	// 0.333333
}

func ExampleDecimal_Quo() {
	fmt.Println(New(7).Quo(New(2), RoundDown))
	// Output:
	// 3
}

func ExampleDecimal_QuoDown() {
	fmt.Println(New(7).QuoDown(New(2)))
	// Output:
	// 3
}

func ExampleDecimal_Quo_differentPrecision() {
	r := MustFromString("12.3").Quo(MustFromString("0.20"), RoundHalfEven)
	fmt.Println(r.String(), r.Precision())
	fmt.Println(r.StringWithTrailingZeros())
	// Output:
	// 61.5 2
	// 61.50
}

func ExampleDecimal_Floor() {
	fmt.Println(MustFromString("-1.2").Floor())
	// Output:
	// -2
}

func ExampleDecimal_Ceil() {
	fmt.Println(MustFromString("-1.2").Ceil())
	// Output:
	// -1
}

func ExampleDecimal_Truncate() {
	fmt.Println(MustFromString("-1.9").Truncate())
	// Output:
	// -1
}

func ExampleDecimal_Round() {
	fmt.Println(MustFromString("2.5").Round())
	fmt.Println(MustFromString("3.5").Round())
	// Output:
	// 2
	// 4
}

func ExampleDecimal_FloorWithPrec() {
	fmt.Println(MustFromString("-1.239").FloorWithPrec(2))
	// Output:
	// -1.24
}

func ExampleDecimal_CeilWithPrec() {
	fmt.Println(MustFromString("-1.239").CeilWithPrec(2))
	// Output:
	// -1.23
}

func ExampleDecimal_TruncateWithPrec() {
	fmt.Println(MustFromString("-1.239").TruncateWithPrec(2))
	// Output:
	// -1.23
}

func ExampleDecimal_RoundWithPrec() {
	fmt.Println(MustFromString("1.245").RoundWithPrec(2))
	// Output:
	// 1.24
}

func ExampleDecimal_IntPart() {
	fmt.Println(MustFromString("12.34").IntPart())
	// Output:
	// 12
}

func ExampleDecimal_Remainder() {
	i, f := MustFromString("12.34").Remainder()
	fmt.Println(i, f)
	// Output:
	// 12 34
}

func ExampleDecimal_IsInteger() {
	fmt.Println(MustFromString("1.000").IsInteger())
	// Output:
	// true
}

func ExampleDecimal_HasFraction() {
	fmt.Println(MustFromString("1.25").HasFraction())
	// Output:
	// true
}

func ExampleDecimal_QuoRem() {
	q, r := MustFromString("-7").QuoRem(New(3))
	fmt.Println(q, r)
	// Output:
	// -2 -1
}

func ExampleDecimal_Mod() {
	fmt.Println(MustFromString("-7").Mod(New(3)))
	// Output:
	// -1
}

func ExampleDecimal_Power() {
	fmt.Println(MustFromString("1.5").Power(3))
	// Output:
	// 3.4
}

func ExampleDecimal_Sqrt() {
	v, err := MustFromString("16.0").Sqrt()
	fmt.Println(v, err == nil)
	// Output:
	// 4 true
}

func ExampleDecimal_ApproxRoot() {
	v, err := MustFromString("3125.0000").ApproxRoot(5)
	fmt.Println(v.StringWithTrailingZeros(), err == nil)
	// Output:
	// 5.0000 true
}

func ExampleDecimal_Log2() {
	fmt.Println(New(8).Log2())
	// Output:
	// 3
}

func ExampleDecimal_RescaleDown() {
	fmt.Println(MustFromString("1.29").RescaleDown(1))
	// Output:
	// 1.2
}

func ExampleDecimal_StripTrailingZeros() {
	fmt.Println(MustFromString("1.2300").StripTrailingZeros())
	// Output:
	// 1.23
}

func ExampleDecimal_SignificantFigures() {
	fmt.Println(MustFromString("123.456").SignificantFigures(4, RoundHalfEven))
	fmt.Println(MustFromString("123.456").SignificantFigures(3, RoundHalfEven))
	fmt.Println(MustFromString("123.456").SignificantFigures(2, RoundHalfEven))
	// Output:
	// 123.5
	// 123
	// 120
}

func ExampleDecimal_MustNonNegative() {
	fmt.Println(New(1).MustNonNegative())
	// Output:
	// 1
}

func ExampleDecimal_Float64() {
	v, exact := MustFromString("0.5").Float64()
	fmt.Println(v, exact)
	// Output:
	// 0.5 true
}

func ExampleDecimal_Float32() {
	v, exact := MustFromString("0.5").Float32()
	fmt.Println(v, exact)
	// Output:
	// 0.5 true
}

func ExampleDecimal_Int64() {
	v, ok := MustFromString("42.0").Int64()
	fmt.Println(v, ok)
	// Output:
	// 42 true
}

func ExampleDecimal_Uint64() {
	v, ok := MustFromString("42.0").Uint64()
	fmt.Println(v, ok)
	// Output:
	// 42 true
}

func ExampleDecimal_Cmp() {
	fmt.Println(New(1).Cmp(New(2)))
	// Output:
	// -1
}

func ExampleDecimal_Equal() {
	fmt.Println(New(1).Equal(New(1)))
	// Output:
	// true
}

func ExampleDecimal_NotEqual() {
	fmt.Println(New(1).NotEqual(New(2)))
	// Output:
	// true
}

func ExampleDecimal_GT() {
	fmt.Println(New(2).GT(New(1)))
	// Output:
	// true
}

func ExampleDecimal_GTE() {
	fmt.Println(New(2).GTE(New(2)))
	// Output:
	// true
}

func ExampleDecimal_LT() {
	fmt.Println(New(1).LT(New(2)))
	// Output:
	// true
}

func ExampleDecimal_LTE() {
	fmt.Println(New(1).LTE(New(1)))
	// Output:
	// true
}

func ExampleDecimal_Sign() {
	fmt.Println(New(-1).Sign())
	// Output:
	// -1
}

func ExampleDecimal_IsNegative() {
	fmt.Println(New(-1).IsNegative())
	// Output:
	// true
}

func ExampleDecimal_IsNil() {
	var d Decimal
	fmt.Println(d.IsNil())
	// Output:
	// true
}

func ExampleDecimal_IsZero() {
	fmt.Println(New(0).IsZero())
	// Output:
	// true
}

func ExampleDecimal_IsNotZero() {
	fmt.Println(New(1).IsNotZero())
	// Output:
	// true
}

func ExampleDecimal_IsPositive() {
	fmt.Println(New(1).IsPositive())
	// Output:
	// true
}

func ExampleDecimal_Neg() {
	fmt.Println(New(1).Neg())
	// Output:
	// -1
}

func ExampleDecimal_Abs() {
	fmt.Println(New(-1).Abs())
	// Output:
	// 1
}

func ExampleDecimal_BigInt() {
	fmt.Println(MustFromString("12.34").BigInt())
	// Output:
	// 1234
}

func ExampleDecimal_BigRat() {
	fmt.Println(MustFromString("12.34").BigRat().RatString())
	// Output:
	// 617/50
}

func ExampleDecimal_BitLen() {
	fmt.Println(New(7).BitLen())
	// Output:
	// 3
}

func ExampleDecimal_Precision() {
	fmt.Println(MustFromString("1.23").Precision())
	// Output:
	// 2
}

func ExampleDecimal_Max() {
	fmt.Println(New(1).Max(New(2)))
	// Output:
	// 2
}

func ExampleDecimal_Min() {
	fmt.Println(New(1).Min(New(2)))
	// Output:
	// 1
}

func ExampleDecimal_StringWithTrailingZeros() {
	fmt.Println(MustFromString("1.2300").StringWithTrailingZeros())
	// Output:
	// 1.2300
}

func ExampleDecimal_String() {
	fmt.Println(MustFromString("1.2300").String())
	// Output:
	// 1.23
}

func ExampleDecimal_MarshalJSON() {
	bz, _ := MustFromString("1.23").MarshalJSON()
	fmt.Println(string(bz))
	// Output:
	// "1.23"
}

func ExampleDecimal_UnmarshalJSON() {
	var d Decimal
	_ = d.UnmarshalJSON([]byte(`"1.23"`))
	fmt.Println(d)
	// Output:
	// 1.23
}

func ExampleDecimal_MarshalYAML() {
	v, _ := MustFromString("1.23").MarshalYAML()
	fmt.Println(v)
	// Output:
	// 1.23
}

func ExampleDecimal_UnmarshalYAML() {
	var d Decimal
	_ = d.UnmarshalYAML(func(target any) error {
		p := target.(*any)
		*p = "1.23"
		return nil
	})
	fmt.Println(d)
	// Output:
	// 1.23
}

func ExampleDecimal_MarshalText() {
	bz, _ := MustFromString("1.23").MarshalText()
	fmt.Println(string(bz))
	// Output:
	// 1.23
}

func ExampleDecimal_UnmarshalText() {
	var d Decimal
	_ = d.UnmarshalText([]byte("1.23"))
	fmt.Println(d)
	// Output:
	// 1.23
}

func ExampleDecimal_MarshalBinary() {
	bz, err := MustFromString("1.23").MarshalBinary()
	fmt.Println(len(bz) > 0 && err == nil)
	// Output:
	// true
}

func ExampleDecimal_UnmarshalBinary() {
	src := MustFromString("1.23")
	bz, _ := src.MarshalBinary()
	var dst Decimal
	_ = dst.UnmarshalBinary(bz)
	fmt.Println(dst)
	// Output:
	// 1.23
}

func ExampleDecimal_Value() {
	v, _ := MustFromString("1.23").Value()
	fmt.Println(v)
	// Output:
	// 1.23
}

func ExampleDecimal_Scan() {
	var d Decimal
	_ = d.Scan("1.23")
	fmt.Println(d)
	// Output:
	// 1.23
}

func ExampleDecimal_Marshal() {
	bz, err := MustFromString("1.23").Marshal()
	fmt.Println(len(bz) > 0 && err == nil)
	// Output:
	// true
}

func ExampleDecimal_MarshalTo() {
	d := MustFromString("1.23")
	buf := make([]byte, d.Size())
	n, err := d.MarshalTo(buf)
	fmt.Println(n > 0 && err == nil)
	// Output:
	// true
}

func ExampleDecimal_Unmarshal() {
	src := MustFromString("1.23")
	bz, _ := src.Marshal()
	var dst Decimal
	_ = dst.Unmarshal(bz)
	fmt.Println(dst)
	// Output:
	// 1.23
}

func ExampleDecimal_Size() {
	fmt.Println(MustFromString("1.23").Size() > 0)
	// Output:
	// true
}
