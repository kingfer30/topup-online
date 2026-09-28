package webmail

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	ouleyxBaseURL  = "https://ouleyx.cc"
	ouleyxLoginURL = "https://ouleyx.cc/api/v1/user/login"
	ouleyxCookie   = "ouyun_user"
)

var (
	ouleyxHTMLTagRe = regexp.MustCompile(`(?s)<[^>]*>`)
	ouleyxRowRe     = regexp.MustCompile(`(?is)<div\s+class="[^"]*\bu-mail-row\b[^"]*"[^>]*\bdata-id="(\d+)"[^>]*>`)
	ouleyxSenderRe  = regexp.MustCompile(`(?is)<span\s+class="u-mail-sender"[^>]*>(.*?)</span>`)
	ouleyxSubjectRe = regexp.MustCompile(`(?is)<span\s+class="u-mail-subject"[^>]*>(.*?)</span>`)
	ouleyxSnippetRe = regexp.MustCompile(`(?is)<span\s+class="u-mail-snippet"[^>]*>(.*?)</span>`)
	ouleyxTimeRe    = regexp.MustCompile(`(?is)<span\s+class="u-mail-time"[^>]*>(.*?)</span>`)

	ouleyxDetailSubjectRe = regexp.MustCompile(`(?is)<h1\s+class="u-detail-subject"[^>]*>(.*?)</h1>`)
	ouleyxDetailNameRe    = regexp.MustCompile(`(?is)<div\s+class="u-meta-name"[^>]*>(.*?)</div>`)
	ouleyxDetailEmailRe   = regexp.MustCompile(`(?is)<div\s+class="u-meta-email"[^>]*>(.*?)</div>`)
	ouleyxDetailTimeRe    = regexp.MustCompile(`(?is)<div\s+class="u-meta-time"[^>]*>(.*?)</div>`)
	ouleyxPlainRe         = regexp.MustCompile(`(?is)<div\s+class="u-mail-plaintext"[^>]*>(.*?)</div>`)
	ouleyxBodyRe          = regexp.MustCompile(`(?is)<div\s+class="u-mail-body"[^>]*>(.*?)</div>`)
	ouleyxStyleScriptRe   = regexp.MustCompile(`(?is)<(script|style)\b[^>]*>.*?</(script|style)>`)

	ouleyxCodePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)一次性代码为[:：\s]*(\d{4,8})`),
		regexp.MustCompile(`(?i)一次性代码[:：\s]*(\d{4,8})`),
		regexp.MustCompile(`(?i)your one-time code is\s*(\d{4,8})`),
		regexp.MustCompile(`(?i)one-time code is\s*(\d{4,8})`),
		regexp.MustCompile(`(?i)verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)temporary openai login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt verification code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)Your OpenAI code is\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your ChatGPT code is\s*(\d{6})`),
		regexp.MustCompile(`(?i)(?:your\s+)?(?:cursor\s+)?(?:login\s+)?code(?:\s+is)?[:\s]+(\d{6})`),
	}
)

type ouleyxLoginRequest struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

type ouleyxLoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Address string `json:"address"`
		Token   string `json:"token"`
	} `json:"data"`
}

// OuleyxMailItem 单封 ouleyx 邮件
type OuleyxMailItem struct {
	ID         string `json:"id"`
	Subject    string `json:"subject"`
	From       string `json:"from"`
	ReceivedAt string `json:"received_at"`
	Preview    string `json:"preview"`
	Mailbox    string `json:"mailbox"`
	Code       string `json:"code"`
	Body       string `json:"body"`
	HtmlBody   string `json:"html_body"`
}

// OuleyxMailDetail ouleyx 邮件详情
type OuleyxMailDetail struct {
	Subject    string `json:"subject"`
	From       string `json:"from"`
	ReceivedAt string `json:"received_at"`
	Body       string `json:"body"`
	HtmlBody   string `json:"html_body"`
	Code       string `json:"code"`
}

// OuleyxFetchData ouleyx 快捷取件结果
type OuleyxFetchData struct {
	Email string           `json:"email"`
	Token string           `json:"token"`
	Inbox []OuleyxMailItem `json:"inbox"`
	Junk  []OuleyxMailItem `json:"junk"`
}

func ouleyxStripHTML(s string) string {
	s = html.UnescapeString(s)
	s = ouleyxHTMLTagRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func extractOuleyxCode(text string) string {
	for _, re := range ouleyxCodePatterns {
		if m := re.FindStringSubmatch(text); len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

func newOuleyxClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func ouleyxDo(client *http.Client, method, url, token, referer, hxTarget string) (string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("HX-Request", "true")
	if hxTarget != "" {
		req.Header.Set("HX-Target", hxTarget)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
		req.Header.Set("HX-Current-URL", referer)
	}
	if token != "" {
		req.Header.Set("Cookie", ouleyxCookie+"="+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, toolsvipSafePrefix(string(body), 200))
	}
	return string(body), nil
}

// ouleyxGetPage 按普通文档请求拉取页面，供邮件 render iframe 同源内容使用
func ouleyxGetPage(client *http.Client, pageURL, token, referer string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if token != "" {
		req.Header.Set("Cookie", ouleyxCookie+"="+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, toolsvipSafePrefix(string(body), 200))
	}
	return string(body), nil
}

// LoginOuleyx 登录 ouleyx 并返回会话 token
func LoginOuleyx(email, password string) (token string, err error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return "", fmt.Errorf("辅助邮箱或密码为空")
	}

	reqBody, err := json.Marshal(ouleyxLoginRequest{Address: email, Password: password})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, ouleyxLoginURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", ouleyxBaseURL)
	req.Header.Set("Referer", ouleyxBaseURL+"/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := newOuleyxClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("登录失败 HTTP %d: %s", resp.StatusCode, toolsvipSafePrefix(string(body), 200))
	}

	var parsed ouleyxLoginResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %w", err)
	}
	if parsed.Code != 0 || strings.TrimSpace(parsed.Data.Token) == "" {
		msg := strings.TrimSpace(parsed.Message)
		if msg == "" || strings.EqualFold(msg, "success") {
			msg = "登录失败"
		}
		return "", fmt.Errorf("%s", msg)
	}
	return strings.TrimSpace(parsed.Data.Token), nil
}

func parseOuleyxList(pageHTML, mailbox string) []OuleyxMailItem {
	starts := ouleyxRowRe.FindAllStringSubmatchIndex(pageHTML, -1)
	if len(starts) == 0 {
		return []OuleyxMailItem{}
	}

	items := make([]OuleyxMailItem, 0, len(starts))
	for i, loc := range starts {
		id := pageHTML[loc[2]:loc[3]]
		end := len(pageHTML)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		chunk := pageHTML[loc[0]:end]

		sender := ""
		if m := ouleyxSenderRe.FindStringSubmatch(chunk); len(m) >= 2 {
			sender = ouleyxStripHTML(m[1])
		}
		preview := ""
		if m := ouleyxSnippetRe.FindStringSubmatch(chunk); len(m) >= 2 {
			preview = strings.TrimSpace(strings.TrimPrefix(ouleyxStripHTML(m[1]), "—"))
			preview = strings.TrimSpace(strings.TrimPrefix(preview, "-"))
		}
		subject := ""
		chunkNoSnippet := ouleyxSnippetRe.ReplaceAllString(chunk, "")
		if m := ouleyxSubjectRe.FindStringSubmatch(chunkNoSnippet); len(m) >= 2 {
			subject = ouleyxStripHTML(m[1])
		}
		receivedAt := ""
		if m := ouleyxTimeRe.FindStringSubmatch(chunk); len(m) >= 2 {
			receivedAt = ouleyxStripHTML(m[1])
		}

		code := extractOuleyxCode(subject + " " + preview)
		items = append(items, OuleyxMailItem{
			ID:         id,
			Subject:    subject,
			From:       sender,
			ReceivedAt: receivedAt,
			Preview:    preview,
			Mailbox:    mailbox,
			Code:       code,
		})
	}
	return items
}

func parseOuleyxDetail(pageHTML string) *OuleyxMailDetail {
	detail := &OuleyxMailDetail{}
	if m := ouleyxDetailSubjectRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		detail.Subject = ouleyxStripHTML(m[1])
	}
	name := ""
	if m := ouleyxDetailNameRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		name = ouleyxStripHTML(m[1])
	}
	fromEmail := ""
	if m := ouleyxDetailEmailRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		fromEmail = ouleyxStripHTML(m[1])
		fromEmail = strings.Trim(fromEmail, "<> ")
	}
	switch {
	case name != "" && fromEmail != "":
		detail.From = name + " <" + fromEmail + ">"
	case fromEmail != "":
		detail.From = fromEmail
	default:
		detail.From = name
	}
	if m := ouleyxDetailTimeRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		detail.ReceivedAt = ouleyxStripHTML(m[1])
	}

	if m := ouleyxPlainRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		detail.Body = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	if detail.Body == "" {
		if m := ouleyxBodyRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
			detail.HtmlBody = strings.TrimSpace(m[1])
			detail.Body = ouleyxStripHTML(m[1])
		}
	}
	detail.Code = extractOuleyxCode(detail.Subject + " " + detail.Body)
	return detail
}

func applyOuleyxRender(detail *OuleyxMailDetail, rendered string) {
	rendered = strings.TrimSpace(rendered)
	if detail == nil || rendered == "" || strings.Contains(rendered, "登录您的企业邮箱") {
		return
	}
	detail.HtmlBody = rendered
	plain := ouleyxStripHTML(ouleyxStyleScriptRe.ReplaceAllString(rendered, " "))
	if plain != "" {
		detail.Body = plain
	}
	if code := extractOuleyxCode(detail.Subject + " " + detail.Body); code != "" {
		detail.Code = code
	}
}

func fetchOuleyxMailbox(client *http.Client, token, mailbox, mailboxLabel string) ([]OuleyxMailItem, error) {
	pageURL := fmt.Sprintf("%s/app/list?mailbox=%s", ouleyxBaseURL, mailbox)
	referer := fmt.Sprintf("%s/app?mailbox=%s", ouleyxBaseURL, mailbox)
	html, err := ouleyxDo(client, http.MethodGet, pageURL, token, referer, "uListBody")
	if err != nil {
		return nil, fmt.Errorf("获取%s失败: %w", mailboxLabel, err)
	}
	return parseOuleyxList(html, mailboxLabel), nil
}

func resolveOuleyxToken(email, password, token string) (string, error) {
	if strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token), nil
	}
	return LoginOuleyx(email, password)
}

// FetchOuleyxMails 登录后拉取收件箱与垃圾箱列表
func FetchOuleyxMails(email, password string) (*OuleyxFetchData, error) {
	token, err := LoginOuleyx(email, password)
	if err != nil {
		return nil, err
	}

	client := newOuleyxClient()
	inbox, err := fetchOuleyxMailbox(client, token, "inbox", "收件箱")
	if err != nil {
		return nil, err
	}
	junk, junkErr := fetchOuleyxMailbox(client, token, "junk", "垃圾箱")
	if junkErr != nil {
		junk = []OuleyxMailItem{}
	}

	return &OuleyxFetchData{
		Email: email,
		Token: token,
		Inbox: inbox,
		Junk:  junk,
	}, nil
}

// FetchOuleyxMailDetail 拉取单封邮件详情
func FetchOuleyxMailDetail(email, password, token, mailID string) (*OuleyxMailDetail, error) {
	mailID = strings.TrimSpace(mailID)
	if mailID == "" {
		return nil, fmt.Errorf("邮件 ID 为空")
	}

	session, err := resolveOuleyxToken(email, password, token)
	if err != nil {
		return nil, err
	}

	client := newOuleyxClient()
	pageURL := fmt.Sprintf("%s/app/emails/%s/detail", ouleyxBaseURL, mailID)
	html, err := ouleyxDo(client, http.MethodGet, pageURL, session, ouleyxBaseURL+"/app?mailbox=inbox", "uDetailScroll")
	if err != nil {
		return nil, fmt.Errorf("获取邮件详情失败: %w", err)
	}
	detail := parseOuleyxDetail(html)

	renderURL := fmt.Sprintf("%s/app/emails/%s/render", ouleyxBaseURL, mailID)
	rendered, renderErr := ouleyxGetPage(client, renderURL, session, pageURL)
	if renderErr == nil {
		applyOuleyxRender(detail, rendered)
	}
	return detail, nil
}

// IsOuleyxHost 判断辅助邮箱地址是否为已对接的 ouleyx
func IsOuleyxHost(host string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(host)), "ouleyx.cc")
}
