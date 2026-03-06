package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	StartDate       string          `json:"startDate"`
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

// checkCards 检查所有 cards_* 表中未出售且有 token 的卡密
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
			// 从 token 中提取 workos_id
			workosID, err := extractWorkosID(card.Token)
			if err != nil {
				logger.SysError(fmt.Sprintf("表 %s 卡密 [ID:%d, Account:%s] 提取 workos_id 失败: %v",
					tableName, card.Id, card.Account, err))
				totalSkip++
				continue
			}

			// 调用 cursor.com 接口检查订阅状态
			usageData, err := fetchUsageSummary(workosID, card.Token)
			now := time.Now().Unix()

			if err != nil {
				logger.SysError(fmt.Sprintf("表 %s 卡密 [ID:%d, Account:%s] 检查订阅失败: %v",
					tableName, card.Id, card.Account, err))
				// 请求失败：标记为检查失败
				updateErr := model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
					"is_check":   2,
					"check_time": now,
				})
				if updateErr != nil {
					logger.SysError(fmt.Sprintf("表 %s 卡密 [ID:%d] 更新检查失败状态出错: %v",
						tableName, card.Id, updateErr))
				}
				totalFail++
				time.Sleep(1 * time.Second)
				continue
			}

			if usageData.MembershipType == "free" {
				// free 账号说明已掉订阅：标记为掉订阅，并直接出售（价格 0）
				zeroPrice := 0.0
				updateErr := model.UpdateCardCheckResult(tableName, card.Id, map[string]interface{}{
					"subscription_status": -1,
					"is_check":            2,
					"check_time":          now,
					"sell_status":         3,
					"sell_price":          zeroPrice,
					"sell_date":           now,
				})
				if updateErr != nil {
					logger.SysError(fmt.Sprintf("表 %s 卡密 [ID:%d] 更新掉订阅状态出错: %v",
						tableName, card.Id, updateErr))
				} else {
					logger.SysLog(fmt.Sprintf("表 %s 卡密 [ID:%d, Account:%s] 已掉订阅(free)，已标记出售",
						tableName, card.Id, card.Account))
				}
				totalFail++
			} else {
				// 正常订阅：更新订阅类型、剩余额度和检查状态
				updates := map[string]interface{}{
					"subscription_type": usageData.MembershipType,
					"is_check":          1,
					"check_time":        now,
				}

				// 计算剩余额度（单位从分转换为美元）
				if usageData.IndividualUsage.Plan.Remaining != nil {
					remainingUSD := float64(*usageData.IndividualUsage.Plan.Remaining) / 100.0
					updates["subscription_credits"] = remainingUSD
				}

				// 解析订阅开始时间（格式如 "2024-01-15T00:00:00Z"）
				if usageData.StartDate != "" {
					if t, err := time.Parse(time.RFC3339, usageData.StartDate); err == nil {
						ts := t.Unix()
						updates["subscription_time"] = ts
					} else if t, err := time.Parse("2006-01-02", usageData.StartDate); err == nil {
						ts := t.Unix()
						updates["subscription_time"] = ts
					}
				}

				updateErr := model.UpdateCardCheckResult(tableName, card.Id, updates)
				if updateErr != nil {
					logger.SysError(fmt.Sprintf("表 %s 卡密 [ID:%d] 更新检查成功状态出错: %v",
						tableName, card.Id, updateErr))
				} else {
					logger.SysLog(fmt.Sprintf("表 %s 卡密 [ID:%d, Account:%s] 检查成功，订阅类型: %s",
						tableName, card.Id, card.Account, usageData.MembershipType))
				}
				totalSuccess++
			}

			// 每张卡间隔 1 秒，避免触发频率限制
			time.Sleep(1 * time.Second)
		}
	}

	logger.SysLog(fmt.Sprintf("卡密订阅检查完成，成功: %d, 掉订阅/失败: %d, 跳过: %d",
		totalSuccess, totalFail, totalSkip))
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

	if usageData.MembershipType == "free" {
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
		"subscription_type": usageData.MembershipType,
		"is_check":          1,
		"check_time":        now,
	}
	if usageData.IndividualUsage.Plan.Remaining != nil {
		remainingUSD := float64(*usageData.IndividualUsage.Plan.Remaining) / 100.0
		updates["subscription_credits"] = remainingUSD
	}
	if usageData.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, usageData.StartDate); err == nil {
			updates["subscription_time"] = t.Unix()
		} else if t, err := time.Parse("2006-01-02", usageData.StartDate); err == nil {
			updates["subscription_time"] = t.Unix()
		}
	}
	return model.UpdateCardCheckResult(tableName, card.Id, updates)
}
