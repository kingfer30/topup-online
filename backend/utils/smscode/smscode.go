// Package smscode 用于从第三方接码平台抓取 Cursor 账号的短信验证码。
// 目前支持三套 phone_link 格式：
//  1. api1997.com   —— 直接 GET phone_link，返回纯文本/HTML 页面
//  2. sms.toolsvip.cc —— 页面本身是静态壳子，真实数据来自 /api/query 接口，返回 JSON
//  3. lurentool.cn  —— 固定入口页，真实数据通过 POST /api/sms/fetch 提交 "账号-密码" 获取
package smscode

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

// 状态常量
const (
	StatusReceived = "received" // 已收到验证码
	StatusWaiting  = "waiting"  // 暂未收到，等待中
	StatusError    = "error"    // 查询失败/接口异常
)

// Result 统一的查询结果
type Result struct {
	Status    string `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	ExpiresAt string `json:"expires_at"`
}

var (
	htmlTagRe = regexp.MustCompile(`(?s)<[^>]*>`)
	// 各平台转发的短信原文基本都是 "123456 is your verification code for Cursor. Do not share it."
	verificationCodeRe = regexp.MustCompile(`(?i)(\d{6})\s+is\s+your\s+verification\s+code`)
	dateTimeRe         = regexp.MustCompile(`\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}`)
)

func httpClient() *http.Client {
	return &http.Client{Timeout: 20 * time.Second}
}

func stripHTML(s string) string {
	return strings.TrimSpace(htmlTagRe.ReplaceAllString(s, " "))
}

// extractCode 从任意文本中提取 6 位验证码，仅在明确出现 "is your verification code" 语境时才提取，
// 避免把日期、到期时间等其他数字误判为验证码。
func extractCode(text string) string {
	if m := verificationCodeRe.FindStringSubmatch(text); len(m) >= 2 {
		return m[1]
	}
	return ""
}

// FetchCode 根据 phone_link 所属平台抓取验证码。
// account/password 仅 lurentool 平台需要用来拼接请求体，其余平台可传空。
func FetchCode(phoneLink, account, password string) (*Result, error) {
	phoneLink = strings.TrimSpace(phoneLink)
	if phoneLink == "" {
		return nil, fmt.Errorf("phone_link 为空")
	}

	u, err := url.Parse(phoneLink)
	if err != nil {
		return nil, fmt.Errorf("phone_link 无效: %w", err)
	}
	host := strings.ToLower(u.Hostname())

	switch {
	case strings.Contains(host, "toolsvip.cc"):
		return fetchToolsvip(phoneLink)
	case strings.Contains(host, "lurentool.cn"):
		return fetchLurentool(account, password)
	case strings.Contains(host, "api1997.com"):
		return fetchApi1997(phoneLink)
	default:
		// 未识别的平台，按纯文本页面尝试通用解析
		return fetchGenericPage(phoneLink)
	}
}

// fetchApi1997 直接 GET 页面，返回纯文本内容，例如：
// "该号将于2026-11-15 00:00:00到期，请立即换绑，如因逾期换绑的，后果自负！！！ 暂时未有邮件...试试重发吧."
func fetchApi1997(link string) (*Result, error) {
	body, err := httpGetBody(link)
	if err != nil {
		return nil, err
	}
	text := stripHTML(body)

	result := &Result{Message: text}
	if code := extractCode(text); code != "" {
		result.Status = StatusReceived
		result.Code = code
		return result, nil
	}

	if m := dateTimeRe.FindString(text); m != "" {
		result.ExpiresAt = m
	}
	result.Status = StatusWaiting
	return result, nil
}

// fetchGenericPage 未识别平台时的兜底解析：直接 GET 页面并做通用文本提取
func fetchGenericPage(link string) (*Result, error) {
	return fetchApi1997(link)
}

type toolsvipQueryResponse struct {
	Success   bool   `json:"success"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	ExpiresAt string `json:"expiresAt"`
}

// fetchToolsvip phone_link 形如 https://sms.toolsvip.cc/?token=xxx，
// 真实数据来自 https://sms.toolsvip.cc/api/query?token=xxx
func fetchToolsvip(link string) (*Result, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("phone_link 无效: %w", err)
	}
	token := u.Query().Get("token")
	if token == "" {
		return nil, fmt.Errorf("phone_link 缺少 token 参数")
	}

	apiURL := fmt.Sprintf("https://%s/api/query?token=%s", u.Hostname(), url.QueryEscape(token))
	body, err := httpGetBody(apiURL)
	if err != nil {
		return nil, err
	}

	var resp toolsvipQueryResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	result := &Result{Message: resp.Message, ExpiresAt: resp.ExpiresAt}
	if resp.Success && resp.Status == "received" {
		result.Status = StatusReceived
		code := extractCode(resp.Message)
		if code == "" {
			// 兜底：消息本身可能就是纯验证码
			if m := regexp.MustCompile(`\d{6}`).FindString(resp.Message); m != "" {
				code = m
			}
		}
		result.Code = code
		return result, nil
	}

	switch resp.Status {
	case "waiting":
		result.Status = StatusWaiting
	case "invalid":
		result.Status = StatusError
	default:
		result.Status = StatusError
	}
	return result, nil
}

type lurentoolRequest struct {
	Input string `json:"input"`
}

type lurentoolResultItem struct {
	Account    string `json:"account"`
	Ok         bool   `json:"ok"`
	Code       string `json:"code"`
	SmsText    string `json:"smsText"`
	ExpireTime string `json:"expireTime"`
	Cached     bool   `json:"cached"`
	CachedAt   string `json:"cachedAt"`
	Message    string `json:"message"`
}

type lurentoolResponse struct {
	Results []lurentoolResultItem `json:"results"`
}

// fetchLurentool 通过 POST https://lurentool.cn/api/sms/fetch 提交 "账号-密码" 获取验证码
func fetchLurentool(account, password string) (*Result, error) {
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if account == "" || password == "" {
		return nil, fmt.Errorf("lurentool 平台需要账号与密码")
	}

	reqBody, err := json.Marshal(lurentoolRequest{Input: account + "-" + password})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://lurentool.cn/api/sms/fetch", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lurentool 接口 HTTP %d: %s", resp.StatusCode, string(body))
	}

	var parsed lurentoolResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if len(parsed.Results) == 0 {
		return nil, fmt.Errorf("lurentool 未返回结果")
	}

	item := parsed.Results[0]
	result := &Result{Message: item.Message, ExpiresAt: item.ExpireTime}
	if item.Ok && item.Code != "" {
		result.Status = StatusReceived
		result.Code = item.Code
		if result.Message == "" {
			result.Message = item.SmsText
		}
		return result, nil
	}

	// lurentool 未收到验证码时统一视为等待中，不区分具体原因
	result.Status = StatusWaiting
	return result, nil
}

func httpGetBody(link string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/html, */*")

	client := httpClient()
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
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}
