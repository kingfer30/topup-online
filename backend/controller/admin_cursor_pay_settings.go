package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

// GetAdminCursorPaySettings 获取 Cursor 付款设置（不含明文代理密码）
func GetAdminCursorPaySettings(c *gin.Context) {
	cfg, err := model.GetCursorPaySettingsPublic()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "读取配置失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    cfg,
	})
}

type updateAdminCursorPaySettingsRequest struct {
	BillingName    string `json:"billing_name"`
	BillingPostal  string `json:"billing_postal"`
	BillingState   string `json:"billing_state"`
	BillingCity    string `json:"billing_city"`
	BillingLine1   string `json:"billing_line1"`
	BillingCountry string `json:"billing_country"`
	ProxyScheme    string `json:"proxy_scheme"`
	ProxyHost      string `json:"proxy_host"`
	ProxyUsername  string `json:"proxy_username"`
	ProxyPassword  string `json:"proxy_password"`
}

// UpdateAdminCursorPaySettings 更新 Cursor 付款设置；代理密码为空时不修改已保存的密码
func UpdateAdminCursorPaySettings(c *gin.Context) {
	var req updateAdminCursorPaySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	req.BillingName = strings.TrimSpace(req.BillingName)
	req.BillingPostal = strings.TrimSpace(req.BillingPostal)
	req.BillingState = strings.TrimSpace(req.BillingState)
	req.BillingCity = strings.TrimSpace(req.BillingCity)
	req.BillingLine1 = strings.TrimSpace(req.BillingLine1)
	req.BillingCountry = strings.ToUpper(strings.TrimSpace(req.BillingCountry))
	req.ProxyScheme = strings.ToLower(strings.TrimSpace(req.ProxyScheme))
	req.ProxyHost = strings.TrimSpace(req.ProxyHost)
	req.ProxyUsername = strings.TrimSpace(req.ProxyUsername)
	req.ProxyPassword = strings.TrimSpace(req.ProxyPassword)

	if req.BillingName == "" {
		req.BillingName = model.DefaultCursorPayBillingName
	}
	if req.BillingPostal == "" {
		req.BillingPostal = model.DefaultCursorPayBillingPostal
	}
	if req.BillingState == "" {
		req.BillingState = model.DefaultCursorPayBillingState
	}
	if req.BillingCity == "" {
		req.BillingCity = model.DefaultCursorPayBillingCity
	}
	if req.BillingLine1 == "" {
		req.BillingLine1 = model.DefaultCursorPayBillingLine1
	}
	if req.BillingCountry == "" {
		req.BillingCountry = model.DefaultCursorPayBillingCountry
	}
	if req.ProxyScheme == "" {
		req.ProxyScheme = "http"
	}

	pairs := [][2]string{
		{model.KeyCursorPayBillingName, req.BillingName},
		{model.KeyCursorPayBillingPostal, req.BillingPostal},
		{model.KeyCursorPayBillingState, req.BillingState},
		{model.KeyCursorPayBillingCity, req.BillingCity},
		{model.KeyCursorPayBillingLine1, req.BillingLine1},
		{model.KeyCursorPayBillingCountry, req.BillingCountry},
		{model.KeyCursorPayProxyScheme, req.ProxyScheme},
		{model.KeyCursorPayProxyHost, req.ProxyHost},
		{model.KeyCursorPayProxyUsername, req.ProxyUsername},
	}
	for _, pair := range pairs {
		if err := model.UpsertSystemConfig(pair[0], pair[1]); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存失败: " + err.Error()})
			return
		}
	}
	if req.ProxyPassword != "" {
		if err := model.UpsertSystemConfig(model.KeyCursorPayProxyPassword, req.ProxyPassword); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存代理密码失败: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
	})
}
