package main

import (
	"bytes"
	"fmt"
)

func main() {
	// Example 1: Using generic Render with map[string]any
	fmt.Println("=== Example 1: Render (dynamic) ===")
	var buf1 bytes.Buffer
	_ = Render(&buf1, Template.Email, map[string]any{
		"User":    map[string]any{"Name": "Alice"},
		"Message": "Welcome!",
	})
	fmt.Println(buf1.String())

	// Example 2: Using type-safe RenderEmail
	fmt.Println("=== Example 2: RenderEmail (type-safe) ===")
	var buf2 bytes.Buffer
	_ = RenderEmail(&buf2, Email{
		User:    EmailUser{Name: "Bob"},
		Message: "Hello from type-safe params!",
	})
	fmt.Println(buf2.String())
}
