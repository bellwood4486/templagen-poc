package typing

import (
	"testing"

	"github.com/bellwood4486/templagen-poc/internal/scan"
)

func TestIsBuiltinType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"string", "string", true},
		{"int", "int", true},
		{"int64", "int64", true},
		{"float64", "float64", true},
		{"bool", "bool", true},
		{"custom", "User", false},
		{"slice", "[]string", false},
		{"map", "map[string]string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinType(tt.typeName)
			if got != tt.want {
				t.Errorf("isBuiltinType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestInferDefaultTypes_BasicFields(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"Name": {
				Name: "Name",
				Kind: scan.KindString,
			},
			"Message": {
				Name: "Message",
				Kind: scan.KindString,
			},
		},
	}

	typed := inferDefaultTypes(schema)

	if len(typed.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(typed.Fields))
	}

	if typed.Fields["Name"].GoType != "string" {
		t.Errorf("Name.GoType = %q, want %q", typed.Fields["Name"].GoType, "string")
	}

	if typed.Fields["Message"].GoType != "string" {
		t.Errorf("Message.GoType = %q, want %q", typed.Fields["Message"].GoType, "string")
	}
}

func TestInferDefaultTypes_StructField(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"User": {
				Name: "User",
				Kind: scan.KindStruct,
				Children: map[string]*scan.Field{
					"Name": {
						Name: "Name",
						Kind: scan.KindString,
					},
					"Age": {
						Name: "Age",
						Kind: scan.KindString,
					},
				},
			},
		},
	}

	typed := inferDefaultTypes(schema)

	if typed.Fields["User"].GoType != "User" {
		t.Errorf("User.GoType = %q, want %q", typed.Fields["User"].GoType, "User")
	}

	if len(typed.Fields["User"].Children) != 2 {
		t.Errorf("User children count = %d, want 2", len(typed.Fields["User"].Children))
	}

	if typed.Fields["User"].Children["Name"].GoType != "string" {
		t.Errorf("User.Name.GoType = %q, want %q", typed.Fields["User"].Children["Name"].GoType, "string")
	}
}

func TestInferDefaultTypes_SliceField(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"Items": {
				Name: "Items",
				Kind: scan.KindSlice,
				Elem: &scan.Field{
					Name: "Items",
					Kind: scan.KindStruct,
					Children: map[string]*scan.Field{
						"ID": {
							Name: "ID",
							Kind: scan.KindString,
						},
						"Title": {
							Name: "Title",
							Kind: scan.KindString,
						},
					},
				},
			},
		},
	}

	typed := inferDefaultTypes(schema)

	if typed.Fields["Items"].GoType != "[]ItemsItem" {
		t.Errorf("Items.GoType = %q, want %q", typed.Fields["Items"].GoType, "[]ItemsItem")
	}

	if len(typed.Fields["Items"].Children) != 2 {
		t.Errorf("Items children count = %d, want 2", len(typed.Fields["Items"].Children))
	}
}

func TestInferDefaultTypes_MapField(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"Meta": {
				Name: "Meta",
				Kind: scan.KindMap,
				Elem: &scan.Field{
					Name: "Meta",
					Kind: scan.KindString,
				},
			},
		},
	}

	typed := inferDefaultTypes(schema)

	if typed.Fields["Meta"].GoType != "map[string]string" {
		t.Errorf("Meta.GoType = %q, want %q", typed.Fields["Meta"].GoType, "map[string]string")
	}
}

func TestExtractNamedTypes(t *testing.T) {
	typed := &TypedSchema{
		Fields: map[string]*TypedField{
			"Items": {
				Name:   "Items",
				GoType: "[]ItemsItem",
				Children: map[string]*TypedField{
					"ID": {
						Name:   "ID",
						GoType: "string",
					},
					"Title": {
						Name:   "Title",
						GoType: "string",
					},
				},
			},
			"User": {
				Name:   "User",
				GoType: "User",
				Children: map[string]*TypedField{
					"Name": {
						Name:   "Name",
						GoType: "string",
					},
				},
			},
		},
		NamedTypes: []*NamedType{},
	}

	extractNamedTypes(typed)

	if len(typed.NamedTypes) != 2 {
		t.Errorf("expected 2 named types, got %d", len(typed.NamedTypes))
	}

	// Check order (should be sorted)
	if typed.NamedTypes[0].Name != "ItemsItem" {
		t.Errorf("first named type = %q, want %q", typed.NamedTypes[0].Name, "ItemsItem")
	}

	if typed.NamedTypes[1].Name != "User" {
		t.Errorf("second named type = %q, want %q", typed.NamedTypes[1].Name, "User")
	}

	// Check fields
	if len(typed.NamedTypes[0].Fields) != 2 {
		t.Errorf("ItemsItem fields count = %d, want 2", len(typed.NamedTypes[0].Fields))
	}
}

func TestResolve_WithoutParams(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"Name": {
				Name: "Name",
				Kind: scan.KindString,
			},
		},
	}

	templateSrc := "{{ .Name }}"

	typed, err := Resolve(schema, templateSrc)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(typed.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(typed.Fields))
	}

	if typed.Fields["Name"].GoType != "string" {
		t.Errorf("Name.GoType = %q, want %q", typed.Fields["Name"].GoType, "string")
	}

	if len(typed.RequiredImports) != 0 {
		t.Errorf("expected no imports, got %d", len(typed.RequiredImports))
	}
}

func TestResolve_WithParamOverride(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"User": {
				Name: "User",
				Kind: scan.KindStruct,
				Children: map[string]*scan.Field{
					"Name": {
						Name: "Name",
						Kind: scan.KindString,
					},
					"Age": {
						Name: "Age",
						Kind: scan.KindString,
					},
				},
			},
		},
	}

	templateSrc := `
{{/* @param User.Age int */}}
{{ .User.Name }} is {{ .User.Age }} years old.
`

	typed, err := Resolve(schema, templateSrc)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// User.Age should be overridden to int
	if typed.Fields["User"].Children["Age"].GoType != "int" {
		t.Errorf("User.Age.GoType = %q, want %q", typed.Fields["User"].Children["Age"].GoType, "int")
	}

	// User.Name should remain string
	if typed.Fields["User"].Children["Name"].GoType != "string" {
		t.Errorf("User.Name.GoType = %q, want %q", typed.Fields["User"].Children["Name"].GoType, "string")
	}
}

func TestResolve_WithSliceStructOverride(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"Items": {
				Name: "Items",
				Kind: scan.KindSlice,
				Elem: &scan.Field{
					Name: "Items",
					Kind: scan.KindStruct,
					Children: map[string]*scan.Field{
						"ID": {
							Name: "ID",
							Kind: scan.KindString,
						},
					},
				},
			},
		},
	}

	templateSrc := `
{{/* @param Items []struct{ID int64; Title string} */}}
{{ range .Items }}{{ .ID }}: {{ .Title }}{{ end }}
`

	typed, err := Resolve(schema, templateSrc)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Items should be []ItemsItem
	if typed.Fields["Items"].GoType != "[]ItemsItem" {
		t.Errorf("Items.GoType = %q, want %q", typed.Fields["Items"].GoType, "[]ItemsItem")
	}

	// Check named type
	found := false
	for _, nt := range typed.NamedTypes {
		if nt.Name == "ItemsItem" {
			found = true
			if nt.Fields["ID"].GoType != "int64" {
				t.Errorf("ItemsItem.ID.GoType = %q, want %q", nt.Fields["ID"].GoType, "int64")
			}
			if nt.Fields["Title"].GoType != "string" {
				t.Errorf("ItemsItem.Title.GoType = %q, want %q", nt.Fields["Title"].GoType, "string")
			}
		}
	}

	if !found {
		t.Error("ItemsItem named type not found")
	}
}

func TestResolve_InvalidParamDirective(t *testing.T) {
	schema := scan.Schema{
		Fields: map[string]*scan.Field{
			"Name": {
				Name: "Name",
				Kind: scan.KindString,
			},
		},
	}

	templateSrc := `
{{/* @param Name struct{Field */}}
{{ .Name }}
`

	_, err := Resolve(schema, templateSrc)
	if err == nil {
		t.Error("expected error for invalid param directive, got nil")
	}
}
