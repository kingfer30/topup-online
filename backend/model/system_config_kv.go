package model

import (
	"errors"
	"fmt"
	"strings"

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

	KeyAPIPlatformSyncEnabled     = "api_platform_sync_enabled"
	KeyAPIPlatformSyncMinutes     = "api_platform_sync_minutes"
	KeyAPIPlatformDiscoverEnabled = "api_platform_discover_enabled"
	KeyAPIPlatformDiscoverMinutes = "api_platform_discover_minutes"
	KeyAPIPlatformTestModels      = "api_platform_test_models"

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
	ModelName        string `json:"model_name"`
	BaseURL          string `json:"base_url"`
	APIKeyConfigured bool   `json:"api_key_configured"`
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

const DefaultAPIPlatformTestModels = "gpt-5.6-terra,gpt-5.6-sol,gpt-5.6-luna,gpt-6-sol,gpt-6.1-sol,gpt-6-astra,gpt-6-luna"

// APIPlatformSettings 后台 API 台子定时与测试模型
type APIPlatformSettings struct {
	SyncEnabled     bool     `json:"sync_enabled"`
	SyncMinutes     int      `json:"sync_minutes"`
	DiscoverEnabled bool     `json:"discover_enabled"`
	DiscoverMinutes int      `json:"discover_minutes"`
	TestModels      []string `json:"test_models"`
}

func parseBoolConfig(value string, fallback bool) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parsePositiveInt(value string, fallback int) int {
	n := 0
	for _, r := range strings.TrimSpace(value) {
		if r < '0' || r > '9' {
			return fallback
		}
		n = n*10 + int(r-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}

// SplitAPIPlatformModels 把逗号分隔的模型列表整理成去重后的顺序
func SplitAPIPlatformModels(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n'
	})
	seen := map[string]struct{}{}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// GetAPIPlatformSettings 读取设置，缺省时使用默认值
func GetAPIPlatformSettings() (*APIPlatformSettings, error) {
	syncEnabled, err := GetSystemConfigValue(KeyAPIPlatformSyncEnabled)
	if err != nil {
		return nil, err
	}
	syncMinutes, err := GetSystemConfigValue(KeyAPIPlatformSyncMinutes)
	if err != nil {
		return nil, err
	}
	discoverEnabled, err := GetSystemConfigValue(KeyAPIPlatformDiscoverEnabled)
	if err != nil {
		return nil, err
	}
	discoverMinutes, err := GetSystemConfigValue(KeyAPIPlatformDiscoverMinutes)
	if err != nil {
		return nil, err
	}
	models, err := GetSystemConfigValue(KeyAPIPlatformTestModels)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(models) == "" {
		models = DefaultAPIPlatformTestModels
	}
	list := SplitAPIPlatformModels(models)
	if len(list) == 0 {
		list = SplitAPIPlatformModels(DefaultAPIPlatformTestModels)
	}
	return &APIPlatformSettings{
		SyncEnabled:     parseBoolConfig(syncEnabled, true),
		SyncMinutes:     parsePositiveInt(syncMinutes, 30),
		DiscoverEnabled: parseBoolConfig(discoverEnabled, true),
		DiscoverMinutes: parsePositiveInt(discoverMinutes, 30),
		TestModels:      list,
	}, nil
}

// SaveAPIPlatformSettings 写入 API 台子设置
func SaveAPIPlatformSettings(cfg APIPlatformSettings) error {
	if cfg.SyncMinutes < 1 {
		cfg.SyncMinutes = 30
	}
	if cfg.DiscoverMinutes < 1 {
		cfg.DiscoverMinutes = 30
	}
	models := SplitAPIPlatformModels(strings.Join(cfg.TestModels, ","))
	if len(models) == 0 {
		models = SplitAPIPlatformModels(DefaultAPIPlatformTestModels)
	}
	pairs := []struct{ key, value string }{
		{KeyAPIPlatformSyncEnabled, boolConfig(cfg.SyncEnabled)},
		{KeyAPIPlatformSyncMinutes, fmt.Sprintf("%d", cfg.SyncMinutes)},
		{KeyAPIPlatformDiscoverEnabled, boolConfig(cfg.DiscoverEnabled)},
		{KeyAPIPlatformDiscoverMinutes, fmt.Sprintf("%d", cfg.DiscoverMinutes)},
		{KeyAPIPlatformTestModels, strings.Join(models, ",")},
	}
	for _, pair := range pairs {
		if err := UpsertSystemConfig(pair.key, pair.value); err != nil {
			return err
		}
	}
	return nil
}

func boolConfig(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
