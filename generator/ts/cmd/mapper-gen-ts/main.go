package main

import (
	"fmt"
	"os"

	typescript "github.com/peacewalker122/mapper/generator/ts"
	protocol "github.com/peacewalker122/mapper/protocol/generator/v1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	request, err := protocol.DecodeRequest(os.Stdin)
	if err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	schema, err := request.ToIR()
	if err != nil {
		return fmt.Errorf("decode schema: %w", err)
	}
	opts := typescript.Options{WithSchema: true}
	if request.Options != nil {
		if value, ok := request.Options["importFrom"]; ok {
			importFrom, valid := value.(string)
			if !valid {
				return fmt.Errorf("option importFrom must be a string")
			}
			opts.ImportFrom = importFrom
		} else {
			opts.ImportFrom = "@mapper/client"
		}
		if value, ok := request.Options["withSchema"]; ok {
			withSchema, valid := value.(bool)
			if !valid {
				return fmt.Errorf("option withSchema must be a boolean")
			}
			opts.WithSchema = withSchema
		}
	} else {
		opts.ImportFrom = "@mapper/client"
	}
	content, err := typescript.Generate(schema, opts)
	if err != nil {
		return err
	}
	return protocol.EncodeResponse(os.Stdout, protocol.Response{
		Protocol: protocol.Version,
		Files: []protocol.Artifact{{
			Path:    schema.Model.Name + ".gen.ts",
			Content: string(content),
		}},
	})
}
