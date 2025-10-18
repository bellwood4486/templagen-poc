// Package gen はテンプレートからGoコードを生成します。
//
// このパッケージは以下の処理を順次実行します:
//   1. テンプレートのスキャン (internal/scan)
//   2. 型の解決 (internal/typing)
//   3. Goコードの生成
//
// 生成されるコードには、Params構造体、Template関数、Render関数が含まれます。
package gen
