# Example 06: Non-ASCII Filename

This example demonstrates how templagen handles template files with non-ASCII filenames (e.g., Japanese).

## Template File

- `templates/メール.tmpl` - A template file with a Japanese filename

## Generated Code

From the Japanese filename `メール.tmpl`, templagen generates:

- **Type name**: `メール`
- **Render function**: `Renderメール(w io.Writer, p メール) error`
- **Embedded variable**: `メールTplSource`

Go supports Unicode identifiers, so these Japanese names work perfectly in Go code.

## Usage

```go
var buf bytes.Buffer
_ = Renderメール(&buf, メール{
    Name: "田中太郎",
})
fmt.Println(buf.String())
```

## Running the Example

```bash
go generate
go run .
```
