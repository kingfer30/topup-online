package apiplatform

import (
	"math"
	"regexp"
	"strings"
)

var upstreamStatusPattern = regexp.MustCompile(`bad response status code (\d+)`)

// QuotaToUSD 把渠道 used_quota 换成与「已使用」列相同的美元数。
func QuotaToUSD(usedQuota int64, quotaPerUnit float64) float64 {
	if quotaPerUnit <= 0 {
		quotaPerUnit = 500000
	}
	return round(float64(usedQuota)/quotaPerUnit, 6)
}

// UnsettledAfter 按消耗减已结算得到未结算，不为负。
func UnsettledAfter(used, settled float64) float64 {
	value := round(used-settled, 6)
	if value < 0 {
		return 0
	}
	return value
}

func round(v float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(v*pow) / pow
}

// ParseTestStatus 从测试响应里取出上游状态。
// 只有正文里的 bad response status code 为 401/403/404 时才需要删模型。
func ParseTestStatus(httpStatus int, success bool, message string) (code int, dropModel bool) {
	if success {
		return 200, false
	}
	if match := upstreamStatusPattern.FindStringSubmatch(message); len(match) == 2 {
		n := 0
		for _, r := range match[1] {
			n = n*10 + int(r-'0')
		}
		return n, n == 401 || n == 403 || n == 404
	}
	if httpStatus == 0 {
		return 0, false
	}
	return httpStatus, false
}

// RemoveModel 从逗号分隔的模型列表里去掉一个模型。
func RemoveModel(models, name string) (string, bool) {
	name = strings.TrimSpace(name)
	parts := splitModels(models)
	out := make([]string, 0, len(parts))
	removed := false
	for _, part := range parts {
		if part == name {
			removed = true
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, ","), removed
}

func splitModels(models string) []string {
	raw := strings.Split(models, ",")
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
