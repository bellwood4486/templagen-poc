package main

import (
	"bytes"
	"fmt"
)

func main() {
	fmt.Println("=== Example: @param directive ===")
	var buf bytes.Buffer
	_ = RenderUser(&buf, User{
		User: UserUser{
			Name:  "Charlie",
			Age:   30,
			Email: strPtr("charlie@example.com"),
		},
		Items: []UserItemsItem{
			{ID: 1, Title: "First Item", Price: 99.99},
			{ID: 2, Title: "Second Item", Price: 149.99},
		},
	})
	fmt.Println(buf.String())
}

func strPtr(s string) *string {
	return &s
}
