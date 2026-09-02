package model

import (
	"errors"

	"gorm.io/gorm"
)

const (
	KeyAIModelName = "ai_model_name"
	KeyAIBaseURL   = "ai_base_url"
	KeyAIAPIKey    = "ai_api_key"

	KeyCursorPayBillingName    = "cursor_pay_billing_name"
	KeyCursorPayBillingPostal  = "cursor_pay_billing_postal"
	KeyCursorPayBillingState   = "cursor_pay_billing_state"
	KeyCursorPayBillingCity    = "cursor_pay_billing_city"
	KeyCursorPayBillingLine1   = "cursor_pay_billing_line1"
	KeyCursorPayBillingCountry = "cursor_pay_billing_country"
	KeyCursorPayProxyScheme    = "cursor_pay_proxy_scheme"
	KeyCursorPayProxyHost      = "cursor_pay_proxy_host"
	KeyCursorPayProxyUsername  = "cursor_pay_proxy_username"
	KeyCursorPayProxyPassword  = "cursor_pay_proxy_password"

	DefaultCursorPayBillingName    = "AIGuoGuo"
	DefaultCursorPayBillingPostal  = "536546"
	DefaultCursorPayBillingState   = "Zhejiang"
	DefaultCursorPayBillingCity    = "Huzhou"
	DefaultCursorPayBillingLine1   = "清河路177号"
	DefaultCursorPayBillingCountry = "CN"
)

// GetSystemConfigValue 按 key 读取配置值，不存在则返回空字符串
func GetSystemConfigValue(key string) (string, error) {
	var sc SystemConfig
	err := DB.Where("`key` = ?", key).First(&sc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return sc.Value, nil
}

// UpsertSystemConfig 写入或更新一条系统配置
func UpsertSystemConfig(key, value string) error {
	var sc SystemConfig
	err := DB.Where("`key` = ?", key).First(&sc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DB.Create(&SystemConfig{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	sc.Value = value
	return DB.Save(&sc).Error
}

// AISettingsPublic 返回给前端的 AI 设置（不含明文 API Key）
type AISettingsPublic struct {
	ModelName          string `json:"model_name"`
	BaseURL            string `json:"base_url"`
	APIKeyConfigured   bool   `json:"api_key_configured"`
}

// GetAISettingsPublic 读取 AI 模型相关配置（用于设置页）
func GetAISettingsPublic() (*AISettingsPublic, error) {
	modelName, err := GetSystemConfigValue(KeyAIModelName)
	if err != nil {
		return nil, err
	}
	baseURL, err := GetSystemConfigValue(KeyAIBaseURL)
	if err != nil {
		return nil, err
	}
	apiKey, err := GetSystemConfigValue(KeyAIAPIKey)
	if err != nil {
		return nil, err
	}
	return &AISettingsPublic{
		ModelName:        modelName,
		BaseURL:          baseURL,
		APIKeyConfigured: apiKey != "",
	}, nil
}

// GetAISettingsForTranslate 读取调用大模型所需的完整配置
func GetAISettingsForTranslate() (modelName, baseURL, apiKey string, err error) {
	modelName, err = GetSystemConfigValue(KeyAIModelName)
	if err != nil {
		return "", "", "", err
	}
	baseURL, err = GetSystemConfigValue(KeyAIBaseURL)
	if err != nil {
		return "", "", "", err
	}
	apiKey, err = GetSystemConfigValue(KeyAIAPIKey)
	if err != nil {
		return "", "", "", err
	}
	return modelName, baseURL, apiKey, nil
}
