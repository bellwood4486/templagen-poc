package magic

import (
	"strings"

	"github.com/bellwood4486/templagen-poc/internal/util"
)

// TypeResolver manages type overrides from @param directives
type TypeResolver struct {
	overrides    map[string]string      // path -> Go type string (e.g., "User.Age" -> "int")
	structFields map[string]map[string]string  // path -> field definitions for struct types
	imports      map[string]struct{}
}

// NewTypeResolver creates a TypeResolver from template source
func NewTypeResolver(src string) (*TypeResolver, error) {
	directives, err := ParseParams(src)
	if err != nil {
		return nil, err
	}

	resolver := &TypeResolver{
		overrides:    make(map[string]string),
		structFields: make(map[string]map[string]string),
		imports:      make(map[string]struct{}),
	}

	for _, dir := range directives {
		// Handle []struct{...} specially to create named types
		if dir.Type.Kind == TypeKindSlice && dir.Type.Elem != nil && dir.Type.Elem.Kind == TypeKindStruct {
			// Create a named type like "ItemsItem" for []struct{...}
			typeName := util.Export(dir.Path) + "Item"
			resolver.overrides[dir.Path] = "[]" + typeName

			// Store struct field definitions
			fields := make(map[string]string)
			for _, field := range dir.Type.Elem.Fields {
				fields[field.Name] = resolver.typeExprToString(field.Type)
			}
			resolver.structFields[dir.Path] = fields
		} else {
			typeStr := resolver.typeExprToString(dir.Type)
			resolver.overrides[dir.Path] = typeStr
		}

		resolver.collectImportsFromExpr(&dir.Type)
	}

	return resolver, nil
}

// GetType returns the type override for a given path, if any
func (r *TypeResolver) GetType(path []string) (string, bool) {
	key := strings.Join(path, ".")
	typ, ok := r.overrides[key]
	return typ, ok
}

// RequiredImports returns the list of imports required by type overrides
func (r *TypeResolver) RequiredImports() []string {
	var result []string
	for imp := range r.imports {
		result = append(result, imp)
	}
	return result
}

// GetAllOverrides returns all type overrides
func (r *TypeResolver) GetAllOverrides() map[string]string {
	return r.overrides
}

// GetStructFields returns struct field definitions for a given path
func (r *TypeResolver) GetStructFields(path string) map[string]string {
	return r.structFields[path]
}

// typeExprToString converts TypeExpr to Go type string
func (r *TypeResolver) typeExprToString(expr TypeExpr) string {
	switch expr.Kind {
	case TypeKindBase:
		return expr.BaseType
	case TypeKindSlice:
		if expr.Elem != nil {
			return "[]" + r.typeExprToString(*expr.Elem)
		}
		return "[]string"
	case TypeKindMap:
		if expr.Elem != nil {
			return "map[string]" + r.typeExprToString(*expr.Elem)
		}
		return "map[string]string"
	case TypeKindPointer:
		if expr.Elem != nil {
			return "*" + r.typeExprToString(*expr.Elem)
		}
		return "*string"
	case TypeKindStruct:
		// For struct, we'll generate inline struct type
		// This is a simplified version - in practice we might want to generate named types
		var fields []string
		for _, f := range expr.Fields {
			fields = append(fields, f.Name + " " + r.typeExprToString(f.Type))
		}
		return "struct{" + strings.Join(fields, "; ") + "}"
	default:
		return "string"
	}
}

// collectImportsFromExpr collects required imports from TypeExpr
func (r *TypeResolver) collectImportsFromExpr(expr *TypeExpr) {
	if expr == nil {
		return
	}

	switch expr.Kind {
	case TypeKindBase:
		if expr.BaseType == "time.Time" {
			r.imports["time"] = struct{}{}
		}
	case TypeKindSlice, TypeKindMap, TypeKindPointer:
		r.collectImportsFromExpr(expr.Elem)
	case TypeKindStruct:
		for _, f := range expr.Fields {
			r.collectImportsFromExpr(&f.Type)
		}
	}
}