package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GotoProRequest 提链请求
type GotoProRequest struct {
	Token            string `json:"token" binding:"required"`
	SubscriptionType string `json:"subscription_type" binding:"required"`
}

// GotoPro 仿照 goto-pro.py 发起完整请求链，获取 Cursor Pro 订阅付款链接
func GotoPro(c *gin.Context) {
	var req GotoProRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 拆分 token：格式为 user_xxx::JWT（分隔符可能被 URL 编码为 %3A%3A）
	token := req.Token
	if decoded, err := url.QueryUnescape(token); err == nil {
		token = decoded
	}
	parts := strings.SplitN(token, "::", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "token 格式错误，应为 user_xxx::JWT",
		})
		return
	}
	workosID := parts[0]
	sessionToken := token

	// 创建带 cookie jar 的 HTTP 客户端
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 通用请求头
	commonHeaders := map[string]string{
		"accept-language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6",
		"cache-control":   "no-cache",
		"pragma":          "no-cache",
		"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
		"sec-ch-ua":       `"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
	}

	// 预热请求列表
	warmupRequests := []struct {
		url     string
		headers map[string]string
	}{
		{
			url: "https://js.stripe.com/basil/stripe.js",
			headers: map[string]string{
				"accept":               "*/*",
				"referer":              "https://cursor.com/",
				"sec-fetch-dest":       "script",
				"sec-fetch-mode":       "no-cors",
				"sec-fetch-site":       "cross-site",
				"sec-fetch-storage-access": "none",
			},
		},
		{
			url: "https://js.stripe.com/v3/controller-with-preconnect-562954400a1d5ed269a33dd401040ace.html",
			headers: map[string]string{
				"accept":               "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"referer":              "https://cursor.com/",
				"priority":             "u=0, i",
				"sec-fetch-dest":       "iframe",
				"sec-fetch-mode":       "navigate",
				"sec-fetch-site":       "cross-site",
				"sec-fetch-storage-access": "none",
				"sec-fetch-user":       "?1",
				"upgrade-insecure-requests": "1",
			},
		},
		{
			url: "https://js.stripe.com/v3/m-outer-3437aaddcdf6922d623e172c2d6f9278.html",
			headers: map[string]string{
				"accept":               "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"referer":              "https://cursor.com/",
				"priority":             "u=0, i",
				"sec-fetch-dest":       "iframe",
				"sec-fetch-mode":       "navigate",
				"sec-fetch-site":       "cross-site",
				"sec-fetch-storage-access": "none",
				"sec-fetch-user":       "?1",
				"upgrade-insecure-requests": "1",
			},
		},
		{
			url: "https://js.stripe.com/v3/fingerprinted/js/shared-9a246c2de00bc656633679d45127b572.js",
			headers: map[string]string{
				"accept":               "*/*",
				"referer":              "https://js.stripe.com/v3/controller-with-preconnect-562954400a1d5ed269a33dd401040ace.html",
				"sec-fetch-dest":       "script",
				"sec-fetch-mode":       "no-cors",
				"sec-fetch-site":       "same-origin",
				"sec-fetch-storage-access": "none",
			},
		},
		{
			url: "https://js.stripe.com/v3/fingerprinted/js/controller-with-preconnect-01e2ee01eaf8224dfc7590fb8fc67129.js",
			headers: map[string]string{
				"accept":               "*/*",
				"referer":              "https://js.stripe.com/v3/controller-with-preconnect-562954400a1d5ed269a33dd401040ace.html",
				"sec-fetch-dest":       "script",
				"sec-fetch-mode":       "no-cors",
				"sec-fetch-site":       "same-origin",
				"sec-fetch-storage-access": "none",
			},
		},
		{
			url: "https://js.stripe.com/v3/fingerprinted/js/m-outer-15a2b40a058ddff1cffdb63779fe3de1.js",
			headers: map[string]string{
				"accept":               "*/*",
				"referer":              "https://js.stripe.com/v3/m-outer-3437aaddcdf6922d623e172c2d6f9278.html",
				"sec-fetch-dest":       "script",
				"sec-fetch-mode":       "no-cors",
				"sec-fetch-site":       "same-origin",
				"sec-fetch-storage-access": "none",
			},
		},
	}

	// 依次执行预热请求（忽略响应内容，只建立连接/cookie）
	for _, req := range warmupRequests {
		httpReq, err := http.NewRequest("GET", req.url, nil)
		if err != nil {
			continue
		}
		for k, v := range commonHeaders {
			httpReq.Header.Set(k, v)
		}
		for k, v := range req.headers {
			httpReq.Header.Set(k, v)
		}
		resp, err := client.Do(httpReq)
		if err == nil {
			resp.Body.Close()
		}
	}

	// 设置 cursor.com 的 cookies
	cursorURL, _ := url.Parse("https://cursor.com")
	jar.SetCookies(cursorURL, []*http.Cookie{
		{Name: "workos_id", Value: workosID},
		{Name: "WorkosCursorSessionToken", Value: sessionToken},
	})

	// 发起最终的 checkout 请求，tier 使用前端传入的订阅类型
	checkoutBody, _ := json.Marshal(map[string]interface{}{
		"tier":                  req.SubscriptionType,
		"allowAutomaticPayment": true,
		"yearly":                false,
	})

	checkoutReq, err := http.NewRequest("POST", "https://cursor.com/api/checkout", bytes.NewReader(checkoutBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "构建请求失败: " + err.Error(),
		})
		return
	}

	for k, v := range commonHeaders {
		checkoutReq.Header.Set(k, v)
	}
	checkoutReq.Header.Set("accept", "*/*")
	checkoutReq.Header.Set("content-type", "application/json")
	checkoutReq.Header.Set("origin", "https://cursor.com")
	checkoutReq.Header.Set("priority", "u=1, i")
	checkoutReq.Header.Set("referer", "https://cursor.com/dashboard?tab=billing")
	checkoutReq.Header.Set("sec-fetch-dest", "empty")
	checkoutReq.Header.Set("sec-fetch-mode", "cors")
	checkoutReq.Header.Set("sec-fetch-site", "same-origin")
	checkoutReq.Header.Set("sec-ch-ua-arch", `"x86"`)
	checkoutReq.Header.Set("sec-ch-ua-bitness", `"64"`)
	checkoutReq.Header.Set("sec-ch-ua-platform-version", `"19.0.0"`)

	resp, err := client.Do(checkoutReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "请求 cursor.com 失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取响应失败: " + err.Error(),
		})
		return
	}

	// 响应可能是 JSON 字符串（带引号），尝试反序列化；失败则直接使用原始文本
	link := strings.TrimSpace(string(bodyBytes))
	var jsonStr string
	if err := json.Unmarshal(bodyBytes, &jsonStr); err == nil {
		link = jsonStr
	}

	if link == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "未获取到付款链接，响应为空",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    link,
	})
}

const halfPriceCheckoutBase = "https://cursor.120.hk"

type HalfPriceCheckoutRequest struct {
	UID   string `json:"uid" binding:"required"`
	Token string `json:"token" binding:"required"`
	Tier  string `json:"tier" binding:"required"`
}

type halfPriceAPIResp struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	URL      string `json:"url"`
	Email    string `json:"email"`
	BillUsed int    `json:"bill_used"`
	BillMax  int    `json:"bill_max"`
}

func normalizeWorkosCookie(raw string) string {
	token := strings.TrimSpace(raw)
	if decoded, err := url.QueryUnescape(token); err == nil && decoded != "" {
		token = decoded
	}
	const key = "WorkosCursorSessionToken="
	if idx := strings.Index(token, key); idx >= 0 {
		token = token[idx+len(key):]
		if i := strings.Index(token, ";"); i >= 0 {
			token = token[:i]
		}
	}
	token = strings.Trim(token, "\"' \t")
	return key + token
}

func postHalfPriceForm(client *http.Client, apiPath string, fields map[string]string) (*halfPriceAPIResp, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, halfPriceCheckoutBase+apiPath, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", halfPriceCheckoutBase)
	req.Header.Set("Referer", halfPriceCheckoutBase+"/?uid="+url.QueryEscape(fields["uid"]))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data halfPriceAPIResp
	if err := json.Unmarshal(body, &data); err != nil {
		preview := strings.TrimSpace(string(body))
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("解析活动接口响应失败: %s", preview)
	}
	return &data, nil
}

// HalfPriceCheckout 走半价活动页预检 + 开单，返回官方支付链接
func HalfPriceCheckout(c *gin.Context) {
	var req HalfPriceCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	uid := strings.TrimSpace(req.UID)
	tier := strings.TrimSpace(req.Tier)
	if uid == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请输入 uid"})
		return
	}
	switch tier {
	case "pro", "pro_plus", "ultra":
	default:
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "套餐仅支持 pro / pro_plus / ultra"})
		return
	}

	cookie := normalizeWorkosCookie(req.Token)
	if strings.TrimPrefix(cookie, "WorkosCursorSessionToken=") == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "token 为空或格式不正确"})
		return
	}

	client := &http.Client{Timeout: 90 * time.Second}

	pre, err := postHalfPriceForm(client, "/api/precheck", map[string]string{
		"uid":    uid,
		"cookie": cookie,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "预检请求失败: " + err.Error()})
		return
	}
	switch pre.Status {
	case "blocked":
		email := strings.TrimSpace(pre.Email)
		msg := "该号无资格"
		if email != "" {
			msg = email + " 已被标记为无资格账号，请更换账号后再试"
		}
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": msg})
		return
	case "limit":
		email := strings.TrimSpace(pre.Email)
		msg := fmt.Sprintf("邮箱开单已达上限 %d/%d，请更换账号", pre.BillUsed, pre.BillMax)
		if email != "" {
			msg = fmt.Sprintf("%s 已开单 %d/%d 次，请更换账号", email, pre.BillUsed, pre.BillMax)
		}
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": msg})
		return
	case "success":
	default:
		msg := strings.TrimSpace(pre.Message)
		if msg == "" {
			msg = "预检未通过，请稍后再试"
		}
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": msg})
		return
	}

	result, err := postHalfPriceForm(client, "/api/checkout", map[string]string{
		"uid":    uid,
		"cookie": cookie,
		"tier":   tier,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "开单请求失败: " + err.Error()})
		return
	}
	if result.Status != "success" {
		msg := strings.TrimSpace(result.Message)
		if msg == "" {
			msg = "申请暂未通过，请稍后再试"
		}
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": msg})
		return
	}
	link := strings.TrimSpace(result.URL)
	if link == "" {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "后台已成功但未返回支付地址"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    link,
	})
}

var (
	halfPriceNavPillRe  = regexp.MustCompile(`(?is)class=["']nav-pill["'][^>]*>([^<]+)`)
	halfPriceQuotaNumRe = regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)
)

func extractHalfPriceQuota(html string) string {
	text := html
	if m := halfPriceNavPillRe.FindStringSubmatch(html); len(m) >= 2 {
		text = m[1]
	}
	if n := halfPriceQuotaNumRe.FindStringSubmatch(text); len(n) >= 3 {
		return n[1] + "/" + n[2]
	}
	return ""
}

func fetchHalfPriceQuota(uid string) string {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return ""
	}
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, halfPriceCheckoutBase+"/?uid="+url.QueryEscape(uid), nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return extractHalfPriceQuota(string(body))
}

// GetHalfPriceQuota 抓取活动页 nav-pill 中的已提交 xx/xx
func GetHalfPriceQuota(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请输入 uid"})
		return
	}
	quota := fetchHalfPriceQuota(uid)
	if quota == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "未能读取活动余量"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": quota})
}
