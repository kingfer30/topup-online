package webmail

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const toolsvipAPIBase = "https://tool.toolsvip.cc/easy-mailbox/emails"

var (
	toolsvipHTMLTagRe    = regexp.MustCompile(`(?s)<[^>]*>`)
	toolsvipCodePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)temporary openai login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt verification code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)enter this temporary verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your OpenAI code is\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your ChatGPT code is\s*(\d{6})`),
	}
)

type toolsvipRawItem struct {
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Body    string `json:"body"`
	Text    string `json:"text"`
	Content string `json:"content"`
	HTML    string `json:"html"`
}

// ToolsvipMailItem 单封 toolsvip 邮件
type ToolsvipMailItem struct {
	Subject  string `json:"subject"`
	Date     string `json:"date"`
	Mailbox  string `json:"mailbox"`
	Code     string `json:"code"`
	HtmlBody string `json:"html_body"`
	Body     string `json:"body"`
}

func toolsvipStripHTML(s string) string {
	return strings.TrimSpace(toolsvipHTMLTagRe.ReplaceAllString(s, " "))
}

func extractToolsvipCode(text string) string {
	for _, re := range toolsvipCodePatterns {
		if m := re.FindStringSubmatch(text); len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

func newToolsvipClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func fetchToolsvipMailbox(client *http.Client, email, password, mailbox string) ([]toolsvipRawItem, error) {
	apiURL := fmt.Sprintf("%s?email=%s&password=%s&mailbox=%s",
		toolsvipAPIBase,
		url.QueryEscape(email),
		url.QueryEscape(password),
		mailbox,
	)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s HTTP %d: %s", mailbox, resp.StatusCode, string(body))
	}

	var items []toolsvipRawItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("JSON 解析失败 (%s): %w", mailbox, err)
	}
	return items, nil
}

// FetchToolsvipMails 获取 toolsvip 收件箱与垃圾箱邮件
func FetchToolsvipMails(email, password string) (inbox, junk []ToolsvipMailItem, err error) {
	client := newToolsvipClient()

	convert := func(items []toolsvipRawItem, mailboxLabel string) []ToolsvipMailItem {
		var result []ToolsvipMailItem
		for _, item := range items {
			rawBody := item.Body + item.Text + item.Content
			htmlBody := item.HTML
			plainBody := toolsvipStripHTML(rawBody + " " + htmlBody)

			// 提取验证码：先 subject，再合并正文
			code := extractToolsvipCode(item.Subject)
			if code == "" {
				code = extractToolsvipCode(plainBody)
			}

			body := strings.TrimSpace(rawBody)
			if body == "" {
				body = plainBody
			}

			result = append(result, ToolsvipMailItem{
				Subject:  item.Subject,
				Date:     item.Date,
				Mailbox:  mailboxLabel,
				Code:     code,
				HtmlBody: htmlBody,
				Body:     body,
			})
		}
		return result
	}

	inboxRaw, err := fetchToolsvipMailbox(client, email, password, "inbox")
	if err != nil {
		return nil, nil, fmt.Errorf("获取收件箱失败: %w", err)
	}
	inbox = convert(inboxRaw, "收件箱")

	junkRaw, err := fetchToolsvipMailbox(client, email, password, "junk")
	if err != nil {
		junk = []ToolsvipMailItem{}
	} else {
		junk = convert(junkRaw, "垃圾箱")
	}

	if inbox == nil {
		inbox = []ToolsvipMailItem{}
	}
	if junk == nil {
		junk = []ToolsvipMailItem{}
	}
	return inbox, junk, nil
}
