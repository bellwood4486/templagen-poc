// Package magic は @param マジックコメントのパースと型オーバーライドを提供します。
//
// このパッケージは以下の機能を提供します:
//   - テンプレート内の @param ディレクティブの抽出
//   - 型表現のパース (基本型、スライス、マップ、ポインタ、構造体)
//   - 型オーバーライドの管理
//
// @param ディレクティブの形式:
//   {{/* @param User.Age int */}}
//   {{/* @param Items []struct{ID int; Name string} */}}
package magic
