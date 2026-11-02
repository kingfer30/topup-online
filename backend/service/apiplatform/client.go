package apiplatform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RemoteChannel 上游渠道列表中的一行
type RemoteChannel struct {
	ID        int
	Name      string
	Status    int
	Models    string
	UsedQuota int64
}

type remoteChannelJSON struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Status    int     `json:"status"`
	Models    string  `json:"models"`
	UsedQuota float64 `json:"used_quota"`
}

func (row remoteChannelJSON) toRemote() RemoteChannel {
	return RemoteChannel{
		ID:        row.ID,
		Name:      row.Name,
		Status:    row.Status,
		Models:    row.Models,
		UsedQuota: int64(row.UsedQuota),
	}
}

// NormalizeBase 去掉尾部斜杠，并补上协议。
func NormalizeBase(base string) (string, error) {
	base = strings.TrimSpace(base)
	base = strings.TrimRight(base, "/")
	if base == "" {
		return "", fmt.Errorf("台子地址为空")
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return base, nil
}

func authToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if len(token) >= 7 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if token == "" {
		return "", fmt.Errorf("API Key 为空")
	}
	return token, nil
}

func request(method, base, token, path string, query url.Values, body any, timeout time.Duration) (int, []byte, error) {
	base, err := NormalizeBase(base)
	if err != nil {
		return 0, nil, err
	}
	token, err = authToken(token)
	if err != nil {
		return 0, nil, err
	}
	rawURL := base + path
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("无法连接台子: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, payload, nil
}

// QuotaPerUnit 读取台子的额度换算单位。
func QuotaPerUnit(base, token string) (float64, error) {
	status, payload, err := request(http.MethodGet, base, token, "/api/status", nil, nil, 30*time.Second)
	if err != nil {
		return 0, err
	}
	if status < 200 || status >= 300 {
		return 0, fmt.Errorf("读取站点状态失败 HTTP %d", status)
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			QuotaPerUnit float64 `json:"quota_per_unit"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return 0, err
	}
	if body.Data.QuotaPerUnit <= 0 {
		return 500000, nil
	}
	return body.Data.QuotaPerUnit, nil
}

// SearchChannels 按页拉完全部搜索结果。
func SearchChannels(base, token, keyword, status string) ([]RemoteChannel, error) {
	page := 1
	const pageSize = 500
	var all []RemoteChannel
	for page <= 40 {
		query := url.Values{}
		query.Set("keyword", keyword)
		query.Set("sort_by", "id")
		query.Set("sort_order", "asc")
		query.Set("p", fmt.Sprintf("%d", page))
		query.Set("page_size", fmt.Sprintf("%d", pageSize))
		if status == "enabled" || status == "disabled" {
			query.Set("status", status)
		}
		httpStatus, payload, err := request(http.MethodGet, base, token, "/api/channel/search", query, nil, 60*time.Second)
		if err != nil {
			return nil, err
		}
		if httpStatus < 200 || httpStatus >= 300 {
			return nil, fmt.Errorf("搜索渠道失败 HTTP %d: %s", httpStatus, trimBody(payload))
		}
		var body struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				Items []remoteChannelJSON `json:"items"`
				Total int                 `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, err
		}
		if !body.Success {
			return nil, fmt.Errorf("搜索渠道失败: %s", body.Message)
		}
		if len(body.Data.Items) == 0 {
			break
		}
		for _, item := range body.Data.Items {
			all = append(all, item.toRemote())
		}
		if page*pageSize >= body.Data.Total {
			break
		}
		page++
	}
	return all, nil
}

// FilterChannels 按名称、ID 范围和状态再滤一次。
func FilterChannels(rows []RemoteChannel, keyword string, minID, maxID int, status string) []RemoteChannel {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	out := make([]RemoteChannel, 0, len(rows))
	for _, row := range rows {
		if keyword != "" && !strings.Contains(strings.ToLower(row.Name), keyword) {
			continue
		}
		if minID > 0 && row.ID < minID {
			continue
		}
		if maxID > 0 && row.ID > maxID {
			continue
		}
		if status == "enabled" && row.Status != 1 {
			continue
		}
		if status == "disabled" && row.Status == 1 {
			continue
		}
		out = append(out, row)
	}
	return out
}

// GetChannel 读取单个上游渠道。
func GetChannel(base, token string, channelID int) (RemoteChannel, error) {
	httpStatus, payload, err := request(http.MethodGet, base, token, fmt.Sprintf("/api/channel/%d", channelID), nil, nil, 30*time.Second)
	if err != nil {
		return RemoteChannel{}, err
	}
	if httpStatus == http.StatusNotFound {
		return RemoteChannel{}, fmt.Errorf("上游渠道 %d 不存在", channelID)
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return RemoteChannel{}, fmt.Errorf("读取渠道失败 HTTP %d: %s", httpStatus, trimBody(payload))
	}
	var body struct {
		Success bool              `json:"success"`
		Message string            `json:"message"`
		Data    remoteChannelJSON `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return RemoteChannel{}, err
	}
	if !body.Success {
		return RemoteChannel{}, fmt.Errorf("读取渠道失败: %s", body.Message)
	}
	return body.Data.toRemote(), nil
}

// TestChannel 调用渠道测试接口，并返回可展示的文案。
func TestChannel(base, token string, channelID int, modelName string) (code int, drop bool, message string, err error) {
	query := url.Values{}
	query.Set("model", modelName)
	query.Set("endpoint_type", "openai-response")
	httpStatus, payload, err := request(http.MethodGet, base, token, fmt.Sprintf("/api/channel/test/%d", channelID), query, nil, 120*time.Second)
	if err != nil {
		return 0, false, "", err
	}
	var body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &body)
	}
	if body.Message == "" {
		body.Message = trimBody(payload)
	}
	code, drop = ParseTestStatus(httpStatus, body.Success, body.Message)
	if body.Success {
		return code, false, "成功", nil
	}
	return code, drop, body.Message, nil
}

// UpdateChannelModels 先改上游模型，成功后由调用方再写本地。
func UpdateChannelModels(base, token string, channelID int, models string) error {
	httpStatus, payload, err := request(http.MethodGet, base, token, fmt.Sprintf("/api/channel/%d", channelID), nil, nil, 30*time.Second)
	if err != nil {
		return err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return fmt.Errorf("读取渠道失败 HTTP %d: %s", httpStatus, trimBody(payload))
	}
	var body struct {
		Success bool           `json:"success"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return err
	}
	if !body.Success || body.Data == nil {
		return fmt.Errorf("读取渠道失败: %s", body.Message)
	}
	delete(body.Data, "status")
	delete(body.Data, "key")
	body.Data["id"] = channelID
	body.Data["models"] = models
	putStatus, putPayload, err := request(http.MethodPut, base, token, "/api/channel/", nil, body.Data, 60*time.Second)
	if err != nil {
		return err
	}
	if putStatus < 200 || putStatus >= 300 {
		return fmt.Errorf("更新渠道模型失败 HTTP %d: %s", putStatus, trimBody(putPayload))
	}
	var putBody struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(putPayload, &putBody); err != nil {
		return err
	}
	if !putBody.Success {
		if putBody.Message == "" {
			putBody.Message = trimBody(putPayload)
		}
		return fmt.Errorf("更新渠道模型失败: %s", putBody.Message)
	}
	if strings.TrimSpace(models) == "" {
		current, err := GetChannel(base, token, channelID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(current.Models) != "" {
			return fmt.Errorf("上游未清空模型列表")
		}
	}
	return nil
}

// UpdateChannelStatus 调用上游修改渠道状态。1 启用，2 手动禁用。
func UpdateChannelStatus(base, token string, channelID, status int) error {
	httpStatus, payload, err := request(http.MethodPost, base, token, fmt.Sprintf("/api/channel/%d/status", channelID), nil, map[string]int{"status": status}, 30*time.Second)
	if err != nil {
		return err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return fmt.Errorf("更新渠道状态失败 HTTP %d: %s", httpStatus, trimBody(payload))
	}
	var body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return err
	}
	if !body.Success {
		if body.Message == "" {
			body.Message = trimBody(payload)
		}
		return fmt.Errorf("更新渠道状态失败: %s", body.Message)
	}
	return nil
}

func trimBody(payload []byte) string {
	text := strings.TrimSpace(string(payload))
	if len(text) > 300 {
		return text[:300]
	}
	return text
}
