package typing

// TypedSchema represents a schema with resolved types
type TypedSchema struct {
	// トップレベルフィールド
	Fields map[string]*TypedField
	// 生成すべき名前付き型のリスト（例: ItemsItem）
	NamedTypes []*NamedType
}

// TypedField represents a field with resolved type
type TypedField struct {
	Name     string                   // フィールド名（エクスポート済み）
	GoType   string                   // 最終的なGo型文字列（例: "int", "[]ItemsItem"）
	Children map[string]*TypedField   // 構造体の子フィールド
}

// NamedType represents a named type to be generated
type NamedType struct {
	Name   string                   // 型名（例: "ItemsItem"）
	Fields map[string]*TypedField   // 構造体フィールド
}