package decimal

import (
	"encoding/json"
	"errors"
	"math/big"
	"math/rand/v2"
	"os"
	"testing"
)

func TestHighDegreeRootReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/roots.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input, Want string
		Degree      int64
		Prec        int
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		got, err := MustFromString(tc.Input).ApproxRootWithPrec(tc.Degree, tc.Prec)
		if err != nil || got.prec != tc.Prec || !got.Equal(MustFromString(tc.Want)) {
			t.Fatalf("root(%s,%d), precision %d: %s, %v; want %s", tc.Input, tc.Degree, tc.Prec, got, err, tc.Want)
		}
	}
	x := new(big.Int).Lsh(big.NewInt(1), 10000)
	got, err := NewFromBigInt(x).ApproxRootWithPrec(1000, 0)
	if err != nil || !got.Equal(New(1024)) {
		t.Fatalf("exact high-degree root: %s %v", got, err)
	}
}

func TestIntegerRootExactCertificate(t *testing.T) {
	rng := rand.New(rand.NewPCG(789, 123))
	for _, degree := range []int64{3, 7, 19, 401, 500, 2000} {
		for range 20 {
			var raw [64]byte
			for i := range raw {
				raw[i] = byte(rng.Uint64())
			}
			x := new(big.Int).SetBytes(raw[:])
			before := new(big.Int).Set(x)
			got, ok := integerRoot(x, degree)
			if !ok {
				t.Fatalf("degree %d did not converge", degree)
			}
			power := new(big.Int).Exp(got, big.NewInt(degree), nil)
			next := new(big.Int).Add(got, big.NewInt(1))
			next.Exp(next, big.NewInt(degree), nil)
			if power.Cmp(x) > 0 || next.Cmp(x) <= 0 || x.Cmp(before) != 0 {
				t.Fatalf("degree %d: invalid floor certificate", degree)
			}
		}
	}
}

func TestHighDegreeMidpointRounding(t *testing.T) {
	for _, degree := range []int64{401, 500, 2000} {
		for _, coefficient := range []int64{5, 15, 25} {
			midpoint := NewWithPrec(coefficient, 1)
			x := Decimal{i: new(big.Int).Exp(midpoint.i, big.NewInt(degree), nil), prec: int(degree)}
			got, err := x.ApproxRootWithPrec(degree, 0)
			want := midpoint.Round()
			if err != nil || !got.Equal(want) {
				t.Fatalf("degree %d midpoint: %s %v, want %s", degree, got, err, want)
			}
		}
	}
	if _, err := New(2).ApproxRootWithPrec(2000, 100); !errors.Is(err, ErrNoConvergence) {
		t.Fatalf("scale budget was bypassed: %v", err)
	}
}

func BenchmarkHighDegreeRoot(b *testing.B) {
	input := New(2)
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchDecimalResult, err = input.ApproxRootWithPrec(500, 50)
		if err != nil {
			b.Fatal(err)
		}
	}
}
