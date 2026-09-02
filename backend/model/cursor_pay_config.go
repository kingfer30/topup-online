package model

import (
	"net/url"
	"strings"
)

// CursorPaySettingsPublic 返回给设置页的 Cursor 付款配置（不含明文代理密码）
type CursorPaySettingsPublic struct {
	BillingName             string `json:"billing_name"`
	BillingPostal           string `json:"billing_postal"`
	BillingState            string `json:"billing_state"`
	BillingCity             string `json:"billing_city"`
	BillingLine1            string `json:"billing_line1"`
	BillingCountry          string `json:"billing_country"`
	ProxyScheme             string `json:"proxy_scheme"`
	ProxyHost               string `json:"proxy_host"`
	ProxyUsername           string `json:"proxy_username"`
	ProxyPasswordConfigured bool   `json:"proxy_password_configured"`
}

// CursorPayBilling Stripe Alipay 账单地址
type CursorPayBilling struct {
	Name       string
	PostalCode string
	State      string
	City       string
	Line1      string
	Country    string
}

func configOrDefault(key, fallback string) string {
	val, err := GetSystemConfigValue(key)
	if err != nil {
		return fallback
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return fallback
	}
	return val
}

// GetCursorPaySettingsPublic 读取设置页展示用配置
func GetCursorPaySettingsPublic() (*CursorPaySettingsPublic, error) {
	password, err := GetSystemConfigValue(KeyCursorPayProxyPassword)
	if err != nil {
		return nil, err
	}
	scheme, err := GetSystemConfigValue(KeyCursorPayProxyScheme)
	if err != nil {
		return nil, err
	}
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	if scheme == "" {
		scheme = "http"
	}
	return &CursorPaySettingsPublic{
		BillingName:             configOrDefault(KeyCursorPayBillingName, DefaultCursorPayBillingName),
		BillingPostal:           configOrDefault(KeyCursorPayBillingPostal, DefaultCursorPayBillingPostal),
		BillingState:            configOrDefault(KeyCursorPayBillingState, DefaultCursorPayBillingState),
		BillingCity:             configOrDefault(KeyCursorPayBillingCity, DefaultCursorPayBillingCity),
		BillingLine1:            configOrDefault(KeyCursorPayBillingLine1, DefaultCursorPayBillingLine1),
		BillingCountry:          configOrDefault(KeyCursorPayBillingCountry, DefaultCursorPayBillingCountry),
		ProxyScheme:             scheme,
		ProxyHost:               configOrDefault(KeyCursorPayProxyHost, ""),
		ProxyUsername:           configOrDefault(KeyCursorPayProxyUsername, ""),
		ProxyPasswordConfigured: strings.TrimSpace(password) != "",
	}, nil
}

// GetCursorPayBilling 读取实际提交 Stripe 时使用的账单地址
func GetCursorPayBilling() CursorPayBilling {
	return CursorPayBilling{
		Name:       configOrDefault(KeyCursorPayBillingName, DefaultCursorPayBillingName),
		PostalCode: configOrDefault(KeyCursorPayBillingPostal, DefaultCursorPayBillingPostal),
		State:      configOrDefault(KeyCursorPayBillingState, DefaultCursorPayBillingState),
		City:       configOrDefault(KeyCursorPayBillingCity, DefaultCursorPayBillingCity),
		Line1:      configOrDefault(KeyCursorPayBillingLine1, DefaultCursorPayBillingLine1),
		Country:    strings.ToUpper(configOrDefault(KeyCursorPayBillingCountry, DefaultCursorPayBillingCountry)),
	}
}

// GetCursorPayProxyURL 组装 Stripe/Alipay 提取流程使用的代理；未配置时返回 nil
func GetCursorPayProxyURL() *url.URL {
	host, err := GetSystemConfigValue(KeyCursorPayProxyHost)
	if err != nil {
		return nil
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return nil
	}

	if strings.Contains(host, "://") {
		parsed, parseErr := url.Parse(host)
		if parseErr != nil || parsed.Host == "" {
			return nil
		}
		return parsed
	}

	scheme, _ := GetSystemConfigValue(KeyCursorPayProxyScheme)
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	if scheme == "" {
		scheme = "http"
	}
	parsed := &url.URL{
		Scheme: scheme,
		Host:   host,
	}
	username, _ := GetSystemConfigValue(KeyCursorPayProxyUsername)
	password, _ := GetSystemConfigValue(KeyCursorPayProxyPassword)
	username = strings.TrimSpace(username)
	if username != "" {
		parsed.User = url.UserPassword(username, password)
	}
	return parsed
}
