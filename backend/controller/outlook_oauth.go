package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/outlook"
)

type outlookFetchRequest struct {
	AccountLine   string `json:"account_line" binding:"required"`
	AccountFormat string `json:"account_format"`
}

type outlookDetailRequest struct {
	AccountLine   string `json:"account_line" binding:"required"`
	AccountFormat string `json:"account_format"`
	Folder        string `json:"folder"`
	SeqNum        uint32 `json:"seq_num"`
	MessageID     string `json:"message_id"` // Graph 取件时的邮件 id
}

// OutlookOauthFetch 解析账号行，优先 Graph 取件，失败回退 IMAP
func OutlookOauthFetch(c *gin.Context) {
	var req outlookFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	if line == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "账号行为空"})
		return
	}

	acc, err := outlook.ParseAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	source := "graph"
	var inbox, junk []outlook.MailItem
	var graphErr error

	graphTokens, err := outlook.RefreshGraphAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
	if err != nil {
		graphErr = err
		logger.SysLog("OutlookOauthFetch: " + acc.Email + " Graph oauth failed: " + err.Error())
	} else {
		logger.SysLog("OutlookOauthFetch: " + acc.Email + " Graph oauth ok")
		inbox, junk, graphErr = outlook.FetchViaGraph(httpClient, graphTokens)
	}

	if graphErr != nil {
		logger.SysLog("OutlookOauthFetch: " + acc.Email + " Graph failed, fallback IMAP: " + graphErr.Error())
		accessTokens, imapTokenErr := outlook.RefreshAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if imapTokenErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "取件失败: Graph=" + graphErr.Error() + "；IMAP OAuth=" + imapTokenErr.Error(),
			})
			return
		}
		inbox, junk, err = outlook.FetchViaIMAP(acc, accessTokens)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "取件失败: Graph=" + graphErr.Error() + "；IMAP=" + err.Error(),
			})
			return
		}
		source = "imap"
		logger.SysLog("OutlookOauthFetch: " + acc.Email + " IMAP ok")
	}

	if inbox == nil {
		inbox = []outlook.MailItem{}
	}
	if junk == nil {
		junk = []outlook.MailItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"email":  acc.Email,
			"source": source,
			"inbox":  inbox,
			"junk":   junk,
		},
	})
}

// OutlookFetchDetail 按序列号 / Graph message id 按需拉取单封邮件正文
func OutlookFetchDetail(c *gin.Context) {
	var req outlookDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	acc, err := outlook.ParseAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	var detail *outlook.MailDetail

	// Graph 路径：有 message_id 时优先走 Graph
	if strings.TrimSpace(req.MessageID) != "" {
		graphTokens, graphErr := outlook.RefreshGraphAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if graphErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Graph OAuth 失败: " + graphErr.Error()})
			return
		}
		detail, err = outlook.FetchBodyByGraphID(httpClient, graphTokens, req.MessageID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "拉取正文失败: " + err.Error()})
			return
		}
	} else {
		if req.Folder == "" || req.SeqNum == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "缺少 folder/seq_num 或 message_id"})
			return
		}
		accessTokens, tokenErr := outlook.RefreshAccessTokens(httpClient, acc.ClientID, acc.RefreshToken)
		if tokenErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "OAuth 授权失败: " + tokenErr.Error()})
			return
		}
		detail, err = outlook.FetchBodyBySeq(acc, accessTokens, req.Folder, req.SeqNum)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "拉取正文失败: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    detail,
	})
}

func firstNonEmptyLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return ""
}
