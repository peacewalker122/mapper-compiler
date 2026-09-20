package plugin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/peacewalker122/mapper/compiler"
	"github.com/peacewalker122/mapper/ir"
	protocol "github.com/peacewalker122/mapper/protocol/generator/v1"
)

type Command struct {
	Path string
	Args []string
}

type Registry struct {
	commands map[string]Command
}

func NewRegistry() Registry {
	return Registry{commands: map[string]Command{}}
}

func (r *Registry) Register(name string, command Command) {
	if r.commands == nil {
		r.commands = map[string]Command{}
	}
	r.commands[name] = command
}

func (r Registry) Resolve(name string) (Command, error) {
	if command, ok := r.commands[name]; ok {
		return command, nil
	}
	path, err := exec.LookPath("mapper-gen-" + name)
	if err != nil {
		return Command{}, fmt.Errorf("resolve plugin %q: %w", name, err)
	}
	return Command{Path: path}, nil
}

type Runner struct {
	Registry Registry
	Timeout  time.Duration
}

func (r Runner) Run(ctx context.Context, config compiler.GeneratorConfig, schema ir.Schema) ([]protocol.Artifact, error) {
	command, err := r.Registry.Resolve(config.Plugin)
	if err != nil {
		return nil, err
	}
	if command.Path == "" {
		return nil, fmt.Errorf("plugin %q has empty command", config.Plugin)
	}
	if r.Timeout <= 0 {
		r.Timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	request := protocol.NewRequest(schema, config.Options)
	var input bytes.Buffer
	if err := protocol.EncodeRequest(&input, request); err != nil {
		return nil, fmt.Errorf("encode plugin %q request: %w", config.Plugin, err)
	}
	cmd := exec.CommandContext(ctx, command.Path, command.Args...)
	cmd.Stdin = &input
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return nil, fmt.Errorf("plugin %q: %w: %s", config.Plugin, err, message)
		}
		return nil, fmt.Errorf("plugin %q: %w", config.Plugin, err)
	}
	response, err := protocol.DecodeResponse(&stdout)
	if err != nil {
		return nil, fmt.Errorf("decode plugin %q response: %w", config.Plugin, err)
	}
	for _, artifact := range response.Files {
		if err := ValidateArtifactPath(artifact.Path); err != nil {
			return nil, fmt.Errorf("plugin %q artifact: %w", config.Plugin, err)
		}
	}
	return response.Files, nil
}

func ValidateArtifactPath(path string) error {
	if path == "" || strings.IndexByte(path, 0) >= 0 {
		return fmt.Errorf("invalid empty or NUL-containing path")
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("path must be relative: %q", path)
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes output directory: %q", path)
	}
	return nil
}

type Output struct {
	Root  string
	Files []protocol.Artifact
}

func WriteOutputs(outputs []Output) error {
	seen := map[string]string{}
	for _, output := range outputs {
		for _, artifact := range output.Files {
			if err := ValidateArtifactPath(artifact.Path); err != nil {
				return err
			}
			target := filepath.Join(output.Root, filepath.Clean(artifact.Path))
			if owner, ok := seen[target]; ok {
				return fmt.Errorf("artifact collision: %s and %s", owner, target)
			}
			seen[target] = target
		}
	}
	for _, output := range outputs {
		for _, artifact := range output.Files {
			target := filepath.Join(output.Root, filepath.Clean(artifact.Path))
			if err := writeAtomic(target, []byte(artifact.Content)); err != nil {
				return err
			}
		}
	}
	return nil
}

func WriteAtomic(path string, data []byte) error {
	return writeAtomic(path, data)
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".mapper-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
