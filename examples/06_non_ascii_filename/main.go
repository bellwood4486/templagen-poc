package main

import (
	"bytes"
	"fmt"
)

func main() {
	// Example: Using template with Japanese filename
	fmt.Println("=== Example: Template with Japanese filename (メール.tmpl) ===")
	var buf bytes.Buffer
	_ = Renderメール(&buf, メール{
		Name: "田中太郎",
	})
	fmt.Println(buf.String())
}
