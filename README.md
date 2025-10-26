# tmpltype

**[English version below](#english)** | **[英語版は下にあります](#english)**

---

## 日本語

Go テンプレートファイルから型安全なテンプレート描画関数を生成する Go コードジェネレータです。

### 概要

`tmpltype` は Go テンプレートファイルを解析し、パラメータの型を自動推論または明示的な型指定により判定して、型安全な Go コードを生成します。これにより実行時の型エラーを排除し、テンプレートパラメータに対する IDE の自動補完機能を提供します。

### 特徴

- **型推論**: テンプレート構文からパラメータの型を自動推論（例: `.User.Name` → `string`）
- **明示的な型ディレクティブ**: `@param` ディレクティブによる複雑な型の指定をサポート
- **型安全性**: 強く型付けされた構造体と描画関数を生成
- **テンプレートのグループ化**: サブディレクトリでテンプレートを論理的にグループ化し、ネストされた名前空間を生成
- **複数テンプレート**: 単一または複数のテンプレートファイルを一度に処理
- **go generate 統合**: Go のコード生成ワークフローにシームレスに統合
- **柔軟な描画**: 型安全な描画と動的な描画の両方のオプションを提供

### インストール

```bash
go install github.com/bellwood4486/tmpltype/cmd/tmpltype@latest
```

プロジェクトに追加する場合:

```bash
go get github.com/bellwood4486/tmpltype
```

### 使い方

#### 基本的な例

1. テンプレートファイル `templates/email.tmpl` を作成:

```html
<h1>Hello {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
```

2. Go ファイル `gen.go` に generate ディレクティブを追加:

```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
```

3. コード生成を実行:

```bash
go generate
```

4. 生成された型安全なコードを使用:

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

#### 高度な例: 型ディレクティブ

複雑な型の場合、テンプレート `templates/user.tmpl` で `@param` ディレクティブを使用します:

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

#### 複数テンプレート

`-dir`オプションを指定すると、ディレクトリ内のすべての`.tmpl`ファイルが自動的に処理されます:

```go
//go:generate tmpltype -dir templates -pkg main -out templates_gen.go
```

このコマンドは以下を自動的にスキャンします:
- `templates/*.tmpl` (フラットなテンプレート)
- `templates/*/*.tmpl` (グループ化されたテンプレート、1階層のみ)

#### テンプレートのグループ化

サブディレクトリでテンプレートを論理的にグループ化:

```
templates/
├── footer.tmpl                  # フラットなテンプレート
├── mail_invite/                 # グループ
│   ├── title.tmpl
│   └── content.tmpl
└── mail_account_created/        # グループ
    ├── title.tmpl
    └── content.tmpl
```

生成されるコード:

```go
var Template = struct {
    Footer             TemplateName  // フラット
    MailInvite struct {              // グループ
        Title   TemplateName
        Content TemplateName
    }
    MailAccountCreated struct {      // グループ
        Title   TemplateName
        Content TemplateName
    }
}

// 使用例
RenderMailInviteTitle(w, MailInviteTitle{...})
Render(w, Template.MailInvite.Title, data)
```

### `@param` ディレクティブリファレンス

`@param` ディレクティブを使用すると、テンプレートパラメータの型を明示的に指定でき、自動型推論を上書きできます。これは特定の整数サイズ、オプショナルフィールド（ポインタ）、構造化データなどの複雑な型に不可欠です。

#### 構文

```go
{{/* @param <フィールドパス> <型> */}}
```

- `<フィールドパス>`: ドット区切りのフィールドパス（例: `User.Name`, `Items`, `Config.Database.Host`）
- `<型>`: Go の型表現（サポートされる型は下記参照）

#### サポートされる型

##### ✅ 完全サポート

**1. 基本型**
```go
{{/* @param Name string */}}
{{/* @param Age int */}}
{{/* @param Count int64 */}}
{{/* @param Price float64 */}}
{{/* @param Active bool */}}
```

サポートされる基本型: `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `bool`, `byte`, `rune`, `any`

**2. ポインタ型（オプショナル/Null許可）**
```go
{{/* @param Email *string */}}
{{/* @param Score *int */}}
{{/* @param Discount *float64 */}}
```

任意の基本型を `*` でラップしてオプショナルにできます。

**3. スライス**
```go
{{/* @param Tags []string */}}
{{/* @param IDs []int */}}
{{/* @param Prices []float64 */}}
```

**4. マップ**
```go
{{/* @param Metadata map[string]string */}}
{{/* @param Counters map[string]int */}}
{{/* @param Settings map[string]bool */}}
```

**注意:** マップのキーは常に `string` である必要があります。他のキー型はサポートされていません。

**5. ネストされた構造体フィールド（ドット記法）**
```go
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}
{{/* @param User.Email string */}}

{{/* @param Config.Database.Host string */}}
{{/* @param Config.Database.Port int */}}
```

ネストされた構造体型を生成:
```go
type All_typesUser struct {
    ID    int64
    Name  string
    Email string
}
```

**6. 構造体のスライス**
```go
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
{{/* @param Records []struct{Name string; Tags []string; Score *int} */}}
```

構造体フィールドはセミコロン (`;`) で区切ります。構造体フィールド内にネストされたスライス/マップを含めることができます。

**7. オプショナルスライス**
```go
{{/* @param OptionalTags *[]string */}}
```

##### ❌ 既知の制限事項

**1. ネストされたスライス/マップ**
```go
// ❌ 動作しません - 無効な構文を生成
{{/* @param Matrix [][]string */}}
{{/* @param Groups map[string][]string */}}
{{/* @param Data []map[string]int */}}
```

**回避策:** 構造体のスライスを使用:
```go
// ✅ 動作します
{{/* @param Groups []struct{Key string; Values []string} */}}
```

**2. トップレベルでのインライン構造体定義**
```go
// ❌ 動作しません - 無効な Go コードを生成
{{/* @param User struct{ID int64; Name string} */}}
```

**回避策:** ドット記法を使用:
```go
// ✅ 動作します
{{/* @param User.ID int64 */}}
{{/* @param User.Name string */}}
```

**3. インライン構造体を持つ深くネストされたパス**
```go
// ❌ 動作しません - ドットを含む型名を生成
{{/* @param Complex.Nested.User struct{ID int64; Name string} */}}
```

**回避策:** 構造をフラット化するか、よりシンプルなフィールドパスを使用してください。

**4. 文字列以外のマップキー**
```go
// ❌ サポートされていません
{{/* @param Lookup map[int]string */}}
```

**5. 構造体フィールド構文**
```go
// ❌ 間違い - カンマは使用できません
{{/* @param Item struct{Name string, ID int} */}}

// ✅ 正しい - セミコロンを使用
{{/* @param Item struct{Name string; ID int} */}}
```

#### ベストプラクティス

✅ **推奨:**
- ネストされた構造にはドット記法を使用: `User.Name`, `Config.Database.Host`
- 複雑なデータのコレクションには `[]struct{...}` を使用
- オプショナルフィールドにはポインタ型 (`*Type`) を使用
- フィールドパスは比較的フラットに（1〜2階層の深さ）
- 構造体フィールドの区切りにはセミコロンを使用

❌ **非推奨:**
- トップレベルでのインライン `struct{...}` を使用しない
- スライス/マップを直接ネストしない（`[][]T`, `map[K][]V`）
- 深いフィールドパスとインライン構造体定義を組み合わせない
- 構造体フィールド定義でカンマを使用しない

#### 完全な例

サポートされるすべての型パターンと制限事項を示す包括的な例については、[`examples/05_all_param_types`](./examples/05_all_param_types) を参照してください。

### コマンドラインオプション

```
tmpltype -dir <directory> -pkg <name> -out <file>

オプション:
  -dir string
        テンプレートディレクトリ（必須）
        dir/*.tmpl (フラット) と dir/*/*.tmpl (グループ、1階層) を自動スキャン
  -pkg string
        出力パッケージ名（必須）
  -out string
        出力 .go ファイルパス（必須）
```

### 動作原理

1. **スキャン**: テンプレートファイルを解析し、フィールドアクセスパターンを抽出（例: `.User.Name`, `.Items[0].ID`）
2. **型解決**:
   - 明示的な `@param` 型ディレクティブを適用
   - テンプレート構文から型を推論（単純なフィールドは文字列、`range` からコレクションを推論）
3. **コード生成**: 以下を生成:
   - 型安全なパラメータ構造体
   - テンプレート解析関数
   - 型安全な `Render<テンプレート名>()` 関数
   - 動的なユースケース用の汎用 `Render()` 関数

### 生成されるコード構造

`tmpltype` は以下の構造でコードを生成します:

#### 共通部分

すべてのテンプレートで共有される要素:

```go
// Code generated by tmpltype; DO NOT EDIT.
package main

import (
    _ "embed"
    "fmt"
    "io"
    "text/template"
)

// 型安全なテンプレート名型
type TemplateName string

// テンプレート名の名前空間
var Template = struct {
    Email TemplateName
}{
    Email: "email",
}

// テンプレートファイルの埋め込み
//go:embed templates/email.tmpl
var emailTplSource string

// テンプレート初期化ヘルパー
func newTemplate(name TemplateName, source string) *template.Template {
    return template.Must(template.New(string(name)).Option("missingkey=error").Parse(source))
}

// すべてのテンプレートのマップ
var templates = map[TemplateName]*template.Template{
    Template.Email: newTemplate(Template.Email, emailTplSource),
}

// テンプレートマップを返す関数
func Templates() map[TemplateName]*template.Template {
    return templates
}

// 汎用描画関数（動的使用向け）
func Render(w io.Writer, name TemplateName, data any) error {
    tmpl, ok := templates[name]
    if !ok {
        return fmt.Errorf("template %q not found", name)
    }
    return tmpl.Execute(w, data)
}
```

#### テンプレートブロック

各テンプレートごとに、型定義と描画関数がグループ化されます:

```go
// ============================================================
// email template
// ============================================================

type EmailUser struct {
    Name string
}

// Email represents parameters for email template
type Email struct {
    User    EmailUser
    Message string
}

// RenderEmail renders the email template
func RenderEmail(w io.Writer, p Email) error {
    tmpl, ok := templates[Template.Email]
    if !ok {
        return fmt.Errorf("template %q not found", Template.Email)
    }
    return tmpl.Execute(w, p)
}
```

#### 複数テンプレートの場合

複数のテンプレートを処理する場合、各テンプレートのブロックがファイル名のアルファベット順に並びます:

```go
// 共通部分...

// ============================================================
// footer template
// ============================================================
type Footer struct { ... }
func RenderFooter(w io.Writer, p Footer) error { ... }

// ============================================================
// header template
// ============================================================
type Header struct { ... }
func RenderHeader(w io.Writer, p Header) error { ... }

// ============================================================
// nav template
// ============================================================
type Nav struct { ... }
func RenderNav(w io.Writer, p Nav) error { ... }
```

この構造により、特定のテンプレートに関連するコードが1箇所にまとまり、デバッグや確認が容易になります。

### サポートされるテンプレート構文

`tmpltype` は以下の Go テンプレート構文パターンをサポートしています。テンプレートスキャナはこれらのパターンを解析して自動的に型を推論します:

#### 1. 基本的なフィールド参照

```go
{{ .Title }}
```

生成される構造体に `string` フィールドを作成します。

#### 2. ネストされたフィールド参照

```go
{{ .User.Name }}
{{ .Author.Email }}
```

`string` フィールドを持つネストされた構造体型を作成します。

#### 3. 条件文（if）

```go
{{ if .Status }}
  <p>Status: {{ .Status }}</p>
{{ end }}
```

条件内のフィールドは、子フィールドがある場合は構造体として推論され、それ以外は `string` として推論されます。

#### 4. with 文と else 句

```go
{{ with .Summary }}
  <p>{{ .Content }}</p>
{{ else }}
  <p>{{ .DefaultMessage }}</p>
{{ end }}
```

ブロック内のドット (`.`) コンテキストを変更します。スキャナはスコープの変更を正しく追跡します。

#### 5. スライスの range

```go
{{ range .Items }}
  <li>{{ .Title }} - {{ .ID }}</li>
{{ end }}
```

`.Items` を range 本体内のフィールドを持つスライス型 `[]struct{...}` として推論します。

#### 6. index 関数によるマップアクセス

```go
{{ index .Meta "key" }}
{{ index .Meta "env" }}
```

`index` 関数を使用する場合、`.Meta` を `map[string]string` として推論します。

#### 7. ネストされた構造（with + range）

```go
{{ with .Project }}
  <h3>{{ .Name }}</h3>
  {{ range .Tasks }}
    <p>{{ .Title }}</p>
  {{ end }}
{{ end }}
```

`with` と `range` を組み合わせて、スライスフィールドを持つネストされた構造体階層を作成します。

#### 8. 深くネストされたパス

```go
{{ .Company.Department.Team.Manager.Name }}
```

完全なパスに従って深くネストされた構造体型を作成します。

#### 完全な例

サポートされるすべての構文パターンを示す完全なテンプレートについては、[`examples/04_comprehensive_template`](./examples/04_comprehensive_template) を参照してください。

### サンプル

完全に動作するサンプルについては、[`examples/`](./examples) ディレクトリを確認してください:

- [`01_basic`](./examples/01_basic): 型推論を使用した基本的な使用法
- [`02_param_directive`](./examples/02_param_directive): 複雑な型に対する `@param` ディレクティブの使用
- [`03_multi_template`](./examples/03_multi_template): 複数テンプレートの一括処理
- [`04_comprehensive_template`](./examples/04_comprehensive_template): サポートされるすべてのテンプレート構文パターンを示す包括的な例
- [`05_all_param_types`](./examples/05_all_param_types): サポートされるすべての `@param` 型と制限事項の完全なリファレンス
- [`07_grouping`](./examples/07_grouping): テンプレートのグループ化（フラットとグループの混在）

サンプルの実行:

```bash
cd examples/01_basic
go generate
go run .
```

### プロジェクト構造

```
.
├── cmd/tmpltype/          # CLI ツールのエントリポイント
├── internal/
│   ├── gen/               # コード生成ロジック
│   ├── scan/              # テンプレートスキャンと解析
│   ├── typing/            # 型推論と解決
│   │   └── magic/         # マジックコメント（@param）の解析
│   └── util/              # ユーティリティ関数
└── examples/              # 使用例
```

### コントリビューション

コントリビューションを歓迎します！プルリクエストを自由に提出してください。

#### 開発

1. リポジトリをクローン:
```bash
git clone https://github.com/bellwood4486/tmpltype.git
cd templagen-poc
```

2. テストの実行:
```bash
go test ./...
```

3. ビルド:
```bash
go build ./cmd/tmpltype
```

### ライセンス

このプロジェクトは MIT ライセンスのもとでライセンスされています - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

### 謝辞

Go の [`text/template`](https://pkg.go.dev/text/template) および [`html/template`](https://pkg.go.dev/html/template) パッケージを使用して構築されています。

---

<a name="english"></a>

## English

A Go code generator that creates type-safe template rendering functions from Go template files.

### Overview

`tmpltype` analyzes your Go template files, automatically infers or uses explicit parameter types, and generates type-safe Go code with structs and render functions. This eliminates runtime type errors and provides IDE autocompletion for template parameters.

### Features

- **Type Inference**: Automatically infers parameter types from template syntax (e.g., `.User.Name` → `string`)
- **Explicit Type Directives**: Support for `@param` directives to specify complex types
- **Type Safety**: Generate strongly-typed structs and render functions
- **Template Grouping**: Organize templates logically in subdirectories with nested namespaces
- **Multiple Templates**: Process single or multiple template files at once
- **go generate Integration**: Seamlessly integrates with Go's code generation workflow
- **Flexible Rendering**: Provides both type-safe and dynamic rendering options

### Installation

```bash
go install github.com/bellwood4486/tmpltype/cmd/tmpltype@latest
```

Or add to your project:

```bash
go get github.com/bellwood4486/tmpltype
```

### Usage

#### Basic Example

1. Create a template file `templates/email.tmpl`:

```html
<h1>Hello {{ .User.Name }}</h1>
<p>{{ .Message }}</p>
```

2. Add a generate directive to your Go file `gen.go`:

```go
package main

//go:generate tmpltype -dir templates -pkg main -out template_gen.go
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

#### Advanced Example: Type Directives

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

#### Multiple Templates

The `-dir` option automatically processes all `.tmpl` files in the directory:

```go
//go:generate tmpltype -dir templates -pkg main -out templates_gen.go
```

This command automatically scans:
- `templates/*.tmpl` (flat templates)
- `templates/*/*.tmpl` (grouped templates, 1 level only)

#### Template Grouping

Organize templates logically in subdirectories:

```
templates/
├── footer.tmpl                  # Flat template
├── mail_invite/                 # Group
│   ├── title.tmpl
│   └── content.tmpl
└── mail_account_created/        # Group
    ├── title.tmpl
    └── content.tmpl
```

Generated code:

```go
var Template = struct {
    Footer             TemplateName  // Flat
    MailInvite struct {              // Group
        Title   TemplateName
        Content TemplateName
    }
    MailAccountCreated struct {      // Group
        Title   TemplateName
        Content TemplateName
    }
}

// Usage
RenderMailInviteTitle(w, MailInviteTitle{...})
Render(w, Template.MailInvite.Title, data)
```

### `@param` Directive Reference

The `@param` directive allows you to explicitly specify types for template parameters, overriding automatic type inference. This is essential for complex types like specific integer sizes, optional fields (pointers), and structured data.

#### Syntax

```go
{{/* @param <FieldPath> <Type> */}}
```

- `<FieldPath>`: Dot-separated field path (e.g., `User.Name`, `Items`, `Config.Database.Host`)
- `<Type>`: Go type expression (see supported types below)

#### Supported Types

##### ✅ Fully Supported

**1. Basic Types**
```go
{{/* @param Name string */}}
{{/* @param Age int */}}
{{/* @param Count int64 */}}
{{/* @param Price float64 */}}
{{/* @param Active bool */}}
```

Supported base types: `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `bool`, `byte`, `rune`, `any`

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

##### ❌ Known Limitations

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

#### Best Practices

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

#### Complete Example

See [`examples/05_all_param_types`](./examples/05_all_param_types) for a comprehensive example demonstrating all supported type patterns and limitations.

### Command Line Options

```
tmpltype -dir <directory> -pkg <name> -out <file>

Options:
  -dir string
        Template directory (required)
        Automatically scans dir/*.tmpl (flat) and dir/*/*.tmpl (grouped, 1 level)
  -pkg string
        Output package name (required)
  -out string
        Output .go file path (required)
```

### How It Works

1. **Scan**: Parse template files and extract field access patterns (e.g., `.User.Name`, `.Items[0].ID`)
2. **Type Resolution**:
   - Apply explicit `@param` type directives
   - Infer types from template syntax (string for simple fields, infer collections from `range`)
3. **Code Generation**: Generate:
   - Type-safe parameter structs
   - Template parsing functions
   - Type-safe `Render<TemplateName>()` functions
   - Generic `Render()` function for dynamic use cases

### Generated Code Structure

`tmpltype` generates code in the following structure:

#### Common Section

Elements shared across all templates:

```go
// Code generated by tmpltype; DO NOT EDIT.
package main

import (
    _ "embed"
    "fmt"
    "io"
    "text/template"
)

// Type-safe template name type
type TemplateName string

// Template namespace for type-safe template names
var Template = struct {
    Email TemplateName
}{
    Email: "email",
}

// Template file embedding
//go:embed templates/email.tmpl
var emailTplSource string

// Template initialization helper
func newTemplate(name TemplateName, source string) *template.Template {
    return template.Must(template.New(string(name)).Option("missingkey=error").Parse(source))
}

// Map of all templates
var templates = map[TemplateName]*template.Template{
    Template.Email: newTemplate(Template.Email, emailTplSource),
}

// Function returning the templates map
func Templates() map[TemplateName]*template.Template {
    return templates
}

// Generic render function (for dynamic use)
func Render(w io.Writer, name TemplateName, data any) error {
    tmpl, ok := templates[name]
    if !ok {
        return fmt.Errorf("template %q not found", name)
    }
    return tmpl.Execute(w, data)
}
```

#### Template Blocks

For each template, type definitions and render function are grouped together:

```go
// ============================================================
// email template
// ============================================================

type EmailUser struct {
    Name string
}

// Email represents parameters for email template
type Email struct {
    User    EmailUser
    Message string
}

// RenderEmail renders the email template
func RenderEmail(w io.Writer, p Email) error {
    tmpl, ok := templates[Template.Email]
    if !ok {
        return fmt.Errorf("template %q not found", Template.Email)
    }
    return tmpl.Execute(w, p)
}
```

#### Multiple Templates

When processing multiple templates, each template block is ordered alphabetically by filename:

```go
// Common section...

// ============================================================
// footer template
// ============================================================
type Footer struct { ... }
func RenderFooter(w io.Writer, p Footer) error { ... }

// ============================================================
// header template
// ============================================================
type Header struct { ... }
func RenderHeader(w io.Writer, p Header) error { ... }

// ============================================================
// nav template
// ============================================================
type Nav struct { ... }
func RenderNav(w io.Writer, p Nav) error { ... }
```

This structure keeps code related to a specific template co-located, making debugging and verification easier.

### Supported Template Syntax

`tmpltype` supports the following Go template syntax patterns. The template scanner analyzes these patterns to infer types automatically:

#### 1. Basic Field Reference

```go
{{ .Title }}
```

Creates a `string` field in the generated struct.

#### 2. Nested Field Reference

```go
{{ .User.Name }}
{{ .Author.Email }}
```

Creates nested struct types with `string` fields.

#### 3. Conditional Statements (if)

```go
{{ if .Status }}
  <p>Status: {{ .Status }}</p>
{{ end }}
```

The field in the condition is inferred as a struct if it has child fields, otherwise as `string`.

#### 4. With Statement and Else Clause

```go
{{ with .Summary }}
  <p>{{ .Content }}</p>
{{ else }}
  <p>{{ .DefaultMessage }}</p>
{{ end }}
```

Changes the dot (`.`) context within the block. The scanner tracks scope changes correctly.

#### 5. Range Over Slice

```go
{{ range .Items }}
  <li>{{ .Title }} - {{ .ID }}</li>
{{ end }}
```

Infers `.Items` as a slice type `[]struct{...}` with fields from the range body.

#### 6. Map Access with Index Function

```go
{{ index .Meta "key" }}
{{ index .Meta "env" }}
```

Infers `.Meta` as `map[string]string` when using the `index` function.

#### 7. Nested Structures (with + range)

```go
{{ with .Project }}
  <h3>{{ .Name }}</h3>
  {{ range .Tasks }}
    <p>{{ .Title }}</p>
  {{ end }}
{{ end }}
```

Combines `with` and `range` to create nested struct hierarchies with slice fields.

#### 8. Deep Nested Paths

```go
{{ .Company.Department.Team.Manager.Name }}
```

Creates deeply nested struct types following the full path.

#### Complete Example

See [`examples/04_comprehensive_template`](./examples/04_comprehensive_template) for a complete template demonstrating all supported syntax patterns.

### Examples

Check the [`examples/`](./examples) directory for complete working examples:

- [`01_basic`](./examples/01_basic): Basic usage with type inference
- [`02_param_directive`](./examples/02_param_directive): Using `@param` directives for complex types
- [`03_multi_template`](./examples/03_multi_template): Processing multiple templates at once
- [`04_comprehensive_template`](./examples/04_comprehensive_template): Comprehensive example demonstrating all supported template syntax patterns
- [`05_all_param_types`](./examples/05_all_param_types): Complete reference for all supported `@param` types and limitations
- [`07_grouping`](./examples/07_grouping): Template grouping (mixed flat and grouped templates)

Run examples:

```bash
cd examples/01_basic
go generate
go run .
```

### Project Structure

```
.
├── cmd/tmpltype/          # CLI tool entry point
├── internal/
│   ├── gen/               # Code generation logic
│   ├── scan/              # Template scanning and parsing
│   ├── typing/            # Type inference and resolution
│   │   └── magic/         # Magic comment (@param) parsing
│   └── util/              # Utility functions
└── examples/              # Usage examples
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

#### Development

1. Clone the repository:
```bash
git clone https://github.com/bellwood4486/tmpltype.git
cd templagen-poc
```

2. Run tests:
```bash
go test ./...
```

3. Build:
```bash
go build ./cmd/tmpltype
```

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Acknowledgments

Built with Go's [`text/template`](https://pkg.go.dev/text/template) and [`html/template`](https://pkg.go.dev/html/template) packages.
