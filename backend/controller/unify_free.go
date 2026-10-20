package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

const (
	unifyAccountsReferer = "https://cursor.com/unify/accounts"
	unifyConfirmReferer  = "https://cursor.com/unify/confirm?free=1"
	unifyMergedStatus    = "INDIVIDUAL_UNIFICATION_MERGE_STATUS_MERGED"
	freeMergeClosedText  = "Cursor 已关闭免费合并，Continue with Free 不可用。这个账号需要直接订阅套餐才能完成关联"
)

type unifyFreeResult struct {
	ID      int    `json:"id"`
	Account string `json:"account"`
	Email   string `json:"email"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type unificationIntentBody struct {
	Enabled bool `json:"enabled"`
	Intent  struct {
		Step             string `json:"step"`
		IntendedXaiEmail string `json:"intendedXaiEmail"`
	} `json:"intent"`
	MergeStatus       string `json:"mergeStatus"`
	LinkedGrokAccount struct {
		Email string `json:"email"`
	} `json:"linkedGrokAccount"`
	ContinueWithFreeEligible bool `json:"continueWithFreeEligible"`
}

// BatchUnifyFree 把勾选的 Cursor 普号合并为 Grok Free，不发起付款。
func BatchUnifyFree(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择要合并的卡密"})
		return
	}
	if req.Category != "cursor" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "仅 cursor 卡密支持合并"})
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
	merged, already, failed := 0, 0, 0
	for _, id := range req.IDs {
		card := byID[id]
		if card == nil {
			failed++
			results = append(results, unifyFreeResult{
				ID: id, Status: "failed", Message: "无 Token 或已删除",
			})
			continue
		}
		item := unifyCursorAccountFree(card)
		switch item.Status {
		case "merged":
			merged++
		case "already":
			already++
		default:
			failed++
		}
		results = append(results, item)
		logger.SysLog(fmt.Sprintf("免费合并 [ID:%d Account:%s] %s %s", card.Id, card.Account, item.Status, item.Message))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("合并完成：新合并 %d，已合并 %d，失败 %d", merged, already, failed),
		"data":    results,
	})
}

func unifyCursorAccountFree(card *model.AccountCard) unifyFreeResult {
	out := unifyFreeResult{ID: card.Id, Account: card.Account, Status: "failed"}
	workosID, sessionToken, err := splitCursorSession(card.Token)
	if err != nil {
		out.Message = err.Error()
		return out
	}

	httpClient := newUnifyHTTPClient()
	setCursorSession(httpClient.Jar, workosID, sessionToken)
	email, already, err := ensureFreeUnification(httpClient)
	if err != nil {
		out.Email = email
		out.Message = err.Error()
		return out
	}
	out.Email = email
	if already {
		out.Status = "already"
		out.Message = "已经是合并后的 Free 账号"
		return out
	}
	out.Status = "merged"
	out.Message = "已合并为 Free"
	return out
}

// ensureFreeUnification 把当前 Cursor 会话合并为 Grok Free。already 为 true 表示本来就已经合并。
func ensureFreeUnification(httpClient *http.Client) (string, bool, error) {
	intent, _, raw, err := readUnificationIntent(httpClient)
	if err != nil {
		return "", false, err
	}
	if intent.MergeStatus == unifyMergedStatus {
		return intentEmail(intent), true, nil
	}

	if intentEmail(intent) == "" {
		status, provisionRaw, provisionErr := postCursorJSON(httpClient, "https://cursor.com/api/dashboard/provision-unify-grok-account", unifyAccountsReferer, "{}")
		if provisionErr != nil {
			return "", false, fmt.Errorf("创建 Grok 账号失败: %w", provisionErr)
		}
		if status != http.StatusOK {
			return "", false, fmt.Errorf("创建 Grok 账号失败: HTTP %d %s", status, clipText(string(provisionRaw)))
		}
		intent, _, raw, err = readUnificationIntent(httpClient)
		if err != nil {
			return "", false, err
		}
		if intent.MergeStatus == unifyMergedStatus {
			return intentEmail(intent), true, nil
		}
	}

	email := intentEmail(intent)
	if email == "" {
		return "", false, fmt.Errorf("未得到 Grok 邮箱: %s", clipText(string(raw)))
	}
	if !intent.ContinueWithFreeEligible {
		return email, false, fmt.Errorf("%s（状态 %s %s）", freeMergeClosedText, intent.MergeStatus, intent.Intent.Step)
	}

	status, raw, err := postCursorJSON(httpClient, "https://cursor.com/api/dashboard/finalize-individual-unification", unifyConfirmReferer, `{"continueWithFree":true}`)
	if err != nil {
		return email, false, fmt.Errorf("确认合并失败: %w", err)
	}
	if status != http.StatusOK {
		if freeMergeClosedBody(raw) {
			return email, false, fmt.Errorf("%s", freeMergeClosedText)
		}
		return email, false, fmt.Errorf("确认合并失败: HTTP %d %s", status, clipText(string(raw)))
	}
	var finalized struct {
		MergeStatus string `json:"mergeStatus"`
	}
	_ = json.Unmarshal(raw, &finalized)
	if finalized.MergeStatus != unifyMergedStatus {
		return email, false, fmt.Errorf("确认合并未完成: %s", clipText(string(raw)))
	}

	dashURL := "https://cursor.com/api/dashboard/open-grok-dashboard?to=" + url.QueryEscape("/dashboard?welcome=true")
	_, _, _ = cursorRequest(httpClient, http.MethodGet, dashURL, "https://cursor.com/unify/confirm?free=1&merging=1", "")
	return email, false, nil
}

func readUnificationIntent(httpClient *http.Client) (unificationIntentBody, int, []byte, error) {
	status, raw, err := postCursorJSON(httpClient, "https://cursor.com/api/dashboard/get-unification-intent", unifyAccountsReferer, "{}")
	if err != nil {
		return unificationIntentBody{}, status, raw, fmt.Errorf("查询合并状态失败: %w", err)
	}
	if status != http.StatusOK {
		return unificationIntentBody{}, status, raw, fmt.Errorf("查询合并状态失败: HTTP %d %s", status, clipText(string(raw)))
	}
	var intent unificationIntentBody
	if err := json.Unmarshal(raw, &intent); err != nil {
		return unificationIntentBody{}, status, raw, fmt.Errorf("解析合并状态失败: %s", clipText(string(raw)))
	}
	return intent, status, raw, nil
}

func freeMergeClosedBody(raw []byte) bool {
	text := strings.ToLower(string(raw))
	return strings.Contains(text, "no longer available") || strings.Contains(text, "free_unification") || strings.Contains(text, "\"error_code\":\"free_")
}

func intentEmail(intent unificationIntentBody) string {
	if email := strings.TrimSpace(intent.LinkedGrokAccount.Email); email != "" {
		return email
	}
	return strings.TrimSpace(intent.Intent.IntendedXaiEmail)
}

func postCursorJSON(httpClient *http.Client, rawURL, referer, body string) (int, []byte, error) {
	return cursorRequest(httpClient, http.MethodPost, rawURL, referer, body)
}

func cursorRequest(httpClient *http.Client, method, rawURL, referer, body string) (int, []byte, error) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("origin", "https://cursor.com")
	req.Header.Set("referer", referer)
	if body != "" {
		req.Header.Set("content-type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, raw, nil
}

func newUnifyHTTPClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	var base http.RoundTripper
	if client.HTTPClient != nil && client.HTTPClient.Transport != nil {
		base = client.HTTPClient.Transport
	}
	profiled, _ := withBrowserProfile(base)
	return &http.Client{Jar: jar, Timeout: 40 * time.Second, Transport: profiled}
}

func setCursorSession(jar http.CookieJar, workosID, sessionToken string) {
	cursorURL, _ := url.Parse("https://cursor.com")
	jar.SetCookies(cursorURL, []*http.Cookie{
		{Name: "workos_id", Value: workosID, Path: "/"},
		{Name: "WorkosCursorSessionToken", Value: sessionToken, Path: "/"},
	})
}

func splitCursorSession(token string) (string, string, error) {
	token = strings.TrimSpace(token)
	if decoded, err := url.QueryUnescape(token); err == nil {
		token = decoded
	}
	parts := strings.SplitN(token, "::", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("token 格式错误，应为 user_xxx::JWT")
	}
	return parts[0], token, nil
}

func clipText(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 240 {
		return s[:240]
	}
	return s
}
