package main

import (
	"bytes"
	"fmt"

	"github.com/bellwood4486/templagen-poc/examples/mailtpl"
)

func main() {
	var buf bytes.Buffer
	_ = mailtpl.RenderAny(&buf, map[string]any{
		"User":    map[string]any{"Name": "Alice"},
		"Message": "Welcome!",
	})

	fmt.Println(buf.String())
}
