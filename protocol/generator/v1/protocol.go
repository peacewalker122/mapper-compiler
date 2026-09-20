package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/peacewalker122/mapper/ir"
)

const Version = 1

type Request struct {
	Protocol int            `json:"protocol"`
	Schema   Schema         `json:"schema"`
	Options  map[string]any `json:"options,omitempty"`
}

type Response struct {
	Protocol int        `json:"protocol"`
	Files    []Artifact `json:"files"`
}

type Artifact struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Schema struct {
	Version uint32 `json:"version"`
	Model   Model  `json:"model"`
}

type Model struct {
	ID     uint64  `json:"id"`
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func NewRequest(schema ir.Schema, options map[string]any) Request {
	return Request{Protocol: Version, Schema: fromIR(schema), Options: options}
}

func (r Request) ToIR() (ir.Schema, error) {
	fields := make([]ir.Field, 0, len(r.Schema.Model.Fields))
	for _, field := range r.Schema.Model.Fields {
		fieldType := ir.FieldType(field.Type)
		switch fieldType {
		case ir.TypeString, ir.TypeInteger, ir.TypeDecimal, ir.TypeBoolean, ir.TypeDateTime:
		default:
			return ir.Schema{}, fmt.Errorf("unsupported field type %q", field.Type)
		}
		fields = append(fields, ir.Field{
			ID:       field.ID,
			Name:     field.Name,
			Type:     fieldType,
			Required: field.Required,
		})
	}
	return ir.Schema{
		Version: r.Schema.Version,
		Model: ir.Model{
			ID:     r.Schema.Model.ID,
			Name:   r.Schema.Model.Name,
			Fields: fields,
		},
	}, nil
}

func EncodeRequest(w io.Writer, request Request) error {
	return encode(w, request)
}

func DecodeRequest(r io.Reader) (Request, error) {
	var request Request
	if err := decode(r, &request); err != nil {
		return Request{}, err
	}
	if request.Protocol != Version {
		return Request{}, fmt.Errorf("unsupported generator protocol %d", request.Protocol)
	}
	return request, nil
}

func EncodeResponse(w io.Writer, response Response) error {
	if response.Protocol == 0 {
		response.Protocol = Version
	}
	return encode(w, response)
}

func DecodeResponse(r io.Reader) (Response, error) {
	var response Response
	if err := decode(r, &response); err != nil {
		return Response{}, err
	}
	if response.Protocol != Version {
		return Response{}, fmt.Errorf("unsupported generator protocol %d", response.Protocol)
	}
	return response, nil
}

func fromIR(schema ir.Schema) Schema {
	fields := make([]Field, 0, len(schema.Model.Fields))
	for _, field := range schema.Model.Fields {
		fields = append(fields, Field{
			ID:       field.ID,
			Name:     field.Name,
			Type:     string(field.Type),
			Required: field.Required,
		})
	}
	return Schema{
		Version: schema.Version,
		Model: Model{
			ID:     schema.Model.ID,
			Name:   schema.Model.Name,
			Fields: fields,
		},
	}
}

func encode(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func decode(r io.Reader, value any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
