package magic

import (
	"testing"
)

func TestParseType_BaseType(t *testing.T) {
	tests := []struct {
		input    string
		wantKind TypeKind
		wantBase string
	}{
		{"string", TypeKindBase, "string"},
		{"int", TypeKindBase, "int"},
		{"int64", TypeKindBase, "int64"},
		{"bool", TypeKindBase, "bool"},
		{"float64", TypeKindBase, "float64"},
		{"time.Time", TypeKindBase, "time.Time"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", got.Kind, tt.wantKind)
			}

			if got.BaseType != tt.wantBase {
				t.Errorf("BaseType = %q, want %q", got.BaseType, tt.wantBase)
			}
		})
	}
}

func TestParseType_Slice(t *testing.T) {
	tests := []struct {
		input    string
		wantElem string
	}{
		{"[]string", "string"},
		{"[]int", "int"},
		{"[]bool", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Kind != TypeKindSlice {
				t.Fatalf("Kind = %v, want %v", got.Kind, TypeKindSlice)
			}

			if got.Elem == nil {
				t.Fatal("Elem is nil")
			}

			if got.Elem.BaseType != tt.wantElem {
				t.Errorf("Elem.BaseType = %q, want %q", got.Elem.BaseType, tt.wantElem)
			}
		})
	}
}

func TestParseType_Map(t *testing.T) {
	tests := []struct {
		input    string
		wantElem string
	}{
		{"map[string]string", "string"},
		{"map[string]int", "int"},
		{"map[string]bool", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Kind != TypeKindMap {
				t.Fatalf("Kind = %v, want %v", got.Kind, TypeKindMap)
			}

			if got.Elem == nil {
				t.Fatal("Elem is nil")
			}

			if got.Elem.BaseType != tt.wantElem {
				t.Errorf("Elem.BaseType = %q, want %q", got.Elem.BaseType, tt.wantElem)
			}
		})
	}
}

func TestParseType_Pointer(t *testing.T) {
	tests := []struct {
		input    string
		wantElem string
	}{
		{"*string", "string"},
		{"*int", "int"},
		{"*bool", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Kind != TypeKindPointer {
				t.Fatalf("Kind = %v, want %v", got.Kind, TypeKindPointer)
			}

			if got.Elem == nil {
				t.Fatal("Elem is nil")
			}

			if got.Elem.BaseType != tt.wantElem {
				t.Errorf("Elem.BaseType = %q, want %q", got.Elem.BaseType, tt.wantElem)
			}
		})
	}
}

func TestParseType_Struct(t *testing.T) {
	tests := []struct {
		input      string
		wantFields int
	}{
		{"struct{Name string}", 1},
		{"struct{ID int64; Title string}", 2},
		{"struct{X int; Y int; Z int}", 3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Kind != TypeKindStruct {
				t.Fatalf("Kind = %v, want %v", got.Kind, TypeKindStruct)
			}

			if len(got.Fields) != tt.wantFields {
				t.Errorf("len(Fields) = %d, want %d", len(got.Fields), tt.wantFields)
			}
		})
	}
}

func TestParseType_Complex(t *testing.T) {
	tests := []struct {
		name string
		input string
	}{
		{"nested slice", "[][]int"},
		{"slice of map", "[]map[string]string"},
		{"map of slice", "map[string][]int"},
		{"pointer to slice", "*[]string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
