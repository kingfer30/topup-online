package controller

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

const (
	stripeCheckoutOrigin   = "https://checkout.stripe.com"
	stripeAPIBase          = "https://api.stripe.com/v1/payment_pages/"
	stripePaymentMethodAPI = "https://api.stripe.com/v1/payment_methods"
	cursorStripePK         = "pk_live_51Lb5LzB4TZWxSIGU4LcaRyvT5xW1Iw8Z3E1iOpuCblBLoLhoq3xQnt2U6sR0kfr6wwTdLdQCykfzNnw778PaO7n200tsRmVe72"
)

var (
	stripeSessionIDRe   = regexp.MustCompile(`cs_(?:live|test)_[A-Za-z0-9]+`)
	stripePublishableRe = regexp.MustCompile(`pk_(?:live|test)_[A-Za-z0-9]+`)
)

// StripeAlipayRequest 用 Stripe Checkout 链接自动提交 Alipay
type StripeAlipayRequest struct {
	CheckoutURL string `json:"checkout_url" binding:"required"`
	Name        string `json:"name"`
	PostalCode  string `json:"postal_code"`
	State       string `json:"state"`
	City        string `json:"city"`
	Line1       string `json:"line1"`
	Country     string `json:"country"`
}

// StripeAlipayResult 返回支付宝跳转页
type StripeAlipayResult struct {
	AlipayURL       string `json:"alipay_url"`
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
	Email           string `json:"email"`
	PaymentIntentID string `json:"payment_intent_id"`
	SessionID       string `json:"session_id"`
}

// SubmitStripeAlipay 切 USD、填账单地址、用 Alipay confirm，返回付款页
func SubmitStripeAlipay(c *gin.Context) {
	var req StripeAlipayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	sessionID := stripeSessionIDRe.FindString(req.CheckoutURL)
	if sessionID == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "未识别到 Stripe Checkout Session ID"})
		return
	}

	billing := stripeBillingFromReq(req)
	httpClient := stripeHTTPClient()

	keys := collectStripePublishableKeys(httpClient, req.CheckoutURL, sessionID)
	session, pk, err := stripeInitWithKeys(httpClient, sessionID, keys)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "初始化结账会话失败: " + err.Error()})
		return
	}
	logger.SysLog("[stripe-alipay] init ok session=" + sessionID + " pk=" + stripePKPreview(pk) + " geo=" + stripeGeoCountry(session) + " currency=" + stripePresentmentCurrency(session) + " pms=" + strings.Join(stripeAvailablePaymentMethods(session), ","))

	if stripePresentmentCurrency(session) != "usd" {
		if err := stripeSwitchCurrency(httpClient, pk, sessionID, "usd"); err != nil {
			logger.SysLog("[stripe-alipay] 切换 USD 失败: " + err.Error())
		}
		session, err = stripeInitPaymentPage(httpClient, pk, sessionID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "切换币种后重新初始化失败: " + err.Error()})
			return
		}
	}

	currency := stripePresentmentCurrency(session)
	if currency != "usd" {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "未能切换到 USD，当前币种: " + currency})
		return
	}

	amount := stripeExpectedAmount(session)
	if amount <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "未能读取应付金额"})
		return
	}

	methods := stripeAvailablePaymentMethods(session)
	if !stripeHasPaymentMethod(methods, "alipay") {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "当前结账页没有 Alipay（geo=" + stripeGeoCountry(session) + " 可用: " + strings.Join(methods, ", ") + "），无法自动提交",
		})
		return
	}

	email := stripeMapString(session, "customer_email")
	confirmed := session
	if stripeAlipayRedirectURL(confirmed) == "" {
		var confirmErr error
		confirmed, confirmErr = stripeConfirmAlipay(httpClient, pk, sessionID, amount, currency, email, billing)
		if confirmErr != nil {
			if retrySession, retryErr := stripeInitPaymentPage(httpClient, pk, sessionID); retryErr == nil && stripeAlipayRedirectURL(retrySession) != "" {
				logger.SysLog("[stripe-alipay] confirm 报错但会话已进入 Alipay 跳转: " + confirmErr.Error())
				confirmed = retrySession
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 500, "message": "提交 Alipay 失败: " + confirmErr.Error()})
				return
			}
		}
	} else {
		logger.SysLog("[stripe-alipay] 会话已有 Alipay 跳转，跳过重复 confirm")
	}

	alipayURL := stripeAlipayRedirectURL(confirmed)
	if alipayURL == "" {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "已提交但未返回支付宝付款地址"})
		return
	}

	pi := stripeMap(confirmed["payment_intent"])
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": StripeAlipayResult{
			AlipayURL:       alipayURL,
			Amount:          amount,
			Currency:        currency,
			Email:           email,
			PaymentIntentID: stripeMapString(pi, "id"),
			SessionID:       sessionID,
		},
	})
}

type stripeBilling struct {
	Name       string
	PostalCode string
	State      string
	City       string
	Line1      string
	Country    string
}

func stripeBillingFromReq(req StripeAlipayRequest) stripeBilling {
	cfg := model.GetCursorPayBilling()
	b := stripeBilling{
		Name:       strings.TrimSpace(req.Name),
		PostalCode: strings.TrimSpace(req.PostalCode),
		State:      strings.TrimSpace(req.State),
		City:       strings.TrimSpace(req.City),
		Line1:      strings.TrimSpace(req.Line1),
		Country:    strings.ToUpper(strings.TrimSpace(req.Country)),
	}
	if b.Name == "" {
		b.Name = cfg.Name
	}
	if b.PostalCode == "" {
		b.PostalCode = cfg.PostalCode
	}
	if b.State == "" {
		b.State = cfg.State
	}
	if b.City == "" {
		b.City = cfg.City
	}
	if b.Line1 == "" {
		b.Line1 = cfg.Line1
	}
	if b.Country == "" {
		b.Country = cfg.Country
	}
	return b
}

func collectStripePublishableKeys(client *http.Client, checkoutURL, sessionID string) []string {
	seen := map[string]bool{}
	var keys []string
	add := func(pk string) {
		pk = strings.TrimSpace(pk)
		if pk == "" || seen[pk] {
			return
		}
		if strings.HasPrefix(sessionID, "cs_live_") && !strings.HasPrefix(pk, "pk_live_") {
			return
		}
		if strings.HasPrefix(sessionID, "cs_test_") && !strings.HasPrefix(pk, "pk_test_") {
			return
		}
		if !strings.Contains(pk, "_51") || len(pk) < 40 {
			return
		}
		seen[pk] = true
		keys = append(keys, pk)
	}

	add(cursorStripePK)

	pageURL := strings.TrimSpace(checkoutURL)
	if pageURL == "" {
		pageURL = stripeCheckoutOrigin + "/c/pay/" + sessionID
	}
	if i := strings.Index(pageURL, "#"); i >= 0 {
		pageURL = pageURL[:i]
	}
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return keys
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := client.Do(req)
	if err != nil {
		return keys
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return keys
	}
	for _, pk := range stripePublishableRe.FindAllString(string(body), -1) {
		add(pk)
	}
	return keys
}

func stripeInitWithKeys(client *http.Client, sessionID string, keys []string) (map[string]any, string, error) {
	if len(keys) == 0 {
		keys = []string{cursorStripePK}
	}
	var last error
	for _, pk := range keys {
		session, err := stripeInitPaymentPage(client, pk, sessionID)
		if err != nil {
			last = err
			logger.SysLog("[stripe-alipay] init 失败 pk=" + stripePKPreview(pk) + " err=" + err.Error())
			continue
		}
		return session, pk, nil
	}
	if last == nil {
		last = fmt.Errorf("没有可用的 publishable key")
	}
	return nil, "", last
}

func stripeInitPaymentPage(client *http.Client, pk, sessionID string) (map[string]any, error) {
	form := url.Values{}
	form.Set("key", pk)
	form.Set("eid", "NA")
	form.Set("browser_locale", "en-US")
	form.Set("browser_timezone", "Asia/Shanghai")
	return stripeFormPost(client, pk, stripeAPIBase+sessionID+"/init", form)
}

func stripeSwitchCurrency(client *http.Client, pk, sessionID, currency string) error {
	attempts := []url.Values{
		{"updated_currency": {currency}},
		{"currency": {currency}},
		{"presentment_currency": {currency}},
	}
	var last error
	for _, extra := range attempts {
		form := url.Values{}
		form.Set("key", pk)
		form.Set("eid", "NA")
		for k, vs := range extra {
			for _, v := range vs {
				form.Set(k, v)
			}
		}
		if _, err := stripeFormPost(client, pk, stripeAPIBase+sessionID, form); err != nil {
			last = err
			logger.SysLog("[stripe-alipay] 切换 USD 尝试失败 " + kOf(extra) + ": " + err.Error())
			continue
		}
		return nil
	}
	if last != nil {
		return last
	}
	return fmt.Errorf("切换币种失败")
}

func kOf(v url.Values) string {
	for k := range v {
		return k
	}
	return ""
}

func stripeHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     false,
		TLSNextProto:          map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
	}
	if proxyURL := model.GetCursorPayProxyURL(); proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
		logger.SysLog("[stripe-alipay] 使用配置代理 " + proxyURL.Scheme + "://" + proxyURL.Host)
	} else if client.HTTPClient != nil {
		if t, ok := client.HTTPClient.Transport.(*http.Transport); ok && t != nil && t.Proxy != nil {
			transport.Proxy = t.Proxy
		}
	}
	return &http.Client{
		Timeout:   45 * time.Second,
		Transport: transport,
	}
}

func stripeRetryableNetErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "tls handshake") ||
		strings.Contains(msg, "unexpected eof")
}

func stripeConfirmAlipay(client *http.Client, pk, sessionID string, amount int, currency, email string, billing stripeBilling) (map[string]any, error) {
	var last error

	pmID, err := stripeCreateAlipayPaymentMethod(client, pk, sessionID, email, billing)
	if err != nil {
		last = err
		logger.SysLog("[stripe-alipay] 创建 payment_method 失败: " + err.Error())
	} else {
		logger.SysLog("[stripe-alipay] 已创建 payment_method=" + pmID)
		if refreshed, rerr := stripeInitPaymentPage(client, pk, sessionID); rerr == nil {
			if amt := stripeExpectedAmount(refreshed); amt > 0 {
				amount = amt
			}
			if stripeAlipayRedirectURL(refreshed) != "" {
				return refreshed, nil
			}
		}
		data, confirmErr := stripeConfirmWithPaymentMethod(client, pk, sessionID, pmID, amount, currency)
		if confirmErr == nil {
			return data, nil
		}
		last = confirmErr
		if stripeConfirmAlreadyDone(confirmErr) {
			if retrySession, retryErr := stripeInitPaymentPage(client, pk, sessionID); retryErr == nil && stripeAlipayRedirectURL(retrySession) != "" {
				return retrySession, nil
			}
			return nil, last
		}
	}

	attempts := []struct {
		name string
		form url.Values
	}{
		{"full", stripeAlipayConfirmForm(pk, sessionID, amount, email, billing, true, true)},
		{"no-amount", stripeAlipayConfirmForm(pk, sessionID, 0, email, billing, true, true)},
		{"country-only", stripeAlipayConfirmForm(pk, sessionID, amount, email, billing, false, true)},
		{"minimal", stripeAlipayConfirmForm(pk, sessionID, 0, "", billing, false, false)},
	}
	for _, attempt := range attempts {
		logger.SysLog(fmt.Sprintf("[stripe-alipay] confirm session=%s amount=%d currency=%s mode=%s", sessionID, amount, currency, attempt.name))
		data, err := stripeFormPost(client, pk, stripeAPIBase+sessionID+"/confirm", attempt.form)
		if err == nil {
			if stripeAlipayRedirectURL(data) != "" {
				return data, nil
			}
			last = fmt.Errorf("confirm 成功但未返回支付宝跳转")
			continue
		}
		last = err
		logger.SysLog("[stripe-alipay] confirm 失败 mode=" + attempt.name + " err=" + err.Error())
		if retrySession, retryErr := stripeInitPaymentPage(client, pk, sessionID); retryErr == nil && stripeAlipayRedirectURL(retrySession) != "" {
			logger.SysLog("[stripe-alipay] confirm 报错后重新 init 已拿到 Alipay 跳转")
			return retrySession, nil
		}
		if stripeConfirmAlreadyDone(err) {
			break
		}
	}
	if last == nil {
		last = fmt.Errorf("提交 Alipay 失败")
	}
	return nil, last
}

func stripeConfirmAlreadyDone(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already been confirmed") || strings.Contains(msg, "has already been")
}

func stripeConfirmWithPaymentMethod(client *http.Client, pk, sessionID, pmID string, amount int, currency string) (map[string]any, error) {
	tryAmount := amount
	var last error
	for i := 0; i < 2; i++ {
		logger.SysLog(fmt.Sprintf("[stripe-alipay] confirm session=%s amount=%d currency=%s mode=payment-method", sessionID, tryAmount, currency))
		form := stripeAlipayConfirmWithPMForm(pk, sessionID, pmID, tryAmount)
		data, err := stripeFormPost(client, pk, stripeAPIBase+sessionID+"/confirm", form)
		if err == nil {
			if stripeAlipayRedirectURL(data) != "" {
				return data, nil
			}
			last = fmt.Errorf("confirm 成功但未返回支付宝跳转")
			return nil, last
		}
		last = err
		logger.SysLog("[stripe-alipay] confirm 失败 mode=payment-method err=" + err.Error())
		if retrySession, retryErr := stripeInitPaymentPage(client, pk, sessionID); retryErr == nil && stripeAlipayRedirectURL(retrySession) != "" {
			logger.SysLog("[stripe-alipay] confirm 报错后重新 init 已拿到 Alipay 跳转")
			return retrySession, nil
		}
		if i == 0 && strings.Contains(strings.ToLower(err.Error()), "amount") {
			if refreshed, rerr := stripeInitPaymentPage(client, pk, sessionID); rerr == nil {
				if amt := stripeExpectedAmount(refreshed); amt > 0 && amt != tryAmount {
					tryAmount = amt
					continue
				}
			}
		}
		break
	}
	if last == nil {
		last = fmt.Errorf("提交 Alipay 失败")
	}
	return nil, last
}

func stripeCreateAlipayPaymentMethod(client *http.Client, pk, sessionID, email string, billing stripeBilling) (string, error) {
	var last error
	for _, postal := range stripePostalCandidates(billing.PostalCode) {
		b := billing
		b.PostalCode = postal
		id, err := stripeCreateAlipayPaymentMethodOnce(client, pk, sessionID, email, b)
		if err == nil {
			if postal != strings.TrimSpace(billing.PostalCode) {
				logger.SysLog("[stripe-alipay] 邮编 " + billing.PostalCode + " 无效，已改用 " + postal)
			}
			return id, nil
		}
		last = err
		if !strings.Contains(strings.ToLower(err.Error()), "postal") {
			return "", err
		}
		logger.SysLog("[stripe-alipay] 创建 payment_method 邮编失败 postal=" + postal + " err=" + err.Error())
	}
	if last == nil {
		last = fmt.Errorf("创建 Alipay payment method 失败")
	}
	return "", last
}

func stripeCreateAlipayPaymentMethodOnce(client *http.Client, pk, sessionID, email string, billing stripeBilling) (string, error) {
	form := url.Values{}
	form.Set("type", "alipay")
	form.Set("key", pk)
	form.Set("guid", "NA")
	form.Set("muid", "NA")
	form.Set("sid", "NA")
	form.Set("billing_details[name]", billing.Name)
	form.Set("billing_details[address][country]", billing.Country)
	form.Set("billing_details[address][line1]", billing.Line1)
	form.Set("billing_details[address][city]", billing.City)
	form.Set("billing_details[address][state]", billing.State)
	form.Set("billing_details[address][postal_code]", billing.PostalCode)
	if email != "" {
		form.Set("billing_details[email]", email)
	}
	form.Set("client_attribution_metadata[checkout_session_id]", sessionID)
	form.Set("client_attribution_metadata[merchant_integration_source]", "checkout")
	form.Set("client_attribution_metadata[merchant_integration_version]", "hosted_checkout")
	form.Set("client_attribution_metadata[payment_method_selection_flow]", "automatic")

	data, err := stripeFormPost(client, pk, stripePaymentMethodAPI, form)
	if err != nil {
		return "", err
	}
	id := stripeMapString(data, "id")
	if id == "" {
		return "", fmt.Errorf("创建 Alipay payment method 未返回 id")
	}
	return id, nil
}

func stripePostalCandidates(configured string) []string {
	configured = strings.TrimSpace(configured)
	seen := map[string]bool{}
	out := make([]string, 0, 6)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		out = append(out, v)
	}
	add(configured)
	for _, v := range []string{"313000", "200000", "100000", "518000", "545452"} {
		add(v)
	}
	return out
}

func stripeAlipayConfirmWithPMForm(pk, sessionID, pmID string, amount int) url.Values {
	form := url.Values{}
	form.Set("key", pk)
	form.Set("eid", "NA")
	form.Set("payment_method", pmID)
	form.Set("expected_payment_method_type", "alipay")
	form.Set("return_url", stripeCheckoutOrigin+"/c/pay/"+sessionID)
	if amount > 0 {
		form.Set("expected_amount", fmt.Sprintf("%d", amount))
	}
	return form
}

func stripeAlipayConfirmForm(pk, sessionID string, amount int, email string, billing stripeBilling, fullAddress, includeEmail bool) url.Values {
	form := url.Values{}
	form.Set("key", pk)
	form.Set("eid", "NA")
	form.Set("expected_payment_method_type", "alipay")
	form.Set("payment_method_data[type]", "alipay")
	form.Set("payment_method_data[guid]", "NA")
	form.Set("payment_method_data[muid]", "NA")
	form.Set("payment_method_data[sid]", "NA")
	form.Set("payment_method_data[billing_details][name]", billing.Name)
	form.Set("payment_method_data[billing_details][address][country]", billing.Country)
	if includeEmail && email != "" {
		form.Set("payment_method_data[billing_details][email]", email)
	}
	if fullAddress {
		form.Set("payment_method_data[billing_details][address][postal_code]", billing.PostalCode)
		form.Set("payment_method_data[billing_details][address][state]", billing.State)
		form.Set("payment_method_data[billing_details][address][city]", billing.City)
		form.Set("payment_method_data[billing_details][address][line1]", billing.Line1)
	}
	if amount > 0 {
		form.Set("expected_amount", fmt.Sprintf("%d", amount))
	}
	form.Set("return_url", stripeCheckoutOrigin+"/c/pay/"+sessionID)
	return form
}

func stripeFormPost(httpClient *http.Client, pk, endpoint string, form url.Values) (map[string]any, error) {
	encoded := form.Encode()
	var last error
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 400 * time.Millisecond)
			logger.SysLog(fmt.Sprintf("[stripe-alipay] 网络重试 %d %s", i+1, endpoint))
		}
		req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+pk)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Origin", stripeCheckoutOrigin)
		req.Header.Set("Referer", stripeCheckoutOrigin+"/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
		req.Header.Set("Connection", "close")
		req.ContentLength = int64(len(encoded))

		resp, err := httpClient.Do(req)
		if err != nil {
			last = err
			if stripeRetryableNetErr(err) {
				continue
			}
			return nil, err
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if err != nil {
			last = err
			if stripeRetryableNetErr(err) {
				continue
			}
			return nil, err
		}

		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			preview := strings.TrimSpace(string(body))
			if len(preview) > 240 {
				preview = preview[:240]
			}
			return nil, fmt.Errorf("解析 Stripe 响应失败: %s", preview)
		}
		if errObj, ok := data["error"].(map[string]any); ok {
			if pi := stripeMap(errObj["payment_intent"]); stripeAlipayRedirectURL(map[string]any{"payment_intent": pi}) != "" {
				logger.SysLog("[stripe-alipay] confirm 返回 error 但已带 Alipay 跳转: " + stripeMapString(errObj, "message"))
				return map[string]any{"payment_intent": pi}, nil
			}
			msg := stripeMapString(errObj, "message")
			code := stripeMapString(errObj, "code")
			param := stripeMapString(errObj, "param")
			if msg == "" {
				msg = strings.TrimSpace(string(body))
				if len(msg) > 240 {
					msg = msg[:240]
				}
			}
			if code != "" || param != "" {
				msg = strings.TrimSpace(msg + " code=" + code + " param=" + param)
			}
			logger.SysLog("[stripe-alipay] stripe error " + msg)
			return nil, fmt.Errorf("%s", msg)
		}
		if resp.StatusCode >= 400 {
			preview := strings.TrimSpace(string(body))
			if len(preview) > 240 {
				preview = preview[:240]
			}
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, preview)
		}
		return data, nil
	}
	if last == nil {
		last = fmt.Errorf("请求 Stripe 失败")
	}
	return nil, last
}

func stripePresentmentCurrency(session map[string]any) string {
	adaptive := stripeMap(session["adaptive_pricing_info"])
	if cur := strings.ToLower(stripeMapString(adaptive, "active_presentment_currency")); cur != "" {
		return cur
	}
	return strings.ToLower(stripeMapString(session, "currency"))
}

func stripeExpectedAmount(session map[string]any) int {
	if summary := stripeMap(session["total_summary"]); summary != nil {
		if due := stripeMapInt(summary, "due"); due > 0 {
			return due
		}
		if total := stripeMapInt(summary, "total"); total > 0 {
			return total
		}
	}
	adaptive := stripeMap(session["adaptive_pricing_info"])
	if amt := stripeMapInt(adaptive, "integration_amount"); amt > 0 {
		return amt
	}
	if pi := stripeMap(session["payment_intent"]); pi != nil {
		if amt := stripeMapInt(pi, "amount"); amt > 0 {
			return amt
		}
	}
	if invoice := stripeMap(session["invoice"]); invoice != nil {
		if due := stripeMapInt(invoice, "amount_due"); due > 0 {
			return due
		}
		if total := stripeMapInt(invoice, "total"); total > 0 {
			return total
		}
	}
	if amt := stripeMapInt(session, "amount_total"); amt > 0 {
		return amt
	}
	return 0
}

func stripeAlipayRedirectURL(session map[string]any) string {
	pi := stripeMap(session["payment_intent"])
	next := stripeMap(pi["next_action"])
	redirect := stripeMap(next["alipay_handle_redirect"])
	var fallback string
	for _, key := range []string{"url", "native_url", "native_data"} {
		v := stripeMapString(redirect, key)
		if strings.HasPrefix(v, "http") {
			return v
		}
		if fallback == "" && v != "" {
			fallback = v
		}
	}
	return fallback
}

func stripeGeoCountry(session map[string]any) string {
	return stripeMapString(stripeMap(session["geocoding"]), "country_code")
}

func stripePaymentMethodTypes(session map[string]any) []string {
	return stripeStringList(session["payment_method_types"])
}

func stripeAvailablePaymentMethods(session map[string]any) []string {
	ordered := stripeStringList(session["ordered_payment_method_types"])
	if len(ordered) > 0 {
		return ordered
	}
	specs := session["payment_method_specs"]
	if raw, ok := specs.([]any); ok && len(raw) > 0 {
		out := make([]string, 0, len(raw))
		seen := map[string]bool{}
		for _, item := range raw {
			t := stripeMapString(stripeMap(item), "type")
			if t == "" || seen[t] {
				continue
			}
			seen[t] = true
			out = append(out, t)
		}
		if len(out) > 0 {
			return out
		}
	}
	return stripePaymentMethodTypes(session)
}

func stripeHasPaymentMethod(methods []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, m := range methods {
		if strings.ToLower(m) == want {
			return true
		}
	}
	return false
}

func stripeStringList(v any) []string {
	raw, _ := v.([]any)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func stripePKPreview(pk string) string {
	if len(pk) <= 18 {
		return pk
	}
	return pk[:12] + "..." + pk[len(pk)-6:]
}

func stripeMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func stripeMapString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return strings.TrimSpace(s)
}

func stripeMapInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}
