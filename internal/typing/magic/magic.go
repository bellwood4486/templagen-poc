package magic

import (
	"fmt"
	"regexp"
	"strings"
)

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
	BaseType string     // for base types: "string", "int", "time.Time"
	Elem     *TypeExpr  // for slice/map/pointer
	Fields   []FieldDef // for struct
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

