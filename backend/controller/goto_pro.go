package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

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

	// 拆分 token：格式为 user_xxx::JWT
	parts := strings.SplitN(req.Token, "::", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "token 格式错误，应为 user_xxx::JWT",
		})
		return
	}
	workosID := parts[0]
	sessionToken := req.Token

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
