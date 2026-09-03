package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/smscode"
)

// cursorCardTable Cursor 账号卡密所在表名
const cursorCardTable = "cards_cursor"

// GetCursorSmsCode 独立取码页专用接口（公开，无需管理员认证）
// GET /api/sms/cursor/query?q=account----pass
// q 为地址栏 ? 后面的原始整串内容（未做任何拆分），由后端统一按 ---- 拆分为 account/pass，
// 避免前端拆分时因密码中包含 # 等字符被浏览器截断导致拆分错误。
// 依据 account 在 cards_cursor 表中查找记录，校验 pass 后取出 phone_link 并抓取短信验证码
func GetCursorSmsCode(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("q"))
	if raw == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "查询参数不能为空"})
		return
	}

	sepIndex := strings.Index(raw, "----")
	if sepIndex == -1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "查询参数格式错误，缺少 ---- 分隔符"})
		return
	}
	account := strings.TrimSpace(raw[:sepIndex])
	pass := strings.TrimSpace(raw[sepIndex+len("----"):])
	if account == "" || pass == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "账号或密码不能为空"})
		return
	}

	card, err := model.GetCardByAccount(cursorCardTable, account)
	if err != nil || card == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "账号不存在或已失效"})
		return
	}

	if card.Password != pass {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "账号或密码错误"})
		return
	}

	if strings.TrimSpace(card.PhoneLink) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该账号未配置接码地址"})
		return
	}

	result, err := smscode.FetchCode(card.PhoneLink, account, pass)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "取码失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"account":    account,
			"status":     result.Status,
			"code":       result.Code,
			"message":    result.Message,
			"expires_at": result.ExpiresAt,
		},
	})
}
