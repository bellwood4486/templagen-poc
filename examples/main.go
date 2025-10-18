package main

import (
	"bytes"
	"fmt"

	"github.com/bellwood4486/templagen-poc/examples/mailtpl"
	"github.com/bellwood4486/templagen-poc/examples/usertpl"
)

func main() {
	// Example 1: Using RenderAny with map[string]any
	fmt.Println("=== Example 1: RenderAny (dynamic) ===")
	var buf1 bytes.Buffer
	_ = mailtpl.RenderAny(&buf1, map[string]any{
		"User":    map[string]any{"Name": "Alice"},
		"Message": "Welcome!",
	})
	fmt.Println(buf1.String())

	// Example 2: Using Render with type-safe Params
	fmt.Println("=== Example 2: Render (type-safe) ===")
	var buf2 bytes.Buffer
	_ = mailtpl.Render(&buf2, mailtpl.Params{
		User:    mailtpl.User{Name: "Bob"},
		Message: "Hello from type-safe params!",
	})
	fmt.Println(buf2.String())

	// Example 3: Using @param with custom types
	fmt.Println("=== Example 3: With @param type override ===")
	var buf3 bytes.Buffer
	_ = usertpl.Render(&buf3, usertpl.Params{
		User: usertpl.User{
			Name:  "Charlie",
			Age:   30,
			Email: strPtr("charlie@example.com"),
		},
		Items: []usertpl.ItemsItem{
			{ID: 1, Title: "First Item", Price: 99.99},
			{ID: 2, Title: "Second Item", Price: 149.99},
		},
	})
	fmt.Println(buf3.String())
}

func strPtr(s string) *string {
	return &s
}
