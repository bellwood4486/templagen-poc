# Example 5: All Param Types

This example demonstrates **all supported types** that can be used with the `@param` directive in templagen.

## Overview

The `@param` directive allows you to override the inferred types for template fields. This example showcases every supported type pattern, from basic types to complex nested structures.

**Purpose:** This example serves as a hands-on reference with working code demonstrating all `@param` type patterns.

> **📖 For complete `@param` directive documentation, see the [main README](../../README.md#param-directive-reference).**

## Table of Contents
- [Quick Start](#quick-start)
- [What's Included](#whats-included)
- [Running the Example](#running-the-example)
- [Understanding Generated Code](#understanding-generated-code)
- [Key Takeaways](#key-takeaways)

## Quick Start

```go
// Basic types
{{/* @param Name string */}}
{{/* @param Age int */}}

// Optional types (pointers)
{{/* @param Email *string */}}

// Collections
{{/* @param Tags []string */}}
{{/* @param Config map[string]string */}}

// Nested structures (use dot notation)
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}

// Slice of structs
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
```

```bash
# Generate and run
go generate
go run .
```

## What's Included

This example demonstrates all working `@param` patterns organized across 6 focused template files:

### Template Files

| Template | Patterns Demonstrated |
|----------|----------------------|
| `basic_types.tmpl` | `string`, `int`, `int64`, `float64`, `bool`, `time.Time` |
| `pointer_types.tmpl` | `*string`, `*int`, `*float64` (optional/nullable fields) |
| `slice_types.tmpl` | `[]string`, `[]int`, `[]float64`, `[]bool` |
| `map_types.tmpl` | `map[string]string`, `map[string]int`, `map[string]float64`, `map[string]bool` |
| `struct_types.tmpl` | Nested fields using dot notation (`User.ID`, `Product.Price`) |
| `complex_types.tmpl` | `[]struct{...}`, `*[]string`, structs with optional fields |

### ✅ Supported Patterns Demonstrated

1. **Basic Types**: `string`, `int`, `int64`, `float64`, `bool`, `time.Time`
2. **Pointer Types**: `*string`, `*int`, `*float64` (optional/nullable fields)
3. **Slices**: `[]string`, `[]int`, `[]float64`, `[]bool`
4. **Maps**: `map[string]string`, `map[string]int`, `map[string]float64`, `map[string]bool`
5. **Nested Struct Fields**: Using dot notation (`User.ID`, `Product.Price`)
6. **Slice of Structs**: `[]struct{ID int64; Title string; Tags []string}`
7. **Optional Slices**: `*[]string`
8. **Structs with Optional Fields**: `[]struct{Name string; Score *int}`

### ❌ Known Limitations (See Main README)

This example intentionally **avoids** patterns that don't work:
- ❌ Nested slices/maps: `[][]string`, `map[string][]string`
- ❌ Top-level inline structs: `struct{...}`
- ❌ Deep paths with inline structs: `A.B.C struct{...}`

For workarounds and detailed explanations, see the [main README](../../README.md#param-directive-reference).

## Running the Example

1. Generate the code:
```bash
go generate
```

2. Run the example:
```bash
go run .
```

## What Gets Generated

The `go generate` command creates `template_gen.go` containing:
- Type-safe struct definitions for each template (e.g., `Basic_types`, `Pointer_types`)
- Dedicated render functions for each template (e.g., `RenderBasic_types()`, `RenderPointer_types()`)
- Proper imports (including `time` for `time.Time`)
- Template map with all compiled templates

## File Structure

```
05_all_param_types/
├── gen.go              # go:generate directive
├── main.go             # Example usage with sample data
├── README.md           # This file
├── template_gen.go     # Generated code (created by go generate)
└── templates/
    ├── basic_types.tmpl     # Basic type @param directives
    ├── pointer_types.tmpl   # Pointer/optional type @param directives
    ├── slice_types.tmpl     # Slice type @param directives
    ├── map_types.tmpl       # Map type @param directives
    ├── struct_types.tmpl    # Nested struct @param directives
    └── complex_types.tmpl   # Complex/nested @param directives
```

## Understanding Generated Code

Each template file generates its own dedicated types and render function:

| Template File | Generated Type | Render Function |
|--------------|----------------|-----------------|
| `basic_types.tmpl` | `Basic_types` | `RenderBasic_types()` |
| `pointer_types.tmpl` | `Pointer_types` | `RenderPointer_types()` |
| `slice_types.tmpl` | `Slice_types` | `RenderSlice_types()` |
| `map_types.tmpl` | `Map_types` | `RenderMap_types()` |
| `struct_types.tmpl` | `Struct_types` | `RenderStruct_types()` |
| `complex_types.tmpl` | `Complex_types` | `RenderComplex_types()` |

### Naming Patterns

The code generator follows these naming patterns:

| Pattern | Example Input | Generated Type |
|---------|--------------|----------------|
| Main struct | `basic_types.tmpl` | `Basic_types` |
| Nested field | `@param User.Name string` | `Basic_typesUser` struct |
| Slice items | `@param Items []struct{...}` | `Basic_typesItemsItem` struct |

Example:
```go
// From struct_types.tmpl
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}

// Generated code
type Struct_typesUser struct {
    ID    int64
    Name  string
    Email string
}

type Struct_types struct {
    User    Struct_typesUser
    Product Struct_typesProduct
}
```

## Key Takeaways

✅ **Use this example to:**
- See working code for all supported `@param` patterns
- Understand how different types are generated
- Test and experiment with type specifications
- Reference specific type patterns quickly (organized by template file)

✅ **Benefits of the modular structure:**
- **Easy reference**: Jump directly to the type category you need
- **Focused learning**: Study one category at a time without distractions
- **Copy-friendly**: Easily copy specific patterns to your own projects
- **Maintainable**: Smaller, focused files are easier to understand and update

📖 **For complete documentation:**
- Type specifications: [Main README - `@param` Directive Reference](../../README.md#param-directive-reference)
- Limitations and workarounds: [Main README - Known Limitations](../../README.md#-known-limitations)
- Best practices: [Main README - Best Practices](../../README.md#best-practices)
