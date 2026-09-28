package webmail

import (
	"os"
	"strings"
	"testing"
)

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

func TestParseLurentoolMailResponse(t *testing.T) {
	raw := `[{"email":"a@b.com","ok":true,"mails":[{"subject":"Complete code challenge","receivedTime":"2026-09-15T12:45:55Z","preview":"Your one-time code is 012888.","isRead":false,"from":"Cursor <no-reply@cursor.sh>","body":"Your one-time code is 012888. This code expires in 10 minutes."}],"source":"reg_email","message":"成功，共 1 封"}]`
	result, err := parseLurentoolMailResponse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if result.Email != "a@b.com" || !result.Ok || len(result.Mails) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	items := convertLurentoolMails(result.Mails, result.Source)
	if items[0].Code != "012888" {
		t.Fatalf("want code 012888, got %q", items[0].Code)
	}
	if items[0].ReceivedAt != "2026-09-15 20:45:55" {
		t.Fatalf("want CST time, got %q", items[0].ReceivedAt)
	}
}

func TestParseOuleyxListAndDetail(t *testing.T) {
	listHTML := `
<div class="u-mail-row unread" data-id="1137456" data-action="select-mail">
    <div class="u-mail-main">
        <span class="u-mail-sender">Microsoft 帐户团队</span>
        <span class="u-mail-subject">
            你的一次性代码 <span class="u-mail-snippet"> — user@ouleyx.cc，你好! 我们已收到你要求获得 Microsoft 帐户所用的一次性代码的…</span>
        </span>
    </div>
    <span class="u-mail-time">18:13</span>
</div>`
	items := parseOuleyxList(listHTML, "收件箱")
	if len(items) != 1 {
		t.Fatalf("want 1 mail, got %d", len(items))
	}
	if items[0].ID != "1137456" || items[0].Subject != "你的一次性代码" || items[0].From != "Microsoft 帐户团队" {
		t.Fatalf("unexpected list item: %+v", items[0])
	}

	detailHTML := `
<h1 class="u-detail-subject">你的一次性代码</h1>
<div class="u-meta-name">Microsoft 帐户团队</div>
<div class="u-meta-email">&lt;account-security-noreply@accountprotection.microsoft.com &gt;</div>
<div class="u-meta-time">2026年9月18日 18:13</div>
<div class="u-mail-plaintext">你的一次性代码为: 461230</div>`
	detail := parseOuleyxDetail(detailHTML)
	if detail.Code != "461230" {
		t.Fatalf("want code 461230, got %q", detail.Code)
	}
	if !strings.Contains(detail.From, "Microsoft 帐户团队") {
		t.Fatalf("unexpected from: %q", detail.From)
	}

	applyOuleyxRender(detail, `<html><body><p>your one-time code is 778899</p></body></html>`)
	if detail.Code != "778899" || !strings.Contains(detail.HtmlBody, "<p>") {
		t.Fatalf("render not applied: %+v", detail)
	}
	applyOuleyxRender(detail, `<h1>登录您的企业邮箱</h1>`)
	if strings.Contains(detail.HtmlBody, "登录您的企业邮箱") {
		t.Fatal("login page should not replace mail html")
	}
}

func TestFetchOuleyxMailsLive(t *testing.T) {
	email := strings.TrimSpace(os.Getenv("OULEYX_EMAIL"))
	password := strings.TrimSpace(os.Getenv("OULEYX_PASSWORD"))
	if email == "" || password == "" {
		t.Skip("set OULEYX_EMAIL and OULEYX_PASSWORD to run live fetch")
	}
	data, err := FetchOuleyxMails(email, password)
	if err != nil {
		t.Fatal(err)
	}
	if data.Email != email || data.Token == "" {
		t.Fatalf("unexpected fetch data: email=%q token_empty=%v", data.Email, data.Token == "")
	}
	if len(data.Inbox) == 0 && len(data.Junk) == 0 {
		t.Log("inbox and junk are empty")
	}
	if len(data.Inbox) > 0 {
		detail, err := FetchOuleyxMailDetail(email, password, data.Token, data.Inbox[0].ID)
		if err != nil {
			t.Fatal(err)
		}
		if detail.Subject == "" && detail.Body == "" {
			t.Fatalf("empty detail: %+v", detail)
		}
	}
}
