package util

import "testing"

func TestExport(t *testing.T) {
    tests := []struct{
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
        {in: "_user", want: "_user", name: "starts with underscore"},
        {in: "applePie", want: "ApplePie", name: "capitalize only first"},
        {in: "éclair", want: "éclair", name: "non-ascii lower"},
        {in: "あいう", want: "あいう", name: "non-ascii japanese"},
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

