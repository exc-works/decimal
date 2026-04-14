package decimal

import (
	"fmt"
)

func ExampleNewFromString() {
	d, err := NewFromString("00123.4500")
	if err != nil {
		panic(err)
	}

	fmt.Println(d.String())
	fmt.Println(d.StringWithTrailingZeros())

	// Output:
	// 123.45
	// 123.4500
}

func ExampleNewFromString_negative() {
	d, err := NewFromString("-00123.4500")
	if err != nil {
		panic(err)
	}

	fmt.Println(d.String())
	fmt.Println(d.StringWithTrailingZeros())

	// Output:
	// -123.45
	// -123.4500
}

func ExampleNewFromString_scientificNotation() {
	d1, err := NewFromString("1.23456e3")
	if err != nil {
		panic(err)
	}
	d2, err := NewFromString("-4.56E-2")
	if err != nil {
		panic(err)
	}

	fmt.Println(d1.String())
	fmt.Println(d2.String())
	fmt.Println(d2.StringWithTrailingZeros())

	// Output:
	// 1234.56
	// -0.0456
	// -0.0456
}

func ExampleDecimal_Add() {
	sum := MustFromString("1.20").Add(MustFromString("2.34"))
	rounded := MustFromString("2.555").Rescale(2, RoundHalfEven)

	fmt.Println(sum.String())
	fmt.Println(rounded.String())

	// Output:
	// 3.54
	// 2.56
}

func ExampleDecimal_Rescale() {
	d := MustFromString("7.5000").Rescale(2, RoundDown)
	bz, err := d.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(d.StringWithTrailingZeros())
	fmt.Println(string(bz))

	// Output:
	// 7.50
	// "7.5"
}
