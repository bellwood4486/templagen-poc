package typing

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/bellwood4486/templagen-poc/internal/magic"
	"github.com/bellwood4486/templagen-poc/internal/scan"
	"github.com/bellwood4486/templagen-poc/internal/util"
)

// Resolve resolves types for a schema with both default inference and @param overrides
func Resolve(schema scan.Schema, templateSrc string) (*TypedSchema, error) {
	// 1. デフォルト型推論
	typed := inferDefaultTypes(schema)

	// 2. @paramによるオーバーライド適用
	resolver, err := magic.NewTypeResolver(templateSrc)
	if err != nil {
		return nil, fmt.Errorf("failed to create type resolver: %w", err)
	}

	// オーバーライドを適用
	applyOverrides(typed, resolver)

	// 3. 必要なimportを収集
	typed.RequiredImports = resolver.RequiredImports()

	// 4. 名前付き型を抽出
	extractNamedTypes(typed)

	return typed, nil
}

// inferDefaultTypes performs default type inference
func inferDefaultTypes(schema scan.Schema) *TypedSchema {
	typed := &TypedSchema{
		Fields:          make(map[string]*TypedField),
		RequiredImports: []string{},
		NamedTypes:      []*NamedType{},
	}

	for name, field := range schema.Fields {
		typed.Fields[name] = inferFieldType([]string{name}, field)
	}

	return typed
}

// inferFieldType infers type for a single field
func inferFieldType(path []string, field *scan.Field) *TypedField {
	typed := &TypedField{
		Name: field.Name,
	}

	switch field.Kind {
	case scan.KindString:
		typed.GoType = "string"

	case scan.KindStruct:
		// 構造体の場合、名前付き型かインライン型か判断
		typeName := util.Export(path[len(path)-1])
		typed.GoType = typeName
		typed.Children = make(map[string]*TypedField)

		// 子フィールドの型推論
		for childName, childField := range field.Children {
			childPath := append(path, childName)
			typed.Children[childName] = inferFieldType(childPath, childField)
		}

	case scan.KindSlice:
		elemType := "string"
		if field.Elem != nil {
			if field.Elem.Kind == scan.KindStruct {
				// スライスの要素が構造体の場合、ItemsItemのような名前付き型
				if len(path) > 0 {
					elemType = util.Export(path[len(path)-1]) + "Item"
				} else {
					elemType = util.Export(field.Name) + "Item"
				}
				// 要素の子フィールドも推論し、その子フィールドを保存
				elem := inferFieldType(path, field.Elem)
				typed.Children = elem.Children
			} else {
				elem := inferFieldType(path, field.Elem)
				elemType = elem.GoType
			}
		}
		typed.GoType = "[]" + elemType

	case scan.KindMap:
		valType := "string"
		if field.Elem != nil {
			elem := inferFieldType(path, field.Elem)
			valType = elem.GoType
		}
		typed.GoType = "map[string]" + valType

	default:
		typed.GoType = "string"
	}

	return typed
}

// applyOverrides applies @param overrides to typed schema
func applyOverrides(typed *TypedSchema, resolver *magic.TypeResolver) {
	// トップレベルフィールドから順に処理
	for name, field := range typed.Fields {
		applyFieldOverride([]string{name}, field, resolver)
	}

	// @paramで定義された構造体型を名前付き型として追加
	for path, typeStr := range resolver.GetAllOverrides() {
		if strings.HasPrefix(typeStr, "[]") && !strings.HasPrefix(typeStr, "[]struct{") && !isBuiltinType(typeStr[2:]) {
			// []ItemsItem のような名前付き型
			if fields := resolver.GetStructFields(path); fields != nil {
				namedType := &NamedType{
					Name:   typeStr[2:], // "ItemsItem"
					Fields: make(map[string]*TypedField),
				}
				for fieldName, fieldType := range fields {
					namedType.Fields[fieldName] = &TypedField{
						Name:   util.Export(fieldName),
						GoType: fieldType,
					}
				}
				typed.NamedTypes = append(typed.NamedTypes, namedType)
			}
		}
	}
}

// applyFieldOverride applies override for a single field recursively
func applyFieldOverride(path []string, field *TypedField, resolver *magic.TypeResolver) {
	// このパスに対するオーバーライドを確認
	if overrideType, ok := resolver.GetType(path); ok {
		field.GoType = overrideType
		// @paramで上書きされた場合、子フィールドは不要
		field.Children = nil
		return
	}

	// 子フィールドに対して再帰的に適用
	if field.Children != nil {
		for childName, childField := range field.Children {
			childPath := append(path, childName)
			applyFieldOverride(childPath, childField, resolver)
		}
	}
}

// extractNamedTypes extracts named struct types
func extractNamedTypes(typed *TypedSchema) {
	namedTypes := make(map[string]*NamedType)

	var extract func(path []string, field *TypedField)
	extract = func(path []string, field *TypedField) {
		// スライスの要素型が名前付き構造体の場合
		if strings.HasPrefix(field.GoType, "[]") {
			elemType := field.GoType[2:]
			if !isBuiltinType(elemType) && !strings.Contains(elemType, "[") &&
			   !strings.Contains(elemType, "map") && !strings.HasPrefix(elemType, "struct{") {
				// すでに登録済みでない場合のみ追加
				if _, exists := namedTypes[elemType]; !exists {
					// scan結果から構造体を探す
					if field.Children != nil && len(field.Children) > 0 {
						namedType := &NamedType{
							Name:   elemType,
							Fields: field.Children,
						}
						namedTypes[elemType] = namedType
					}
				}
			}
		}

		// 構造体型の場合
		if field.GoType != "" && !isBuiltinType(field.GoType) &&
		   !strings.Contains(field.GoType, "[") && !strings.Contains(field.GoType, "map") &&
		   field.GoType != "Params" && field.Children != nil && len(field.Children) > 0 {
			if _, exists := namedTypes[field.GoType]; !exists {
				namedType := &NamedType{
					Name:   field.GoType,
					Fields: field.Children,
				}
				namedTypes[field.GoType] = namedType
			}
		}

		// 子フィールドも再帰的に処理
		if field.Children != nil {
			for childName, childField := range field.Children {
				childPath := append(path, childName)
				extract(childPath, childField)
			}
		}
	}

	// トップレベルから処理
	for name, field := range typed.Fields {
		extract([]string{name}, field)
	}

	// マップから配列に変換（順序を安定させる）
	names := slices.Sorted(maps.Keys(namedTypes))
	for _, name := range names {
		typed.NamedTypes = append(typed.NamedTypes, namedTypes[name])
	}
}

func isBuiltinType(typeName string) bool {
	builtins := []string{
		"string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune", "any",
		"time.Time",
	}
	for _, b := range builtins {
		if typeName == b {
			return true
		}
	}
	return false
}