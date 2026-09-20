package plugin

import "testing"

func TestRegistryPrefersRegisteredCommand(t *testing.T) {
	registry := NewRegistry()
	registry.Register("go", Command{Path: "/custom/mapper-gen-go"})
	command, err := registry.Resolve("go")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if command.Path != "/custom/mapper-gen-go" {
		t.Fatalf("path = %q", command.Path)
	}
}
