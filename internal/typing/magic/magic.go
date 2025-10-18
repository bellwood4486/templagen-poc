package magic

import (
	"fmt"
	"regexp"
	"strings"
)

// TypeKind は型表現の種類を表す
type TypeKind int

const (
	TypeKindBase TypeKind = iota
	TypeKindSlice
	TypeKindMap
	TypeKindPointer
	TypeKindStruct
)

// TypeExpr はパース済みの型表現を表す
type TypeExpr struct {
	Kind     TypeKind
	BaseType string     // 基本型用: "string", "int", "time.Time"
	Elem     *TypeExpr  // スライス/マップ/ポインタ用
	Fields   []FieldDef // 構造体用
}

// FieldDef は構造体型のフィールドを表す
type FieldDef struct {
	Name string
	Type TypeExpr
}

// ParamDirective は @param ディレクティブを表す
type ParamDirective struct {
	Path string   // 例: "User.Age"
	Type TypeExpr // パース済みの型
	Line int      // テンプレート内の行番号
}

var paramRegex = regexp.MustCompile(`\{\{/\*\s*@param\s+(\S+)\s+(.+?)\s*\*/\}\}`)

// ParseParams はテンプレートソースから @param ディレクティブを抽出する
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

