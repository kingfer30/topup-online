package controller

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
)

var csubPattern = regexp.MustCompile(`csub_[A-Za-z0-9]+`)

// CancelSubscription 用 Cursor 会话换 Grok sso，再取消当前付费订阅。
func CancelSubscription(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择要取消订阅的卡密"})
		return
	}
	if req.Category != "cursor" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "仅 cursor 卡密支持取消订阅"})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}
	cards, err := model.GetCardsByIds(tableName, req.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询卡密失败: " + err.Error()})
		return
	}
	byID := make(map[int]*model.AccountCard, len(cards))
	for _, card := range cards {
		byID[card.Id] = card
	}

	results := make([]unifyFreeResult, 0, len(req.IDs))
	okCount, failCount := 0, 0
	for _, id := range req.IDs {
		card := byID[id]
		if card == nil {
			failCount++
			results = append(results, unifyFreeResult{ID: id, Status: "failed", Message: "无 Token 或已删除"})
			continue
		}
		item := cancelCursorSubscription(card)
		if item.Status == "failed" {
			failCount++
		} else {
			okCount++
		}
		results = append(results, item)
		logger.SysLog(fmt.Sprintf("取消订阅 [ID:%d Account:%s] %s %s", card.Id, card.Account, item.Status, item.Message))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("取消完成：成功 %d，失败 %d", okCount, failCount),
		"data":    results,
	})
}

func cancelCursorSubscription(card *model.AccountCard) unifyFreeResult {
	out := unifyFreeResult{ID: card.Id, Account: card.Account, Status: "failed"}
	workosID, sessionToken, err := splitCursorSession(card.Token)
	if err != nil {
		out.Message = err.Error()
		return out
	}

	httpClient := newUnifyHTTPClient()
	setCursorSession(httpClient.Jar, workosID, sessionToken)
	bearer, err := cursorCommerceBearer(httpClient)
	if err != nil {
		out.Message = err.Error()
		return out
	}
	subs, err := listGrokSubscriptions(httpClient, bearer)
	if err != nil {
		out.Message = err.Error()
		return out
	}
	if len(subs) == 0 {
		out.Message = "没有可取消的付费订阅"
		return out
	}

	var canceled []string
	var skipped []string
	var problems []string
	for _, sub := range subs {
		label := sub.Plan
		if label == "" {
			label = sub.ID
		} else {
			label = label + " " + sub.ID
		}
		if sub.Status != grokStatusActive {
			skipped = append(skipped, label+" 状态 "+grokStatusName(sub.Status))
			continue
		}
		if sub.Platform != 0 && sub.Platform != grokPlatformStripe {
			skipped = append(skipped, label+" 渠道 "+grokPlatformName(sub.Platform))
			continue
		}
		if err := cancelGrokAcquisition(httpClient, bearer, sub.ID); err != nil {
			problems = append(problems, label+": "+err.Error())
			continue
		}
		canceled = append(canceled, label)
	}
	if len(canceled) == 0 && len(problems) == 0 {
		out.Message = "没有可取消的有效订阅：" + strings.Join(skipped, "; ")
		return out
	}
	if len(canceled) == 0 {
		out.Message = strings.Join(problems, "; ")
		return out
	}
	out.Status = "canceled"
	out.Message = "已提交取消 " + strings.Join(canceled, ",")
	if len(skipped) > 0 {
		out.Message += "；跳过 " + strings.Join(skipped, "; ")
	}
	if len(problems) > 0 {
		out.Message += "；失败 " + strings.Join(problems, "; ")
	}
	return out
}

func cursorCommerceBearer(httpClient *http.Client) (string, error) {
	status, raw, err := postCursorJSON(httpClient, "https://cursor.com/api/dashboard/client-action", "https://cursor.com/dashboard", `{"action":"xaiCommerceUnificationBearerToken","args":{"use":"billing"}}`)
	if err != nil {
		return "", fmt.Errorf("获取 commerce 凭证失败: %w", err)
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("获取 commerce 凭证失败: HTTP %d %s", status, clipText(string(raw)))
	}
	var wrap struct {
		Success     bool   `json:"success"`
		InfoMessage string `json:"infoMessage"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return "", fmt.Errorf("解析 commerce 凭证失败: %s", clipText(string(raw)))
	}
	var info struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(wrap.InfoMessage), &info); err != nil || info.Token == "" {
		return "", fmt.Errorf("commerce 凭证为空: %s", clipText(wrap.InfoMessage))
	}
	return info.Token, nil
}

const (
	grokStatusActive   = 1
	grokPlatformStripe = 1
)

type grokSub struct {
	ID       string
	Status   int
	Platform int
	Plan     string
}

func listGrokSubscriptions(httpClient *http.Client, bearer string) ([]grokSub, error) {
	status, raw, err := grokRequest(httpClient, http.MethodPost, "https://grok.com/commerce.subscription.v1.SubscriptionService/GetSubscriptions", "https://grok.com/dashboard/billing", "", bearer)
	if err != nil {
		return nil, fmt.Errorf("查询订阅失败: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("查询订阅失败: HTTP %d %s", status, clipText(string(raw)))
	}
	text := string(raw)
	if !strings.Contains(text, "grpc-status:0") && !csubPattern.MatchString(text) {
		return nil, fmt.Errorf("查询订阅失败: %s", clipText(text))
	}
	return parseGrokSubscriptions(raw), nil
}

func parseGrokSubscriptions(raw []byte) []grokSub {
	msg := firstGRPCPayload(raw)
	var subs []grokSub
	for _, field := range protoFields(msg) {
		if field.num != 1 || field.wt != 2 {
			continue
		}
		sub := grokSub{Plan: planName(field.raw)}
		var state []byte
		for _, inner := range protoFields(field.raw) {
			if inner.num == 1 && inner.wt == 2 {
				sub.ID = string(inner.raw)
			}
			if inner.num == 2 && inner.wt == 2 {
				state = inner.raw
			}
		}
		for _, st := range protoFields(state) {
			if st.num == 1 && st.wt == 0 {
				sub.Status = st.numVal
			}
			if st.num == 6 && st.wt == 0 {
				sub.Platform = st.numVal
			}
		}
		if strings.HasPrefix(sub.ID, "csub_") {
			subs = append(subs, sub)
		}
	}
	return subs
}

func planName(raw []byte) string {
	text := string(raw)
	for _, name := range []string{"Premium", "Super", "Ultra", "Plus"} {
		if strings.Contains(text, name) {
			return name
		}
	}
	return ""
}

func grokStatusName(status int) string {
	switch status {
	case 1:
		return "ACTIVE"
	case 2:
		return "GRACE"
	case 3:
		return "TERMINATED"
	case 4:
		return "EXPIRING"
	case 5:
		return "PAUSED"
	default:
		return fmt.Sprintf("%d", status)
	}
}

func grokPlatformName(platform int) string {
	switch platform {
	case 1:
		return "STRIPE"
	case 2:
		return "APPLE"
	case 3:
		return "APPLE_IAP"
	case 4:
		return "GOOGLE"
	case 7:
		return "STRIPE_LEGACY"
	default:
		return fmt.Sprintf("%d", platform)
	}
}

type protoField struct {
	num    int
	wt     int
	numVal int
	raw    []byte
}

func protoFields(buf []byte) []protoField {
	var fields []protoField
	i := 0
	for i < len(buf) {
		tag, n := readVarint(buf[i:])
		if n == 0 {
			break
		}
		i += n
		wt := tag & 7
		num := tag >> 3
		switch wt {
		case 0:
			val, m := readVarint(buf[i:])
			if m == 0 {
				return fields
			}
			i += m
			fields = append(fields, protoField{num: num, wt: wt, numVal: val})
		case 2:
			length, m := readVarint(buf[i:])
			if m == 0 || i+m+length > len(buf) {
				return fields
			}
			i += m
			fields = append(fields, protoField{num: num, wt: wt, raw: buf[i : i+length]})
			i += length
		default:
			return fields
		}
	}
	return fields
}

func readVarint(buf []byte) (int, int) {
	value := 0
	for i, b := range buf {
		if i > 4 {
			return 0, 0
		}
		value |= int(b&0x7f) << (7 * i)
		if b&0x80 == 0 {
			return value, i + 1
		}
	}
	return 0, 0
}

func firstGRPCPayload(raw []byte) []byte {
	if len(raw) < 5 || raw[0] != 0 {
		return raw
	}
	n := int(raw[1])<<24 | int(raw[2])<<16 | int(raw[3])<<8 | int(raw[4])
	if n >= 0 && 5+n <= len(raw) {
		return raw[5 : 5+n]
	}
	return raw
}

func cancelGrokAcquisition(httpClient *http.Client, bearer, subscriptionID string) error {
	payload := append(protoString(1, subscriptionID), protoString(2, newIDKey())...)
	status, raw, err := grokRequestBody(httpClient, http.MethodPost, "https://grok.com/commerce.billing.v1.AcquisitionService/CancelAcquisition", "https://grok.com/dashboard/billing", grpcFrame(payload), "application/grpc-web+proto", bearer)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("HTTP %d %s", status, clipText(string(raw)))
	}
	text := string(raw)
	if strings.Contains(text, "SUBSCRIPTION_ALREADY_CANCELED") {
		return nil
	}
	if strings.Contains(text, "SUBSCRIPTION_STORE_MANAGED") {
		return fmt.Errorf("订阅由应用商店管理，不能在这里取消")
	}
	if strings.Contains(text, "grpc-status:0") || strings.Contains(text, "effectiveAt") || strings.Contains(text, "effective_at") {
		return nil
	}
	if strings.Contains(text, "error") || strings.Contains(text, "Error") {
		return fmt.Errorf("%s", clipText(text))
	}
	return nil
}

func grokRequest(httpClient *http.Client, method, rawURL, referer, jsonBody, bearer string) (int, []byte, error) {
	contentType := ""
	var body []byte
	if jsonBody != "" {
		contentType = "application/json"
		body = []byte(jsonBody)
	} else if bearer != "" && method == http.MethodPost {
		contentType = "application/grpc-web+proto"
		body = []byte{0, 0, 0, 0, 0}
	}
	return grokRequestBody(httpClient, method, rawURL, referer, body, contentType, bearer)
}

func grokRequestBody(httpClient *http.Client, method, rawURL, referer string, body []byte, contentType, bearer string) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("origin", "https://grok.com")
	req.Header.Set("referer", referer)
	if contentType != "" {
		req.Header.Set("content-type", contentType)
	}
	if bearer != "" {
		req.Header.Set("authorization", "Bearer "+bearer)
		req.Header.Set("x-grpc-web", "1")
		req.Header.Set("x-user-agent", "connect-es/2.2.0")
		req.Header.Set("x-xai-app", "cursor")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, raw, err
}

func grpcFrame(payload []byte) []byte {
	frame := make([]byte, 5+len(payload))
	frame[0] = 0
	n := len(payload)
	frame[1] = byte(n >> 24)
	frame[2] = byte(n >> 16)
	frame[3] = byte(n >> 8)
	frame[4] = byte(n)
	copy(frame[5:], payload)
	return frame
}

func protoString(field int, value string) []byte {
	tag := byte(field<<3 | 2)
	out := []byte{tag}
	out = appendVarint(out, len(value))
	return append(out, value...)
}

func appendVarint(dst []byte, v int) []byte {
	for v >= 0x80 {
		dst = append(dst, byte(v)|0x80)
		v >>= 7
	}
	return append(dst, byte(v))
}

func newIDKey() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
