# Example 07: Template Grouping

This example demonstrates **template grouping** - organizing templates logically in subdirectories with nested namespaces.

## Overview

tmpltype supports organizing templates into groups using subdirectories. This allows you to structure related templates together while generating type-safe code with nested namespaces. You can also mix flat templates with grouped templates in the same project.

**Purpose:** This example shows how to organize templates in subdirectories and demonstrates the generated nested namespace structure.

> **📖 For template grouping documentation, see the [main README](../../README.md#template-grouping).**

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

The example will render templates from both flat and grouped structures, demonstrating how to use both approaches.

## What's Included

This example demonstrates template grouping with a mixed structure:

### Flat Template
- `footer.tmpl` - A standalone template for page footers

### Grouped Templates
- `01_mail_invite/` - Invitation email templates
  - `title.tmpl` - Email subject line
  - `content.tmpl` - Email body
- `02_mail_account_created/` - Account creation email templates
  - `title.tmpl` - Email subject line
  - `content.tmpl` - Email body
- `03_mail_article_created/` - Article notification email templates
  - `title.tmpl` - Email subject line
  - `content.tmpl` - Email body

## Running the Example

1. Generate the code:
```bash
go generate
```

2. Run the example:
```bash
go run .
```

The output will show:
- Mail invite templates (title and content)
- Mail account created templates (title and content)
- Mail article created template (using generic Render)
- Footer template (flat structure)
- List of all available templates

## What Gets Generated

The `go generate` command creates `template_gen.go` containing:
- Type-safe struct definitions for each template
- Render functions for each template (e.g., `RenderMailInviteTitle()`, `RenderFooter()`)
- A nested `Template` struct with grouped namespaces
- Generic `Render()` function for dynamic template selection
- `Templates()` function to list all available templates

## File Structure

```
07_grouping/
├── gen.go              # go:generate directive
├── main.go             # Example usage
├── README.md           # This file
├── template_gen.go     # Generated code (created by go generate)
└── templates/
    ├── footer.tmpl                      # Flat template
    ├── 01_mail_invite/                  # Group
    │   ├── title.tmpl
    │   └── content.tmpl
    ├── 02_mail_account_created/         # Group
    │   ├── title.tmpl
    │   └── content.tmpl
    └── 03_mail_article_created/         # Group
        ├── title.tmpl
        └── content.tmpl
```

## Template Organization

### Why Template Grouping?

Template grouping provides several benefits:

✅ **Logical organization**: Group related templates together (e.g., all email templates for a specific event)
✅ **Namespace management**: Avoid name conflicts by grouping templates in subdirectories
✅ **Easier navigation**: Find templates quickly in a well-organized structure
✅ **Flexibility**: Mix flat and grouped templates as needed

### Code Generation

Templates are generated with a single command:
```go
//go:generate sh -c "go run ../../cmd/tmpltype -in $(echo templates/*.tmpl templates/*/*.tmpl | tr ' ' ',') -pkg main -out template_gen.go"
```

This generates:
- Type-safe render functions for all templates
- Nested namespace structure for grouped templates
- Flat structure for standalone templates

## Understanding Generated Code

### Nested Template Namespace

The generator creates a nested struct to organize template names:

```go
var Template = struct {
    Footer             TemplateName  // Flat template
    MailInvite struct {              // Group: 01_mail_invite/
        Title   TemplateName
        Content TemplateName
    }
    MailAccountCreated struct {      // Group: 02_mail_account_created/
        Title   TemplateName
        Content TemplateName
    }
    MailArticleCreated struct {      // Group: 03_mail_article_created/
        Title   TemplateName
        Content TemplateName
    }
}{
    Footer: "footer",
    MailInvite: struct {
        Title   TemplateName
        Content TemplateName
    }{
        Title:   "01_mail_invite/title",
        Content: "01_mail_invite/content",
    },
    // ... (other groups)
}
```

### Type-Safe Render Functions

Each template gets its own type-safe render function:

| Template | Generated Type | Render Function |
|----------|----------------|-----------------|
| `footer.tmpl` | `Footer` | `RenderFooter()` |
| `01_mail_invite/title.tmpl` | `MailInviteTitle` | `RenderMailInviteTitle()` |
| `01_mail_invite/content.tmpl` | `MailInviteContent` | `RenderMailInviteContent()` |
| `02_mail_account_created/title.tmpl` | `MailAccountCreatedTitle` | `RenderMailAccountCreatedTitle()` |
| `02_mail_account_created/content.tmpl` | `MailAccountCreatedContent` | `RenderMailAccountCreatedContent()` |
| `03_mail_article_created/title.tmpl` | `MailArticleCreatedTitle` | `RenderMailArticleCreatedTitle()` |
| `03_mail_article_created/content.tmpl` | `MailArticleCreatedContent` | `RenderMailArticleCreatedContent()` |

### Usage Examples

**Type-safe rendering (recommended):**
```go
var buf bytes.Buffer
_ = RenderMailInviteTitle(&buf, MailInviteTitle{
    SiteName:    "MyApp",
    InviterName: "Alice",
})
```

**Generic rendering (for dynamic template selection):**
```go
var buf bytes.Buffer
_ = Render(&buf, Template.MailInvite.Title, MailInviteTitle{
    SiteName:    "MyApp",
    InviterName: "Alice",
})
```

### Naming Patterns

The code generator follows these naming patterns:

| Pattern | Example | Generated Name |
|---------|---------|----------------|
| Flat template | `footer.tmpl` | `Footer` (type), `RenderFooter()` (function) |
| Grouped template | `01_mail_invite/title.tmpl` | `MailInviteTitle` (type), `RenderMailInviteTitle()` (function) |
| Template path | `01_mail_invite/title.tmpl` | `Template.MailInvite.Title` (constant) |

Note: Numeric prefixes (e.g., `01_`, `02_`) are removed from generated names for cleaner identifiers.

## Key Takeaways

✅ **Use this example to:**
- Learn how to organize templates in subdirectories
- Understand how grouped templates generate nested namespaces
- See how to mix flat and grouped templates
- Learn both type-safe and generic rendering approaches

✅ **Best practices demonstrated:**
- Group related templates in subdirectories
- Use numeric prefixes for ordering templates in directories
- Mix flat and grouped templates as needed
- Use type-safe render functions for compile-time safety
- Use generic `Render()` for dynamic template selection

📖 **For more information:**
- Template grouping: [Main README - Template Grouping](../../README.md#template-grouping)
- Multiple template management: [Example 3 - Multi Template](../03_multi_template/)
- Basic template usage: [Example 1 - Basic](../01_basic/)
