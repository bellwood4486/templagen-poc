package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bellwood4486/templagen-poc/internal/gen"
)

func main() {
	in := flag.String("in", "", "input template file")
	pkg := flag.String("pkg", "", "output package name")
	out := flag.String("out", "", "output .go file path")
	flag.Parse()

	if *in == "" || *pkg == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: templagen --in <file> --pkg <name> --out <file>")
		os.Exit(2)
	}

	src, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to read input template: %w", err))
		os.Exit(1)
	}

	outDir := filepath.Dir(*out)
	relPath, err := filepath.Rel(outDir, *in)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to get relative path: %w", err))
		os.Exit(1)
	}

	code, err := gen.Emit(gen.Unit{
		Pkg:           *pkg,
		SourcePath:    relPath,
		SourceLiteral: string(src),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to emit: %w", err))
		os.Exit(1)
	}

	if err := os.WriteFile(*out, []byte(code), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
