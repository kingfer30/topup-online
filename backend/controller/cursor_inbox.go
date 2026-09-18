package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/outlook"
	"github.com/kingfer30/topup-online/utils/webmail"
)

// publicMailItem 公开取件页统一邮件条目
type publicMailItem struct {
	ID         string `json:"id"`
	Subject    string `json:"subject"`
	From       string `json:"from"`
	ReceivedAt string `json:"received_at"`
	Preview    string `json:"preview"`
	Mailbox    string `json:"mailbox"`
	Code       string `json:"code"`
	Body       string `json:"body"`
	HtmlBody   string `json:"html_body"`
	SeqNum     uint32 `json:"seq_num"`
	Folder     string `json:"folder"`
	ViewHref   string `json:"view_href,omitempty"`
}

type publicMailFetchData struct {
	Account string           `json:"account"`
	Email   string           `json:"email"`
	Source  string           `json:"source"`
	Inbox   []publicMailItem `json:"inbox"`
	Junk    []publicMailItem `json:"junk"`
	Message string           `json:"message"`
}

func emptyPublicMails() []publicMailItem {
	return []publicMailItem{}
}

func writePublicMailOK(c *gin.Context, data publicMailFetchData) {
	if data.Inbox == nil {
		data.Inbox = emptyPublicMails()
	}
	if data.Junk == nil {
		data.Junk = emptyPublicMails()
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    data,
	})
}

func detectPublicMailProvider(codeLink string) string {
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

func outlookItemsToPublic(items []outlook.MailItem, mailbox string) []publicMailItem {
	out := make([]publicMailItem, 0, len(items))
	for _, item := range items {
		out = append(out, publicMailItem{
			ID:         item.ID,
			Subject:    item.Subject,
			From:       item.From,
			ReceivedAt: item.ReceivedAt,
			Preview:    item.Preview,
			Mailbox:    mailbox,
			Code:       item.Code,
			Body:       item.Body,
			HtmlBody:   item.HtmlBody,
			SeqNum:     item.SeqNum,
			Folder:     item.Folder,
		})
	}
	return out
}

func ouleyxItemsToPublic(items []webmail.OuleyxMailItem) []publicMailItem {
	out := make([]publicMailItem, 0, len(items))
	for _, item := range items {
		out = append(out, publicMailItem{
			ID:         item.ID,
			Subject:    item.Subject,
			From:       item.From,
			ReceivedAt: item.ReceivedAt,
			Preview:    item.Preview,
			Mailbox:    item.Mailbox,
			Code:       item.Code,
			Body:       item.Body,
			HtmlBody:   item.HtmlBody,
		})
	}
	return out
}

func fetchNativeOutlookByCard(card *model.AccountCard) (inbox, junk []outlook.MailItem, email, source string, ok bool, err error) {
	mail, ferr := model.GetMicrosoftMailByCard(cursorCardTable, card.Id)
	if ferr != nil || mail == nil {
		return nil, nil, "", "", false, nil
	}
	if strings.TrimSpace(mail.Token) == "" || strings.TrimSpace(mail.ClientId) == "" {
		return nil, nil, "", "", false, nil
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

	source = "graph"
	var graphErr error
	graphTokens, tokenErr := outlook.RefreshGraphAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
	if tokenErr != nil {
		graphErr = tokenErr
		logger.SysLog("GetCursorEmailMails: " + acc.Email + " Graph oauth failed: " + tokenErr.Error())
	} else {
		inbox, junk, graphErr = outlook.FetchViaGraph(httpClient, graphTokens)
	}
	if graphErr != nil {
		logger.SysLog("GetCursorEmailMails: " + acc.Email + " Graph failed, fallback IMAP: " + graphErr.Error())
		accessTokens, imapTokenErr := outlook.RefreshAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if imapTokenErr != nil {
			return nil, nil, acc.Email, "", true, imapTokenErr
		}
		inbox, junk, err = outlook.FetchViaIMAP(acc, accessTokens)
		if err != nil {
			return nil, nil, acc.Email, "", true, err
		}
		source = "imap"
	}
	return inbox, junk, acc.Email, source, true, nil
}

func fetchQuickMailByCard(card *model.AccountCard) (data publicMailFetchData, ok bool, err error) {
	provider := detectPublicMailProvider(card.CodeLink)
	if provider == "" {
		return data, false, nil
	}
	email := strings.TrimSpace(card.Account)
	password := card.MailPassword
	if email == "" || strings.TrimSpace(password) == "" {
		return data, false, nil
	}

	data.Email = email
	data.Source = provider
	switch provider {
	case "lqqq":
		inbox, junk, ferr := webmail.FetchLqqqMails(email, password)
		if ferr != nil {
			return data, true, ferr
		}
		data.Inbox = make([]publicMailItem, 0, len(inbox))
		for _, item := range inbox {
			data.Inbox = append(data.Inbox, publicMailItem{
				Subject:    item.Subject,
				ReceivedAt: item.Date,
				Mailbox:    item.Mailbox,
				Code:       item.Code,
				ViewHref:   item.ViewHref,
			})
		}
		data.Junk = make([]publicMailItem, 0, len(junk))
		for _, item := range junk {
			data.Junk = append(data.Junk, publicMailItem{
				Subject:    item.Subject,
				ReceivedAt: item.Date,
				Mailbox:    item.Mailbox,
				Code:       item.Code,
				ViewHref:   item.ViewHref,
			})
		}
	case "toolsvip":
		inbox, junk, ferr := webmail.FetchToolsvipMails(email, password)
		if ferr != nil {
			return data, true, ferr
		}
		data.Inbox = make([]publicMailItem, 0, len(inbox))
		for _, item := range inbox {
			data.Inbox = append(data.Inbox, publicMailItem{
				Subject:    item.Subject,
				ReceivedAt: item.Date,
				Mailbox:    item.Mailbox,
				Code:       item.Code,
				Body:       item.Body,
				HtmlBody:   item.HtmlBody,
			})
		}
		data.Junk = make([]publicMailItem, 0, len(junk))
		for _, item := range junk {
			data.Junk = append(data.Junk, publicMailItem{
				Subject:    item.Subject,
				ReceivedAt: item.Date,
				Mailbox:    item.Mailbox,
				Code:       item.Code,
				Body:       item.Body,
				HtmlBody:   item.HtmlBody,
			})
		}
	case "lurentool":
		fetched, ferr := webmail.FetchLurentoolMails(email, password, 0)
		if ferr != nil {
			return data, true, ferr
		}
		if fetched != nil {
			data.Email = fetched.Email
			data.Message = fetched.Message
			data.Inbox = make([]publicMailItem, 0, len(fetched.Mails))
			for _, item := range fetched.Mails {
				data.Inbox = append(data.Inbox, publicMailItem{
					Subject:    item.Subject,
					From:       item.From,
					ReceivedAt: item.ReceivedAt,
					Preview:    item.Preview,
					Mailbox:    "inbox",
					Code:       item.Code,
					Body:       item.Body,
				})
			}
		}
	}
	return data, true, nil
}

// GetCursorEmailMails 公开邮箱取件：优先原生 Graph/IMAP，没有原生记录再走 code_link 快捷取件
// GET /api/email/cursor/query?q=<urlencoded account----pass>
func GetCursorEmailMails(c *gin.Context) {
	card, ok := loadCursorPublicCard(c)
	if !ok {
		return
	}

	inbox, junk, email, source, hasNative, err := fetchNativeOutlookByCard(card)
	if hasNative {
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Failed to fetch mail: " + err.Error()})
			return
		}
		writePublicMailOK(c, publicMailFetchData{
			Account: card.Account,
			Email:   email,
			Source:  source,
			Inbox:   outlookItemsToPublic(inbox, "inbox"),
			Junk:    outlookItemsToPublic(junk, "junk"),
		})
		return
	}

	data, hasQuick, qerr := fetchQuickMailByCard(card)
	if hasQuick {
		if qerr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Failed to fetch mail: " + qerr.Error()})
			return
		}
		data.Account = card.Account
		writePublicMailOK(c, data)
		return
	}

	writePublicMailOK(c, publicMailFetchData{
		Account: card.Account,
		Email:   card.Account,
		Message: "No mailbox is available.",
	})
}

// GetCursorEmailMailDetail 公开邮箱取件正文（仅原生取件）
// GET /api/email/cursor/detail?q=...&message_id=&folder=&seq_num=
func GetCursorEmailMailDetail(c *gin.Context) {
	card, ok := loadCursorPublicCard(c)
	if !ok {
		return
	}

	mail, err := model.GetMicrosoftMailByCard(cursorCardTable, card.Id)
	if err != nil || mail == nil || strings.TrimSpace(mail.Token) == "" || strings.TrimSpace(mail.ClientId) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Native mailbox is not available."})
		return
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

	messageID := strings.TrimSpace(c.Query("message_id"))
	folder := strings.TrimSpace(c.Query("folder"))
	var seqNum uint32
	if raw := strings.TrimSpace(c.Query("seq_num")); raw != "" {
		n, nerr := strconv.ParseUint(raw, 10, 32)
		if nerr == nil {
			seqNum = uint32(n)
		}
	}

	var detail *outlook.MailDetail
	if messageID != "" {
		graphTokens, graphErr := outlook.RefreshGraphAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if graphErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Graph OAuth failed: " + graphErr.Error()})
			return
		}
		detail, err = outlook.FetchBodyByGraphID(httpClient, graphTokens, messageID)
	} else {
		if folder == "" || seqNum == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Missing folder/seq_num or message_id"})
			return
		}
		accessTokens, tokenErr := outlook.RefreshAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if tokenErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "OAuth failed: " + tokenErr.Error()})
			return
		}
		detail, err = outlook.FetchBodyBySeq(acc, accessTokens, folder, seqNum)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Failed to load message: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    detail,
	})
}

// GetCursorRecoveryMails 公开辅助邮箱取件：字段为空直接返回空列表，有值则走已实现的快捷取码
// GET /api/recovery-mail/cursor/query?q=<urlencoded account----pass>
func GetCursorRecoveryMails(c *gin.Context) {
	card, ok := loadCursorPublicCard(c)
	if !ok {
		return
	}

	email := strings.TrimSpace(card.RecoveryMail)
	password := strings.TrimSpace(card.RecoveryMailPass)
	host := strings.TrimSpace(card.RecoveryMailHost)
	if email == "" || password == "" || host == "" {
		writePublicMailOK(c, publicMailFetchData{Account: card.Account})
		return
	}
	if !webmail.IsOuleyxHost(host) {
		writePublicMailOK(c, publicMailFetchData{Account: card.Account, Email: email})
		return
	}

	fetched, err := webmail.FetchOuleyxMails(email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Failed to fetch mail: " + err.Error()})
		return
	}

	writePublicMailOK(c, publicMailFetchData{
		Account: card.Account,
		Email:   fetched.Email,
		Source:  "ouleyx",
		Inbox:   ouleyxItemsToPublic(fetched.Inbox),
		Junk:    ouleyxItemsToPublic(fetched.Junk),
	})
}

// GetCursorRecoveryMailDetail 公开辅助邮箱正文（ouleyx）
// GET /api/recovery-mail/cursor/detail?q=...&mail_id=
func GetCursorRecoveryMailDetail(c *gin.Context) {
	card, ok := loadCursorPublicCard(c)
	if !ok {
		return
	}

	email := strings.TrimSpace(card.RecoveryMail)
	password := strings.TrimSpace(card.RecoveryMailPass)
	host := strings.TrimSpace(card.RecoveryMailHost)
	mailID := strings.TrimSpace(c.Query("mail_id"))
	if email == "" || password == "" || host == "" || mailID == "" || !webmail.IsOuleyxHost(host) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Unable to load this message."})
		return
	}

	detail, err := webmail.FetchOuleyxMailDetail(email, password, "", mailID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    detail,
	})
}
