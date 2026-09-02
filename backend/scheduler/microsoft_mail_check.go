package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/outlook"
)

const (
	msMailCheckIntervalDays = 15
	msMailCheckPageSize     = 500 // 分页大小：未检查全量 / 到期复查均按页拉取
	msMailCheckRetryHours   = 6
)

// MicrosoftMailCheckScheduler 微软邮箱 refresh_token 保活定时器
type MicrosoftMailCheckScheduler struct {
	interval time.Duration
	running  bool
	stopChan chan bool
	mu       sync.Mutex
}

var globalMsMailCheckScheduler *MicrosoftMailCheckScheduler

// NewMicrosoftMailCheckScheduler 创建调度器，intervalMinutes 为扫表轮询间隔
func NewMicrosoftMailCheckScheduler(intervalMinutes int) *MicrosoftMailCheckScheduler {
	s := &MicrosoftMailCheckScheduler{
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan bool),
	}
	globalMsMailCheckScheduler = s
	return s
}

// Start 启动定时任务
func (s *MicrosoftMailCheckScheduler) Start() {
	if s.running {
		logger.SysLog("MicrosoftMailCheckScheduler is already running")
		return
	}
	s.running = true
	logger.SysLog("MicrosoftMailCheckScheduler started")

	go s.checkMails()

	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.checkMails()
			case <-s.stopChan:
				ticker.Stop()
				logger.SysLog("MicrosoftMailCheckScheduler stopped")
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *MicrosoftMailCheckScheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false
	s.stopChan <- true
}

func (s *MicrosoftMailCheckScheduler) checkMails() {
	if constants.GetDebugEnabled() {
		logger.SysLog("调试模式(DEBUG=true)，跳过微软邮箱 refresh_token 检查")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog("开始检查微软邮箱 refresh_token...")

	uncheckedSuccess, uncheckedFail := s.checkUncheckedMails()
	dueSuccess, dueFail := s.checkDueMails()

	logger.SysLog(fmt.Sprintf(
		"微软邮箱检查完成，未检查全量 成功:%d 失败:%d；到期复查 成功:%d 失败:%d",
		uncheckedSuccess, uncheckedFail, dueSuccess, dueFail,
	))
}

// checkUncheckedMails 对 is_check=-1 的记录做全量检查（分页拉取直到扫完）
func (s *MicrosoftMailCheckScheduler) checkUncheckedMails() (successCount, failCount int) {
	page := 0
	for {
		mails, err := model.GetUncheckedMicrosoftMailsForCheck(msMailCheckPageSize)
		if err != nil {
			logger.SysError("查询未检查微软邮箱失败: " + err.Error())
			return
		}
		if len(mails) == 0 {
			if page == 0 {
				logger.SysLog("没有未检查的微软邮箱")
			}
			return
		}
		page++
		logger.SysLog(fmt.Sprintf("未检查全量第 %d 页，本页 %d 条", page, len(mails)))

		ok, fail := processMicrosoftMailChecks(mails)
		successCount += ok
		failCount += fail

		// 不足一页说明已扫完；成功/失败都会改写 is_check，下一页自然不会再命中
		if len(mails) < msMailCheckPageSize {
			return
		}
	}
}

// checkDueMails 对已检查过且 next_check_time 到期的记录做复查（分页直到本轮到期队列为空）
func (s *MicrosoftMailCheckScheduler) checkDueMails() (successCount, failCount int) {
	page := 0
	for {
		mails, err := model.GetDueMicrosoftMailsForCheck(msMailCheckPageSize)
		if err != nil {
			logger.SysError("查询到期复查微软邮箱失败: " + err.Error())
			return
		}
		if len(mails) == 0 {
			if page == 0 {
				logger.SysLog("没有到期需复查的微软邮箱")
			}
			return
		}
		page++
		logger.SysLog(fmt.Sprintf("到期复查第 %d 页，本页 %d 条", page, len(mails)))

		ok, fail := processMicrosoftMailChecks(mails)
		successCount += ok
		failCount += fail

		if len(mails) < msMailCheckPageSize {
			return
		}
	}
}

func processMicrosoftMailChecks(mails []*model.MicrosoftMail) (successCount, failCount int) {
	for _, mail := range mails {
		if err := CheckSingleMicrosoftMail(mail); err != nil {
			logger.SysError(fmt.Sprintf("微软邮箱检查失败 [ID:%d, Account:%s]: %v", mail.Id, mail.Account, err))
			failCount++
		} else {
			successCount++
		}
		time.Sleep(500 * time.Millisecond)
	}
	return
}

// CheckSingleMicrosoftMail 刷新单条邮箱的 refresh_token 并回写
func CheckSingleMicrosoftMail(mail *model.MicrosoftMail) error {
	if mail == nil {
		return fmt.Errorf("mail 为空")
	}
	if mail.Token == "" || mail.ClientId == "" {
		_ = model.ApplyMicrosoftMailCheckFailure(mail.Id, true, 0)
		return fmt.Errorf("缺少 token 或 client_id")
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	result, err := outlook.RefreshAccessTokenWithRotation(httpClient, mail.ClientId, mail.Token)
	if err != nil {
		permanent := outlook.IsInvalidGrant(err)
		_ = model.ApplyMicrosoftMailCheckFailure(mail.Id, permanent, msMailCheckRetryHours)
		return err
	}

	newRT := result.RefreshToken
	if newRT == "" {
		newRT = mail.Token
	}
	if err := model.ApplyMicrosoftMailCheckSuccess(mail.Id, newRT, msMailCheckIntervalDays); err != nil {
		return fmt.Errorf("回写检查结果失败: %w", err)
	}
	return nil
}
