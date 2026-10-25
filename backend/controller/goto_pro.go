package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

const (
	cursorStripePaymentUA = "stripe.js/c264a67020; stripe-js-v3/c264a67020; checkout"
	// 固定拒付卡，只用来把 Checkout confirm 成 open 账单，把账号钉回老 Stripe 通道。
	cursorPinCardNumber = "4532015112830366"
	cursorPinCardMonth  = "12"
	cursorPinCardYear   = "2029"
	cursorPinCardCVC    = "123"
)

// GotoProRequest 提链请求
type GotoProRequest struct {
	Token            string `json:"token" binding:"required"`
	SubscriptionType string `json:"subscription_type" binding:"required"`
}

// GotoPro 先用 premium 年付拒付卡钉回老通道，再创建所选档位的月付 Checkout 链接。
func GotoPro(c *gin.Context) {
	var req GotoProRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	link, email, err := createMonthlyCheckoutLink(req.Token, req.SubscriptionType)
	if err != nil {
		status := http.StatusInternalServerError
		code := 500
		if strings.Contains(err.Error(), "token 格式") || strings.Contains(err.Error(), "不支持的订阅类型") || strings.Contains(err.Error(), "不能走强开提链") {
			status = http.StatusBadRequest
			code = 400
		}
		c.JSON(status, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    link,
		"email":   email,
	})
}

func createMonthlyCheckoutLink(rawToken, subscriptionType string) (string, string, error) {
	token, err := normalizeCursorCookieToken(rawToken)
	if err != nil {
		return "", "", err
	}
	tier, err := monthlyCheckoutTier(subscriptionType)
	if err != nil {
		return "", "", err
	}
	if err := rejectMergedAccount(rawToken); err != nil {
		return "", "", err
	}

	httpClient := cursorAuthHTTPClient(40*time.Second, true)
	email := cursorAccountEmail(httpClient, token)
	accountEmail := email
	if accountEmail == "user@example.com" {
		accountEmail = ""
	}
	if err := pinOldStripeChannel(httpClient, token, email); err != nil {
		return "", "", err
	}
	link, err := cursorCheckoutLink(httpClient, token, tier, false)
	if err != nil {
		return "", "", err
	}
	if strings.Contains(link, "stripe.com") {
		return link, accountEmail, nil
	}
	if strings.Contains(strings.ToLower(link), "unify") {
		return "", "", fmt.Errorf("账号没钉住老通道，结账走到了 unify")
	}
	return "", "", fmt.Errorf("未拿到 Stripe 结账链接: %s", clipGotoText(link))
}

func rejectMergedAccount(rawToken string) error {
	workosID, sessionToken, err := splitCursorSession(rawToken)
	if err != nil {
		return err
	}
	httpClient := newUnifyHTTPClient()
	setCursorSession(httpClient.Jar, workosID, sessionToken)
	intent, _, _, err := readUnificationIntent(httpClient)
	if err != nil {
		return err
	}
	if intent.MergeStatus == unifyMergedStatus {
		return fmt.Errorf("账号合并过了，不能走强开提链")
	}
	return nil
}

func monthlyCheckoutTier(subscriptionType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(subscriptionType)) {
	case "pro":
		return "pro", nil
	case "pro+", "pro_plus", "proplus":
		return "pro_plus", nil
	case "ultra":
		return "ultra", nil
	default:
		return "", fmt.Errorf("不支持的订阅类型: %s", strings.TrimSpace(subscriptionType))
	}
}

func normalizeCursorCookieToken(raw string) (string, error) {
	token := strings.TrimSpace(raw)
	token = strings.Trim(token, `"'`)
	if strings.Contains(token, "WorkosCursorSessionToken=") {
		token = strings.SplitN(token, "WorkosCursorSessionToken=", 2)[1]
		token = strings.SplitN(token, ";", 2)[0]
		token = strings.TrimSpace(token)
	}
	decoded := token
	if unescaped, err := url.QueryUnescape(token); err == nil && unescaped != "" {
		decoded = unescaped
	}
	if !strings.HasPrefix(decoded, "user_") || !strings.Contains(decoded, "::") {
		return "", fmt.Errorf("token 格式错误，应为 user_xxx::JWT")
	}
	if strings.Contains(strings.ToLower(token), "%3a") {
		return token, nil
	}
	return url.QueryEscape(decoded), nil
}

func cursorAccountEmail(httpClient *http.Client, token string) string {
	status, body, err := cursorJSON(httpClient, token, "/api/auth/me", map[string]any{})
	if err != nil || status != http.StatusOK {
		return "user@example.com"
	}
	var me struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &me); err != nil || strings.TrimSpace(me.Email) == "" {
		return "user@example.com"
	}
	return strings.TrimSpace(me.Email)
}

func pinOldStripeChannel(httpClient *http.Client, token, email string) error {
	link, err := cursorCheckoutLink(httpClient, token, "premium", true)
	if err != nil {
		return fmt.Errorf("钉老通道失败: %w", err)
	}
	if !strings.Contains(link, "stripe.com") {
		return fmt.Errorf("钉老通道失败，premium 年付没有返回 Stripe 链接: %s", clipGotoText(link))
	}
	sessionID, err := checkoutSessionID(link)
	if err != nil {
		return fmt.Errorf("钉老通道失败: %w", err)
	}
	if err := confirmCheckoutByDeclinedCard(httpClient, sessionID, email); err != nil {
		return fmt.Errorf("钉老通道失败: %w", err)
	}
	time.Sleep(2 * time.Second)
	status, body, err := cursorJSON(httpClient, token, "/api/dashboard/list-invoices", map[string]any{})
	if err != nil {
		return fmt.Errorf("钉老通道失败，查询账单出错: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("钉老通道失败，查询账单 HTTP %d: %s", status, clipGotoText(string(body)))
	}
	var listed struct {
		Invoices []struct {
			Status string `json:"status"`
		} `json:"invoices"`
	}
	if err := json.Unmarshal(body, &listed); err != nil {
		return fmt.Errorf("钉老通道失败，账单响应无法解析")
	}
	for _, invoice := range listed.Invoices {
		if invoice.Status == "open" {
			return nil
		}
	}
	return fmt.Errorf("钉老通道失败，没有 open 账单")
}

func cursorCheckoutLink(httpClient *http.Client, token, tier string, yearly bool) (string, error) {
	status, body, err := cursorJSON(httpClient, token, "/api/checkout", map[string]any{
		"tier":                  tier,
		"yearly":                yearly,
		"allowAutomaticPayment": false,
	})
	if err != nil {
		return "", err
	}
	link := parseCheckoutURL(body)
	if status != http.StatusOK {
		return "", fmt.Errorf("checkout HTTP %d: %s", status, clipGotoText(link))
	}
	if link == "" {
		return "", fmt.Errorf("checkout 响应为空")
	}
	return link, nil
}

func parseCheckoutURL(body []byte) string {
	var asString string
	if err := json.Unmarshal(body, &asString); err == nil {
		return strings.TrimSpace(asString)
	}
	var asObject struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &asObject); err == nil && strings.TrimSpace(asObject.URL) != "" {
		return strings.TrimSpace(asObject.URL)
	}
	return strings.TrimSpace(string(body))
}

func checkoutSessionID(checkoutURL string) (string, error) {
	const marker = "/pay/"
	index := strings.Index(checkoutURL, marker)
	if index < 0 {
		return "", fmt.Errorf("不是 Stripe Checkout 链接")
	}
	sessionID := checkoutURL[index+len(marker):]
	if cut := strings.IndexAny(sessionID, "#?"); cut >= 0 {
		sessionID = sessionID[:cut]
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", fmt.Errorf("无法解析 Checkout Session")
	}
	return sessionID, nil
}

func confirmCheckoutByDeclinedCard(httpClient *http.Client, sessionID, email string) error {
	status, body, err := pinStripeFormPost(httpClient, "https://api.stripe.com/v1/payment_pages/"+url.PathEscape(sessionID)+"/init", url.Values{
		"browser_locale": {"en-US"},
	})
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("init 失败: HTTP %d %s", status, clipGotoText(string(body)))
	}
	var page struct {
		Eid          string `json:"eid"`
		TotalSummary struct {
			Total json.Number `json:"total"`
		} `json:"total_summary"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		return fmt.Errorf("init 响应无法解析")
	}
	amount := page.TotalSummary.Total.String()
	if page.Eid == "" || amount == "" {
		return fmt.Errorf("init 响应缺少 eid 或金额")
	}

	pmStatus, pmBody, err := pinStripeFormPost(httpClient, "https://api.stripe.com/v1/payment_methods", url.Values{
		"type":                                  {"card"},
		"card[number]":                          {cursorPinCardNumber},
		"card[exp_month]":                       {cursorPinCardMonth},
		"card[exp_year]":                        {cursorPinCardYear},
		"card[cvc]":                             {cursorPinCardCVC},
		"billing_details[name]":                 {"Cursor User"},
		"billing_details[email]":                {email},
		"billing_details[address][country]":     {"CN"},
		"billing_details[address][postal_code]": {"100000"},
		"billing_details[address][city]":        {"Beijing"},
		"billing_details[address][state]":       {"Beijing"},
		"billing_details[address][line1]":       {"1 Main St"},
		"payment_user_agent":                    {cursorStripePaymentUA},
	})
	if err != nil {
		return err
	}
	if pmStatus != http.StatusOK {
		return fmt.Errorf("创建 Card PaymentMethod 失败: HTTP %d %s", pmStatus, clipGotoText(string(pmBody)))
	}
	var pm struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(pmBody, &pm); err != nil || pm.ID == "" {
		return fmt.Errorf("创建 Card PaymentMethod 失败")
	}

	confirmStatus, confirmBody, err := pinStripeFormPost(httpClient, "https://api.stripe.com/v1/payment_pages/"+url.PathEscape(sessionID)+"/confirm", url.Values{
		"eid":                          {page.Eid},
		"payment_method":               {pm.ID},
		"expected_amount":              {amount},
		"expected_payment_method_type": {"card"},
		"return_url":                   {"https://cursor.com/dashboard"},
	})
	if err != nil {
		return err
	}
	if confirmStatus != http.StatusOK && confirmStatus != http.StatusPaymentRequired {
		return fmt.Errorf("confirm 失败: HTTP %d %s", confirmStatus, clipGotoText(string(confirmBody)))
	}
	return nil
}

func cursorJSON(httpClient *http.Client, token, path string, body any) (int, []byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://cursor.com"+path, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Cookie", "WorkosCursorSessionToken="+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://cursor.com")
	req.Header.Set("Referer", "https://cursor.com/dashboard")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	return doGotoRequest(httpClient, req)
}

func pinStripeFormPost(httpClient *http.Client, endpoint string, fields url.Values) (int, []byte, error) {
	form := url.Values{}
	for key, values := range fields {
		for _, value := range values {
			form.Add(key, value)
		}
	}
	if form.Get("key") == "" {
		form.Set("key", cursorStripePK)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://checkout.stripe.com")
	req.Header.Set("Referer", "https://checkout.stripe.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	return doGotoRequest(httpClient, req)
}

func doGotoRequest(httpClient *http.Client, req *http.Request) (int, []byte, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

func clipGotoText(text string) string {
	text = strings.TrimSpace(text)
	if len(text) > 180 {
		return text[:180]
	}
	return text
}

type HalfPriceCheckoutRequest struct {
	Token string `json:"token" binding:"required"`
}

type welcomeBackOfferResp struct {
	CanActivate      bool   `json:"canActivate"`
	PromoTypeID      string `json:"promoTypeId"`
	PreviousPlanTier string `json:"previousPlanTier"`
	CouponID         string `json:"couponId"`
}

func parseWorkosSessionToken(raw string) (workosID, sessionToken string, err error) {
	token := strings.TrimSpace(raw)
	if decoded, uerr := url.QueryUnescape(token); uerr == nil && decoded != "" {
		token = decoded
	}
	parts := strings.SplitN(token, "::", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("token 格式错误，应为 user_xxx::JWT")
	}
	return strings.TrimSpace(parts[0]), token, nil
}

func cursorAuthHTTPClient(timeout time.Duration, followRedirect bool) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if proxyURL := model.GetCursorPayProxyURL(); proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	} else if client.HTTPClient != nil {
		if t, ok := client.HTTPClient.Transport.(*http.Transport); ok && t != nil && t.Proxy != nil {
			transport.Proxy = t.Proxy
		}
	}
	c := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	if !followRedirect {
		c.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return c
}

func setCursorWelcomeBackHeaders(req *http.Request, workosID, sessionToken string, htmlNav bool) {
	encodedToken := strings.ReplaceAll(sessionToken, "::", "%3A%3A")
	req.Header.Set("cookie", fmt.Sprintf("workos_id=%s; WorkosCursorSessionToken=%s", workosID, encodedToken))
	req.Header.Set("accept-language", "en")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("origin", "https://cursor.com")
	req.Header.Set("referer", "https://cursor.com/activate/welcome-back/offer")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36")
	req.Header.Set("sec-fetch-site", "same-origin")
	if htmlNav {
		req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		req.Header.Set("sec-fetch-dest", "document")
		req.Header.Set("sec-fetch-mode", "navigate")
		req.Header.Set("sec-fetch-user", "?1")
		req.Header.Set("upgrade-insecure-requests", "1")
	} else {
		req.Header.Set("accept", "*/*")
		req.Header.Set("sec-fetch-dest", "empty")
		req.Header.Set("sec-fetch-mode", "cors")
	}
}

func fetchWelcomeBackOffer(httpClient *http.Client, workosID, sessionToken string) (*welcomeBackOfferResp, error) {
	req, err := http.NewRequest(http.MethodGet, "https://cursor.com/api/auth/welcome-back-offer", nil)
	if err != nil {
		return nil, err
	}
	setCursorWelcomeBackHeaders(req, workosID, sessionToken, false)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview := strings.TrimSpace(string(body))
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("offer 接口 HTTP %d: %s", resp.StatusCode, preview)
	}
	var offer welcomeBackOfferResp
	if err := json.Unmarshal(body, &offer); err != nil {
		return nil, fmt.Errorf("解析 offer 响应失败")
	}
	return &offer, nil
}

func fetchWelcomeBackCheckoutLink(httpClient *http.Client, workosID, sessionToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://cursor.com/api/auth/welcome-back?checkout=1", nil)
	if err != nil {
		return "", err
	}
	setCursorWelcomeBackHeaders(req, workosID, sessionToken, true)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	link := strings.TrimSpace(resp.Header.Get("Location"))
	if link == "" {
		link = strings.TrimSpace(string(body))
		var jsonStr string
		if err := json.Unmarshal(body, &jsonStr); err == nil {
			link = strings.TrimSpace(jsonStr)
		}
	}
	if strings.HasPrefix(link, "/") {
		link = "https://cursor.com" + link
	}
	if !strings.Contains(link, "checkout.stripe.com") {
		preview := strings.TrimSpace(string(body))
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return "", fmt.Errorf("未跳转到 Stripe（HTTP %d）: %s", resp.StatusCode, preview)
	}
	return link, nil
}

// HalfPriceCheckout 通过官方 welcome-back 接口提半价链，302 到 Stripe 结账页
func HalfPriceCheckout(c *gin.Context) {
	var req HalfPriceCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	workosID, sessionToken, err := parseWorkosSessionToken(req.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	offerClient := cursorAuthHTTPClient(30*time.Second, true)
	offer, offerErr := fetchWelcomeBackOffer(offerClient, workosID, sessionToken)
	if offerErr != nil {
		logger.SysLog("半价提链 offer 预检失败，继续尝试 checkout: " + offerErr.Error())
	} else if !offer.CanActivate {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "当前账号无法激活半价优惠"})
		return
	}

	checkoutClient := cursorAuthHTTPClient(45*time.Second, false)
	link, err := fetchWelcomeBackCheckoutLink(checkoutClient, workosID, sessionToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "官方半价提链失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    link,
	})
}
