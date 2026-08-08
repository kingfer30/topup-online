package outlook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const msRedirectURI = "https://login.live.com/oauth20_desktop.srf"

// imapScope 是 Outlook/Office365 IMAP 通过 OAuth2(XOAUTH2) 访问所需的权限
const imapScope = "https://outlook.office.com/IMAP.AccessAsUser.All offline_access"

// graphScope 是 Microsoft Graph 读邮件所需权限（IMAP 被禁用时的回退路径）
const graphScope = "https://graph.microsoft.com/Mail.Read offline_access"
const graphDefaultScope = "https://graph.microsoft.com/.default"

// tokenStrategy 描述一种换取 access_token 的方式（端点 + scope）
type tokenStrategy struct {
	endpoint string
	scope    string
}

// RefreshTokenResult 刷新结果（含轮换后的 refresh_token）
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

// imapTokenStrategies 按优先级排列的取 token 策略。
// 现代账号（可用于 Graph 的凭据）需要走 microsoftonline + IMAP scope 才能拿到带 IMAP 权限的 token；
// 老账号则走 login.live.com（无 scope）。逐个尝试以最大化兼容性。
var imapTokenStrategies = []tokenStrategy{
	{endpoint: "https://login.microsoftonline.com/consumers/oauth2/v2.0/token", scope: imapScope},
	{endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/token", scope: imapScope},
	{endpoint: "https://login.live.com/oauth20_token.srf", scope: imapScope},
	{endpoint: "https://login.live.com/oauth20_token.srf", scope: ""},
}

// graphTokenStrategies Graph 取件用的 token 策略（与 IMAP scope 是不同资源，必须单独换票）
var graphTokenStrategies = []tokenStrategy{
	{endpoint: "https://login.microsoftonline.com/consumers/oauth2/v2.0/token", scope: graphScope},
	{endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/token", scope: graphScope},
	{endpoint: "https://login.microsoftonline.com/consumers/oauth2/v2.0/token", scope: graphDefaultScope},
	{endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/token", scope: graphDefaultScope},
}

// RefreshAccessTokens 用 refresh_token 尝试多种端点/scope 组合换取候选 access_token。
// 返回去重后的候选 token 列表，供 IMAP XOAUTH2 逐个尝试。
func RefreshAccessTokens(httpClient *http.Client, clientID, refreshToken string) ([]string, error) {
	return refreshAccessTokensWithStrategies(httpClient, clientID, refreshToken, imapTokenStrategies)
}

// RefreshGraphAccessTokens 换取 Graph API 用的 access_token（与 IMAP token 资源不同）
func RefreshGraphAccessTokens(httpClient *http.Client, clientID, refreshToken string) ([]string, error) {
	return refreshAccessTokensWithStrategies(httpClient, clientID, refreshToken, graphTokenStrategies)
}

func refreshAccessTokensWithStrategies(httpClient *http.Client, clientID, refreshToken string, strategies []tokenStrategy) ([]string, error) {
	var tokens []string
	seen := map[string]struct{}{}
	var lastErr error

	for _, s := range strategies {
		result, err := refreshOnce(httpClient, s.endpoint, s.scope, clientID, refreshToken)
		if err != nil {
			lastErr = err
			continue
		}
		if result == nil || result.AccessToken == "" {
			continue
		}
		if _, ok := seen[result.AccessToken]; ok {
			continue
		}
		seen[result.AccessToken] = struct{}{}
		tokens = append(tokens, result.AccessToken)
	}

	if len(tokens) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("未能获取任何 access_token")
	}
	return tokens, nil
}

// RefreshAccessTokenWithRotation 优先用 consumers + IMAP scope 刷新，并返回新的 refresh_token（若响应含有）。
// 用于库存保活：必须把新的 refresh_token 回写数据库。
func RefreshAccessTokenWithRotation(httpClient *http.Client, clientID, refreshToken string) (*RefreshTokenResult, error) {
	var lastErr error
	for _, s := range imapTokenStrategies {
		result, err := refreshOnce(httpClient, s.endpoint, s.scope, clientID, refreshToken)
		if err != nil {
			lastErr = err
			continue
		}
		if result == nil || result.AccessToken == "" {
			continue
		}
		// 若未返回新 RT，保留旧值，调用方可决定是否覆盖
		if result.RefreshToken == "" {
			result.RefreshToken = refreshToken
		}
		return result, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("未能获取任何 access_token")
}

// IsInvalidGrant 判断错误是否为 refresh_token 永久失效
func IsInvalidGrant(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid_grant")
}

func refreshOnce(httpClient *http.Client, endpoint, scope, clientID, refreshToken string) (*RefreshTokenResult, error) {
	vals := url.Values{
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
		"redirect_uri":  {msRedirectURI},
	}
	if scope != "" {
		vals.Set("scope", scope)
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, fmt.Errorf("构建 OAuth 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MS OAuth 请求失败: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var j map[string]interface{}
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("解析 OAuth 响应失败: %w", err)
	}

	if errDesc, ok := j["error_description"].(string); ok && errDesc != "" {
		return nil, fmt.Errorf("OAuth 授权失败: %s", errDesc)
	}
	if errCode, ok := j["error"].(string); ok && errCode != "" {
		return nil, fmt.Errorf("OAuth 错误: %s", errCode)
	}

	token, _ := j["access_token"].(string)
	if token == "" {
		preview := string(data)
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return nil, fmt.Errorf("OAuth 响应无 access_token: %s", preview)
	}

	newRT, _ := j["refresh_token"].(string)
	return &RefreshTokenResult{
		AccessToken:  token,
		RefreshToken: newRT,
	}, nil
}
