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
// GET /api/sms/cursor/query?account=xxx&pass=xxx
// 依据 account 在 cards_cursor 表中查找记录，校验 pass 后取出 phone_link 并抓取短信验证码
func GetCursorSmsCode(c *gin.Context) {
	account := strings.TrimSpace(c.Query("account"))
	pass := strings.TrimSpace(c.Query("pass"))

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
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "账号不存在或已失效"})
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
		"data":    result,
	})
}
