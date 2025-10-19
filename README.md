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

## `@param` Directive Reference

The `@param` directive allows you to explicitly specify types for template parameters, overriding automatic type inference. This is essential for complex types like specific integer sizes, optional fields (pointers), and structured data.

### Syntax

```go
{{/* @param <FieldPath> <Type> */}}
```

- `<FieldPath>`: Dot-separated field path (e.g., `User.Name`, `Items`, `Config.Database.Host`)
- `<Type>`: Go type expression (see supported types below)

### Supported Types

#### ✅ Fully Supported

**1. Basic Types**
```go
{{/* @param Name string */}}
{{/* @param Age int */}}
{{/* @param Count int64 */}}
{{/* @param Price float64 */}}
{{/* @param Active bool */}}
{{/* @param CreatedAt time.Time */}}  // Automatically imports "time"
```

Supported base types: `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `bool`, `byte`, `rune`, `any`, `time.Time`

**2. Pointer Types (Optional/Nullable)**
```go
{{/* @param Email *string */}}
{{/* @param Score *int */}}
{{/* @param Discount *float64 */}}
```

Any base type can be wrapped with `*` to make it optional.

**3. Slices**
```go
{{/* @param Tags []string */}}
{{/* @param IDs []int */}}
{{/* @param Prices []float64 */}}
```

**4. Maps**
```go
{{/* @param Metadata map[string]string */}}
{{/* @param Counters map[string]int */}}
{{/* @param Settings map[string]bool */}}
```

**Note:** Map keys must always be `string`. Other key types are not supported.

**5. Nested Struct Fields (Dot Notation)**
```go
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}
{{/* @param User.Email string */}}

{{/* @param Config.Database.Host string */}}
{{/* @param Config.Database.Port int */}}
```

Generates nested struct types:
```go
type All_typesUser struct {
    ID    int64
    Name  string
    Email string
}
```

**6. Slice of Structs**
```go
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
{{/* @param Records []struct{Name string; Tags []string; Score *int} */}}
```

Struct fields are separated by semicolons (`;`). Can include nested slices/maps within struct fields.

**7. Optional Slices**
```go
{{/* @param OptionalTags *[]string */}}
```

#### ❌ Known Limitations

**1. Nested Slices/Maps**
```go
// ❌ Does NOT work - generates invalid syntax
{{/* @param Matrix [][]string */}}
{{/* @param Groups map[string][]string */}}
{{/* @param Data []map[string]int */}}
```

**Workaround:** Use slice of structs:
```go
// ✅ Works
{{/* @param Groups []struct{Key string; Values []string} */}}
```

**2. Inline Struct Definitions at Top Level**
```go
// ❌ Does NOT work - generates invalid Go code
{{/* @param User struct{ID int64; Name string} */}}
```

**Workaround:** Use dot notation:
```go
// ✅ Works
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}
```

**3. Deeply Nested Paths with Inline Structs**
```go
// ❌ Does NOT work - generates type names with dots
{{/* @param Complex.Nested.User struct{ID int64; Name string} */}}
```

**Workaround:** Flatten the structure or use simpler field paths.

**4. Non-String Map Keys**
```go
// ❌ Not supported
{{/* @param Lookup map[int]string */}}
```

**5. Struct Field Syntax**
```go
// ❌ Wrong - commas not allowed
{{/* @param Item struct{Name string, ID int} */}}

// ✅ Correct - use semicolons
{{/* @param Item struct{Name string; ID int} */}}
```

### Best Practices

✅ **DO:**
- Use dot notation for nested structures: `User.Name`, `Config.Database.Host`
- Use `[]struct{...}` for collections of complex data
- Use pointer types (`*Type`) for optional fields
- Keep field paths relatively flat (1-2 levels deep)
- Use semicolons to separate struct fields

❌ **DON'T:**
- Don't use inline `struct{...}` at the top level
- Don't nest slices/maps directly (`[][]T`, `map[K][]V`)
- Don't combine deep field paths with inline struct definitions
- Don't use commas in struct field definitions

### Complete Example

See [`examples/05_all_param_types`](./examples/05_all_param_types) for a comprehensive example demonstrating all supported type patterns and limitations.

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

## Supported Template Syntax

`templagen` supports the following Go template syntax patterns. The template scanner analyzes these patterns to infer types automatically:

### 1. Basic Field Reference

```go
{{ .Title }}
```

Creates a `string` field in the generated struct.

### 2. Nested Field Reference

```go
{{ .User.Name }}
{{ .Author.Email }}
```

Creates nested struct types with `string` fields.

### 3. Conditional Statements (if)

```go
{{ if .Status }}
  <p>Status: {{ .Status }}</p>
{{ end }}
```

The field in the condition is inferred as a struct if it has child fields, otherwise as `string`.

### 4. With Statement and Else Clause

```go
{{ with .Summary }}
  <p>{{ .Content }}</p>
{{ else }}
  <p>{{ .DefaultMessage }}</p>
{{ end }}
```

Changes the dot (`.`) context within the block. The scanner tracks scope changes correctly.

### 5. Range Over Slice

```go
{{ range .Items }}
  <li>{{ .Title }} - {{ .ID }}</li>
{{ end }}
```

Infers `.Items` as a slice type `[]struct{...}` with fields from the range body.

### 6. Map Access with Index Function

```go
{{ index .Meta "key" }}
{{ index .Meta "env" }}
```

Infers `.Meta` as `map[string]string` when using the `index` function.

### 7. Nested Structures (with + range)

```go
{{ with .Project }}
  <h3>{{ .Name }}</h3>
  {{ range .Tasks }}
    <p>{{ .Title }}</p>
  {{ end }}
{{ end }}
```

Combines `with` and `range` to create nested struct hierarchies with slice fields.

### 8. Deep Nested Paths

```go
{{ .Company.Department.Team.Manager.Name }}
```

Creates deeply nested struct types following the full path.

### Complete Example

See [`examples/04_comprehensive_template`](./examples/04_comprehensive_template) for a complete template demonstrating all supported syntax patterns.

## Examples

Check the [`examples/`](./examples) directory for complete working examples:

- [`01_basic`](./examples/01_basic): Basic usage with type inference
- [`02_param_directive`](./examples/02_param_directive): Using `@param` directives for complex types
- [`03_multi_template`](./examples/03_multi_template): Processing multiple templates at once
- [`04_comprehensive_template`](./examples/04_comprehensive_template): Comprehensive example demonstrating all supported template syntax patterns
- [`05_all_param_types`](./examples/05_all_param_types): Complete reference for all supported `@param` types and limitations

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