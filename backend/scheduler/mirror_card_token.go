package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kingfer30/topup-online/model"
	crypto "github.com/kingfer30/topup-online/utils/cypto"
	"github.com/kingfer30/topup-online/utils/logger"
)

// MirrorCardTokenScheduler 镜像卡密 Token 定时获取器
type MirrorCardTokenScheduler struct {
	interval time.Duration // 轮询间隔
	running  bool          // 是否正在运行
	stopChan chan bool     // 停止信号
	mu       sync.Mutex    // 互斥锁，防止同时执行
}

var globalScheduler *MirrorCardTokenScheduler

// NewMirrorCardTokenScheduler 创建镜像卡密 Token 定时获取器
func NewMirrorCardTokenScheduler(intervalMinutes int) *MirrorCardTokenScheduler {
	scheduler := &MirrorCardTokenScheduler{
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan bool),
	}
	globalScheduler = scheduler
	return scheduler
}

// Start 启动定时任务
func (s *MirrorCardTokenScheduler) Start() {
	if s.running {
		logger.SysLog("MirrorCardTokenScheduler is already running")
		return
	}

	s.running = true
	logger.SysLog("MirrorCardTokenScheduler started")

	// 立即执行一次
	go s.fetchTokens()

	// 启动定时器
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.fetchTokens()
			case <-s.stopChan:
				ticker.Stop()
				logger.SysLog("MirrorCardTokenScheduler stopped")
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *MirrorCardTokenScheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false
	s.stopChan <- true
}

// fetchTokens 获取卡密的 Token
func (s *MirrorCardTokenScheduler) fetchTokens() {
	// 使用互斥锁防止同时执行
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog("开始获取镜像卡密 Token...")

	// 查询状态正常且 token 为空的卡密（每次处理 10 条）
	cards, err := model.GetMirrorCardsWithoutToken(10)
	if err != nil {
		logger.SysError(fmt.Sprintf("查询镜像卡密失败: %v", err))
		return
	}

	if len(cards) == 0 {
		logger.SysLog("没有需要获取 Token 的镜像卡密")
		return
	}

	logger.SysLog(fmt.Sprintf("找到 %d 个需要获取 Token 的镜像卡密", len(cards)))

	// 获取 ChatShare 节点 URL 列表
	nodeURLs := getNodeURLs()
	if len(nodeURLs) == 0 {
		logger.SysError("未配置 ChatShare 节点 URL")
		return
	}

	successCount := 0
	failCount := 0

	// 遍历卡密，逐个获取 Token
	for _, card := range cards {
		success := false

		// 尝试所有节点，直到成功
		for _, nodeURL := range nodeURLs {
			logger.SysLog(fmt.Sprintf("正在为卡密 [ID:%d, Username:%s] 从节点 %s 获取 Token...",
				card.ID, card.Username, nodeURL))

			// 调用 ChatShare 登录接口
			resp, err := crypto.CallChatShareLogin(card.Username, card.Password, nodeURL)
			if err != nil {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 登录失败: %v",
					card.ID, card.Username, nodeURL, err))
				continue // 尝试下一个节点
			}

			// 检查登录是否成功
			if !resp.IsSuccess {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 登录失败: %s",
					card.ID, card.Username, nodeURL, resp.Msg))
				continue // 尝试下一个节点
			}

			// 解析 Token（RespData 可能是 JSON 字符串）
			token := extractToken(resp.RespData)
			if token == "" {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 获取 Token 失败: 响应数据为空",
					card.ID, card.Username, nodeURL))
				continue // 尝试下一个节点
			}

			// 更新数据库
			if err := model.UpdateMirrorCardToken(card.ID, token, nodeURL); err != nil {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 更新 Token 到数据库失败: %v",
					card.ID, card.Username, err))
				continue // 尝试下一个节点
			}

			logger.SysLog(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 获取 Token 成功",
				card.ID, card.Username, nodeURL))
			successCount++
			success = true
			break // 成功后跳出节点循环
		}

		if !success {
			failCount++
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从所有节点获取 Token 均失败",
				card.ID, card.Username))
		}

		// 为了避免请求过快，每个卡密之间间隔 1 秒
		time.Sleep(1 * time.Second)
	}

	logger.SysLog(fmt.Sprintf("镜像卡密 Token 获取完成，成功: %d, 失败: %d", successCount, failCount))
}

// FetchTokenForCard 为指定的卡密获取 Token（手动触发）
func (s *MirrorCardTokenScheduler) FetchTokenForCard(cardID int) error {
	// 使用互斥锁防止同时执行
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog(fmt.Sprintf("手动触发获取卡密 [ID:%d] 的 Token...", cardID))

	// 查询指定的卡密
	card, err := model.GetMirrorCardById(cardID)
	if err != nil {
		return fmt.Errorf("查询镜像卡密失败: %v", err)
	}

	// 获取 ChatShare 节点 URL 列表
	nodeURLs := getNodeURLs()
	if len(nodeURLs) == 0 {
		return fmt.Errorf("未配置 ChatShare 节点 URL")
	}

	// 尝试所有节点，直到成功
	for _, nodeURL := range nodeURLs {
		logger.SysLog(fmt.Sprintf("正在为卡密 [ID:%d, Username:%s] 从节点 %s 获取 Token...",
			card.ID, card.Username, nodeURL))

		// 调用 ChatShare 登录接口
		resp, err := crypto.CallChatShareLogin(card.Username, card.Password, nodeURL)
		if err != nil {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 登录失败: %v",
				card.ID, card.Username, nodeURL, err))
			continue // 尝试下一个节点
		}

		// 检查登录是否成功
		if !resp.IsSuccess {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 登录失败: %s",
				card.ID, card.Username, nodeURL, resp.Msg))
			continue // 尝试下一个节点
		}

		// 解析 Token（RespData 可能是 JSON 字符串）
		token := extractToken(resp.RespData)
		if token == "" {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 获取 Token 失败: 响应数据为空",
				card.ID, card.Username, nodeURL))
			continue // 尝试下一个节点
		}

		// 更新数据库
		if err := model.UpdateMirrorCardToken(card.ID, token, nodeURL); err != nil {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 更新 Token 到数据库失败: %v",
				card.ID, card.Username, err))
			continue // 尝试下一个节点
		}

		logger.SysLog(fmt.Sprintf("卡密 [ID:%d, Username:%s] 从节点 %s 获取 Token 成功",
			card.ID, card.Username, nodeURL))
		return nil // 成功
	}

	return fmt.Errorf("从所有节点获取 Token 均失败")
}

// TriggerFetchForCard 全局方法：触发为指定卡密获取 Token
func TriggerFetchForCard(cardID int) error {
	if globalScheduler == nil {
		return fmt.Errorf("scheduler 未初始化")
	}
	// 在后台异步执行，不阻塞当前请求
	go func() {
		if err := globalScheduler.FetchTokenForCard(cardID); err != nil {
			logger.SysError(fmt.Sprintf("手动触发获取卡密 [ID:%d] Token 失败: %v", cardID, err))
		}
	}()
	return nil
}

// extractToken 从响应数据中提取 Token
func extractToken(respData string) string {
	if respData == "" {
		return ""
	}

	// 尝试解析为 JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(respData), &data); err == nil {
		// 如果是 JSON，尝试提取 token 字段
		if token, ok := data["token"].(string); ok {
			return token
		}
	}

	// 如果不是 JSON 或没有 token 字段，直接返回原始数据
	return strings.TrimSpace(respData)
}

// getNodeURLs 获取 ChatShare 节点 URL 列表
// 可以从配置文件或数据库读取，这里先硬编码几个常用节点
func getNodeURLs() []string {
	// TODO: 可以从环境变量或数据库读取节点列表
	return []string{
		"https://node1.chatshare.biz",
		"https://node2.chatshare.biz",
		"https://node3.chatshare.biz",
	}
}
