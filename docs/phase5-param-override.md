# フェーズ5: @param型上書き機能 設計書

## 1. 概要

テンプレート内の特殊コメント `{{/* @param Path Type */}}` を使用して、自動推論された型を明示的に上書きする機能を実装する。

## 2. 機能要件

### 2.1 構文

```
{{/* @param Path Type */}}
```

- `Path`: フィールドパス（例: `User.Age`, `Items`, `Meta.Tags`）
- `Type`: Go型表現（例: `int`, `[]string`, `map[string]bool`）

### 2.2 サポート型

#### 基本型
- `string`
- `int`, `int64`, `int32` など
- `float64`, `float32`
- `bool`
- `time.Time`（自動で`import "time"`追加）

#### コンテナ型
- `[]T` - スライス（例: `[]string`, `[]int`）
- `map[string]T` - マップ（キーは`string`固定）
- `*T` - ポインタ（例: `*string`, `*User`）

#### 複合型
- `struct{...}` - インライン構造体定義

### 2.3 使用例

```html
{{/* @param User.Age int */}}
{{/* @param User.Email *string */}}
{{/* @param Items []struct{ID int64; Title string; Price float64} */}}
{{/* @param Meta map[string]string */}}
{{/* @param CreatedAt time.Time */}}

<h1>Hello {{ .User.Name }}</h1>
<p>Age: {{ .User.Age }}</p>
{{ if .User.Email }}<p>Email: {{ .User.Email }}</p>{{ end }}
```

## 3. 実装設計

### 3.1 パッケージ構成

```
internal/magic/
├── magic.go        # メインロジック
├── magic_test.go   # テスト
├── parser.go       # 型パーサー
└── parser_test.go  # パーサーテスト
```

### 3.2 主要インターフェース

```go
// internal/magic/magic.go

// ParamDirective は@param指令を表す
type ParamDirective struct {
    Path string   // "User.Age"
    Type TypeExpr // 解析済み型表現
}

// TypeExpr は型表現を表す
type TypeExpr struct {
    Kind     TypeKind
    BaseType string       // "string", "int", "time.Time" など
    Elem     *TypeExpr    // スライス/マップ/ポインタの要素型
    Fields   []FieldDef   // struct{}の場合のフィールド定義
}

type TypeKind int
const (
    TypeKindBase TypeKind = iota
    TypeKindSlice
    TypeKindMap
    TypeKindPointer
    TypeKindStruct
)

type FieldDef struct {
    Name string
    Type TypeExpr
}

// ParseParams はテンプレートソースから@param指令を抽出
func ParseParams(src string) ([]ParamDirective, error)

// ApplyOverrides はスキーマ木に@param指令を適用
func ApplyOverrides(schema *scan.Schema, directives []ParamDirective) error

// RequiredImports は使用された型から必要なimportを返す
func RequiredImports(directives []ParamDirective) []string
```

### 3.3 処理フロー

1. **抽出フェーズ**: 正規表現で`{{/* @param ... */}}`を検出
2. **パースフェーズ**: 型文字列を`TypeExpr`構造体に変換
3. **適用フェーズ**: スキーマ木のノードを型情報で上書き
4. **生成フェーズ**: 上書き後のスキーマから型定義生成

## 4. 型パーサー設計

### 4.1 文法（簡易版）

```
Type       := BaseType | SliceType | MapType | PointerType | StructType
BaseType   := identifier ("." identifier)*
SliceType  := "[]" Type
MapType    := "map[string]" Type
PointerType:= "*" Type
StructType := "struct{" Fields "}"
Fields     := Field (";" Field)*
Field      := identifier Type
```

### 4.2 パーサー実装方針

- 手書き再帰下降パーサー
- トークナイザは簡易実装（正規表現ベース）
- エラーメッセージは位置情報付き

## 5. スキーマ木への適用

### 5.1 適用ルール

1. **パス解決**: ドット区切りでスキーマ木をたどる
2. **型変換**: `TypeExpr`を`scan.Field`の`Kind`と関連フィールドに変換
3. **検証**:
   - 存在しないパスは警告（エラーにはしない）
   - 型の不整合は警告

### 5.2 型変換マッピング

| TypeExpr | scan.Kind | 備考 |
|----------|-----------|------|
| BaseType(string等) | KindString | 基本型はすべてKindString扱い（gen側で型文字列保持） |
| SliceType | KindSlice | Elemに要素型 |
| MapType | KindMap | Elemに値型 |
| StructType | KindStruct | Childrenにフィールド |

## 6. gen.Emitとの統合

### 6.1 変更点

```go
// internal/gen/emit.go の変更

func Emit(u Unit) (string, error) {
    // 1. テンプレートスキャン
    sch, err := scan.ScanTemplate(u.SourceLiteral)

    // 2. @param抽出と適用（新規追加）
    directives, err := magic.ParseParams(u.SourceLiteral)
    if err != nil {
        return "", fmt.Errorf("failed to parse @param: %w", err)
    }

    if err := magic.ApplyOverrides(&sch, directives); err != nil {
        return "", fmt.Errorf("failed to apply @param: %w", err)
    }

    // 3. 必要なimport追加（新規追加）
    extraImports := magic.RequiredImports(directives)
    for _, imp := range extraImports {
        imports[imp] = struct{}{}
    }

    // 以降既存処理...
}
```

### 6.2 Field構造体の拡張

```go
// internal/scan/scan.go の変更

type Field struct {
    Name     string
    Kind     Kind
    TypeName string            // 新規: 具体的な型名（"int", "time.Time"など）
    Elem     *Field
    Children map[string]*Field
}
```

## 7. テスト計画

### 7.1 単体テスト

#### magic.ParseParams
- 単一@param抽出
- 複数@param抽出
- 無効な構文のエラー処理
- コメント内の位置（先頭、中間、末尾）

#### 型パーサー
- 基本型: `string`, `int`, `bool`
- スライス: `[]string`, `[][]int`
- マップ: `map[string]int`, `map[string][]bool`
- ポインタ: `*string`, `*[]int`
- 構造体: `struct{Name string}`, `struct{ID int64; Items []string}`
- ネスト: `[]map[string]*struct{X int}`

#### ApplyOverrides
- 単純フィールドの上書き
- ネストフィールドの上書き
- スライス要素の型変更
- 存在しないパスの警告

### 7.2 統合テスト

```go
// internal/gen/emit_test.go に追加

func TestEmit_WithParamOverride(t *testing.T) {
    src := `
{{/* @param User.Age int */}}
{{/* @param Items []struct{ID int64; Title string} */}}
{{ .User.Name }} is {{ .User.Age }} years old.
{{ range .Items }}{{ .ID }}: {{ .Title }}{{ end }}
`
    // Params型にAge int、Items []struct{...}が生成されることを確認
}
```

### 7.3 E2Eテスト（フェーズ8で実装）

```bash
# テンプレートファイルに@param付き
# go generate実行
# 生成コードのコンパイル確認
# 実行時の型チェック
```

## 8. エラー処理方針

### 8.1 エラーレベル

1. **Fatal**: パース失敗、無効な型表現 → 生成中断
2. **Warning**: 存在しないパス → ログ出力して続行
3. **Info**: 型の暗黙変換 → デバッグ情報

### 8.2 エラーメッセージ例

```
Error: Invalid type expression at line 3: "map[int]string" - map key must be string
Warning: Path "NonExistent.Field" not found in template
Info: Overriding "User.Age" from string to int
```

## 9. 実装優先順位

1. **Phase 1**: 基本型のみ（string, int, bool）
2. **Phase 2**: スライス、マップ追加
3. **Phase 3**: ポインタ、time.Time追加
4. **Phase 4**: struct{}サポート

## 10. 互換性とマイグレーション

- @paramなしでも従来通り動作（後方互換性維持）
- 段階的に@paramを追加可能
- 将来的に型推論の改善で@param不要化も視野に

## 11. 制限事項

- マップのキーは`string`固定
- インターフェース型は未サポート
- ジェネリクス型は未サポート
- 関数型は未サポート

## 12. 今後の拡張案

- `@param`でデフォルト値指定: `{{/* @param Age int = 0 */}}`
- 型エイリアス: `{{/* @type UserID = int64 */}}`
- 外部型参照: `{{/* @param User myapp.User */}}`（--root-typeと連携）