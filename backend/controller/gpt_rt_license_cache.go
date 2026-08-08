package controller

import (
	"encoding/json"
	"fmt"
	"time"

	redis "github.com/kingfer30/topup-online/config/cache"
)

const (
	gptRtLicQuotaKeyPrefix   = "gpt_rt_lic:quota:"
	gptRtLicRevokedKeyPrefix = "gpt_rt_lic:revoked:"
	gptRtLicQuotaCacheTTL    = 5 * time.Second
)

// gptRtLicenseQuotaCache 配额缓存结构
type gptRtLicenseQuotaCache struct {
	UsedCount      int `json:"used_count"`
	AvailableCount int `json:"available_count"`
	HeldCount      int `json:"held_count"`
}

func gptRtLicQuotaKey(id int) string {
	return fmt.Sprintf("%s%d", gptRtLicQuotaKeyPrefix, id)
}

func gptRtLicRevokedKey(licenseKey string) string {
	return gptRtLicRevokedKeyPrefix + licenseKey
}

// SetQuotaCache 写入配额短缓存
func SetQuotaCache(id int, used, available, held int) {
	if !redis.RedisEnabled {
		return
	}
	b, err := json.Marshal(gptRtLicenseQuotaCache{
		UsedCount:      used,
		AvailableCount: available,
		HeldCount:      held,
	})
	if err != nil {
		return
	}
	_ = redis.Set(gptRtLicQuotaKey(id), string(b), gptRtLicQuotaCacheTTL)
}

// GetQuotaCache 读取配额短缓存；未命中返回 false
func GetQuotaCache(id int) (used, available, held int, ok bool) {
	if !redis.RedisEnabled {
		return 0, 0, 0, false
	}
	s, err := redis.Get(gptRtLicQuotaKey(id))
	if err != nil || s == "" {
		return 0, 0, 0, false
	}
	var q gptRtLicenseQuotaCache
	if err := json.Unmarshal([]byte(s), &q); err != nil {
		return 0, 0, 0, false
	}
	return q.UsedCount, q.AvailableCount, q.HeldCount, true
}

// InvalidateQuotaCache 删除配额缓存
func InvalidateQuotaCache(id int) {
	if !redis.RedisEnabled {
		return
	}
	_ = redis.Del(gptRtLicQuotaKey(id))
}

// SetRevoked 标记许可证已禁用（撤销缓存）
func SetRevoked(licenseKey string) {
	if !redis.RedisEnabled || licenseKey == "" {
		return
	}
	_ = redis.Set(gptRtLicRevokedKey(licenseKey), "1", 0)
}

// ClearRevoked 清除禁用标记
func ClearRevoked(licenseKey string) {
	if !redis.RedisEnabled || licenseKey == "" {
		return
	}
	_ = redis.Del(gptRtLicRevokedKey(licenseKey))
}

// IsRevoked 是否在撤销缓存中
func IsRevoked(licenseKey string) bool {
	if !redis.RedisEnabled || licenseKey == "" {
		return false
	}
	n, err := redis.Exists(gptRtLicRevokedKey(licenseKey))
	return err == nil && n > 0
}
