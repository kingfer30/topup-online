package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/utils/webmail"
)

type webmailFetchRequest struct {
	AccountLine   string `json:"account_line" binding:"required"`
	AccountFormat string `json:"account_format"`
	Top           int    `json:"top"`
}

// LqqqFetch 获取 lqqq 邮件列表
func LqqqFetch(c *gin.Context) {
	var req webmailFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	email, password, err := webmail.ParseWebMailAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	inbox, junk, err := webmail.FetchLqqqMails(email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "取件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"email": email,
			"inbox": inbox,
			"junk":  junk,
		},
	})
}

type lqqqDetailRequest struct {
	AccountLine   string `json:"account_line" binding:"required"`
	AccountFormat string `json:"account_format"`
	ViewHref      string `json:"view_href" binding:"required"`
}

// LqqqDetail 拉取 lqqq 单封邮件详情
func LqqqDetail(c *gin.Context) {
	var req lqqqDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	email, password, err := webmail.ParseWebMailAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	detail, err := webmail.FetchLqqqMailDetail(email, password, req.ViewHref)
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

// ToolsvipFetch 获取 toolsvip 邮件列表
func ToolsvipFetch(c *gin.Context) {
	var req webmailFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	email, password, err := webmail.ParseWebMailAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	inbox, junk, err := webmail.FetchToolsvipMails(email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "取件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"email": email,
			"inbox": inbox,
			"junk":  junk,
		},
	})
}

// LurentoolFetch 通过 lurentool 快捷取件接口拉取邮件
func LurentoolFetch(c *gin.Context) {
	var req webmailFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	email, password, err := webmail.ParseWebMailAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	data, err := webmail.FetchLurentoolMails(email, password, req.Top)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "取件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    data,
	})
}

// OuleyxFetch 登录 ouleyx 并拉取收件箱/垃圾箱列表
func OuleyxFetch(c *gin.Context) {
	var req webmailFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	email, password, err := webmail.ParseWebMailAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	data, err := webmail.FetchOuleyxMails(email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "取件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    data,
	})
}

type ouleyxDetailRequest struct {
	AccountLine   string `json:"account_line" binding:"required"`
	AccountFormat string `json:"account_format"`
	Token         string `json:"token"`
	MailID        string `json:"mail_id" binding:"required"`
}

// OuleyxDetail 拉取 ouleyx 单封邮件详情
func OuleyxDetail(c *gin.Context) {
	var req ouleyxDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	line := firstNonEmptyLine(req.AccountLine)
	email, password, err := webmail.ParseWebMailAccountLine(line, req.AccountFormat)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	detail, err := webmail.FetchOuleyxMailDetail(email, password, req.Token, req.MailID)
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
