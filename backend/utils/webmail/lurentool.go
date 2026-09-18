package webmail

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	lurentoolMailFetchURL   = "https://lurentool.cn/api/mail/fetch"
	lurentoolMailDefaultTop = 20
	lurentoolMailMaxTop     = 50
)

var (
	lurentoolCST          = time.FixedZone("CST", 8*3600)
	lurentoolHTMLTagRe    = regexp.MustCompile(`(?s)<[^>]*>`)
	lurentoolCodePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)your one-time code is\s*(\d{4,8})`),
		regexp.MustCompile(`(?i)one-time code is\s*(\d{4,8})`),
		regexp.MustCompile(`(?i)verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)temporary openai login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt verification code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)temporary chatgpt login code\s*(?:is\s*)?[:\s]*(\d{6})`),
		regexp.MustCompile(`(?i)enter this temporary verification code to continue:\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your OpenAI code is\s*(\d{6})`),
		regexp.MustCompile(`(?i)Your ChatGPT code is\s*(\d{6})`),
		regexp.MustCompile(`(?i)(?:your\s+)?(?:cursor\s+)?(?:login\s+)?code(?:\s+is)?[:\s]+(\d{6})`),
	}
)

type lurentoolMailRequest struct {
	Input string `json:"input"`
	Top   int    `json:"top"`
}

type lurentoolRawMail struct {
	Subject      string `json:"subject"`
	ReceivedTime string `json:"receivedTime"`
	Preview      string `json:"preview"`
	IsRead       bool   `json:"isRead"`
	From         string `json:"from"`
	Body         string `json:"body"`
}

type lurentoolMailResult struct {
	Email   string             `json:"email"`
	Ok      bool               `json:"ok"`
	Mails   []lurentoolRawMail `json:"mails"`
	Source  string             `json:"source"`
	Message string             `json:"message"`
}

type lurentoolMailResponse struct {
	Results []lurentoolMailResult `json:"results"`
}

// LurentoolMailItem 单封 lurentool 邮件
type LurentoolMailItem struct {
	Subject    string `json:"subject"`
	From       string `json:"from"`
	ReceivedAt string `json:"received_at"`
	Preview    string `json:"preview"`
	Body       string `json:"body"`
	IsRead     bool   `json:"is_read"`
	Code       string `json:"code"`
	Source     string `json:"source"`
}

// LurentoolFetchData lurentool 快捷取件结果
type LurentoolFetchData struct {
	Email   string              `json:"email"`
	Ok      bool                `json:"ok"`
	Source  string              `json:"source"`
	Message string              `json:"message"`
	Mails   []LurentoolMailItem `json:"mails"`
}

func lurentoolStripHTML(s string) string {
	return strings.TrimSpace(lurentoolHTMLTagRe.ReplaceAllString(s, " "))
}

func extractLurentoolCode(text string) string {
	for _, re := range lurentoolCodePatterns {
		if m := re.FindStringSubmatch(text); len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

func formatLurentoolTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.In(lurentoolCST).Format("2006-01-02 15:04:05")
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.In(lurentoolCST).Format("2006-01-02 15:04:05")
	}
	return s
}

func normalizeLurentoolTop(top int) int {
	if top <= 0 {
		return lurentoolMailDefaultTop
	}
	if top > lurentoolMailMaxTop {
		return lurentoolMailMaxTop
	}
	return top
}

func convertLurentoolMails(raw []lurentoolRawMail, source string) []LurentoolMailItem {
	items := make([]LurentoolMailItem, 0, len(raw))
	for _, mail := range raw {
		body := strings.TrimSpace(mail.Body)
		preview := strings.TrimSpace(mail.Preview)
		plain := lurentoolStripHTML(mail.Subject + " " + preview + " " + body)
		code := extractLurentoolCode(plain)
		if body == "" {
			body = preview
		}
		items = append(items, LurentoolMailItem{
			Subject:    strings.TrimSpace(mail.Subject),
			From:       strings.TrimSpace(mail.From),
			ReceivedAt: formatLurentoolTime(mail.ReceivedTime),
			Preview:    preview,
			Body:       body,
			IsRead:     mail.IsRead,
			Code:       code,
			Source:     source,
		})
	}
	return items
}

func parseLurentoolMailResponse(body []byte) (*lurentoolMailResult, error) {
	var wrapped lurentoolMailResponse
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Results) > 0 {
		return &wrapped.Results[0], nil
	}

	var list []lurentoolMailResult
	if err := json.Unmarshal(body, &list); err == nil && len(list) > 0 {
		return &list[0], nil
	}

	var single lurentoolMailResult
	if err := json.Unmarshal(body, &single); err == nil && (single.Email != "" || len(single.Mails) > 0 || single.Message != "") {
		return &single, nil
	}

	return nil, fmt.Errorf("解析响应失败: %s", toolsvipSafePrefix(string(body), 200))
}

func buildLurentoolMailInput(email, password string) string {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" {
		return ""
	}
	if password == "" {
		return email
	}
	return email + "----" + password
}

func newLurentoolClient() *http.Client {
	return &http.Client{Timeout: 45 * time.Second}
}

// FetchLurentoolMails 通过 lurentool 快捷取件接口拉取邮件
func FetchLurentoolMails(email, password string, top int) (*LurentoolFetchData, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return nil, fmt.Errorf("邮箱或密码为空")
	}

	reqBody, err := json.Marshal(lurentoolMailRequest{
		Input: buildLurentoolMailInput(email, password),
		Top:   normalizeLurentoolTop(top),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, lurentoolMailFetchURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://lurentool.cn")
	req.Header.Set("Referer", "https://lurentool.cn/mail")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := newLurentoolClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("取件过于频繁，请稍后再试")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lurentool 接口 HTTP %d: %s", resp.StatusCode, toolsvipSafePrefix(string(body), 200))
	}

	raw, err := parseLurentoolMailResponse(body)
	if err != nil {
		return nil, err
	}

	data := &LurentoolFetchData{
		Email:   strings.TrimSpace(raw.Email),
		Ok:      raw.Ok,
		Source:  strings.TrimSpace(raw.Source),
		Message: strings.TrimSpace(raw.Message),
		Mails:   convertLurentoolMails(raw.Mails, strings.TrimSpace(raw.Source)),
	}
	if data.Email == "" {
		data.Email = email
	}
	if data.Mails == nil {
		data.Mails = []LurentoolMailItem{}
	}
	if !data.Ok {
		if data.Message == "" {
			data.Message = "取件失败"
		}
		return data, fmt.Errorf("%s", data.Message)
	}
	return data, nil
}
