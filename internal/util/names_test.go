package util

import "testing"

func TestExport(t *testing.T) {
	tests := []struct {
		in   string
		want string
		name string
	}{
		{in: "", want: "", name: "empty"},
		{in: "user", want: "User", name: "lower ascii"},
		{in: "User", want: "User", name: "already upper"},
		{in: "a", want: "A", name: "single lower"},
		{in: "Z", want: "Z", name: "single upper"},
		{in: "1user", want: "1user", name: "starts with digit"},
		{in: "_user", want: "User", name: "starts with underscore"},
		{in: "applePie", want: "ApplePie", name: "capitalize only first"},
		{in: "éclair", want: "éclair", name: "non-ascii lower"},
		{in: "あいう", want: "あいう", name: "non-ascii japanese"},
		{in: "user_name", want: "UserName", name: "snake_case"},
		{in: "mail_invite", want: "MailInvite", name: "snake_case two words"},
		{in: "mail_account_created", want: "MailAccountCreated", name: "snake_case three words"},
		{in: "__double_underscore", want: "DoubleUnderscore", name: "leading double underscore"},
		{in: "user__name", want: "UserName", name: "double underscore middle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Export(tt.in)
			if got != tt.want {
				t.Fatalf("Export(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}
