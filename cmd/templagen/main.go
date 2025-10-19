package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bellwood4486/templagen-poc/internal/gen"
)

func main() {
	in := flag.String("in", "", "input pattern (glob supported)")
	pkg := flag.String("pkg", "", "output package name")
	out := flag.String("out", "", "output .go file path")
	exclude := flag.String("exclude", "", "exclude pattern (optional)")
	flag.Parse()

	if *in == "" || *pkg == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: templagen -in <pattern> -pkg <name> -out <file>")
		os.Exit(2)
	}

	// 入力ファイルのリストを取得
	files, err := resolveInputFiles(*in, *exclude)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to resolve input files: %w", err))
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no input files found")
		os.Exit(1)
	}

	// 複数のテンプレートを処理
	units := make([]gen.Unit, 0, len(files))
	outDir := filepath.Dir(*out)

	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("failed to read %s: %w", file, err))
			os.Exit(1)
		}

		relPath, err := filepath.Rel(outDir, file)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("failed to get relative path for %s: %w", file, err))
			os.Exit(1)
		}

		units = append(units, gen.Unit{
			Pkg:           *pkg,
			SourcePath:    relPath,
			SourceLiteral: string(src),
			TemplateName:  gen.ExtractTemplateName(file),
		})
	}

	// コード生成
	code, err := gen.Emit(units)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to emit: %w", err))
		os.Exit(1)
	}

	if err := os.WriteFile(*out, []byte(code), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveInputFiles は入力パターンから実際のファイルパスのリストを返す
func resolveInputFiles(pattern string, exclude string) ([]string, error) {
	var files []string

	// カンマ区切りのチェック
	if strings.Contains(pattern, ",") {
		// カンマ区切りの場合
		for _, p := range strings.Split(pattern, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				files = append(files, p)
			}
		}
	} else {
		// globパターンの場合
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		files = matches
	}

	// 除外パターンの適用
	if exclude != "" {
		var filtered []string
		for _, file := range files {
			matched, err := filepath.Match(exclude, filepath.Base(file))
			if err != nil {
				return nil, err
			}
			if !matched {
				filtered = append(filtered, file)
			}
		}
		files = filtered
	}

	return files, nil
}
