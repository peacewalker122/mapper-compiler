package plugin

import (
	"path/filepath"
	"testing"
)

func TestValidateArtifactPath(t *testing.T) {
	for _, path := range []string{"generated.go", "nested/generated.ts"} {
		if err := ValidateArtifactPath(path); err != nil {
			t.Fatalf("ValidateArtifactPath(%q): %v", path, err)
		}
	}
	for _, path := range []string{"", "../escape", filepath.Join("nested", "..", "..", "escape")} {
		if err := ValidateArtifactPath(path); err == nil {
			t.Fatalf("ValidateArtifactPath(%q) accepted unsafe path", path)
		}
	}
}
