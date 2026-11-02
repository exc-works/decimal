package decimal

import (
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"
)

type mustSucceedCase struct {
	ID      string `json:"id"`
	Op      string `json:"op"`
	Input   string `json:"input"`
	Prec    int    `json:"prec"`
	Want    string `json:"want"`
	Degree  int64  `json:"degree,omitempty"`
	Default bool   `json:"default,omitempty"`
}

func TestMustSucceedCorpus(t *testing.T) {
	data, err := os.ReadFile("testdata/must_succeed.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Version       int               `json:"version"`
		TargetRelease string            `json:"target_release"`
		Cases         []mustSucceedCase `json:"cases"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Version != 1 || corpus.TargetRelease == "" || len(corpus.Cases) < 45 {
		t.Fatal("missing or unexpected must-succeed corpus version")
	}
	seen := make(map[string]bool)
	ops := make(map[string]bool)
	for _, tc := range corpus.Cases {
		if tc.ID == "" || seen[tc.ID] {
			t.Fatalf("missing or duplicate case id %q", tc.ID)
		}
		seen[tc.ID], ops[tc.Op] = true, true
		t.Run(tc.ID, func(t *testing.T) {
			input, err := NewFromString(tc.Input)
			if err != nil {
				t.Fatal(err)
			}
			before := captureDecimalState(input)
			var got Decimal
			switch tc.Op {
			case "sqrt":
				if tc.Default {
					got, err = input.Sqrt()
				} else {
					got, err = input.SqrtWithPrec(tc.Prec)
				}
			case "root":
				if tc.Default {
					got, err = input.ApproxRoot(tc.Degree)
				} else {
					got, err = input.ApproxRootWithPrec(tc.Degree, tc.Prec)
				}
			case "ln":
				if tc.Default {
					got, err = input.Ln()
				} else {
					got, err = input.LnWithPrec(tc.Prec)
				}
			case "log2":
				if tc.Default {
					got = input.Log2()
				} else {
					got, err = input.Log2WithPrec(tc.Prec)
				}
			case "log10":
				if tc.Default {
					got, err = input.Log10()
				} else {
					got, err = input.Log10WithPrec(tc.Prec)
				}
			case "exp":
				if tc.Default {
					got, err = input.Exp()
				} else {
					got, err = input.ExpWithPrec(tc.Prec)
				}
			default:
				t.Fatalf("unknown must-succeed operation %q", tc.Op)
			}
			if err != nil {
				t.Fatalf("must-succeed case returned an error: %v", err)
			}
			if got.prec != tc.Prec || got.i == nil || captureDecimalState(input) != before {
				t.Fatal("wrong output scale, uninitialized result, or mutated input")
			}
			want, ok := new(big.Rat).SetString(tc.Want)
			if !ok {
				t.Fatal("invalid reference text")
			}
			want.Mul(want, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tc.Prec)), nil)))
			if !want.IsInt() || got.i.Cmp(want.Num()) != 0 {
				t.Fatalf("must-succeed coefficient differs from reference at scale %d", tc.Prec)
			}
			x, ok := new(big.Rat).SetString(tc.Input)
			if !ok {
				t.Fatal("invalid independent input")
			}
			if tc.Op == "sqrt" || tc.Op == "root" {
				degree := tc.Degree
				if tc.Op == "sqrt" {
					degree = 2
				}
				certifyRootRounding(t, x, degree, tc.Prec, got)
			} else {
				if tc.Op != "exp" && x.Sign() <= 0 {
					t.Fatal("invalid logarithm case")
				}
				certified := independentlyRoundedTranscendental(t, tc.Op, x, tc.Prec)
				if got.i.Cmp(certified) != 0 {
					t.Fatalf("result/reference disagrees with independent certified interval at scale %d", tc.Prec)
				}
			}
		})
	}
	for _, op := range []string{"sqrt", "root", "ln", "log2", "log10", "exp"} {
		if !ops[op] {
			t.Errorf("must-succeed corpus lost %s coverage", op)
		}
	}
	for _, id := range []string{"root-high-degree", "root-high-degree-negative", "root-high-degree-irrational", "exp-negative-huge", "ln-one-precision-boundary", "root-zero-precision-boundary", "exp-zero-precision-boundary", "log2-large-dyadic"} {
		if !seen[id] {
			t.Errorf("must-succeed corpus lost committed regression %s", id)
		}
	}
}
