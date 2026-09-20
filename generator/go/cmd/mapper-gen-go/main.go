package main

import (
	"fmt"
	"os"

	golang "github.com/peacewalker122/mapper/generator/go"
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
	pkg := "generated"
	if value, ok := request.Options["package"]; ok {
		var valid bool
		pkg, valid = value.(string)
		if !valid || pkg == "" {
			return fmt.Errorf("option package must be a non-empty string")
		}
	}
	content, err := golang.Generate(schema, pkg)
	if err != nil {
		return err
	}
	return protocol.EncodeResponse(os.Stdout, protocol.Response{
		Protocol: protocol.Version,
		Files: []protocol.Artifact{{
			Path:    schema.Model.Name + ".gen.go",
			Content: string(content),
		}},
	})
}
