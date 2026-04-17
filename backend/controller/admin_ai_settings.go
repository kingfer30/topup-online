package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

// GetAdminAISettings 获取 AI 模型设置（不含 API Key 明文）
func GetAdminAISettings(c *gin.Context) {
	cfg, err := model.GetAISettingsPublic()
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

type updateAdminAISettingsRequest struct {
	ModelName string `json:"model_name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
}

// UpdateAdminAISettings 更新 AI 模型设置；api_key 为空字符串时不修改已保存的密钥
func UpdateAdminAISettings(c *gin.Context) {
	var req updateAdminAISettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	req.ModelName = strings.TrimSpace(req.ModelName)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.APIKey = strings.TrimSpace(req.APIKey)

	if err := model.UpsertSystemConfig(model.KeyAIModelName, req.ModelName); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存模型名称失败: " + err.Error()})
		return
	}
	if err := model.UpsertSystemConfig(model.KeyAIBaseURL, req.BaseURL); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存 Base URL 失败: " + err.Error()})
		return
	}
	if req.APIKey != "" {
		if err := model.UpsertSystemConfig(model.KeyAIAPIKey, req.APIKey); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存 API Key 失败: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
	})
}
