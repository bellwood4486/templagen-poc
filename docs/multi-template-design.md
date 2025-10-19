# 複数テンプレートファイル対応 設計書（シンプル版）

## 1. 概要

templagenを複数テンプレートファイルに対応させる。シンプルさを最優先し、1つの入力パターンから1つの統合ファイルを生成する設計とする。

## 2. 基本コンセプト

### 2.1 シンプルな動作原理
- **入力**: 1つ以上のテンプレートファイル（glob/ディレクトリ/個別指定）
- **処理**: 全テンプレートをスキャンして型を推論
- **出力**: 1つの統合Goファイル

### 2.2 ユースケース
```
templates/
├── user.tmpl        # ユーザー詳細画面
├── user_list.tmpl   # ユーザー一覧画面
├── user_edit.tmpl   # ユーザー編集画面
└── gen.go           # go:generate 1行だけ
```

## 3. 設計方針

### 3.1 基本原則
1. **統一的な扱い**: 単一ファイルも複数ファイルも同じように処理
2. **出力は常に1ファイル**: 複雑な出力パターンは不要
3. **自動命名**: ファイル名から型名・関数名を自動生成
4. **最小限のオプション**: 本当に必要なものだけ

## 4. 命名規則

### 4.1 自動命名規則
テンプレートファイル名から自動的に型名・関数名を生成：

| テンプレート | 生成される要素 |
|------------|--------------|
| `user.tmpl` | `User`構造体, `RenderUser()` |
| `user_list.tmpl` | `UserList`構造体, `RenderUserList()` |
| `email.tmpl` | `Email`構造体, `RenderEmail()` |

### 4.2 変換ルール
```
ファイル名 → 型名の変換：
- user.tmpl        → User
- user_list.tmpl   → UserList
- user-detail.tmpl → UserDetail
- 01_header.tmpl   → Header (数字プレフィックスは削除)
```

### 4.3 生成される構造

```go
// templates_gen.go
package templates

import (
    _ "embed"
    "io"
    "text/template"
)

//go:embed user.tmpl
var userTplSource string

//go:embed user_list.tmpl
var userListTplSource string

// テンプレートごとのパラメータ型
type User struct {
    Name  string
    Email string
}

type UserList struct {
    Users []User
    Total int
}

// グローバルなTemplates関数
func Templates() map[string]*template.Template {
    return map[string]*template.Template{
        "user":      template.Must(template.New("user").Parse(userTplSource)),
        "user_list": template.Must(template.New("user_list").Parse(userListTplSource)),
    }
}

// 個別のレンダリング関数
func RenderUser(w io.Writer, p User) error {
    return Templates()["user"].Execute(w, p)
}

func RenderUserList(w io.Writer, p UserList) error {
    return Templates()["user_list"].Execute(w, p)
}

// 汎用レンダリング関数
func Render(w io.Writer, name string, data any) error {
    tmpl, ok := Templates()[name]
    if !ok {
        return fmt.Errorf("template %q not found", name)
    }
    return tmpl.Execute(w, data)
}
```

## 5. CLIインターフェース

### 5.1 シンプルなコマンド

```bash
# 単一ファイル（現状と同じ）
templagen -in user.tmpl -pkg templates -out templates_gen.go

# 複数ファイル（glob）
templagen -in "*.tmpl" -pkg templates -out templates_gen.go

# ディレクトリ内の全.tmplファイル
templagen -in "./templates/*.tmpl" -pkg templates -out templates_gen.go

# 複数ファイル（個別指定）
templagen -in "user.tmpl,user_list.tmpl" -pkg templates -out templates_gen.go
```

### 5.2 フラグ

| フラグ | 説明 | デフォルト |
|-------|------|-----------|
| `-in` | 入力パターン（glob対応） | 必須 |
| `-pkg` | 出力パッケージ名 | 必須 |
| `-out` | 出力ファイル | 必須 |
| `-exclude` | 除外パターン（オプション） | - |

### 5.3 使用例

```bash
# go:generateでの使用
//go:generate templagen -in "*.tmpl" -pkg templates -out templates_gen.go

# テストテンプレートを除外
//go:generate templagen -in "*.tmpl" -exclude "*_test.tmpl" -pkg templates -out templates_gen.go
```

## 6. 実装計画

### 実装ステップ
1. **glob展開対応**: `-in`フラグでワイルドカードを受け付ける
2. **複数ファイル処理**: 複数テンプレートを順次処理して型情報を収集
3. **名前空間分離**: テンプレートごとに独立した型定義を生成
4. **統合出力**: すべてを1つのファイルにまとめて出力
5. **除外パターン**: `-exclude`オプションの実装（オプション）

## 7. 実装詳細

### 7.1 内部処理フロー

```
1. 入力パターンからファイルリストを取得
   - glob展開
   - 除外パターン適用

2. 各テンプレートを処理
   - テンプレート名の決定（ファイル名から）
   - AST解析とフィールド収集
   - @paramディレクティブの処理
   - 型推論と型解決

3. 統合コード生成
   - 全テンプレートのembed宣言
   - テンプレートごとの型定義
   - Templates()マップ関数
   - 個別Render関数
   - 汎用Render関数
```

### 7.2 型の名前衝突回避

同じディレクトリ内で型名が衝突する場合の対処：

```go
// user.tmpl と admin.tmpl の両方に User型がある場合
type UserUser struct { ... }    // user.tmpl の User
type AdminUser struct { ... }   // admin.tmpl の User
```

または、ネストした構造として生成：

```go
type User struct {
    Name string
    // user.tmpl特有のフィールド
}

type Admin struct {
    User User  // 共通部分を埋め込み
    // admin.tmpl特有のフィールド
}
```

## 8. 考慮事項

### 8.1 エラーハンドリング
- 一部のテンプレートでパースエラーが発生した場合は、該当ファイルを報告して処理を中断
- 型名衝突が発生した場合は、テンプレート名をプレフィックスとして自動付与

### 8.2 パフォーマンス
- テンプレートのパースと型解析は並行処理可能
- 生成されるTemplates()マップは初回アクセス時に遅延初期化することも検討

### 8.3 型の共有
- 将来的には共通型の自動抽出を実装可能
- 現時点では各テンプレートが独立した型を持つシンプルな実装

## 9. FAQ

### Q: 単一ファイルの場合も動作は変わりますか？
A: 変わりません。単一ファイルは複数ファイルの特殊ケース（要素数1）として扱われます。

### Q: なぜ個別ファイル出力モードを廃止したのですか？
A: シンプルさのため。複数の出力ファイルを管理するより、1つの統合ファイルの方が扱いやすく、実装もシンプルです。

### Q: 型名の衝突はどう解決されますか？
A: テンプレート名を自動的にプレフィックスとして付与します（例：`UserUser`, `AdminUser`）。

### Q: Templates()関数の使い方は？
A: `Templates()["user"]`でテンプレートを取得できます。また、個別の`RenderUser()`関数も生成されるので、型安全に使えます。

## 10. まとめ

シンプルさを優先した設計により：
- **使い方が直感的**: 入力パターンと出力ファイルを指定するだけ
- **実装がシンプル**: 複雑なモード分岐がない
- **出力が予測可能**: 常に1つの統合ファイルが生成される
- **型安全**: テンプレートごとに専用の型とRender関数

この設計により、複数テンプレートの管理が簡単になり、開発効率が向上します。