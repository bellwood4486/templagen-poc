package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bellwood4486/tmpltype/internal/gen"
)

func main() {
	dir := flag.String("dir", "", "template directory (required)")
	pkg := flag.String("pkg", "", "output package name (required)")
	out := flag.String("out", "", "output .go file path (required)")
	flag.Parse()

	if *dir == "" || *pkg == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: tmpltype -dir <directory> -pkg <name> -out <file>")
		os.Exit(2)
	}

	// ディレクトリの存在確認
	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory not found: %s\n", *dir)
		os.Exit(1)
	}

	// テンプレートファイルをスキャン
	files, err := scanTemplateFiles(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to scan directory: %w", err))
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no .tmpl files found in %s/\n", *dir)
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

	// コード生成（basedirを渡す）
	code, err := gen.Emit(units, *dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to emit: %w", err))
		os.Exit(1)
	}

	if err := os.WriteFile(*out, []byte(code), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// scanTemplateFiles はディレクトリから.tmplファイルをスキャンする
// dir/*.tmpl (フラット) と dir/*/*.tmpl (グループ) のみを対象とする
func scanTemplateFiles(dir string) ([]string, error) {
	var files []string

	// フラットなテンプレート: dir/*.tmpl
	flatPattern := filepath.Join(dir, "*.tmpl")
	flatFiles, err := filepath.Glob(flatPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan flat templates: %w", err)
	}
	files = append(files, flatFiles...)

	// グループ化されたテンプレート: dir/*/*.tmpl (1階層のみ)
	groupPattern := filepath.Join(dir, "*", "*.tmpl")
	groupFiles, err := filepath.Glob(groupPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan grouped templates: %w", err)
	}
	files = append(files, groupFiles...)

	return files, nil
}
