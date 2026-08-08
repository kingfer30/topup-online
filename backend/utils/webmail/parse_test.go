package webmail

import "testing"

func TestParseWebMailAccountLine(t *testing.T) {
	cases := []struct {
		format   string
		line     string
		wantMail string
		wantPass string
	}{
		{"1", "a@b.com----mailpass", "a@b.com", "mailpass"},
		{"1", "a@b.com|mailpass", "a@b.com", "mailpass"},
		{"2", "a@b.com----gptpass----mailpass", "a@b.com", "mailpass"},
		{"3", "a@b.com----mailpass----gptpass", "a@b.com", "mailpass"},
		{"2", "a@b.com----gpt----mail--pass", "a@b.com", "mail--pass"},
		{"3", "a@b.com----mail--pass----gpt", "a@b.com", "mail--pass"},
	}
	for _, tc := range cases {
		email, pass, err := ParseWebMailAccountLine(tc.line, tc.format)
		if err != nil {
			t.Fatalf("format=%s line=%q: %v", tc.format, tc.line, err)
		}
		if email != tc.wantMail || pass != tc.wantPass {
			t.Fatalf("format=%s line=%q: got %q/%q want %q/%q", tc.format, tc.line, email, pass, tc.wantMail, tc.wantPass)
		}
	}
}
