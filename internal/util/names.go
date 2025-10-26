package util

import "strings"

// Export は与えられた識別子を Go のエクスポートされた識別子に変換します。
//
// アンダースコア区切りの名前をキャメルケースに変換し、先頭を大文字にします。
// ただし、先頭がアンダースコアの場合は元の名前をそのまま返します。
// この関数は、テンプレートの変数名などから Go の構造体フィールド名を
// 自動生成するときに利用します。
//
// 例:
//   - "user" -> "User"
//   - "mail_invite" -> "MailInvite"
//   - "mail_account_created" -> "MailAccountCreated"
//   - "_user" -> "_user" (先頭がアンダースコアの場合は変換しない)
//   - "applePie" -> "ApplePie" (アンダースコアがない場合は先頭のみ大文字)
func Export(name string) string {
	if name == "" {
		return name
	}

	// 先頭がアンダースコアの場合は変換しない
	if strings.HasPrefix(name, "_") {
		return name
	}

	// アンダースコアが含まれている場合はキャメルケースに変換
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		var result strings.Builder
		for _, part := range parts {
			if part == "" {
				continue
			}
			r := []rune(part)
			if r[0] >= 'a' && r[0] <= 'z' {
				r[0] = r[0] - 'a' + 'A'
			}
			result.WriteString(string(r))
		}
		return result.String()
	}

	// アンダースコアがない場合は先頭のみ大文字に
	r := []rune(name)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] = r[0] - 'a' + 'A'
	}

	return string(r)
}
