package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
)

// MirrorCardSyncScheduler 镜像卡密用户信息同步定时器
type MirrorCardSyncScheduler struct {
	interval time.Duration // 轮询间隔
	running  bool          // 是否正在运行
	stopChan chan bool     // 停止信号
	mu       sync.Mutex    // 互斥锁，防止同时执行
}

// UserDetailResponse 用户详情响应结构
type UserDetailResponse struct {
	IsSuccess bool               `json:"isSuccess"`
	Msg       string             `json:"msg"`
	RespData  UserDetailRespData `json:"respData"`
}

// UserDetailRespData 用户详情数据
type UserDetailRespData struct {
	UserName        string `json:"user_name"`
	UID             int    `json:"uid"`
	GPTFreqLimit    int    `json:"gpt_freq_limit"`
	MJFreqLimit     int    `json:"mj_freq_limit"`
	ClaudeFreqLimit int    `json:"claude_freq_limit"`
	GeminiFreqLimit int    `json:"gemini_freq_limit"`
	APIFreqLimit    int    `json:"api_freq_limit"`
	AvailCount      int    `json:"avail_count"`
	ExpireTime      string `json:"expire_time"`
}

var globalSyncScheduler *MirrorCardSyncScheduler

// NewMirrorCardSyncScheduler 创建镜像卡密用户信息同步定时器
func NewMirrorCardSyncScheduler(intervalMinutes int) *MirrorCardSyncScheduler {
	scheduler := &MirrorCardSyncScheduler{
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan bool),
	}
	globalSyncScheduler = scheduler
	return scheduler
}

// Start 启动定时任务
func (s *MirrorCardSyncScheduler) Start() {
	if s.running {
		logger.SysLog("MirrorCardSyncScheduler is already running")
		return
	}

	s.running = true
	logger.SysLog("MirrorCardSyncScheduler started")

	// 立即执行一次
	go s.syncUserInfo()

	// 启动定时器
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.syncUserInfo()
			case <-s.stopChan:
				ticker.Stop()
				logger.SysLog("MirrorCardSyncScheduler stopped")
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *MirrorCardSyncScheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false
	s.stopChan <- true
}

// syncUserInfo 同步用户信息
func (s *MirrorCardSyncScheduler) syncUserInfo() {
	// 使用互斥锁防止同时执行
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog("开始同步镜像卡密用户信息...")

	// 查询已启用且 token 不为空且未过期的卡密（每次处理 20 条）
	cards, err := model.GetMirrorCardsWithValidToken(20)
	if err != nil {
		logger.SysError(fmt.Sprintf("查询镜像卡密失败: %v", err))
		return
	}

	if len(cards) == 0 {
		logger.SysLog("没有需要同步用户信息的镜像卡密")
		return
	}

	logger.SysLog(fmt.Sprintf("找到 %d 个需要同步用户信息的镜像卡密", len(cards)))

	successCount := 0
	failCount := 0

	// 遍历卡密，逐个同步用户信息
	for _, card := range cards {
		logger.SysLog(fmt.Sprintf("正在同步卡密 [ID:%d, Username:%s, Token:%s] 的用户信息...",
			card.ID, card.Username, card.Token))

		// 调用外部接口获取用户信息
		userDetail, err := s.fetchUserDetail(card.NodeURL, card.Token)
		if err != nil {
			// 检查是否是登录态失效（401）
			if err.Error() == "登录态失效(401)" {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 登录态失效，清空 Token",
					card.ID, card.Username))
				// 清空 token
				if clearErr := model.ClearMirrorCardToken(card.ID); clearErr != nil {
					logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 清空 Token 失败: %v",
						card.ID, card.Username, clearErr))
				}
			} else {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 获取用户信息失败: %v",
					card.ID, card.Username, err))
			}
			failCount++
			// 继续处理下一个卡密
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// 解析过期时间
		var expireTime *time.Time
		if userDetail.ExpireTime != "" {
			t, err := time.Parse(time.RFC3339, userDetail.ExpireTime)
			if err != nil {
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 解析过期时间失败: %v",
					card.ID, card.Username, err))
				// 解析失败也继续，不影响其他字段的更新
			} else {
				expireTime = &t
			}
		}

		// 更新数据库
		err = model.UpdateMirrorCardUserInfo(
			card.ID,
			fmt.Sprintf("%d", userDetail.UID),
			userDetail.AvailCount,
			userDetail.GPTFreqLimit,
			userDetail.ClaudeFreqLimit,
			userDetail.GeminiFreqLimit,
			userDetail.MJFreqLimit,
			expireTime,
		)
		if err != nil {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 更新用户信息到数据库失败: %v",
				card.ID, card.Username, err))
			failCount++
		} else {
			logger.SysLog(fmt.Sprintf("卡密 [ID:%d, Username:%s] 同步用户信息成功 [UID:%d, AvailCount:%d, ExpireTime:%s]",
				card.ID, card.Username, userDetail.UID, userDetail.AvailCount, userDetail.ExpireTime))
			successCount++
		}

		// 为了避免请求过快，每个卡密之间间隔 500ms
		time.Sleep(500 * time.Millisecond)
	}

	logger.SysLog(fmt.Sprintf("镜像卡密用户信息同步完成，成功: %d, 失败: %d", successCount, failCount))
}

// fetchUserDetail 调用外部接口获取用户详情
func (s *MirrorCardSyncScheduler) fetchUserDetail(nodeURL, token string) (*UserDetailRespData, error) {
	if nodeURL == "" {
		return nil, fmt.Errorf("节点 URL 为空")
	}
	if token == "" {
		return nil, fmt.Errorf("token 为空")
	}

	// 构建请求 URL
	url := fmt.Sprintf("%s/share-login/v1/user/home/detail", nodeURL)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置 Cookie
	req.Header.Set("Cookie", fmt.Sprintf("token=%s", token))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码，401 表示登录态失效
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("登录态失效(401)")
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var response UserDetailResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	// 检查请求是否成功
	if !response.IsSuccess {
		return nil, fmt.Errorf("接口返回失败: %s", response.Msg)
	}

	return &response.RespData, nil
}

// SyncUserInfoForCard 为指定的卡密同步用户信息（手动触发）
func (s *MirrorCardSyncScheduler) SyncUserInfoForCard(cardID int) error {
	// 使用互斥锁防止同时执行
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog(fmt.Sprintf("手动触发同步卡密 [ID:%d] 的用户信息...", cardID))

	// 查询指定的卡密
	card, err := model.GetMirrorCardById(cardID)
	if err != nil {
		return fmt.Errorf("查询镜像卡密失败: %v", err)
	}

	// 检查 token 是否为空
	if card.Token == "" {
		return fmt.Errorf("卡密 Token 为空，无法同步用户信息")
	}

	// 调用外部接口获取用户信息
	userDetail, err := s.fetchUserDetail(card.NodeURL, card.Token)
	if err != nil {
		// 检查是否是登录态失效（401）
		if err.Error() == "登录态失效(401)" {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 登录态失效，清空 Token",
				card.ID, card.Username))
			// 清空 token
			if clearErr := model.ClearMirrorCardToken(card.ID); clearErr != nil {
				return fmt.Errorf("清空 Token 失败: %v", clearErr)
			}
			return fmt.Errorf("登录态失效，已清空 Token")
		}
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 解析过期时间
	var expireTime *time.Time
	if userDetail.ExpireTime != "" {
		t, err := time.Parse(time.RFC3339, userDetail.ExpireTime)
		if err != nil {
			logger.SysError(fmt.Sprintf("卡密 [ID:%d, Username:%s] 解析过期时间失败: %v",
				card.ID, card.Username, err))
		} else {
			expireTime = &t
		}
	}

	// 更新数据库
	err = model.UpdateMirrorCardUserInfo(
		card.ID,
		fmt.Sprintf("%d", userDetail.UID),
		userDetail.AvailCount,
		userDetail.GPTFreqLimit,
		userDetail.ClaudeFreqLimit,
		userDetail.GeminiFreqLimit,
		userDetail.MJFreqLimit,
		expireTime,
	)
	if err != nil {
		return fmt.Errorf("更新用户信息到数据库失败: %v", err)
	}

	logger.SysLog(fmt.Sprintf("卡密 [ID:%d, Username:%s] 同步用户信息成功", card.ID, card.Username))
	return nil
}

// TriggerSyncForCard 全局方法：触发为指定卡密同步用户信息
func TriggerSyncForCard(cardID int) error {
	if globalSyncScheduler == nil {
		return fmt.Errorf("sync scheduler 未初始化")
	}
	// 在后台异步执行，不阻塞当前请求
	go func() {
		if err := globalSyncScheduler.SyncUserInfoForCard(cardID); err != nil {
			logger.SysError(fmt.Sprintf("手动触发同步卡密 [ID:%d] 用户信息失败: %v", cardID, err))
		}
	}()
	return nil
}
