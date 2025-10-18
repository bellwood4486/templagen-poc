// Package scan はGoテンプレートをスキャンしてスキーマを推論します。
//
// このパッケージは text/template のASTを解析し、テンプレート内のフィールド参照を追跡して
// スキーマ木を構築します。推論される型は以下の通り:
//   - 葉のフィールド: string
//   - range で使用されるフィールド: []struct{...}
//   - index で使用されるフィールド: map[string]string
//
// スキャン結果は internal/typing パッケージで型解決されます。
package scan
