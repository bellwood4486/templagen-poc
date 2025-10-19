package multi_templates

import (
	"bytes"
	"strings"
	"testing"
)

func TestMultiTemplates(t *testing.T) {
	t.Run("RenderHeader", func(t *testing.T) {
		var buf bytes.Buffer
		subtitle := "Welcome to our site"
		err := RenderHeader(&buf, Header{
			Title:    "My Website",
			Subtitle: &subtitle,
		})
		if err != nil {
			t.Fatalf("RenderHeader failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "My Website") {
			t.Errorf("expected output to contain 'My Website', got: %s", output)
		}
		if !strings.Contains(output, "Welcome to our site") {
			t.Errorf("expected output to contain 'Welcome to our site', got: %s", output)
		}
	})

	t.Run("RenderFooter", func(t *testing.T) {
		var buf bytes.Buffer
		err := RenderFooter(&buf, Footer{
			Year:        2024,
			CompanyName: "Example Corp",
			Links: []FooterLinksItem{
				{Text: "About", URL: "/about"},
				{Text: "Contact", URL: "/contact"},
			},
		})
		if err != nil {
			t.Fatalf("RenderFooter failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2024") {
			t.Errorf("expected output to contain '2024', got: %s", output)
		}
		if !strings.Contains(output, "Example Corp") {
			t.Errorf("expected output to contain 'Example Corp', got: %s", output)
		}
	})

	t.Run("RenderNav", func(t *testing.T) {
		var buf bytes.Buffer
		err := RenderNav(&buf, Nav{
			CurrentUser: NavCurrentUser{
				Name:    "John Doe",
				IsAdmin: true,
			},
			Items: []NavItemsItem{
				{Name: "Home", Link: "/", Active: true},
				{Name: "Products", Link: "/products", Active: false},
			},
		})
		if err != nil {
			t.Fatalf("RenderNav failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "John Doe") {
			t.Errorf("expected output to contain 'John Doe', got: %s", output)
		}
		if !strings.Contains(output, "(Admin)") {
			t.Errorf("expected output to contain '(Admin)', got: %s", output)
		}
		if !strings.Contains(output, `class="active"`) {
			t.Errorf("expected output to contain 'class=\"active\"', got: %s", output)
		}
	})

	t.Run("Templates function", func(t *testing.T) {
		templates := Templates()
		if len(templates) != 3 {
			t.Errorf("expected 3 templates, got %d", len(templates))
		}

		expectedNames := []string{"header", "footer", "nav"}
		for _, name := range expectedNames {
			if _, ok := templates[name]; !ok {
				t.Errorf("expected template %q to exist", name)
			}
		}
	})

	t.Run("Generic Render function", func(t *testing.T) {
		var buf bytes.Buffer
		err := Render(&buf, "header", map[string]any{
			"Title":    "Test Title",
			"Subtitle": nil,
		})
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Test Title") {
			t.Errorf("expected output to contain 'Test Title', got: %s", output)
		}
	})

	t.Run("Generic Render with invalid template", func(t *testing.T) {
		var buf bytes.Buffer
		err := Render(&buf, "nonexistent", map[string]any{})
		if err == nil {
			t.Error("expected error for nonexistent template")
		}
		if !strings.Contains(err.Error(), "nonexistent") {
			t.Errorf("expected error message to mention 'nonexistent', got: %v", err)
		}
	})
}