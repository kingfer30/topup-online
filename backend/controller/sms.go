package controller

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/smscode"
)

// cursorCardTable Cursor 账号卡密所在表名
const cursorCardTable = "cards_cursor"

// parseCursorPublicQuery 解析公开取码页 q=account----pass，失败时已写入响应
func parseCursorPublicQuery(c *gin.Context) (account, pass string, ok bool) {
	rawParam := strings.TrimSpace(c.Query("q"))
	if rawParam == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Query parameter is required."})
		return "", "", false
	}

	// c.Query 已经做过一次标准 URL 解码（对应前端 axios 发请求时的编码），
	// 这里再做一次显式解码，还原 admin 生成链接时对 account/pass 做的那一次编码。
	raw, err := url.QueryUnescape(rawParam)
	if err != nil {
		raw = rawParam
	}

	sepIndex := strings.Index(raw, "----")
	if sepIndex == -1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Invalid query format."})
		return "", "", false
	}
	account = strings.TrimSpace(raw[:sepIndex])
	pass = strings.TrimSpace(raw[sepIndex+len("----"):])
	if account == "" || pass == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Account or password is required."})
		return "", "", false
	}
	return account, pass, true
}

// loadCursorPublicCard 按公开取码链接定位并校验 cards_cursor 记录，失败时已写入响应
func loadCursorPublicCard(c *gin.Context) (*model.AccountCard, bool) {
	account, pass, ok := parseCursorPublicQuery(c)
	if !ok {
		return nil, false
	}

	card, err := model.GetCardByAccount(cursorCardTable, account)
	if err != nil || card == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "Account not found or inactive."})
		return nil, false
	}
	if card.Password != pass {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Incorrect account or password."})
		return nil, false
	}
	return card, true
}

// GetCursorSmsCode 独立取码页专用接口（公开，无需管理员认证）
// GET /api/sms/cursor/query?q=<urlencoded account----pass>
// 生成取码链接时（admin 后台）account/pass 已分别做过一次 URL 编码，
// 前端不做解码，原样把 q 传回来，这里统一做最终的 URL 解码还原真实内容，再按 ---- 拆分为 account/pass。
// 这样可以避免密码中包含 # & 等特殊字符被浏览器当作分隔符处理导致的拆分错误。
// 依据 account 在 cards_cursor 表中查找记录，校验 pass 后取出 phone_link 并抓取短信验证码
func GetCursorSmsCode(c *gin.Context) {
	card, ok := loadCursorPublicCard(c)
	if !ok {
		return
	}

	if strings.TrimSpace(card.PhoneLink) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "This account has no SMS inbox configured."})
		return
	}

	result, err := smscode.FetchCode(card.PhoneLink, card.Account, card.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "Failed to fetch code: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, smsCodeSuccessBody(card.Account, result))
}

// GetCardSmsCode 管理端根据卡密 ID 抓取短信验证码（需要管理员认证），用于列表页"接码-短信接码"弹窗
// GET /api/admin/cards/:id/sms-code?category=xxx
func GetCardSmsCode(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "缺少卡密类别参数"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的卡密ID"})
		return
	}

	tableName := model.GetTableNameByCategory(category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	card, err := model.GetCardById(tableName, id)
	if err != nil || card == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "卡密不存在"})
		return
	}

	if strings.TrimSpace(card.PhoneLink) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该账号未配置接码地址"})
		return
	}

	result, err := smscode.FetchCode(card.PhoneLink, card.Account, card.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "取码失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, smsCodeSuccessBody(card.Account, result))
}

// smsCodeSuccessBody 统一的取码成功响应结构，供公开取码接口与管理端接口共用
func smsCodeSuccessBody(account string, result *smscode.Result) gin.H {
	return gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"account":    account,
			"status":     result.Status,
			"code":       result.Code,
			"message":    result.Message,
			"expires_at": result.ExpiresAt,
		},
	}
}
