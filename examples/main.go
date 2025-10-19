package main

import (
	"bytes"
	"fmt"

	"github.com/bellwood4486/templagen-poc/examples/mailtpl"
	"github.com/bellwood4486/templagen-poc/examples/multi_templates"
	"github.com/bellwood4486/templagen-poc/examples/usertpl"
)

func main() {
	// Example 1: Using generic Render with map[string]any
	fmt.Println("=== Example 1: Render (dynamic) ===")
	var buf1 bytes.Buffer
	_ = mailtpl.Render(&buf1, "email", map[string]any{
		"User":    map[string]any{"Name": "Alice"},
		"Message": "Welcome!",
	})
	fmt.Println(buf1.String())

	// Example 2: Using type-safe RenderEmail
	fmt.Println("=== Example 2: RenderEmail (type-safe) ===")
	var buf2 bytes.Buffer
	_ = mailtpl.RenderEmail(&buf2, mailtpl.Email{
		User:    mailtpl.EmailUser{Name: "Bob"},
		Message: "Hello from type-safe params!",
	})
	fmt.Println(buf2.String())

	// Example 3: Using @param with custom types
	fmt.Println("=== Example 3: With @param type override ===")
	var buf3 bytes.Buffer
	_ = usertpl.RenderUser(&buf3, usertpl.User{
		User: usertpl.UserUser{
			Name:  "Charlie",
			Age:   30,
			Email: strPtr("charlie@example.com"),
		},
		Items: []usertpl.UserItemsItem{
			{ID: 1, Title: "First Item", Price: 99.99},
			{ID: 2, Title: "Second Item", Price: 149.99},
		},
	})
	fmt.Println(buf3.String())

	// Example 4: Multi-template support
	fmt.Println("=== Example 4: Multi-template support ===")

	// Use Templates() map function
	templates := multi_templates.Templates()
	fmt.Printf("Available templates: %d\n", len(templates))
	for name := range templates {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()

	// Use type-safe render functions
	var headerBuf bytes.Buffer
	_ = multi_templates.RenderHeader(&headerBuf, multi_templates.Header{
		Title:    "Multi-Template Demo",
		Subtitle: strPtr("Showing multiple templates in one package"),
	})
	fmt.Println("Header output:")
	fmt.Println(headerBuf.String())

	// Use generic Render function
	var navBuf bytes.Buffer
	_ = multi_templates.Render(&navBuf, "nav", multi_templates.Nav{
		CurrentUser: multi_templates.NavCurrentUser{
			Name:    "Admin User",
			IsAdmin: true,
		},
		Items: []multi_templates.NavItemsItem{
			{Name: "Dashboard", Link: "/dashboard", Active: true},
			{Name: "Settings", Link: "/settings", Active: false},
		},
	})
	fmt.Println("Nav output:")
	fmt.Println(navBuf.String())
}

func strPtr(s string) *string {
	return &s
}
