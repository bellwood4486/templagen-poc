package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bellwood4486/tmpltype/internal/gen"
)

func main() {
	in := flag.String("in", "", "input pattern (glob supported)")
	pkg := flag.String("pkg", "", "output package name")
	out := flag.String("out", "", "output .go file path")
	exclude := flag.String("exclude", "", "exclude pattern (optional)")
	flag.Parse()

	if *in == "" || *pkg == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: tmpltype -in <pattern> -pkg <name> -out <file>")
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

	// カンマ区切りでパターンを分割
	patterns := []string{pattern}
	if strings.Contains(pattern, ",") {
		patterns = strings.Split(pattern, ",")
	}

	// 各パターンをグロブ展開
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// グロブパターンとして展開を試みる
		matches, err := filepath.Glob(p)
		if err != nil {
			return nil, err
		}

		// マッチした場合は展開結果を使用、マッチしない場合はそのまま使用
		if len(matches) > 0 {
			files = append(files, matches...)
		} else {
			files = append(files, p)
		}
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
