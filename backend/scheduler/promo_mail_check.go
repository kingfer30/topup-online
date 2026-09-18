package scheduler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/outlook"
	"github.com/kingfer30/topup-online/utils/webmail"
)

const (
	promoMailCardTable        = "cards_cursor"
	promoMailPageSize         = 50
	promoMailSubjectMatch     = "come back to cursor for 50% off"
	promoMailBodyMatch        = "get 50% off"
	promoMailWrongPasswdText  = "邮箱或者密码有误"
	promoMailMissingInfoText  = "邮箱获取失败，请检查是否有信息没填对"
	promoMailAccountClosedTxt = "your cursor account has been closed"
)

// promoMailAccountInvalidTexts 快捷取件服务返回以下文案时，视为账号信息本身有问题（而非网络等临时故障）
var promoMailAccountInvalidTexts = []string{
	promoMailWrongPasswdText,
	promoMailMissingInfoText,
}

// isAccountInvalidError 判断快捷取件错误是否明确指向账号信息有误
func isAccountInvalidError(errMsg string) bool {
	for _, kw := range promoMailAccountInvalidTexts {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}
	return false
}

// PromoMailCheckScheduler 检测 cards_cursor 表中已配置快捷邮件取件的账号，
// 收件箱/垃圾箱是否收到 Cursor "50% off" 召回优惠邮件，命中则标记
type PromoMailCheckScheduler struct {
	interval time.Duration
	running  bool
	stopChan chan bool
	mu       sync.Mutex
}

var globalPromoMailCheckScheduler *PromoMailCheckScheduler

// NewPromoMailCheckScheduler 创建调度器，intervalMinutes 为轮询间隔
func NewPromoMailCheckScheduler(intervalMinutes int) *PromoMailCheckScheduler {
	s := &PromoMailCheckScheduler{
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan bool),
	}
	globalPromoMailCheckScheduler = s
	return s
}

// Start 启动定时任务
func (s *PromoMailCheckScheduler) Start() {
	if s.running {
		logger.SysLog("PromoMailCheckScheduler is already running")
		return
	}
	s.running = true
	logger.SysLog("PromoMailCheckScheduler started")

	// 启动时立即执行一次
	go s.checkAll()

	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.checkAll()
			case <-s.stopChan:
				ticker.Stop()
				logger.SysLog("PromoMailCheckScheduler stopped")
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *PromoMailCheckScheduler) Stop() {
	if !s.running {
		return
	}
	s.running = false
	s.stopChan <- true
}

// checkAll 按 id 游标分页扫描 cards_cursor 表中未标记过的快捷取件账号
func (s *PromoMailCheckScheduler) checkAll() {
	if constants.GetDebugEnabled() {
		logger.SysLog("调试模式(DEBUG=true)，跳过 Cursor 50% off 召回邮件检测")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	logger.SysLog(fmt.Sprintf("开始检测 %s 表 Cursor 50%% off 召回邮件...", promoMailCardTable))

	total, hit, fail, miss, skipped, closedCnt := 0, 0, 0, 0, 0, 0
	afterID := 0
	for {
		cards, err := model.GetQuickMailUncheckedCards(promoMailCardTable, afterID, promoMailPageSize)
		if err != nil {
			logger.SysError("查询待检测卡密失败: " + err.Error())
			return
		}
		if len(cards) == 0 {
			break
		}

		for _, card := range cards {
			total++
			now := time.Now().Unix()
			matched, info, closed, skip, err := CheckSinglePromoMail(card)
			switch {
			case skip:
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Account:%s] 标记为不再检测半价召回邮件: %v", card.Id, card.Account, err))
				updates := map[string]interface{}{
					"promo_50off_check_time": now,
					"promo_50off_skip":       1,
				}
				if err != nil {
					updates["promo_50off_last_error"] = truncateTo300(err.Error())
				}
				if updErr := model.UpdateCardCheckResult(promoMailCardTable, card.Id, updates); updErr != nil {
					logger.SysError(fmt.Sprintf("卡密 [ID:%d] 标记停止检测失败: %v", card.Id, updErr))
				}
				skipped++
			case err != nil:
				logger.SysError(fmt.Sprintf("卡密 [ID:%d, Account:%s] 检测优惠邮件失败: %v", card.Id, card.Account, err))
				if updErr := model.UpdateCardCheckResult(promoMailCardTable, card.Id, map[string]interface{}{
					"promo_50off_check_time": now,
					"promo_50off_last_error": truncateTo300(err.Error()),
				}); updErr != nil {
					logger.SysError(fmt.Sprintf("卡密 [ID:%d] 记录检测失败信息失败: %v", card.Id, updErr))
				}
				fail++
			case matched:
				updates := map[string]interface{}{
					"is_promo_50off":         1,
					"promo_50off_info":       info,
					"promo_50off_check_time": now,
					"promo_50off_last_error": "",
				}
				if closed {
					updates["promo_50off_skip"] = 1
					updates["status"] = model.CardStatusDisabled
				}
				if err := model.UpdateCardCheckResult(promoMailCardTable, card.Id, updates); err != nil {
					logger.SysError(fmt.Sprintf("卡密 [ID:%d] 标记优惠邮件失败: %v", card.Id, err))
				} else {
					logger.SysLog(fmt.Sprintf("卡密 [ID:%d, Account:%s] 命中 Cursor 50%% off 召回邮件: %s", card.Id, card.Account, info))
					hit++
					if closed {
						logger.SysLog(fmt.Sprintf("卡密 [ID:%d, Account:%s] 同时发现账号关闭邮件，已禁用并停止检测", card.Id, card.Account))
						closedCnt++
					}
				}
			default:
				if updErr := model.UpdateCardCheckResult(promoMailCardTable, card.Id, map[string]interface{}{
					"promo_50off_check_time": now,
					"promo_50off_last_error": "",
				}); updErr != nil {
					logger.SysError(fmt.Sprintf("卡密 [ID:%d] 记录检测时间失败: %v", card.Id, updErr))
				}
				miss++
			}
			// 每个邮箱间隔 1 秒，避免触发取件服务频率限制
			time.Sleep(1 * time.Second)
		}

		afterID = cards[len(cards)-1].Id
		if len(cards) < promoMailPageSize {
			break
		}
	}

	logger.SysLog(fmt.Sprintf("Cursor 50%% off 召回邮件检测完成，共检测 %d 条，命中 %d 条，未命中 %d 条，失败 %d 条，停止检测 %d 条，关户禁用 %d 条", total, hit, miss, fail, skipped, closedCnt))
}

// detectQuickMailProvider 根据 code_link 判断快捷取件服务商
func detectQuickMailProvider(codeLink string) string {
	link := strings.ToLower(strings.TrimSpace(codeLink))
	switch {
	case strings.Contains(link, "lqqq.cc"):
		return "lqqq"
	case strings.Contains(link, "toolsvip.cc"):
		return "toolsvip"
	case strings.Contains(link, "lurentool.cn"):
		return "lurentool"
	default:
		return ""
	}
}

// matchPromoContent 标题包含指定文案 或 正文包含指定文案，任一命中即视为匹配
func matchPromoContent(subject, body string) bool {
	s := strings.ToLower(subject)
	b := strings.ToLower(body)
	return strings.Contains(s, promoMailSubjectMatch) || strings.Contains(b, promoMailBodyMatch)
}

// isAccountClosedMail 标题或正文包含 "Your Cursor account has been closed" 即视为账号已关闭通知
func isAccountClosedMail(subject, body string) bool {
	s := strings.ToLower(subject)
	b := strings.ToLower(body)
	return strings.Contains(s, promoMailAccountClosedTxt) || strings.Contains(b, promoMailAccountClosedTxt)
}

// promoScanItem 统一各服务商邮件条目结构，供 scanPromoItems 统一扫描
type promoScanItem struct {
	Subject string
	Body    string
	Mailbox string
	Date    string
}

// scanPromoItems 在给定邮件列表中查找 Cursor 50% off 召回邮件；
// 一旦命中，再在同一批邮件中查找是否存在 "Your Cursor account has been closed" 账号关闭通知，
// closed=true 表示账号已被关闭，调用方应据此停止后续检测并将账号标记为禁用。
func scanPromoItems(items []promoScanItem) (matched bool, info string, closed bool) {
	for _, item := range items {
		if matchPromoContent(item.Subject, item.Body) {
			matched = true
			info = formatPromoInfo(item.Mailbox, item.Date, item.Subject)
			break
		}
	}
	if !matched {
		return false, "", false
	}
	for _, item := range items {
		if isAccountClosedMail(item.Subject, item.Body) {
			closed = true
			break
		}
	}
	return matched, info, closed
}

// CheckSinglePromoMail 检测单张卡密的收件箱/垃圾箱是否存在 Cursor 50% off 召回邮件
// 返回 matched 及命中详情（位置|日期|标题）；closed 表示命中半价邮件的同时还发现账号已关闭通知
// （"Your Cursor account has been closed"），调用方应据此停止检测并将账号禁用；
// skip 表示该记录应永久停止检测，供外部（定时任务/手动触发）复用
func CheckSinglePromoMail(card *model.AccountCard) (matched bool, info string, closed bool, skip bool, err error) {
	if card == nil {
		return false, "", false, false, fmt.Errorf("card 为空")
	}
	email := strings.TrimSpace(card.Account)
	password := card.MailPassword
	if email == "" || password == "" {
		return false, "", false, false, fmt.Errorf("邮箱账号或邮箱密码为空")
	}

	switch detectQuickMailProvider(card.CodeLink) {
	case "lqqq":
		matched, info, closed, err = checkPromoMailLqqq(email, password)
	case "toolsvip":
		matched, info, closed, err = checkPromoMailToolsvip(email, password)
	case "lurentool":
		matched, info, closed, err = checkPromoMailLurentool(email, password)
	default:
		return false, "", false, false, fmt.Errorf("暂不支持的快捷取件服务: %s", card.CodeLink)
	}

	if err != nil {
		// 快捷取件失败（无论具体原因），只要该卡密在 MicrosoftMail 表中存在可用 refresh_token，就尝试原生取件兜底
		return checkPromoMailNative(card, err)
	}
	return matched, info, closed, false, err
}

// checkPromoMailNative 快捷取件失败时的原生取件兜底：
// 无论快捷取件失败的具体原因，只要该卡密在 MicrosoftMail 表中存在记录且 refresh_token/client_id 齐全，
// 就尝试走 Graph（失败则回退 IMAP）原生取件，取件结果（命中/未命中）以原生结果为准；
// 仅当快捷取件的错误明确指向账号信息有误（如"邮箱或者密码有误"/"邮箱获取失败，请检查是否有信息没填对"），
// 且原生邮件也不可用（无记录/缺 token/token 失效）时，才标记该卡密永久停止检测；
// 其余情况（如网络抖动等临时故障）下原生邮件不可用，仅按普通失败处理，不冷却、下次继续重试。
func checkPromoMailNative(card *model.AccountCard, quickMailErr error) (matched bool, info string, closed bool, skip bool, err error) {
	accountInvalid := isAccountInvalidError(quickMailErr.Error())

	mail, ferr := model.GetMicrosoftMailByCard(promoMailCardTable, card.Id)
	if ferr != nil {
		if accountInvalid {
			return false, "", false, true, fmt.Errorf("%v；且未找到关联的原生邮箱记录: %v", quickMailErr, ferr)
		}
		return false, "", false, false, quickMailErr
	}
	if strings.TrimSpace(mail.Token) == "" || strings.TrimSpace(mail.ClientId) == "" {
		if accountInvalid {
			return false, "", false, true, fmt.Errorf("%v；原生邮箱记录缺少 token 或 client_id", quickMailErr)
		}
		return false, "", false, false, quickMailErr
	}

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	acc := &outlook.Account{
		Email:        mail.Account,
		Password:     mail.Password,
		RefreshToken: mail.Token,
		ClientID:     mail.ClientId,
	}

	var inbox, junk []outlook.MailItem
	graphTokens, graphErr := outlook.RefreshGraphAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
	if graphErr == nil {
		inbox, junk, err = outlook.FetchViaGraph(httpClient, graphTokens)
	} else {
		err = graphErr
	}
	if err != nil {
		graphFailErr := err
		accessTokens, imapTokenErr := outlook.RefreshAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if imapTokenErr != nil {
			if accountInvalid {
				return false, "", false, true, fmt.Errorf("%v；原生 token 不可用: Graph=%v；IMAP OAuth=%v", quickMailErr, graphFailErr, imapTokenErr)
			}
			return false, "", false, false, fmt.Errorf("%v；原生取件亦失败: Graph=%v；IMAP OAuth=%v", quickMailErr, graphFailErr, imapTokenErr)
		}
		inbox, junk, err = outlook.FetchViaIMAP(acc, accessTokens)
		if err != nil {
			if accountInvalid {
				return false, "", false, true, fmt.Errorf("%v；原生取件不可用: Graph=%v；IMAP=%v", quickMailErr, graphFailErr, err)
			}
			return false, "", false, false, fmt.Errorf("%v；原生取件亦失败: Graph=%v；IMAP=%v", quickMailErr, graphFailErr, err)
		}
	}

	items := make([]promoScanItem, 0, len(inbox)+len(junk))
	for _, item := range inbox {
		items = append(items, promoScanItem{Subject: item.Subject, Body: item.Body + " " + item.HtmlBody, Mailbox: "收件箱(原生)", Date: item.ReceivedAt})
	}
	for _, item := range junk {
		items = append(items, promoScanItem{Subject: item.Subject, Body: item.Body + " " + item.HtmlBody, Mailbox: "垃圾箱(原生)", Date: item.ReceivedAt})
	}
	matched, info, closed = scanPromoItems(items)
	return matched, info, closed, false, nil
}

func checkPromoMailLqqq(email, password string) (matched bool, info string, closed bool, err error) {
	inbox, junk, err := webmail.FetchLqqqMails(email, password)
	if err != nil {
		return false, "", false, err
	}
	items := make([]promoScanItem, 0, len(inbox)+len(junk))
	for _, item := range inbox {
		items = append(items, promoScanItem{Subject: item.Subject, Mailbox: item.Mailbox, Date: item.Date})
	}
	for _, item := range junk {
		items = append(items, promoScanItem{Subject: item.Subject, Mailbox: item.Mailbox, Date: item.Date})
	}
	matched, info, closed = scanPromoItems(items)
	return matched, info, closed, nil
}

func checkPromoMailToolsvip(email, password string) (matched bool, info string, closed bool, err error) {
	inbox, junk, err := webmail.FetchToolsvipMails(email, password)
	if err != nil {
		return false, "", false, err
	}
	items := make([]promoScanItem, 0, len(inbox)+len(junk))
	for _, item := range inbox {
		items = append(items, promoScanItem{Subject: item.Subject, Body: item.Body + " " + item.HtmlBody, Mailbox: item.Mailbox, Date: item.Date})
	}
	for _, item := range junk {
		items = append(items, promoScanItem{Subject: item.Subject, Body: item.Body + " " + item.HtmlBody, Mailbox: item.Mailbox, Date: item.Date})
	}
	matched, info, closed = scanPromoItems(items)
	return matched, info, closed, nil
}

func checkPromoMailLurentool(email, password string) (matched bool, info string, closed bool, err error) {
	data, err := webmail.FetchLurentoolMails(email, password, 20)
	if err != nil {
		return false, "", false, err
	}
	items := make([]promoScanItem, 0, len(data.Mails))
	for _, item := range data.Mails {
		mailbox := "收件箱"
		if item.Source != "" {
			mailbox = item.Source
		}
		items = append(items, promoScanItem{
			Subject: item.Subject,
			Body:    item.Body + " " + item.Preview,
			Mailbox: mailbox,
			Date:    item.ReceivedAt,
		})
	}
	matched, info, closed = scanPromoItems(items)
	return matched, info, closed, nil
}

// formatPromoInfo 拼接命中信息并按 promo_50off_info 字段长度（varchar(300)）截断，避免写库报错
func formatPromoInfo(mailbox, date, subject string) string {
	return truncateTo300(fmt.Sprintf("%s | %s | %s", mailbox, date, subject))
}

// truncateTo300 按字段长度（varchar(300)）截断文本，避免写库报错
func truncateTo300(s string) string {
	runes := []rune(s)
	if len(runes) > 300 {
		return string(runes[:300])
	}
	return s
}
