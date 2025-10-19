# Example 4: Comprehensive Template Features

This example demonstrates **comprehensive Go template features** organized across multiple focused template files.

## Overview

This example showcases the major features of Go's `text/template` package in a practical way. Instead of cramming everything into a single large template, it demonstrates how to organize template features into separate, manageable files.

**Purpose:** This example serves as a hands-on reference for understanding Go template capabilities and how to structure multiple templates effectively.

> **📖 For supported template syntax, see the [main README](../../README.md#supported-template-syntax).**

## Table of Contents
- [Quick Start](#quick-start)
- [What's Included](#whats-included)
- [Running the Example](#running-the-example)
- [Template Organization](#template-organization)
- [Understanding Generated Code](#understanding-generated-code)
- [Key Takeaways](#key-takeaways)

## Quick Start

```bash
# Generate and run
go generate
go run .
```

The example will render 4 HTML templates, each demonstrating specific template features.

## What's Included

This example demonstrates 8 key template features organized across 4 template files:

### Template 1: `basic_fields.tmpl`
- ✅ **Basic field reference**: `{{ .Title }}`
- ✅ **Nested field reference**: `{{ .Author.Name }}`, `{{ .Author.Email }}`

### Template 2: `control_flow.tmpl`
- ✅ **Conditional rendering**: `{{ if .Status }}...{{ end }}`
- ✅ **With statement and else clause**: `{{ with .Summary }}...{{ else }}...{{ end }}`

### Template 3: `collections.tmpl`
- ✅ **Range over slice**: `{{ range .Items }}...{{ end }}`
- ✅ **Map access with index function**: `{{ index .Meta "key" }}`

### Template 4: `advanced.tmpl`
- ✅ **Nested structures (with + range)**: `{{ with .Project }}{{ range .Tasks }}...{{ end }}{{ end }}`
- ✅ **Deep nested path**: `{{ .Company.Department.Team.Manager.Name }}`

## Running the Example

1. Generate the code:
```bash
go generate
```

2. Run the example:
```bash
go run .
```

The output will show all 4 templates rendered with sample data, demonstrating each feature in action.

## What Gets Generated

The `go generate` command creates `template_gen.go` containing:
- Type-safe struct definitions for each template
- `RenderBasic_fields()`, `RenderControl_flow()`, `RenderCollections()`, `RenderAdvanced()` functions
- Template map with all compiled templates

## File Structure

```
04_comprehensive_template/
├── gen.go              # go:generate directive
├── main.go             # Example usage with sample data
├── README.md           # This file
├── template_gen.go     # Generated code (created by go generate)
└── templates/
    ├── basic_fields.tmpl    # Basic & nested field references
    ├── control_flow.tmpl    # If, with, else statements
    ├── collections.tmpl     # Range, map access
    └── advanced.tmpl        # Complex nested structures
```

## Template Organization

### Why Multiple Templates?

This example demonstrates the benefits of organizing template features into separate files:

✅ **Focused learning**: Each template demonstrates 2 related features
✅ **Easier reference**: Find specific feature examples quickly
✅ **Maintainability**: Smaller files are easier to understand and modify
✅ **Real-world pattern**: Shows how to structure multiple templates in a project

### Single Template Generation

All templates are generated in one go:
```go
//go:generate go run ../../cmd/templagen -in "templates/*.tmpl" -pkg main -out template_gen.go
```

This generates type-safe functions for all templates in the `templates/` directory.

## Understanding Generated Code

The code generator creates separate types for each template:

| Template File | Generated Type | Render Function |
|--------------|----------------|-----------------|
| `basic_fields.tmpl` | `Basic_fields` | `RenderBasic_fields()` |
| `control_flow.tmpl` | `Control_flow` | `RenderControl_flow()` |
| `collections.tmpl` | `Collections` | `RenderCollections()` |
| `advanced.tmpl` | `Advanced` | `RenderAdvanced()` |

Each template gets its own dedicated parameter type with only the fields it needs.

### Naming Patterns

For nested structures, the generator uses prefixes:

```go
// From advanced.tmpl
{{ .Project.Name }}
{{ .Project.Tasks }}

// Generated code
type AdvancedProject struct {
    Name  string
    Tasks []AdvancedTasksItem
}

type AdvancedTasksItem struct {
    Title  string
    Status string
}

type Advanced struct {
    Project AdvancedProject
    Company AdvancedCompany
}
```

## Key Takeaways

✅ **Use this example to:**
- Learn Go template features through focused, practical examples
- Understand how to organize multiple templates in a project
- See how templagen generates type-safe code for multiple templates

✅ **Best practices demonstrated:**
- Separate templates by feature/concern
- Use wildcards (`*.tmpl`) to generate from multiple templates
- Keep templates focused and manageable

📖 **For more information:**
- Template syntax support: [Main README - Supported Template Syntax](../../README.md#supported-template-syntax)
- Parameter type customization: [Example 2 - Param Directive](../02_param_directive/)
- Multiple template management: [Example 3 - Multi Template](../03_multi_template/)
