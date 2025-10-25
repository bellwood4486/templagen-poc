package main

import (
	"bytes"
	"fmt"
)

func main() {
	fmt.Println("=== Example: Multi-template support ===")

	// Use Templates() map function
	templates := Templates()
	fmt.Printf("Available templates: %d\n", len(templates))
	for name := range templates {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()

	// Use type-safe render functions
	var headerBuf bytes.Buffer
	_ = RenderHeader(&headerBuf, Header{
		Title:    "Multi-Template Demo",
		Subtitle: strPtr("Showing multiple templates in one package"),
	})
	fmt.Println("Header output:")
	fmt.Println(headerBuf.String())

	// Use generic Render function
	var navBuf bytes.Buffer
	_ = Render(&navBuf, Template.Nav, Nav{
		CurrentUser: NavCurrentUser{
			Name:    "Admin User",
			IsAdmin: true,
		},
		Items: []NavItemsItem{
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
