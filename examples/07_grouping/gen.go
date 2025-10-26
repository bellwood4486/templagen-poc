package main

//go:generate sh -c "go run ../../cmd/tmpltype -in $(echo templates/*.tmpl templates/*/*.tmpl | tr ' ' ',') -pkg main -out template_gen.go"
