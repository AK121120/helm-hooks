// Package main provides the helm-hooks post-renderer binary.
// It reads Helm-rendered YAML from stdin, enhances hook resources,
// and outputs modified YAML to stdout.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/agk/helm-hooks/internal/hook"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	// Handle version command
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v") {
		printVersion()
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "helm-hooks: %v\n", err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("helm-hooks %s\n", Version)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
	fmt.Printf("  Build Date: %s\n", BuildDate)
}

func run() error {
	// Read all YAML from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	// Process the YAML through hook enhancement
	output, err := hook.Process(input)
	if err != nil {
		return fmt.Errorf("processing hooks: %w", err)
	}

	// Write result to stdout
	_, err = os.Stdout.Write(output)
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
