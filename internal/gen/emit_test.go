package gen_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bellwood4486/templagen-poc/internal/gen"
)

func parseCode(t *testing.T, code string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "gen.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v\ncode:\n%s", err, code)
	}
	return f
}

func hasImport(f *ast.File, path string, name string) bool {
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, "\"")
		var n string
		if imp.Name != nil {
			n = imp.Name.Name
		}
		if p == path {
			if name == "" && n == "" {
				return true
			}
			if name == n {
				return true
			}
		}
	}
	return false
}

func findType(f *ast.File, name string) *ast.StructType {
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, s := range gd.Specs {
			ts := s.(*ast.TypeSpec)
			if ts.Name.Name == name {
				st, _ := ts.Type.(*ast.StructType)
				return st
			}
		}
	}
	return nil
}

func findFunc(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
			return fd
		}
	}
	return nil
}

func TestEmit_BasicScaffoldAndTypes(t *testing.T) {
	u := gen.Unit{
		Pkg:           "x",
		SourcePath:    "tpl.tmpl",
		SourceLiteral: "{{ .User.Name }}\n{{ .Message }}\n",
	}

	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Quick string checks
	if !strings.Contains(code, "//go:embed "+u.SourcePath) {
		t.Fatalf("missing go:embed for %q\n%s", u.SourcePath, code)
	}
	if !strings.Contains(code, "Option(\"missingkey=error\")") {
		t.Fatalf("missing Template Option missingkey=error\n%s", code)
	}

	// AST checks
	f := parseCode(t, code)
	if f.Name.Name != u.Pkg {
		t.Fatalf("package name = %s; want %s", f.Name.Name, u.Pkg)
	}
	if !hasImport(f, "embed", "_") {
		t.Fatalf("import embed as blank not found")
	}
	if !hasImport(f, "io", "") || !hasImport(f, "text/template", "") {
		t.Fatalf("imports io or text/template not found")
	}

	// var tplSource string
	varFound := false
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, s := range gd.Specs {
			vs := s.(*ast.ValueSpec)
			if len(vs.Names) == 1 && vs.Names[0].Name == "tplSource" {
				if _, ok := vs.Type.(*ast.Ident); ok {
					varFound = true
					break
				}
			}
		}
	}
	if !varFound {
		t.Fatalf("var tplSource string not found")
	}

	// type User struct{ Name string }
	user := findType(f, "User")
	if user == nil || user.Fields == nil || len(user.Fields.List) == 0 {
		t.Fatalf("type User struct not found or empty")
	}
	if len(user.Fields.List) != 1 || len(user.Fields.List[0].Names) != 1 || user.Fields.List[0].Names[0].Name != "Name" {
		t.Fatalf("User fields unexpected")
	}
	if id, ok := user.Fields.List[0].Type.(*ast.Ident); !ok || id.Name != "string" {
		t.Fatalf("User.Name type != string")
	}

	// type Params { Message string; User User } with sorted order
	params := findType(f, "Params")
	if params == nil || params.Fields == nil || len(params.Fields.List) != 2 {
		t.Fatalf("Params fields unexpected")
	}
	if params.Fields.List[0].Names[0].Name != "Message" {
		t.Fatalf("Params first field = %s; want Message", params.Fields.List[0].Names[0].Name)
	}
	if id, ok := params.Fields.List[0].Type.(*ast.Ident); !ok || id.Name != "string" {
		t.Fatalf("Params.Message type != string")
	}
	if params.Fields.List[1].Names[0].Name != "User" {
		t.Fatalf("Params second field = %s; want User", params.Fields.List[1].Names[0].Name)
	}
	if id, ok := params.Fields.List[1].Type.(*ast.Ident); !ok || id.Name != "User" {
		t.Fatalf("Params.User type != User")
	}

	// Render and RenderAny signatures
	render := findFunc(f, "Render")
	if render == nil || render.Type == nil || render.Type.Params == nil || render.Type.Results == nil {
		t.Fatalf("Render signature not found")
	}
	if len(render.Type.Params.List) != 2 || len(render.Type.Results.List) != 1 {
		t.Fatalf("Render parameters/results unexpected")
	}
	// w io.Writer
	if se, ok := render.Type.Params.List[0].Type.(*ast.SelectorExpr); !ok || se.Sel.Name != "Writer" {
		t.Fatalf("Render first param not io.Writer")
	}
	if id, ok := render.Type.Params.List[1].Type.(*ast.Ident); !ok || id.Name != "Params" {
		t.Fatalf("Render second param not Params")
	}
	if id, ok := render.Type.Results.List[0].Type.(*ast.Ident); !ok || id.Name != "error" {
		t.Fatalf("Render result not error")
	}

	renderAny := findFunc(f, "RenderAny")
	if renderAny == nil || len(renderAny.Type.Params.List) != 2 {
		t.Fatalf("RenderAny signature not found")
	}
	if id, ok := renderAny.Type.Params.List[1].Type.(*ast.Ident); !ok || id.Name != "any" {
		t.Fatalf("RenderAny second param not any")
	}
}

func TestEmit_RangeAndIndex_TypesAndOrder(t *testing.T) {
	u := gen.Unit{
		Pkg:           "x",
		SourcePath:    "email.tmpl",
		SourceLiteral: "{{ range .Items }}{{ .Title }}{{ .ID }}{{ end }}\n{{ index .Meta \"env\" }}\n",
	}
	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}
	f := parseCode(t, code)

	// type ItemsItem with fields Title, ID (order sorted)
	it := findType(f, "ItemsItem")
	if it == nil || it.Fields == nil || len(it.Fields.List) != 2 {
		t.Fatalf("ItemsItem struct unexpected")
	}
	if it.Fields.List[0].Names[0].Name != "ID" || it.Fields.List[1].Names[0].Name != "Title" {
		t.Fatalf("ItemsItem fields not sorted as expected: got %s, %s", it.Fields.List[0].Names[0].Name, it.Fields.List[1].Names[0].Name)
	}

	params := findType(f, "Params")
	if params == nil || len(params.Fields.List) != 2 {
		t.Fatalf("Params unexpected")
	}
	if params.Fields.List[0].Names[0].Name != "Items" {
		t.Fatalf("Params first field = %s; want Items", params.Fields.List[0].Names[0].Name)
	}
	if at, ok := params.Fields.List[0].Type.(*ast.ArrayType); !ok {
		t.Fatalf("Params.Items not a slice")
	} else {
		if id, ok := at.Elt.(*ast.Ident); !ok || id.Name != "ItemsItem" {
			t.Fatalf("Params.Items element not ItemsItem")
		}
	}
	if params.Fields.List[1].Names[0].Name != "Meta" {
		t.Fatalf("Params second field = %s; want Meta", params.Fields.List[1].Names[0].Name)
	}
	if mt, ok := params.Fields.List[1].Type.(*ast.MapType); !ok {
		t.Fatalf("Params.Meta not a map")
	} else {
		if k, ok := mt.Key.(*ast.Ident); !ok || k.Name != "string" {
			t.Fatalf("Params.Meta key not string")
		}
		if v, ok := mt.Value.(*ast.Ident); !ok || v.Name != "string" {
			t.Fatalf("Params.Meta value not string")
		}
	}
}

func TestEmit_Golden_Simple(t *testing.T) {
	u := gen.Unit{Pkg: "x", SourcePath: "tpl.tmpl", SourceLiteral: "{{ .User.Name }}\n{{ .Message }}\n"}
	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}
	goldenPath := filepath.Join("testdata", "simple.golden")
	b, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden: %v", err)
	}
	want := string(b)
	if code != want {
		// On mismatch, it helps to see a unified-ish diff. Keep it short.
		t.Fatalf("golden mismatch\n--- want\n%s\n--- got\n%s", want, code)
	}
}

func TestEmit_CompilesInTempModule(t *testing.T) {
	if runtime.GOOS == "js" || runtime.GOOS == "wasip1" {
		t.Skip("skip on restricted platforms")
	}

	u := gen.Unit{Pkg: "x", SourcePath: "tpl.tmpl", SourceLiteral: "Hello {{ .Message }}"}
	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	dir := t.TempDir()
	// Create module
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmpmod\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write the template file for go:embed
	if err := os.WriteFile(filepath.Join(dir, u.SourcePath), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write generated code
	if err := os.WriteFile(filepath.Join(dir, "gen.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	// Ensure build cache is writable within sandbox
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(dir, ".gocache"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
}

func TestEmit_WithParamOverride_BasicTypes(t *testing.T) {
	src := `
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{ .User.Name }} is {{ .User.Age }} years old.
{{ if .User.Email }}Email: {{ .User.Email }}{{ end }}
`
	u := gen.Unit{
		Pkg:           "x",
		SourcePath:    "tpl.tmpl",
		SourceLiteral: src,
	}

	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	f := parseCode(t, code)

	// Check User struct has Age int and Email *string
	user := findType(f, "User")
	if user == nil {
		t.Fatal("User type not found")
	}

	foundAge := false
	foundEmail := false
	for _, field := range user.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		if name == "Age" {
			if id, ok := field.Type.(*ast.Ident); ok && id.Name == "int" {
				foundAge = true
			}
		}
		if name == "Email" {
			if st, ok := field.Type.(*ast.StarExpr); ok {
				if id, ok := st.X.(*ast.Ident); ok && id.Name == "string" {
					foundEmail = true
				}
			}
		}
	}

	if !foundAge {
		t.Error("User.Age int not found")
	}
	if !foundEmail {
		t.Error("User.Email *string not found")
	}
}

func TestEmit_WithParamOverride_SliceType(t *testing.T) {
	src := `
{{/* @param Items []struct{ID int64; Title string} */}}
{{ range .Items }}{{ .ID }}: {{ .Title }}{{ end }}
`
	u := gen.Unit{
		Pkg:           "x",
		SourcePath:    "tpl.tmpl",
		SourceLiteral: src,
	}

	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	f := parseCode(t, code)

	// Check ItemsItem has ID int64 and Title string
	item := findType(f, "ItemsItem")
	if item == nil {
		t.Fatal("ItemsItem type not found")
	}

	foundID := false
	foundTitle := false
	for _, field := range item.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		if name == "ID" {
			if id, ok := field.Type.(*ast.Ident); ok && id.Name == "int64" {
				foundID = true
			}
		}
		if name == "Title" {
			if id, ok := field.Type.(*ast.Ident); ok && id.Name == "string" {
				foundTitle = true
			}
		}
	}

	if !foundID {
		t.Error("ItemsItem.ID int64 not found")
	}
	if !foundTitle {
		t.Error("ItemsItem.Title string not found")
	}
}

func TestEmit_WithParamOverride_TimeImport(t *testing.T) {
	src := `
{{/* @param CreatedAt time.Time */}}
Created: {{ .CreatedAt }}
`
	u := gen.Unit{
		Pkg:           "x",
		SourcePath:    "tpl.tmpl",
		SourceLiteral: src,
	}

	code, err := gen.Emit(u)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	f := parseCode(t, code)

	// Check time import
	if !hasImport(f, "time", "") {
		t.Error("import time not found")
	}

	// Check Params has CreatedAt time.Time
	params := findType(f, "Params")
	if params == nil {
		t.Fatal("Params type not found")
	}

	foundCreatedAt := false
	for _, field := range params.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		if field.Names[0].Name == "CreatedAt" {
			if se, ok := field.Type.(*ast.SelectorExpr); ok {
				if x, ok := se.X.(*ast.Ident); ok && x.Name == "time" && se.Sel.Name == "Time" {
					foundCreatedAt = true
				}
			}
		}
	}

	if !foundCreatedAt {
		t.Error("Params.CreatedAt time.Time not found")
	}
}
