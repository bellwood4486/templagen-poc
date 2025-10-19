package gen

// Unit は単一のテンプレート処理単位
type Unit struct {
	Pkg           string // 出力パッケージ名
	SourcePath    string // 埋め込むテンプレファイルのパス
	SourceLiteral string // テンプレ本文
	TemplateName  string // テンプレート名（複数ファイル対応用）
}

// Emit は単一テンプレート用の互換性ラッパー
// 内部的にはEmitMultiに委譲する
func Emit(u Unit) (string, error) {
	return EmitMulti([]Unit{u})
}
