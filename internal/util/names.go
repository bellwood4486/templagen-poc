package util

// Export は与えられた識別子を Go のエクスポートされた識別子に変換します。
//
// 先頭のルーンが小文字の ASCII 英字であれば大文字に変換し、
// それ以外の場合は入力文字列をそのまま返します。
// この関数は、テンプレートの変数名などから Go の構造体フィールド名を
// 自動生成するときに利用します（例: "user" -> "User"）。
func Export(name string) string {
	if name == "" {
		return name
	}
	r := []rune(name)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] = r[0] - 'a' + 'A'
	}

	return string(r)
}
