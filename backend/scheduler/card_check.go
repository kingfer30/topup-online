package scheduler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

// CardCheckScheduler 卡密订阅检查定时器
type CardCheckScheduler struct {
	interval time.Duration // 轮询间隔
	running  bool          // 是否正在运行
	stopChan chan bool     // 停止信号
	mu       sync.Mutex    // 互斥锁，防止同时执行
}

// UsageSummaryResponse cursor.com/api/usage-summary 响应结构
type UsageSummaryResponse struct {
	MembershipType  string          `json:"membershipType"`
	BillingCycleStart string          `json:"billingCycleStart"`
	IndividualUsage IndividualUsage `json:"individualUsage"`
}

// IndividualUsage 个人用量数据
type IndividualUsage struct {
	Plan PlanUsage `json:"plan"`
}

// PlanUsage 套餐用量数据
type PlanUsage struct {
	Remaining *int64 `json:"remaining"`
}

type chatGptOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type chatGptAccountsCheckResponse struct {
	AccountOrdering []interface{} `json:"account_ordering"`
}

type chatGptSubscriptionsResponse struct {
	PlanType    string `json:"plan_type"`
	ActiveUntil string `json:"active_until"`
}

// ErrChatGPTAccountBanned chatgpt.com 接口返回 403 且账号已删除/停用
var ErrChatGPTAccountBanned = errors.New("chatgpt account banned")

var chatGPTAccountBannedKeywords = []string{
	"do not have an account",
	"has been deleted",
	"been deleted",
	"deactivated",
	"account because it has been deleted",
}

// normalizeMembershipType 将 Cursor API 返回的订阅类型规范化为系统内使用的值
func normalizeMembershipType(raw string) string {
	t := strings.TrimSpace(strings.ToLower(raw))
	switch t {
	case "", "free":
		return "free"
	case "pro", "chatgpt_pro":
		return "pro"
	case "pro_plus", "proplus", "pro+":
		return "pro_plus"
	case "pro_x5", "prox5":
		return "pro_x5"
	case "pro_x20", "prox20":
		return "pro_x20"
	case "ultra":
		return "ultra"
	case "go":
		return "go"
	case "plus":
		return "plus"
	case "team":
		return "team"
	default:
		return t
	}
}

// isPaidSubscriptionType 是否为付费订阅类型（手动标记为成品时使用）
func isPaidSubscriptionType(subscriptionType string) bool {
	switch normalizeMembershipType(subscriptionType) {
	case "pro", "pro_plus", "pro_x5", "pro_x20", "ultra", "go", "plus", "team":
		return true
	default:
		return false
	}
}

// shouldPreserveManualSubscription 手动升级为成品后 API 仍可能短暂返回 free，此时保留库内订阅类型
func shouldPreserveManualSubscription(card *model.AccountCard, apiType string) bool {
	if apiType != "free" {
		return false
	}
	return card.AccountType == 2 && card.SubscriptionStatus == 1 && isPaidSubscriptionType(card.SubscriptionType)
}

var globalCardCheckScheduler *CardCheckScheduler

// NewCardCheckScheduler 创建卡密订阅检查定时器
func NewCardCheckScheduler(intervalMinutes int) *CardCheckScheduler {
	s := &CardCheckScheduler{
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan bool),
	}
	globalCardCheckScheduler = s
	return s
}

// Start 启动定时任务
func (s *CardCheckScheduler) Start() {
	if s.running {
		logger.SysLog("CardCheckScheduler is already running")
		return
	}

	s.running = true
	logger.SysLog("CardCheckScheduler started")

	// 启动时立即执行一次
	go s.checkCards()

	// 启动定时器
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.checkCards()
			case <-s.stopChan:
				ticker.Stop()
				logger.SysLog("CardCheckScheduler stopped")
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *CardCheckScheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false
	s.stopChan <- true
}

// checkCards 检查所有 cards_* 表中未出售成品且有 token 的卡密（排除 subscription_status=-2）
func (s *CardCheckScheduler) checkCards() {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog("开始检查卡密订阅状态...")

	// 获取所有 cards_* 表名
	tableNames, err := model.GetAllCardTableNames()
	if err != nil {
		logger.SysError(fmt.Sprintf("获取 cards_* 表名失败: %v", err))
		return
	}

	if len(tableNames) == 0 {
		logger.SysLog("没有找到任何 cards_* 表")
		return
	}

	totalSuccess := 0
	totalFail := 0
	totalSkip := 0

	for _, tableName := range tableNames {
		// 每次每个表最多处理 50 条
		cards, err := model.GetUnsoldCardsWithToken(tableName, 50)
		if err != nil {
			logger.SysError(fmt.Sprintf("查询表 %s 的卡密失败: %v", tableName, err))
			continue
		}

		if len(cards) == 0 {
			continue
		}

		logger.SysLog(fmt.Sprintf("表 %s 找到 %d 条待检查卡密", tableName, len(cards)))

		for _, card := range cards {
			// 按表名分流：cards_chatgpt* 走 ChatGPT 检查，其余走 Cursor（workos_id）检查
			if err := CheckSingleCard(tableName, card); err != nil {
				logger.SysError(fmt.Sprintf("表 %s 卡密 [ID:%d, Account:%s] 检查失败: %v",
+					tableName, card.Id, card.Account, err))
				if strings.Contains(err.Error(), "提取 workos_id") {
					totalSkip++
				} else {
					totalFail++
				}
			} else {
				logger.SysLog(fmt.Sprintf("表 %s 卡密 [ID:%d, Account:%s] 检查成功",
					tableName, card.Id, card.Account))
				totalSuccess++
			}

			// 每张卡间隔 1 秒，避免触发频率限制
			time.Sleep(1 * time.Second)
		}
	}

	logger.SysLog(fmt.Sprintf("卡密订阅检查完成，成功: %d, 掉订阅/失败: %d, 跳过: %d",
		totalSuccess, totalFail, totalSkip))
}

// ExtractWorkosID 从 Cursor token 中提取 workos_id（导出供外部调用）
func ExtractWorkosID(token string) (string, error) {
	return extractWorkosID(token)
}

// extractWorkosID 从 Cursor token 中提取 workos_id
// Cursor token 格式通常为 user_XXXXX%3A%3ATOKEN 或 user_XXXXX::TOKEN
func extractWorkosID(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token 为空")
	}

	// URL 解码
	decoded, err := url.QueryUnescape(token)
	if err != nil {
		// 解码失败则使用原始 token
		decoded = token
	}

	// 按 "::" 分割
	parts := strings.SplitN(decoded, "::", 2)
	if len(parts) < 2 || parts[0] == "" {
		return "", fmt.Errorf("token 格式不符合预期，无法提取 workos_id (token 前缀: %s)", safePrefix(decoded, 20))
	}

	workosID := parts[0]
	return workosID, nil
}

// safePrefix 安全截取字符串前 n 个字符
func safePrefix(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// fetchUsageSummary 调用 cursor.com/api/usage-summary 获取订阅信息
func fetchUsageSummary(workosID, token string) (*UsageSummaryResponse, error) {
	apiURL := "https://cursor.com/api/usage-summary"

	// cookie 中始终使用解码后的 token（将 %3A%3A 还原为 ::）
	decodedToken, err := url.QueryUnescape(token)
	if err != nil {
		decodedToken = token
	}
	cookie := fmt.Sprintf("workos_id=%s; WorkosCursorSessionToken=%s; ", workosID, decodedToken)
	logger.SysLog(fmt.Sprintf("cookie: %s", cookie))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("origin", "https://cursor.com")
	req.Header.Set("referer", "https://cursor.com/dashboard?tab=billing")
	req.Header.Set("cookie", cookie)

	// // 使用 HTTP 代理
	// proxyURL, _ := url.Parse("http://xiaoguo:Ji6dft4Cqd9l_eX6h3@199.119.138.75:1080")
	// transport := &http.Transport{
	// 	Proxy: http.ProxyURL(proxyURL),
	// }
	// client := &http.Client{
	// 	Timeout:   15 * time.Second,
	// 	Transport: transport,
	// }
	aClient := client.HTTPClient

	resp, err := aClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("接口返回非 200 状态码: %d, 响应: %s", resp.StatusCode, safePrefix(string(body), 200))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result UsageSummaryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %v, 响应: %s", err, safePrefix(string(body), 200))
	}

	return &result, nil
}

// CheckSingleCard 对单张卡密执行订阅检查并将结果写入数据库，供外部直接调用
func CheckSingleCard(tableName string, card *model.AccountCard) error {
	if isChatGPTCardTable(tableName) {
		return checkSingleChatGPTCard(tableName, card)
	}

	workosID, err := extractWorkosID(card.Token)
	if err != nil {
		return fmt.Errorf("提取 workos_id 失败: %v", err)
	}

	usageData, apiErr := fetchUsageSummary(workosID, card.Token)
	now := time.Now().Unix()

	if apiErr != nil {
		// 请求失败：标记检查失败
		_ = model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
			"is_check":   2,
			"check_time": now,
		})
		return apiErr
	}

	membershipType := normalizeMembershipType(usageData.MembershipType)
	if membershipType == "free" {
		if shouldPreserveManualSubscription(card, membershipType) {
			return model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
				"is_check":   2,
				"check_time": now,
			})
		}
		zeroPrice := 0.0
		return model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
			"subscription_status": -1,
			"is_check":            2,
			"check_time":          now,
			"sell_status":         3,
			"sell_price":          zeroPrice,
			"sell_date":           now,
		})
	}

	updates := map[string]interface{}{
		"subscription_type": membershipType,
		"is_check":          1,
		"check_time":        now,
	}
	if usageData.IndividualUsage.Plan.Remaining != nil {
		remainingUSD := float64(*usageData.IndividualUsage.Plan.Remaining) / 100.0
		updates["subscription_credits"] = remainingUSD
	}
	if usageData.BillingCycleStart != "" {
		if t, err := time.Parse(time.RFC3339, usageData.BillingCycleStart); err == nil {
			updates["subscription_time"] = t.Unix()
		} else if t, err := time.Parse("2006-01-02", usageData.BillingCycleStart); err == nil {
			updates["subscription_time"] = t.Unix()
		}
	}
	return model.UpdateCardCheckResult(tableName, card.Id, updates)
}

func isChatGPTCardTable(tableName string) bool {
	t := strings.ToLower(strings.TrimSpace(tableName))
	return strings.Contains(t, "chatgpt")
}

func checkSingleChatGPTCard(tableName string, card *model.AccountCard) error {
	now := time.Now().Unix()
	accessToken, newRefreshToken, err := fetchChatGPTTokens(card.Token)
	if err != nil {
		_ = model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
			"is_check":   2,
			"check_time": now,
		})
		return fmt.Errorf("chatgpt oauth 刷新失败: %v", err)
	}

	accountID, err := fetchChatGPTAccountID(accessToken)
	if err != nil {
		if errors.Is(err, ErrChatGPTAccountBanned) {
			return markChatGPTCardBanned(tableName, card.Id, newRefreshToken)
		}
		_ = model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
			"is_check":   2,
			"check_time": now,
			"token":      newRefreshToken,
		})
		return fmt.Errorf("chatgpt accounts/check 失败: %v", err)
	}

	planType, activeUntil, err := fetchChatGPTSubscription(accessToken, accountID)
	if err != nil {
		if errors.Is(err, ErrChatGPTAccountBanned) {
			return markChatGPTCardBanned(tableName, card.Id, newRefreshToken)
		}
		_ = model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
			"is_check":   2,
			"check_time": now,
			"token":      newRefreshToken,
		})
		return fmt.Errorf("chatgpt subscriptions 失败: %v", err)
	}

	membershipType := normalizeMembershipType(planType)
	updates := map[string]interface{}{
		"token":      newRefreshToken,
		"is_check":   1,
		"check_time": now,
	}

	if membershipType == "plus" {
		updates["subscription_type"] = "plus"
		updates["subscription_status"] = 1
		updates["account_type"] = 2
		updates["sell_status"] = 1
		if activeUntil > 0 {
			updates["subscription_expired_time"] = activeUntil
		}
		if activeUntil > now {
			updates["subscription_time"] = now
		}
	} else {
		updates["subscription_type"] = "free"
		updates["subscription_status"] = -1
	}

	return model.UpdateCardCheckResult(tableName, card.Id, updates)
}

func markChatGPTCardBanned(tableName string, cardID int, newRefreshToken string) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":              model.CardStatusBanned,
		"subscription_type":   "free",
		"subscription_status": -1,
		"is_check":            1,
		"check_time":          now,
	}
	if newRefreshToken != "" {
		updates["token"] = newRefreshToken
	}
	return model.UpdateCardCheckResult(tableName, cardID, updates)
}

func isChatGPTAccountBannedResponse(statusCode int, body string) bool {
	if statusCode != http.StatusForbidden {
		return false
	}
	return bodyIndicatesChatGPTAccountBanned(body)
}

func bodyIndicatesChatGPTAccountBanned(body string) bool {
	lower := strings.ToLower(body)
	for _, kw := range chatGPTAccountBannedKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	var parsed struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(body), &parsed) == nil {
		for _, msg := range []string{parsed.Detail, parsed.Message, parsed.Error.Message} {
			msgLower := strings.ToLower(strings.TrimSpace(msg))
			if msgLower == "" {
				continue
			}
			for _, kw := range chatGPTAccountBannedKeywords {
				if strings.Contains(msgLower, kw) {
					return true
				}
			}
		}
	}
	return false
}

func parseChatGPTAPIError(statusCode int, body string) error {
	if isChatGPTAccountBannedResponse(statusCode, body) {
		return ErrChatGPTAccountBanned
	}
	return fmt.Errorf("HTTP %d: %s", statusCode, safePrefix(body, 200))
}

func fetchChatGPTTokens(refreshToken string) (accessToken string, newRefreshToken string, err error) {
	reqBody := map[string]string{
		"client_id":     "app_LlGpXReQgckcGGUo2JrYvtJK",
		"grant_type":    "refresh_token",
		"redirect_uri":  "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback",
		"refresh_token": refreshToken,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://auth0.openai.com/oauth/token", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://auth0.openai.com")
	req.Header.Set("referer", "https://auth0.openai.com/")
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, safePrefix(string(rawBody), 200))
	}
	var parsed chatGptOAuthResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return "", "", fmt.Errorf("解析 oauth/token 响应失败: %v", err)
	}
	if parsed.AccessToken == "" || parsed.RefreshToken == "" {
		return "", "", fmt.Errorf("oauth/token 返回缺少 token 字段")
	}
	return parsed.AccessToken, parsed.RefreshToken, nil
}

func fetchChatGPTAccountID(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://chatgpt.com/backend-api/accounts/check/v4-2023-04-27?timezone_offset_min=-480", nil)
	if err != nil {
		return "", err
	}
	req.Header = buildChatGPTHeaders(accessToken)
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(resp.Body)
	bodyText := string(rawBody)
	if resp.StatusCode != http.StatusOK {
		return "", parseChatGPTAPIError(resp.StatusCode, bodyText)
	}
	var parsed chatGptAccountsCheckResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return "", fmt.Errorf("解析 accounts/check 响应失败: %v", err)
	}
	if len(parsed.AccountOrdering) == 0 {
		return "", fmt.Errorf("account_ordering 为空")
	}
	accountID := strings.TrimSpace(fmt.Sprintf("%v", parsed.AccountOrdering[0]))
	if accountID == "" || accountID == "<nil>" {
		return "", fmt.Errorf("account_ordering[0] 无效")
	}
	return accountID, nil
}

func fetchChatGPTSubscription(accessToken, accountID string) (planType string, activeUntil int64, err error) {
	reqURL := "https://chatgpt.com/backend-api/subscriptions?account_id=" + url.QueryEscape(accountID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header = buildChatGPTHeaders(accessToken)
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(resp.Body)
	bodyText := string(rawBody)
	if resp.StatusCode != http.StatusOK {
		trimmed := strings.TrimSpace(bodyText)
		if resp.StatusCode == http.StatusNotFound && strings.Contains(strings.ToLower(trimmed), "no subscription found for account") {
			// 无订阅视为 free，而不是检查失败
			return "free", 0, nil
		}
		return "", 0, parseChatGPTAPIError(resp.StatusCode, bodyText)
	}
	var parsed chatGptSubscriptionsResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return "", 0, fmt.Errorf("解析 subscriptions 响应失败: %v", err)
	}
	return parsed.PlanType, parseChatGPTActiveUntil(parsed.ActiveUntil), nil
}

func buildChatGPTHeaders(accessToken string) http.Header {
	h := make(http.Header)
	h.Set("accept", "*/*")
	h.Set("accept-language", "en-US,en;q=0.9")
	h.Set("authorization", "Bearer "+accessToken)
	h.Set("origin", "https://chatgpt.com")
	h.Set("referer", "https://chatgpt.com/")
	h.Set("sec-fetch-dest", "empty")
	h.Set("sec-fetch-mode", "cors")
	h.Set("sec-fetch-site", "same-origin")
	h.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	return h
}

func parseChatGPTActiveUntil(raw string) int64 {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0
	}
	if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
		if ts > 1e12 {
			return ts / 1000
		}
		return ts
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.Unix()
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t.Unix()
	}
	return 0
}

// EnableOnDemandSpend 为指定卡密开启按需付费（hardLimit=200）
func EnableOnDemandSpend(workosID, token string) error {
	// cookie 中始终使用解码后的 token
	decodedToken, err := url.QueryUnescape(token)
	if err != nil {
		decodedToken = token
	}
	cookie := fmt.Sprintf("workos_id=%s; WorkosCursorSessionToken=%s; ", workosID, decodedToken)

	apiURL := "https://cursor.com/api/dashboard/enable-on-demand-spend"
	body := strings.NewReader(`{"hardLimit":200}`)
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("origin", "https://cursor.com")
	req.Header.Set("referer", "https://cursor.com/dashboard?tab=billing")
	req.Header.Set("sec-ch-ua", `"Not(A:Brand";v="8", "Chromium";v="144", "Google Chrome";v="144"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("cookie", cookie)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("接口返回非 200 状态码: %d, 响应: %s", resp.StatusCode, safePrefix(string(respBody), 200))
	}
	return nil
}
