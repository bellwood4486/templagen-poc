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

This example demonstrates all working `@param` patterns in a single template:

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
- Type-safe struct definitions for all `@param` types
- `RenderAll_types()` function for type-safe rendering
- Proper imports (including `time` for `time.Time`)

## File Structure

```
05_all_param_types/
├── gen.go              # go:generate directive
├── main.go             # Example usage with sample data
├── README.md           # This file
├── template_gen.go     # Generated code (created by go generate)
└── templates/
    └── all_types.tmpl  # Template with @param directives
```

## Understanding Generated Code

The code generator follows these naming patterns:

| Pattern | Example Input | Generated Type |
|---------|--------------|----------------|
| Main struct | `all_types.tmpl` | `All_types` |
| Nested field | `@param User.Name string` | `All_typesUser` struct |
| Slice items | `@param Items []struct{...}` | `All_typesItemsItem` struct |

Example:
```go
// From template
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}

// Generated code
type All_typesUser struct {
    ID   int64
    Name string
}

type All_types struct {
    User All_typesUser
    // ...
}
```

## Key Takeaways

✅ **Use this example to:**
- See working code for all supported `@param` patterns
- Understand how different types are generated
- Test and experiment with type specifications

📖 **For complete documentation:**
- Type specifications: [Main README - `@param` Directive Reference](../../README.md#param-directive-reference)
- Limitations and workarounds: [Main README - Known Limitations](../../README.md#-known-limitations)
- Best practices: [Main README - Best Practices](../../README.md#best-practices)
