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

// toolsvipErrorResponse 接口异常/包裹响应时可能返回的对象结构（而非数组）
type toolsvipErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Msg     string            `json:"msg"`
	Data    []toolsvipRawItem `json:"data"`
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
	if err := json.Unmarshal(body, &items); err == nil {
		return items, nil
	}

	// 接口未按预期返回数组：尝试按对象解析（常见于报错信息或 {data: [...]} 包裹结构）
	var wrapped toolsvipErrorResponse
	if jsonErr := json.Unmarshal(body, &wrapped); jsonErr == nil {
		if len(wrapped.Data) > 0 {
			return wrapped.Data, nil
		}
		if msg := firstNonEmptyToolsvip(wrapped.Error, wrapped.Message, wrapped.Msg); msg != "" {
			return nil, fmt.Errorf("%s 接口返回: %s", mailbox, msg)
		}
		// 未命中已知错误字段的空对象，视为该邮箱暂无邮件
		return []toolsvipRawItem{}, nil
	}

	return nil, fmt.Errorf("JSON 解析失败 (%s): %s", mailbox, toolsvipSafePrefix(string(body), 200))
}

func firstNonEmptyToolsvip(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func toolsvipSafePrefix(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
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
