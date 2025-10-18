# フェーズ5: @param実装ガイド

## 実装手順

### Step 1: internal/magic/types.go - 型定義

```go
package magic

// TypeKind represents the kind of type expression
type TypeKind int

const (
	TypeKindBase TypeKind = iota
	TypeKindSlice
	TypeKindMap
	TypeKindPointer
	TypeKindStruct
)

// TypeExpr represents a parsed type expression
type TypeExpr struct {
	Kind     TypeKind
	BaseType string      // for base types: "string", "int", "time.Time"
	Elem     *TypeExpr   // for slice/map/pointer
	Fields   []FieldDef  // for struct
}

// FieldDef represents a field in struct type
type FieldDef struct {
	Name string
	Type TypeExpr
}

// ParamDirective represents a @param directive
type ParamDirective struct {
	Path string   // e.g., "User.Age"
	Type TypeExpr // parsed type
	Line int      // line number in template
}
```

### Step 2: internal/magic/parser.go - 型パーサー実装

```go
package magic

import (
	"fmt"
	"strings"
	"unicode"
)

// parseType parses a type string into TypeExpr
func parseType(s string) (TypeExpr, error) {
	p := &typeParser{input: s, pos: 0}
	return p.parseType()
}

type typeParser struct {
	input string
	pos   int
}

func (p *typeParser) parseType() (TypeExpr, error) {
	p.skipWhitespace()

	// Check for slice
	if strings.HasPrefix(p.remaining(), "[]") {
		p.pos += 2
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindSlice, Elem: &elem}, nil
	}

	// Check for map
	if strings.HasPrefix(p.remaining(), "map[string]") {
		p.pos += 11
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindMap, Elem: &elem}, nil
	}

	// Check for pointer
	if strings.HasPrefix(p.remaining(), "*") {
		p.pos++
		elem, err := p.parseType()
		if err != nil {
			return TypeExpr{}, err
		}
		return TypeExpr{Kind: TypeKindPointer, Elem: &elem}, nil
	}

	// Check for struct
	if strings.HasPrefix(p.remaining(), "struct{") {
		return p.parseStruct()
	}

	// Parse base type
	return p.parseBaseType()
}

func (p *typeParser) parseBaseType() (TypeExpr, error) {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if !unicode.IsLetter(rune(ch)) && !unicode.IsDigit(rune(ch)) && ch != '.' && ch != '_' {
			break
		}
		p.pos++
	}

	if start == p.pos {
		return TypeExpr{}, fmt.Errorf("expected type at position %d", p.pos)
	}

	baseType := p.input[start:p.pos]
	return TypeExpr{Kind: TypeKindBase, BaseType: baseType}, nil
}

func (p *typeParser) parseStruct() (TypeExpr, error) {
	// Skip "struct{"
	p.pos += 7

	var fields []FieldDef
	for {
		p.skipWhitespace()

		if p.pos >= len(p.input) {
			return TypeExpr{}, fmt.Errorf("unexpected end of struct")
		}

		if p.input[p.pos] == '}' {
			p.pos++
			break
		}

		// Parse field name
		name := p.parseIdentifier()
		if name == "" {
			return TypeExpr{}, fmt.Errorf("expected field name at position %d", p.pos)
		}

		p.skipWhitespace()

		// Parse field type
		fieldType, err := p.parseType()
		if err != nil {
			return TypeExpr{}, fmt.Errorf("invalid field type for %s: %w", name, err)
		}

		fields = append(fields, FieldDef{Name: name, Type: fieldType})

		p.skipWhitespace()

		// Check for separator or end
		if p.pos < len(p.input) && p.input[p.pos] == ';' {
			p.pos++
			continue
		}

		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
			break
		}
	}

	return TypeExpr{Kind: TypeKindStruct, Fields: fields}, nil
}

func (p *typeParser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if !unicode.IsLetter(rune(ch)) && !unicode.IsDigit(rune(ch)) && ch != '_' {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *typeParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *typeParser) remaining() string {
	if p.pos >= len(p.input) {
		return ""
	}
	return p.input[p.pos:]
}
```

### Step 3: internal/magic/magic.go - メイン処理

```go
package magic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bellwood4486/templagen-poc/internal/scan"
)

var paramRegex = regexp.MustCompile(`\{\{/\*\s*@param\s+(\S+)\s+(.+?)\s*\*/\}\}`)

// ParseParams extracts @param directives from template source
func ParseParams(src string) ([]ParamDirective, error) {
	var directives []ParamDirective

	lines := strings.Split(src, "\n")
	lineNum := 0

	for _, line := range lines {
		lineNum++
		matches := paramRegex.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			if len(match) != 3 {
				continue
			}

			path := match[1]
			typeStr := match[2]

			typeExpr, err := parseType(typeStr)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid type expression %q: %w", lineNum, typeStr, err)
			}

			directives = append(directives, ParamDirective{
				Path: path,
				Type: typeExpr,
				Line: lineNum,
			})
		}
	}

	return directives, nil
}

// ApplyOverrides applies @param directives to schema tree
func ApplyOverrides(schema *scan.Schema, directives []ParamDirective) error {
	for _, dir := range directives {
		if err := applyOverride(schema, dir); err != nil {
			// Warning, not fatal
			fmt.Printf("Warning: %v\n", err)
		}
	}
	return nil
}

func applyOverride(schema *scan.Schema, dir ParamDirective) error {
	parts := strings.Split(dir.Path, ".")
	if len(parts) == 0 {
		return fmt.Errorf("empty path in @param directive")
	}

	// Navigate to target field
	field := findField(schema, parts)
	if field == nil {
		return fmt.Errorf("path %q not found in template", dir.Path)
	}

	// Apply type override
	applyTypeToField(field, dir.Type)

	return nil
}

func findField(schema *scan.Schema, parts []string) *scan.Field {
	if len(parts) == 0 || schema.Fields == nil {
		return nil
	}

	field := schema.Fields[parts[0]]
	if field == nil {
		return nil
	}

	for i := 1; i < len(parts); i++ {
		if field.Children == nil {
			return nil
		}
		field = field.Children[parts[i]]
		if field == nil {
			return nil
		}
	}

	return field
}

func applyTypeToField(field *scan.Field, typeExpr TypeExpr) {
	switch typeExpr.Kind {
	case TypeKindBase:
		field.Kind = scan.KindString
		field.TypeName = typeExpr.BaseType

	case TypeKindSlice:
		field.Kind = scan.KindSlice
		if typeExpr.Elem != nil {
			if field.Elem == nil {
				field.Elem = &scan.Field{Name: field.Name + "Item"}
			}
			applyTypeToField(field.Elem, *typeExpr.Elem)
		}

	case TypeKindMap:
		field.Kind = scan.KindMap
		if typeExpr.Elem != nil {
			if field.Elem == nil {
				field.Elem = &scan.Field{Name: field.Name + "Value"}
			}
			applyTypeToField(field.Elem, *typeExpr.Elem)
		}

	case TypeKindPointer:
		// Store pointer info in TypeName
		if typeExpr.Elem != nil {
			applyTypeToField(field, *typeExpr.Elem)
			field.TypeName = "*" + field.TypeName
		}

	case TypeKindStruct:
		field.Kind = scan.KindStruct
		if field.Children == nil {
			field.Children = make(map[string]*scan.Field)
		}

		// Apply struct fields
		for _, f := range typeExpr.Fields {
			childField := &scan.Field{Name: f.Name}
			applyTypeToField(childField, f.Type)
			field.Children[strings.ToLower(f.Name)] = childField
		}
	}
}

// RequiredImports returns list of imports needed for the types
func RequiredImports(directives []ParamDirective) []string {
	imports := make(map[string]struct{})

	for _, dir := range directives {
		collectImports(&dir.Type, imports)
	}

	var result []string
	for imp := range imports {
		result = append(result, imp)
	}

	return result
}

func collectImports(typeExpr *TypeExpr, imports map[string]struct{}) {
	if typeExpr == nil {
		return
	}

	switch typeExpr.Kind {
	case TypeKindBase:
		if typeExpr.BaseType == "time.Time" {
			imports["time"] = struct{}{}
		}

	case TypeKindSlice, TypeKindMap, TypeKindPointer:
		collectImports(typeExpr.Elem, imports)

	case TypeKindStruct:
		for _, f := range typeExpr.Fields {
			collectImports(&f.Type, imports)
		}
	}
}
```

### Step 4: internal/scan/scan.go の拡張

```go
// Field 構造体に TypeName フィールドを追加
type Field struct {
	Name     string
	Kind     Kind
	TypeName string            // 追加: 具体的な型名
	Elem     *Field
	Children map[string]*Field
}
```

### Step 5: internal/gen/emit.go の修正

```go
// goTypeOf 関数を修正して TypeName を使用
func goTypeOf(name string, f *scan.Field) string {
	// TypeName が設定されていればそれを優先
	if f.TypeName != "" {
		return f.TypeName
	}

	// 既存のロジック
	switch f.Kind {
	case scan.KindString:
		return "string"
	// ...
	}
}
```

## テストケース例

### internal/magic/magic_test.go

```go
package magic

import (
	"testing"

	"github.com/bellwood4486/templagen-poc/internal/scan"
)

func TestParseParams_Simple(t *testing.T) {
	src := `
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{ .User.Name }}
`
	directives, err := ParseParams(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(directives) != 2 {
		t.Errorf("expected 2 directives, got %d", len(directives))
	}

	// Check first directive
	if directives[0].Path != "User.Age" {
		t.Errorf("unexpected path: %s", directives[0].Path)
	}
	if directives[0].Type.Kind != TypeKindBase || directives[0].Type.BaseType != "int" {
		t.Errorf("unexpected type for User.Age")
	}

	// Check second directive
	if directives[1].Path != "User.Email" {
		t.Errorf("unexpected path: %s", directives[1].Path)
	}
	if directives[1].Type.Kind != TypeKindPointer {
		t.Errorf("expected pointer type for User.Email")
	}
}

func TestApplyOverrides(t *testing.T) {
	// Create sample schema
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"User": {
				Name: "User",
				Kind: scan.KindStruct,
				Children: map[string]*scan.Field{
					"Age": {
						Name: "Age",
						Kind: scan.KindString,
					},
				},
			},
		},
	}

	directives := []ParamDirective{
		{
			Path: "User.Age",
			Type: TypeExpr{Kind: TypeKindBase, BaseType: "int"},
		},
	}

	err := ApplyOverrides(&schema, directives)
	if err != nil {
		t.Fatal(err)
	}

	// Check override was applied
	ageField := schema.Fields["User"].Children["Age"]
	if ageField.TypeName != "int" {
		t.Errorf("expected TypeName to be 'int', got %s", ageField.TypeName)
	}
}
```

## 実装チェックリスト

- [ ] `internal/magic/types.go` - 型定義
- [ ] `internal/magic/parser.go` - 型パーサー
- [ ] `internal/magic/parser_test.go` - パーサーテスト
- [ ] `internal/magic/magic.go` - メイン処理
- [ ] `internal/magic/magic_test.go` - 統合テスト
- [ ] `internal/scan/scan.go` - Field構造体拡張
- [ ] `internal/gen/emit.go` - TypeName対応
- [ ] `internal/gen/emit_test.go` - @param統合テスト

## 実装の注意点

1. **エラー処理**: @paramパースエラーは致命的、パス不在は警告
2. **後方互換性**: TypeNameが空の場合は既存ロジックで動作
3. **import管理**: time.Time使用時のみ自動追加
4. **テスト**: 各コンポーネントの単体テストを先に実装