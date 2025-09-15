package scan_test

import (
	"testing"

	"github.com/bellwood4486/templagen-poc/internal/scan"
)

func TestScanTemplate_SimpleFieldsAndNested(t *testing.T) {
	src := `
{{ .User.Name }}	
{{ .Message }}
`
	sch, err := scan.ScanTemplate(src)
	if err != nil {
		t.Fatal(err)
	}

	user := getTop(t, sch, "User")
	assertKind(t, user, scan.KindStruct)
	name := getChild(t, user, "Name")
	assertKind(t, name, scan.KindString)

	msg := getTop(t, sch, "Message")
	assertKind(t, msg, scan.KindString)
}

func getTop(t *testing.T, s scan.Schema, name string) *scan.Field {
	t.Helper()
	f := s.Fields[name]
	if f == nil {
		t.Fatalf("top-level field %q not found", name)
	}
	return f
}

func getChild(t *testing.T, f *scan.Field, name string) *scan.Field {
	t.Helper()
	if f.Children == nil {
		t.Fatalf("field %q has no children", f.Name)
	}
	ch, ok := f.Children[name]
	if !ok || ch == nil {
		t.Fatalf("child %q not found under %q", name, f.Name)
	}
	return ch
}

func assertKind(t *testing.T, got *scan.Field, want scan.Kind) {
	t.Helper()
	if got.Kind != want {
		t.Fatalf("kind mismatch: got=%v want=%v", got.Kind, want)
	}
}
