package outlook

import (
	"fmt"
	"io"
	"strings"
	"time"

	imap "github.com/emersion/go-imap"
	imapclient "github.com/emersion/go-imap/client"
)

const outlookIMAPPort = 993

// MailItem 单封邮件摘要
type MailItem struct {
	ID         string `json:"id"` // Graph message id；IMAP 时为空
	SeqNum     uint32 `json:"seq_num"`
	Folder     string `json:"folder"`
	Subject    string `json:"subject"`
	From       string `json:"from"`
	ReceivedAt string `json:"received_at"`
	Preview    string `json:"preview"`
	Body       string `json:"body"`
	HtmlBody   string `json:"html_body"`
	Code       string `json:"code"`
	IsRead     bool   `json:"is_read"`
}

var junkFolderCandidates = []string{
	"Junk Email",
	"Spam",
	"Bulk Mail",
	"Junk",
	"垃圾邮件",
}

var outlookIMAPHosts = []string{
	"outlook.office365.com",
	"outlook.live.com",
}

type xoauth2Client struct {
	payload []byte
}

func (x *xoauth2Client) Start() (string, []byte, error) {
	return "XOAUTH2", x.payload, nil
}

func (x *xoauth2Client) Next(_ []byte) ([]byte, error) {
	return []byte{}, nil
}

func buildXOAuth2Payload(email, token string) []byte {
	return []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", email, token))
}

func dialOutlookIMAP(host string) (*imapclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, outlookIMAPPort)
	c, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("IMAP 连接 %s 失败: %w", host, err)
	}
	return c, nil
}

// authenticateIMAP 依次用候选 access_token 尝试 XOAUTH2。
// 注意：每次失败后必须重连，同一 TCP 会话上连续 AUTHENTICATE 会残留半认证状态。
func authenticateIMAP(host string, acc *Account, accessTokens []string) (*imapclient.Client, error) {
	var lastErr error
	for _, token := range accessTokens {
		if token == "" {
			continue
		}
		c, err := dialOutlookIMAP(host)
		if err != nil {
			lastErr = err
			continue
		}
		payload := buildXOAuth2Payload(acc.Email, token)
		if authErr := c.Authenticate(&xoauth2Client{payload: payload}); authErr != nil {
			lastErr = authErr
			c.Logout() //nolint:errcheck
			continue
		}
		// 探测 INBOX：部分账号 AUTHENTICATE 看似成功，但 SELECT 报
		// "User is authenticated but not connected"（IMAP 协议未开通）
		if _, selErr := c.Select("INBOX", true); selErr != nil {
			lastErr = selErr
			c.Logout() //nolint:errcheck
			continue
		}
		return c, nil
	}

	if acc.Password != "" {
		c, err := dialOutlookIMAP(host)
		if err != nil {
			return nil, err
		}
		if loginErr := c.Login(acc.Email, acc.Password); loginErr != nil {
			c.Logout() //nolint:errcheck
			if lastErr != nil {
				return nil, fmt.Errorf("XOAUTH2 失败(%v)且密码登录失败: %w", lastErr, loginErr)
			}
			return nil, fmt.Errorf("IMAP 登录失败: %w", loginErr)
		}
		return c, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("XOAUTH2 认证失败: %w", lastErr)
	}
	return nil, fmt.Errorf("XOAUTH2 认证失败且密码为空，无法登录")
}

// FetchViaIMAP 通过 IMAP 获取收件箱与垃圾箱（多主机 + XOAUTH2 优先，失败回退密码）
func FetchViaIMAP(acc *Account, accessTokens []string) (inbox []MailItem, junk []MailItem, err error) {
	var lastErr error
	for _, host := range outlookIMAPHosts {
		c, authErr := authenticateIMAP(host, acc, accessTokens)
		if authErr != nil {
			lastErr = authErr
			continue
		}

		inbox, _ = fetchIMAPFolder(c, "INBOX")
		junkFolder := resolveJunkFolder(c)
		if junkFolder != "" {
			junk, _ = fetchIMAPFolder(c, junkFolder)
		}
		if junk == nil {
			junk = []MailItem{}
		}
		c.Logout() //nolint:errcheck
		return inbox, junk, nil
	}
	if lastErr != nil {
		return nil, nil, lastErr
	}
	return nil, nil, fmt.Errorf("IMAP 登录失败")
}

func resolveJunkFolder(c *imapclient.Client) string {
	for _, name := range junkFolderCandidates {
		if _, err := c.Select(name, true); err == nil {
			return name
		}
	}
	mailboxes := make(chan *imap.MailboxInfo, 30)
	done := make(chan error, 1)
	go func() { done <- c.List("", "*", mailboxes) }()
	for m := range mailboxes {
		if m == nil {
			continue
		}
		low := strings.ToLower(m.Name)
		if strings.Contains(low, "junk") || strings.Contains(low, "spam") ||
			strings.Contains(low, "bulk") || strings.Contains(low, "垃圾") {
			<-done
			return m.Name
		}
	}
	<-done
	return ""
}

// envInfo 临时存放 envelope 阶段解析结果
type envInfo struct {
	seqNum     uint32
	subject    string
	fromStr    string
	receivedAt string
}

func fetchIMAPFolder(c *imapclient.Client, folder string) ([]MailItem, error) {
	mbox, err := c.Select(folder, true)
	if err != nil {
		return nil, fmt.Errorf("选择文件夹 %s 失败: %w", folder, err)
	}
	if mbox.Messages == 0 {
		return []MailItem{}, nil
	}

	from := uint32(1)
	if mbox.Messages > 20 {
		from = mbox.Messages - 19
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, mbox.Messages)

	// ── 第一步：只取 envelope，速度极快 ──────────────────────────────
	msgs1 := make(chan *imap.Message, 25)
	errCh1 := make(chan error, 1)
	go func() { errCh1 <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchInternalDate}, msgs1) }()

	var envs []envInfo
	for msg := range msgs1 {
		if msg == nil || msg.Envelope == nil {
			continue
		}
		fromStr := ""
		if len(msg.Envelope.From) > 0 {
			addr := msg.Envelope.From[0]
			if addr.PersonalName != "" {
				fromStr = addr.PersonalName + " <" + addr.Address() + ">"
			} else {
				fromStr = addr.Address()
			}
		}
		var receivedAt string
		if !msg.InternalDate.IsZero() {
			receivedAt = msg.InternalDate.Local().Format("2006-01-02 15:04:05")
		} else if msg.Envelope.Date != (time.Time{}) {
			receivedAt = msg.Envelope.Date.Local().Format("2006-01-02 15:04:05")
		}
		envs = append(envs, envInfo{
			seqNum:     msg.SeqNum,
			subject:    msg.Envelope.Subject,
			fromStr:    fromStr,
			receivedAt: receivedAt,
		})
	}
	if err := <-errCh1; err != nil {
		return nil, fmt.Errorf("IMAP Fetch envelope 失败: %w", err)
	}

	// ── 第二步：仅对标题含关键词的邮件拉取正文 ───────────────────────
	type bodyResult struct {
		plain    string
		htmlBody string
	}
	bodyMap := make(map[uint32]bodyResult)

	codeSeqset := new(imap.SeqSet)
	for _, e := range envs {
		if HasCodeKeyword(e.subject) {
			codeSeqset.AddNum(e.seqNum)
		}
	}

	if len(codeSeqset.Set) > 0 {
		section := &imap.BodySectionName{}
		msgs2 := make(chan *imap.Message, 25)
		errCh2 := make(chan error, 1)
		go func() { errCh2 <- c.Fetch(codeSeqset, []imap.FetchItem{section.FetchItem()}, msgs2) }()

		for msg := range msgs2 {
			if msg == nil {
				continue
			}
			var raw []byte
			if body := msg.GetBody(section); body != nil {
				raw, _ = io.ReadAll(body)
			}
			if len(raw) == 0 {
				for _, lit := range msg.Body {
					if lit == nil {
						continue
					}
					if b, _ := io.ReadAll(lit); len(b) > 0 {
						raw = b
						break
					}
				}
			}
			plain, htmlBody := ParseBodies(raw)
			bodyMap[msg.SeqNum] = bodyResult{plain: plain, htmlBody: htmlBody}
		}
		<-errCh2
	}

	// ── 合并结果 ─────────────────────────────────────────────────────
	var items []MailItem
	for _, e := range envs {
		br := bodyMap[e.seqNum]
		bodyText := collapseSpaces(br.plain)
		if bodyText == "" {
			bodyText = collapseSpaces(stripHTML(br.htmlBody))
		}
		preview := truncateRunes(bodyText, 200)
		code := ExtractCode(e.subject, br.plain+"\n"+stripHTML(br.htmlBody))

		items = append(items, MailItem{
			SeqNum:     e.seqNum,
			Folder:     folder,
			Subject:    e.subject,
			From:       e.fromStr,
			ReceivedAt: e.receivedAt,
			Preview:    preview,
			Body:       bodyText,
			HtmlBody:   br.htmlBody,
			Code:       code,
			IsRead:     false,
		})
	}

	// 倒序（最新在前）
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}

	return items, nil
}

// MailDetail 单封邮件完整正文
type MailDetail struct {
	Body     string `json:"body"`
	HtmlBody string `json:"html_body"`
	Code     string `json:"code"`
}

// FetchBodyBySeq 按序列号拉取单封邮件正文（供详情按需加载）
func FetchBodyBySeq(acc *Account, accessTokens []string, folder string, seqNum uint32) (*MailDetail, error) {
	var lastErr error
	for _, host := range outlookIMAPHosts {
		c, authErr := authenticateIMAP(host, acc, accessTokens)
		if authErr != nil {
			lastErr = authErr
			continue
		}

		if _, err := c.Select(folder, true); err != nil {
			lastErr = fmt.Errorf("选择文件夹 %s 失败: %w", folder, err)
			c.Logout() //nolint:errcheck
			continue
		}

		seqset := new(imap.SeqSet)
		seqset.AddNum(seqNum)

		section := &imap.BodySectionName{}
		msgs := make(chan *imap.Message, 2)
		errCh := make(chan error, 1)
		go func() { errCh <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, msgs) }()

		var detail *MailDetail
		for msg := range msgs {
			if msg == nil {
				continue
			}
			var raw []byte
			if body := msg.GetBody(section); body != nil {
				raw, _ = io.ReadAll(body)
			}
			if len(raw) == 0 {
				for _, lit := range msg.Body {
					if lit == nil {
						continue
					}
					if b, _ := io.ReadAll(lit); len(b) > 0 {
						raw = b
						break
					}
				}
			}
			plain, htmlBody := ParseBodies(raw)
			bodyText := collapseSpaces(plain)
			if bodyText == "" {
				bodyText = collapseSpaces(stripHTML(htmlBody))
			}
			subject := ""
			if msg.Envelope != nil {
				subject = msg.Envelope.Subject
			}
			code := ExtractCode(subject, plain+"\n"+stripHTML(htmlBody))
			detail = &MailDetail{
				Body:     bodyText,
				HtmlBody: htmlBody,
				Code:     code,
			}
		}
		<-errCh
		c.Logout() //nolint:errcheck

		if detail == nil {
			lastErr = fmt.Errorf("未找到序列号 %d 的邮件", seqNum)
			continue
		}
		return detail, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("未找到序列号 %d 的邮件", seqNum)
}
