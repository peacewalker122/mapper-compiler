package plugin

import (
	"os"
	"path/filepath"
	"testing"

	protocol "github.com/peacewalker122/mapper/protocol/generator/v1"
)

func TestWriteOutputs(t *testing.T) {
	root := t.TempDir()
	if err := WriteOutputs([]Output{{Root: root, Files: []protocol.Artifact{{Path: "nested/generated.go", Content: "package generated\n"}}}}); err != nil {
		t.Fatalf("WriteOutputs: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "nested", "generated.go"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "package generated\n" {
		t.Fatalf("content = %q", data)
	}
}
