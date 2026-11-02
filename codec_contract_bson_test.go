//go:build bson

package decimal

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestContractBSONValidZeroAndFailureState(t *testing.T) {
	n := NewNullDecimal(Decimal{})
	typ, data, err := n.MarshalBSONValue()
	text, ok := (bson.RawValue{Type: bson.Type(typ), Value: data}).StringValueOK()
	if err != nil || !ok || text != "0" || !n.Valid || !n.Decimal.IsNil() {
		t.Fatalf("valid BSON zero: type=%v text=%q valid=%v error=%v", typ, text, n.Valid, err)
	}
	for _, initial := range []Decimal{Decimal{}, MustFromString("1.20")} {
		for _, valid := range []bool{false, true} {
			n = NullDecimal{Decimal: initial.Clone(), Valid: valid}
			before := captureDecimalState(n.Decimal)
			typ, data, err := bson.MarshalValue("bad")
			if err != nil {
				t.Fatal(err)
			}
			err = n.UnmarshalBSONValue(byte(typ), data)
			if !errors.Is(err, ErrUnmarshal) || !errors.Is(err, ErrInvalidFormat) || n.Valid != valid || captureDecimalState(n.Decimal) != before {
				t.Fatalf("BSON failure changed state or lost cause: %v", err)
			}
		}
	}
}
