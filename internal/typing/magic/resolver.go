package magic

import (
	"strings"

	"github.com/bellwood4486/templagen-poc/internal/util"
)

// TypeResolver は @param ディレクティブからの型オーバーライドを管理する
type TypeResolver struct {
	overrides    map[string]string      // パス -> Go型文字列 (例: "User.Age" -> "int")
	structFields map[string]map[string]string  // パス -> 構造体型のフィールド定義
}

// NewTypeResolver はテンプレートソースからTypeResolverを作成する
func NewTypeResolver(src string) (*TypeResolver, error) {
	directives, err := ParseParams(src)
	if err != nil {
		return nil, err
	}

	resolver := &TypeResolver{
		overrides:    make(map[string]string),
		structFields: make(map[string]map[string]string),
	}

	for _, dir := range directives {
		// []struct{...} を特別に扱い、名前付き型を作成
		if dir.Type.Kind == TypeKindSlice && dir.Type.Elem != nil && dir.Type.Elem.Kind == TypeKindStruct {
			// []struct{...} に対して "ItemsItem" のような名前付き型を作成
			typeName := util.Export(dir.Path) + "Item"
			resolver.overrides[dir.Path] = "[]" + typeName

			// 構造体フィールド定義を保存
			fields := make(map[string]string)
			for _, field := range dir.Type.Elem.Fields {
				fields[field.Name] = resolver.typeExprToString(field.Type)
			}
			resolver.structFields[dir.Path] = fields
		} else {
			typeStr := resolver.typeExprToString(dir.Type)
			resolver.overrides[dir.Path] = typeStr
		}
	}

	return resolver, nil
}

// GetType は指定されたパスに対する型オーバーライドがあれば返す
func (r *TypeResolver) GetType(path []string) (string, bool) {
	key := strings.Join(path, ".")
	typ, ok := r.overrides[key]
	return typ, ok
}

// GetAllOverrides はすべての型オーバーライドを返す
func (r *TypeResolver) GetAllOverrides() map[string]string {
	return r.overrides
}

// GetStructFields は指定されたパスの構造体フィールド定義を返す
func (r *TypeResolver) GetStructFields(path string) map[string]string {
	return r.structFields[path]
}

// typeExprToString はTypeExprをGo型文字列に変換する
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
		// 構造体の場合、インライン構造体型を生成
		// これは簡略版 - 実際には名前付き型を生成することも検討
		var fields []string
		for _, f := range expr.Fields {
			fields = append(fields, f.Name + " " + r.typeExprToString(f.Type))
		}
		return "struct{" + strings.Join(fields, "; ") + "}"
	default:
		return "string"
	}
}

