package webmail

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const lqqqBaseURL = "https://ms.lqqq.cc/web/"

var (
	lqqqSubjectRe      = regexp.MustCompile(`(?is)<div\s+class=["']email-subject["'][^>]*>(.*?)</div>`)
	lqqqDateRe         = regexp.MustCompile(`(?is)<(?:div|time)\s+class=["']email-date["'][^>]*>([^<]*)</(?:div|time)>`)
	lqqqCardRe         = regexp.MustCompile(`(?is)<article\s+class=["'][^"']*email-row[^"']*["'][^>]*>`)
	lqqqPanelRe        = regexp.MustCompile(`(?is)<div\s+class=["']mail-panel["']`)
	lqqqTitleRe        = regexp.MustCompile(`(?is)<div\s+class=["']mail-panel-header["'][^>]*>.*?<h1[^>]*>([^<]*)</h1>`)
	lqqqViewOpenRe     = regexp.MustCompile(`(?is)<a\b[^>]*\bemail-open\b[^>]*>`)
	lqqqViewHrefRe     = regexp.MustCompile(`(?is)href=["'](/?[^"']*show_email[^"']*)["']`)
	lqqqHrefAttrRe     = regexp.MustCompile(`(?is)href=["']([^"']+)["']`)
	lqqqHTMLTagRe      = regexp.MustCompile(`(?s)<[^>]*>`)
	lqqqDetailIframeRe = regexp.MustCompile(`(?is)<iframe[^>]*\bmessage-body\b[^>]*\ssrcdoc=["'](.*?)["']`)
	lqqqPreheaderRe    = regexp.MustCompile(`(?is)<div\s+class=["']preheader["'][^>]*>([^<]*)</div>`)
	lqqqH1Re           = regexp.MustCompile(`(?is)<h1[^>]*>\s*(\d{6})\s*</h1>`)
	lqqqHeadRe         = regexp.MustCompile(`(?is)<head[^>]*>.*?</head>`)
	lqqqCodePatterns   = []*regexp.Regexp{
		regexp.MustCompile(`(?i)verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)temporary openai login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt verification code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)enter this temporary verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your OpenAI code is\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your ChatGPT code is\s*(\d{6})`),
	}
)

// LqqqMailItem 单封 lqqq 邮件
type LqqqMailItem struct {
	Subject  string `json:"subject"`
	Date     string `json:"date"`
	Mailbox  string `json:"mailbox"`
	Code     string `json:"code"`
	ViewHref string `json:"view_href"`
}

// LqqqMailDetail lqqq 邮件详情
type LqqqMailDetail struct {
	Body     string `json:"body"`
	HtmlBody string `json:"html_body"`
	Code     string `json:"code"`
}

// ParseWebMailAccountLine 解析 Web 邮箱账号行，仅返回取件所需的邮箱与邮箱密码。
// GPT 密码字段（格式 2/3）会被忽略。
//
//	1: 邮箱----邮箱密码（兼容 邮箱|邮箱密码）
//	2: 邮箱----GPT密码----邮箱密码
//	3: 邮箱----邮箱密码----GPT密码
func ParseWebMailAccountLine(line, format string) (email, password string, err error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", fmt.Errorf("账号行为空")
	}

	format = normalizeWebMailFormat(format)
	parts := splitWebMailLine(line, format)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("账号行格式错误，字段不足")
	}

	email = strings.TrimSpace(parts[0])
	password, err = webMailMailboxPassword(format, parts)
	if err != nil {
		return "", "", err
	}

	if email == "" {
		return "", "", fmt.Errorf("邮箱不能为空")
	}
	if password == "" {
		return "", "", fmt.Errorf("邮箱密码不能为空")
	}
	return email, password, nil
}

// webMailMailboxPassword 按格式取出邮箱密码，忽略 GPT 密码段
func webMailMailboxPassword(format string, parts []string) (string, error) {
	switch format {
	case "2":
		if len(parts) < 3 {
			return "", fmt.Errorf("格式2 需要 邮箱----GPT密码----邮箱密码")
		}
		// parts[1] 为 GPT 密码，取件不使用
		return strings.TrimSpace(parts[2]), nil
	case "3":
		if len(parts) < 2 {
			return "", fmt.Errorf("格式3 需要 邮箱----邮箱密码----GPT密码")
		}
		// parts[2] 为 GPT 密码（若存在），取件不使用
		return strings.TrimSpace(parts[1]), nil
	default:
		return strings.TrimSpace(parts[1]), nil
	}
}

func splitWebMailLine(line, format string) []string {
	if format == "1" && strings.Contains(line, "|") && !strings.Contains(line, "----") {
		return strings.SplitN(line, "|", 2)
	}
	switch format {
	case "2", "3":
		// 最多分 3 段，避免邮箱密码内含 ---- 被误切
		return strings.SplitN(line, "----", 3)
	default:
		return strings.SplitN(line, "----", 2)
	}
}

func normalizeWebMailFormat(f string) string {
	switch f {
	case "f2", "2", "4":
		return "2"
	case "f3", "3", "5":
		return "3"
	default:
		return "1"
	}
}

func lqqqStripHTML(s string) string {
	return strings.TrimSpace(lqqqHTMLTagRe.ReplaceAllString(s, " "))
}

func extractLqqqCode(text string) string {
	for _, re := range lqqqCodePatterns {
		if m := re.FindStringSubmatch(text); len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

func isOTPSubject(subject string) bool {
	s := strings.ToLower(strings.TrimSpace(subject))
	return strings.Contains(s, "temporary openai login code") ||
		strings.Contains(s, "temporary openai verification") ||
		strings.Contains(s, "temporary chatgpt login code") ||
		strings.Contains(s, "temporary chatgpt verification") ||
		strings.Contains(s, "your chatgpt code") ||
		strings.Contains(s, "your openai code") ||
		strings.Contains(s, "chatgpt log-in code") ||
		(strings.Contains(s, "chatgpt") && strings.Contains(s, "verification code")) ||
		(strings.Contains(s, "openai") && strings.Contains(s, "code"))
}

func lqqqViewHref(chunk string) string {
	if tag := lqqqViewOpenRe.FindString(chunk); tag != "" {
		if m := lqqqHrefAttrRe.FindStringSubmatch(tag); len(m) >= 2 {
			return strings.TrimSpace(m[1])
		}
	}
	if m := lqqqViewHrefRe.FindStringSubmatch(chunk); len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func newLqqqClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func lqqqGet(client *http.Client, pageURL, referer string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	if referer != "" {
		req.Header.Set("Referer", referer)
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
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return string(body), nil
}

func extractCodeFromLqqqMailBody(bodyHTML string) string {
	if m := lqqqPreheaderRe.FindStringSubmatch(bodyHTML); len(m) >= 2 {
		if code := extractLqqqCode(lqqqStripHTML(m[1])); code != "" {
			return code
		}
	}
	if m := lqqqH1Re.FindStringSubmatch(bodyHTML); len(m) >= 2 {
		return m[1]
	}
	return extractLqqqCode(lqqqStripHTML(bodyHTML))
}

func extractCodeFromLqqqDetail(pageHTML string) string {
	if m := lqqqDetailIframeRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		if code := extractCodeFromLqqqMailBody(html.UnescapeString(m[1])); code != "" {
			return code
		}
	}
	noHead := lqqqHeadRe.ReplaceAllString(pageHTML, "")
	return extractLqqqCode(lqqqStripHTML(noHead))
}

func extractLqqqDetailContent(pageHTML string) (htmlBody, plainBody, code string) {
	code = extractCodeFromLqqqDetail(pageHTML)
	if m := lqqqDetailIframeRe.FindStringSubmatch(pageHTML); len(m) >= 2 {
		htmlBody = html.UnescapeString(m[1])
		plainBody = lqqqStripHTML(htmlBody)
		return htmlBody, plainBody, code
	}
	plainBody = lqqqStripHTML(pageHTML)
	return "", plainBody, code
}

// FetchLqqqMailDetail 按详情页链接拉取邮件正文
func FetchLqqqMailDetail(email, password, viewHref string) (*LqqqMailDetail, error) {
	viewHref = strings.TrimSpace(viewHref)
	if viewHref == "" {
		return nil, fmt.Errorf("缺少邮件详情链接")
	}

	client := newLqqqClient()
	pageURL := lqqqBaseURL + url.PathEscape(email) + "----" + url.PathEscape(password)

	viewURL, err := url.Parse(viewHref)
	if err != nil {
		return nil, fmt.Errorf("详情链接无效: %w", err)
	}
	if !viewURL.IsAbs() {
		base, err := url.Parse(pageURL)
		if err != nil {
			return nil, err
		}
		viewURL = base.ResolveReference(viewURL)
	}

	html, err := lqqqGet(client, viewURL.String(), pageURL)
	if err != nil {
		return nil, fmt.Errorf("拉取详情失败: %w", err)
	}

	htmlBody, plainBody, code := extractLqqqDetailContent(html)
	return &LqqqMailDetail{
		Body:     plainBody,
		HtmlBody: htmlBody,
		Code:     code,
	}, nil
}

func fetchDetailCode(client *http.Client, pageURL, href string) string {
	viewURL, err := url.Parse(href)
	if err != nil {
		return ""
	}
	if !viewURL.IsAbs() {
		base, err := url.Parse(pageURL)
		if err != nil {
			return ""
		}
		viewURL = base.ResolveReference(viewURL)
	}
	html, err := lqqqGet(client, viewURL.String(), pageURL)
	if err != nil {
		return ""
	}
	return extractCodeFromLqqqDetail(html)
}

func parseLqqqSection(client *http.Client, pageURL, sectionHTML, title string) []LqqqMailItem {
	var items []LqqqMailItem

	cardStarts := lqqqCardRe.FindAllStringIndex(sectionHTML, -1)
	for i, idx := range cardStarts {
		start := idx[0]
		end := len(sectionHTML)
		if i+1 < len(cardStarts) {
			end = cardStarts[i+1][0]
		}
		chunk := sectionHTML[start:end]

		subject := ""
		if m := lqqqSubjectRe.FindStringSubmatch(chunk); len(m) >= 2 {
			subject = strings.TrimSpace(lqqqStripHTML(m[1]))
		}
		if subject == "" {
			continue
		}

		date := ""
		if m := lqqqDateRe.FindStringSubmatch(chunk); len(m) >= 2 {
			date = strings.TrimSpace(m[1])
		}

		viewHref := lqqqViewHref(chunk)
		code := ""
		if isOTPSubject(subject) && viewHref != "" {
			code = fetchDetailCode(client, pageURL, viewHref)
		}

		items = append(items, LqqqMailItem{
			Subject:  subject,
			Date:     date,
			Mailbox:  title,
			Code:     code,
			ViewHref: viewHref,
		})
	}
	return items
}

func forEachLqqqPanel(pageHTML string, fn func(sectionHTML, title string)) {
	panelStarts := lqqqPanelRe.FindAllStringIndex(pageHTML, -1)
	if len(panelStarts) == 0 {
		fn(pageHTML, "")
		return
	}
	for i, idx := range panelStarts {
		start := idx[0]
		end := len(pageHTML)
		if i+1 < len(panelStarts) {
			end = panelStarts[i+1][0]
		} else if close := strings.Index(pageHTML[start:], "</section>"); close >= 0 {
			end = start + close
		}
		body := pageHTML[start:end]
		title := ""
		if tm := lqqqTitleRe.FindStringSubmatch(body); len(tm) >= 2 {
			title = strings.TrimSpace(tm[1])
		}
		fn(body, title)
	}
}

// FetchLqqqMails 获取 lqqq 收件箱与垃圾箱邮件
func FetchLqqqMails(email, password string) (inbox, junk []LqqqMailItem, err error) {
	client := newLqqqClient()
	pageURL := lqqqBaseURL + url.PathEscape(email) + "----" + url.PathEscape(password)

	pageHTML, err := lqqqGet(client, pageURL, "")
	if err != nil {
		return nil, nil, fmt.Errorf("访问 lqqq 页面失败: %w", err)
	}

	forEachLqqqPanel(pageHTML, func(sectionHTML, title string) {
		items := parseLqqqSection(client, pageURL, sectionHTML, title)
		titleLower := strings.ToLower(title)
		if title == "" || strings.Contains(titleLower, "inbox") || strings.Contains(titleLower, "收件") {
			inbox = append(inbox, items...)
		} else {
			junk = append(junk, items...)
		}
	})

	if inbox == nil {
		inbox = []LqqqMailItem{}
	}
	if junk == nil {
		junk = []LqqqMailItem{}
	}
	return inbox, junk, nil
}
