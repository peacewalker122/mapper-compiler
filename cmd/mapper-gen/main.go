package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peacewalker122/mapper/compiler"
	"github.com/peacewalker122/mapper/plugin"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: mapper-gen <validate|generate> ...")
	}
	switch args[0] {
	case "validate":
		return validate(args[1:])
	case "generate":
		return generate(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func validate(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: mapper-gen validate mapper.yaml")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := compiler.ParseAndValidate(data); err != nil {
		return err
	}
	return nil
}

func generate(args []string) error {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	input := flags.String("input", "", "mapper YAML path")
	lock := flags.String("lock", "", "lock YAML path")
	timeout := flags.Duration("timeout", 2*time.Minute, "per-plugin timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return errors.New("generate requires --input")
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	parsed, err := compiler.ParseAndValidate(data)
	if err != nil {
		return err
	}
	if len(parsed.Generators) == 0 {
		return errors.New("generate requires at least one generators entry")
	}
	lockPath := *lock
	if lockPath == "" {
		lockPath = strings.TrimSuffix(*input, filepath.Ext(*input)) + ".lock.yaml"
	}
	lockData, err := os.ReadFile(lockPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read lock: %w", err)
	}
	compiled, err := (&compiler.Compiler{}).Compile(compiler.CompileRequest{
		Source:   data,
		LockFile: lockData,
	})
	if err != nil {
		return err
	}

	runner := plugin.Runner{Registry: plugin.NewRegistry(), Timeout: *timeout}
	outputs := make([]plugin.Output, 0, len(parsed.Generators))
	for _, config := range parsed.Generators {
		files, err := runner.Run(context.Background(), config, compiled.Schema)
		if err != nil {
			return err
		}
		outputs = append(outputs, plugin.Output{Root: config.Out, Files: files})
	}
	if err := plugin.WriteOutputs(outputs); err != nil {
		return fmt.Errorf("write generated artifacts: %w", err)
	}
	if err := plugin.WriteAtomic(lockPath, compiled.LockFile); err != nil {
		return fmt.Errorf("write lock: %w", err)
	}
	return nil
}
