package v1

import (
	"bytes"
	"testing"

	"github.com/peacewalker122/mapper/ir"
)

func TestRequestRoundTrip(t *testing.T) {
	request := NewRequest(ir.Schema{
		Version: 1,
		Model: ir.Model{
			ID:   1001,
			Name: "subscriber",
			Fields: []ir.Field{
				{ID: 1002, Name: "msisdn", Type: ir.TypeString, Required: true},
			},
		},
	}, map[string]any{"package": "mapping"})
	var encoded bytes.Buffer
	if err := EncodeRequest(&encoded, request); err != nil {
		t.Fatalf("EncodeRequest: %v", err)
	}
	decoded, err := DecodeRequest(&encoded)
	if err != nil {
		t.Fatalf("DecodeRequest: %v", err)
	}
	got, err := decoded.ToIR()
	if err != nil {
		t.Fatalf("ToIR: %v", err)
	}
	if got.Model.ID != 1001 || got.Model.Fields[0].ID != 1002 || decoded.Options["package"] != "mapping" {
		t.Fatalf("decoded request = %#v", decoded)
	}
}
