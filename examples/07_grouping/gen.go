//go:build ignore

package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "../../cmd/tmpltype",
		"-in", "templates/*/*.tmpl",
		"-pkg", "main",
		"-out", "template_gen.go",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
