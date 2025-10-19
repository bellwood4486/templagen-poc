# templagen

A Go code generator that creates type-safe template rendering functions from Go template files.

## Overview

`templagen` analyzes your Go template files, automatically infers or uses explicit parameter types, and generates type-safe Go code with structs and render functions. This eliminates runtime type errors and provides IDE autocompletion for template parameters.

## Features

- **Type Inference**: Automatically infers parameter types from template syntax (e.g., `.User.Name` → `string`)
- **Explicit Type Directives**: Support for `@param` directives to specify complex types
- **Type Safety**: Generate strongly-typed structs and render functions
- **Multiple Templates**: Process single or multiple template files at once
- **go generate Integration**: Seamlessly integrates with Go's code generation workflow
- **Flexible Rendering**: Provides both type-safe and dynamic rendering options

## Installation

```bash
go install github.com/bellwood4486/templagen-poc/cmd/templagen@latest
```

Or add to your project:

```bash
go get github.com/bellwood4486/templagen-poc
```

## Usage

### Basic Example

1. Create a template file `templates/email.tmpl`:

```html
<h1>Hello {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
```

2. Add a generate directive to your Go file `gen.go`:

```go
package main

//go:generate templagen -in templates/email.tmpl -pkg main -out template_gen.go
```

3. Run code generation:

```bash
go generate
```

4. Use the generated type-safe code:

```go
package main

import (
    "bytes"
    "fmt"
)

func main() {
    var buf bytes.Buffer
    _ = RenderEmail(&buf, Email{
        User:    EmailUser{Name: "Alice"},
        Message: "Welcome!",
    })
    fmt.Println(buf.String())
}
```

### Advanced Example: Type Directives

For complex types, use `@param` directives in your template `templates/user.tmpl`:

```html
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
<div class="user-profile">
  <h1>{{ .User.Name }}</h1>
  <p>Age: {{ .User.Age }}</p>
  {{ if .User.Email }}<p>Email: {{ .User.Email }}</p>{{ end }}
</div>

<div class="items">
  <h2>Items</h2>
  <ul>
  {{ range .Items }}
    <li>#{{ .ID }}: {{ .Title }} - ${{ .Price }}</li>
  {{ end }}
  </ul>
</div>
```

### Multiple Templates

Process multiple template files at once:

```go
//go:generate templagen -in "templates/*.tmpl" -pkg main -out templates_gen.go
```

Or specify files explicitly:

```go
//go:generate templagen -in "header.tmpl,footer.tmpl,nav.tmpl" -pkg main -out templates_gen.go
```

## Command Line Options

```
templagen -in <pattern> -pkg <name> -out <file> [-exclude <pattern>]

Options:
  -in string
        Input pattern (glob supported, e.g., "*.tmpl" or "templates/*.tmpl")
        Multiple files can be specified with comma separation
  -pkg string
        Output package name
  -out string
        Output .go file path
  -exclude string
        Exclude pattern (optional, applied to file basenames)
```

## How It Works

1. **Scan**: Parse template files and extract field access patterns (e.g., `.User.Name`, `.Items[0].ID`)
2. **Type Resolution**:
   - Apply explicit `@param` type directives
   - Infer types from template syntax (string for simple fields, infer collections from `range`)
3. **Code Generation**: Generate:
   - Type-safe parameter structs
   - Template parsing functions
   - Type-safe `Render<TemplateName>()` functions
   - Generic `Render()` function for dynamic use cases

## Generated Code Structure

For each template, `templagen` generates:

```go
// Parameter struct (type-safe)
type Email struct {
    User    EmailUser
    Message string
}

// Template function
func EmailTemplate() *template.Template { ... }

// Type-safe render function
func RenderEmail(w io.Writer, params Email) error { ... }

// Generic render function (for dynamic use)
func Render(w io.Writer, name string, params map[string]any) error { ... }
```

## Examples

Check the [`examples/`](./examples) directory for complete working examples:

- [`01_basic`](./examples/01_basic): Basic usage with type inference
- [`02_param_directive`](./examples/02_param_directive): Using `@param` directives for complex types
- [`03_multi_template`](./examples/03_multi_template): Processing multiple templates at once

Run examples:

```bash
cd examples/01_basic
go generate
go run .
```

## Project Structure

```
.
├── cmd/templagen/          # CLI tool entry point
├── internal/
│   ├── gen/               # Code generation logic
│   ├── scan/              # Template scanning and parsing
│   ├── typing/            # Type inference and resolution
│   │   └── magic/         # Magic comment (@param) parsing
│   └── util/              # Utility functions
└── examples/              # Usage examples
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development

1. Clone the repository:
```bash
git clone https://github.com/bellwood4486/templagen-poc.git
cd templagen-poc
```

2. Run tests:
```bash
go test ./...
```

3. Build:
```bash
go build ./cmd/templagen
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with Go's [`text/template`](https://pkg.go.dev/text/template) and [`html/template`](https://pkg.go.dev/html/template) packages.