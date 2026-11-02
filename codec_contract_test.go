package decimal

import (
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

type decimalState struct {
	coefficient string
	scale       int
	nilValue    bool
}

func captureDecimalState(d Decimal) decimalState {
	s := decimalState{scale: d.prec, nilValue: d.IsNil()}
	if d.i != nil {
		s.coefficient = d.i.String()
	}
	return s
}

func contractXMLDecode(text string, decode func(*xml.Decoder, xml.StartElement) error) error {
	dec := xml.NewDecoder(strings.NewReader(text))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	return decode(dec, token.(xml.StartElement))
}

func contractBadYAML(out any) error {
	if resolved, ok := out.(*any); ok {
		*resolved = "bad"
		return nil
	}
	return out.(interface{ UnmarshalText([]byte) error }).UnmarshalText([]byte("bad"))
}

func TestContractDecodeFailurePreservesStateAndCause(t *testing.T) {
	for _, tc := range []struct {
		name string
		d    func(*Decimal) error
		n    func(*NullDecimal) error
		text bool
	}{
		{"json text", func(d *Decimal) error { return d.UnmarshalJSON([]byte(`"bad"`)) }, func(n *NullDecimal) error { return n.UnmarshalJSON([]byte(`"bad"`)) }, true},
		{"json empty string", func(d *Decimal) error { return d.UnmarshalJSON([]byte(`""`)) }, func(n *NullDecimal) error { return n.UnmarshalJSON([]byte(`""`)) }, true},
		{"json syntax", func(d *Decimal) error { return d.UnmarshalJSON([]byte(`{"x":`)) }, func(n *NullDecimal) error { return n.UnmarshalJSON([]byte(`{"x":`)) }, false},
		{"json type", func(d *Decimal) error { return d.UnmarshalJSON([]byte(`[]`)) }, func(n *NullDecimal) error { return n.UnmarshalJSON([]byte(`[]`)) }, false},
		{"json number range", func(d *Decimal) error { return d.UnmarshalJSON([]byte(`1e2000000000`)) }, func(n *NullDecimal) error { return n.UnmarshalJSON([]byte(`1e2000000000`)) }, true},
		{"text", func(d *Decimal) error { return d.UnmarshalText([]byte("bad")) }, func(n *NullDecimal) error { return n.UnmarshalText([]byte("bad")) }, true},
		{"param", func(d *Decimal) error { return d.UnmarshalParam("bad") }, func(n *NullDecimal) error { return n.UnmarshalParam("bad") }, true},
		{"sql text", func(d *Decimal) error { return d.Scan("bad") }, func(n *NullDecimal) error { return n.Scan("bad") }, true},
		{"sql empty string", func(d *Decimal) error { return d.Scan("") }, func(n *NullDecimal) error { return n.Scan("") }, true},
		{"sql empty bytes", func(d *Decimal) error { return d.Scan([]byte{}) }, func(n *NullDecimal) error { return n.Scan([]byte{}) }, true},
		{"sql nonfinite", func(d *Decimal) error { return d.Scan(math.Inf(1)) }, func(n *NullDecimal) error { return n.Scan(math.Inf(1)) }, false},
		{"yaml text", func(d *Decimal) error { return d.UnmarshalYAML(contractBadYAML) }, func(n *NullDecimal) error { return n.UnmarshalYAML(contractBadYAML) }, true},
		{"xml text", func(d *Decimal) error { return contractXMLDecode("<v>bad</v>", d.UnmarshalXML) }, func(n *NullDecimal) error { return contractXMLDecode("<v>bad</v>", n.UnmarshalXML) }, true},
		{"xml syntax", func(d *Decimal) error { return contractXMLDecode("<v>", d.UnmarshalXML) }, func(n *NullDecimal) error { return contractXMLDecode("<v>", n.UnmarshalXML) }, false},
		{"xml attr", func(d *Decimal) error { return d.UnmarshalXMLAttr(xml.Attr{Value: "bad"}) }, func(n *NullDecimal) error { return n.UnmarshalXMLAttr(xml.Attr{Value: "bad"}) }, true},
		{"binary prefix", func(d *Decimal) error { return d.UnmarshalBinary([]byte{0}) }, nil, false},
		{"binary coefficient", func(d *Decimal) error { return d.UnmarshalBinary([]byte{0, 0, 0, 0, 255}) }, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkError := func(err error) {
				t.Helper()
				if !errors.Is(err, ErrUnmarshal) || tc.text && !errors.Is(err, ErrInvalidFormat) {
					t.Fatalf("missing error category: %v", err)
				}
			}
			for _, initial := range []Decimal{Decimal{}, MustFromString("1.20"), NewWithPrec(0, 4)} {
				d := initial.Clone()
				before := captureDecimalState(d)
				checkError(tc.d(&d))
				if got := captureDecimalState(d); got != before {
					t.Fatalf("Decimal failure changed state: %+v -> %+v", before, got)
				}
				if tc.n == nil {
					continue
				}
				for _, valid := range []bool{false, true} {
					n := NullDecimal{Decimal: initial.Clone(), Valid: valid}
					before := captureDecimalState(n.Decimal)
					checkError(tc.n(&n))
					if got := captureDecimalState(n.Decimal); got != before || n.Valid != valid {
						t.Fatalf("NullDecimal failure changed state: %+v valid=%v -> %+v valid=%v", before, valid, got, n.Valid)
					}
				}
			}
		})
	}
}

func TestContractDecodePreservesTypedErrors(t *testing.T) {
	var d Decimal
	err := d.UnmarshalJSON([]byte(`{"x":`))
	var syntax *json.SyntaxError
	if !errors.Is(err, ErrUnmarshal) || !errors.As(err, &syntax) {
		t.Fatalf("JSON syntax cause lost: %v", err)
	}
	err = d.UnmarshalJSON([]byte(`[]`))
	var wrongType *json.UnmarshalTypeError
	if !errors.Is(err, ErrUnmarshal) || !errors.As(err, &wrongType) {
		t.Fatalf("JSON type cause lost: %v", err)
	}
	err = contractXMLDecode("<v>", d.UnmarshalXML)
	var xmlSyntax *xml.SyntaxError
	if !errors.Is(err, ErrUnmarshal) || !errors.As(err, &xmlSyntax) {
		t.Fatalf("XML syntax cause lost: %v", err)
	}
	cause := errors.New("callback failure")
	err = d.UnmarshalYAML(func(any) error { return cause })
	if !errors.Is(err, cause) || !errors.Is(err, ErrUnmarshal) {
		t.Fatalf("YAML callback cause lost: %v", err)
	}
}

func TestContractNullDecimalValidZeroEncodings(t *testing.T) {
	n := NewNullDecimal(Decimal{})
	before := captureDecimalState(n.Decimal)
	b, err := n.MarshalJSON()
	if err != nil || string(b) != `"0"` {
		t.Fatalf("JSON valid zero: %s, %v", b, err)
	}
	b, err = xml.Marshal(n)
	if err != nil || string(b) != "<NullDecimal>0</NullDecimal>" {
		t.Fatalf("XML valid zero: %s, %v", b, err)
	}
	a, err := n.MarshalXMLAttr(xml.Name{Local: "amount"})
	if err != nil || a.Value != "0" || a.Name.Local != "amount" {
		t.Fatalf("XML attribute valid zero: %+v, %v", a, err)
	}
	v, err := n.Value()
	if err != nil || v != "0" {
		t.Fatalf("SQL valid zero: %v, %v", v, err)
	}
	v, err = n.MarshalYAML()
	if err != nil || v != "0" {
		t.Fatalf("YAML valid zero: %v, %v", v, err)
	}
	b, err = n.MarshalText()
	if err != nil || string(b) != "0" {
		t.Fatalf("Text valid zero: %s, %v", b, err)
	}
	if got := captureDecimalState(n.Decimal); got != before || !n.Valid {
		t.Fatal("encoding mutated nullable state")
	}
}

func TestContractJSONWhitespaceNull(t *testing.T) {
	for _, input := range []string{"null", " \r\nnull\t "} {
		d := MustFromString("1.20")
		before := captureDecimalState(d)
		if err := d.UnmarshalJSON([]byte(input)); err != nil || captureDecimalState(d) != before {
			t.Fatalf("Decimal null changed state: %v", err)
		}
		n := NewNullDecimal(MustFromString("1.20"))
		if err := n.UnmarshalJSON([]byte(input)); err != nil || n.Valid || !n.Decimal.IsNil() || n.Decimal.prec != 0 {
			t.Fatalf("NullDecimal null did not clear state: %v", err)
		}
	}
}

func TestContractHookErrorLayers(t *testing.T) {
	hook := DecodeHook()
	for _, data := range []any{"", []byte{}, "bad", []byte("bad"), json.Number("bad")} {
		for _, target := range []reflect.Type{reflect.TypeFor[Decimal](), reflect.TypeFor[NullDecimal]()} {
			got, err := hook(reflect.TypeOf(data), target, data)
			if target == reflect.TypeFor[NullDecimal]() && (data == "" || reflect.TypeOf(data) == reflect.TypeFor[[]byte]() && len(data.([]byte)) == 0) {
				if err != nil || got.(NullDecimal).Valid || !got.(NullDecimal).Decimal.IsNil() {
					t.Fatalf("hook empty nullable: %v, %v", got, err)
				}
				continue
			}
			if !errors.Is(err, ErrInvalidFormat) || !errors.Is(err, ErrUnmarshal) {
				t.Fatalf("hook error categories: %v", err)
			}
		}
	}
}

func TestContractNullInputPolicies(t *testing.T) {
	for _, tc := range []struct {
		name string
		d    func(*Decimal) error
		n    func(*NullDecimal) error
		keep bool
	}{
		{"json", func(d *Decimal) error { return d.UnmarshalJSON([]byte("null")) }, func(n *NullDecimal) error { return n.UnmarshalJSON([]byte("null")) }, true},
		{"yaml callback", func(d *Decimal) error { return d.UnmarshalYAML(yamlTestDecoder(nil, nil)) }, func(n *NullDecimal) error { return n.UnmarshalYAML(yamlTestDecoder(nil, nil)) }, true},
		{"sql", func(d *Decimal) error { return d.Scan(nil) }, func(n *NullDecimal) error { return n.Scan(nil) }, false},
		{"xml element", func(d *Decimal) error { return contractXMLDecode("<v> </v>", d.UnmarshalXML) }, func(n *NullDecimal) error { return contractXMLDecode("<v> </v>", n.UnmarshalXML) }, false},
		{"xml attr", func(d *Decimal) error { return d.UnmarshalXMLAttr(xml.Attr{Value: " "}) }, func(n *NullDecimal) error { return n.UnmarshalXMLAttr(xml.Attr{Value: " "}) }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := MustFromString("1.20")
			before := captureDecimalState(d)
			if err := tc.d(&d); err != nil {
				t.Fatal(err)
			}
			if tc.keep && captureDecimalState(d) != before || !tc.keep && (!d.IsNil() || d.prec != 0) {
				t.Fatal("wrong Decimal null policy")
			}
			n := NewNullDecimal(MustFromString("1.20"))
			if err := tc.n(&n); err != nil || n.Valid || !n.Decimal.IsNil() || n.Decimal.prec != 0 {
				t.Fatalf("wrong nullable null policy: %v", err)
			}
		})
	}
	d := MustFromString("1.20")
	if err := d.UnmarshalBinary(nil); err != nil || d.IsNil() || !d.IsZero() || d.prec != 0 {
		t.Fatalf("binary empty input must initialize zero: %v", err)
	}
	for _, decode := range []func(*NullDecimal) error{
		func(n *NullDecimal) error { return n.UnmarshalText(nil) },
		func(n *NullDecimal) error { return n.UnmarshalParam("") },
	} {
		n := NewNullDecimal(MustFromString("1.20"))
		if err := decode(&n); err != nil || n.Valid || !n.Decimal.IsNil() || n.Decimal.prec != 0 {
			t.Fatalf("empty nullable text must clear: %v", err)
		}
	}
}

func TestContractBinaryRejectsUnsignedScale(t *testing.T) {
	data, err := One.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	binary.BigEndian.PutUint32(data, math.MaxUint32)
	d := MustFromString("1.20")
	before := captureDecimalState(d)
	if err := d.UnmarshalBinary(data); !errors.Is(err, ErrUnmarshal) || captureDecimalState(d) != before {
		t.Fatalf("oversized unsigned scale must fail without changing state: %v", err)
	}
}
