package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	in := flag.String("in", "", "input template file")
	pkg := flag.String("pkg", "", "output package name")
	out := flag.String("out", "", "output .go file path")
	flag.Parse()

	if *in == "" || *pkg == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: templagen --in <file> --pkg <name> --out <file>")
		os.Exit(2)
	}

	fmt.Printf("templagen called with: in=%s pkg=%s out=%s\n", *in, *pkg, *out)
}
