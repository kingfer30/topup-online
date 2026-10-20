package controller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

const (
	xpassPaywallName       = "dynamic::productType::PRODUCT_TYPE_XPASS"
	billingLine1           = "13 Northeast Revere Avenue"
	billingCity            = "Bend"
	billingState           = "OR"
	billingPostal          = "97701"
	billingCountry         = "US"
	billingPlatformStripe  = 1
	intervalUnitMonth      = 2
	stripeJSBuild          = "2bae21f786"
	stripePaymentElementUA = "stripe.js/" + stripeJSBuild + "; stripe-js-v3/" + stripeJSBuild + "; payment-element"
)

var (
	subscribeTierName = map[string]string{
		"plus":    "Plus",
		"premium": "Premium",
		"super":   "Super",
		"ultra":   "Ultra",
	}
	subscribeTierSKU = map[string]string{
		"plus":    "sku_ToYBWq565pLU",
		"premium": "sku_iFHd43ge7daR",
		"super":   "sku_WWvZ1VL9fabh",
		"ultra":   "",
	}
	stripeKeyPattern       = regexp.MustCompile(`pk_live_[A-Za-z0-9]+`)
	longDigitPattern       = regexp.MustCompile(`\d{12,19}`)
	subscribeSecretPattern = regexp.MustCompile(`pps-session\.v1\.[A-Za-z0-9+/=._-]+|eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}(?:\.[A-Za-z0-9_-]+)?|user_[A-Za-z0-9]+::\S+|pk_(?:live|test)_[A-Za-z0-9]+`)
)

// SubscribeWindow 只换出该卡密对应档位的付款页地址，不提交支付。
func SubscribeWindow(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		ID       int    `json:"id" binding:"required"`
		Tier     string `json:"tier" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	tier := strings.ToLower(strings.TrimSpace(req.Tier))
	if _, ok := subscribeTierName[tier]; !ok {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "不支持的订阅档位"})
		return
	}
	if req.Category != "cursor" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "仅 cursor 卡密支持订阅"})
		return
	}
	tableName := model.GetTableNameByCategory(req.Category)
	card, err := model.GetCardById(tableName, req.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询卡密失败: " + err.Error()})
		return
	}
	if strings.TrimSpace(card.Token) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该卡密没有 Token"})
		return
	}
	workosID, sessionToken, err := splitCursorSession(card.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	payURL, err := cursorPaywallURL(workosID, sessionToken, tier)
	if err != nil {
		subscribeLog(card.Id, card.Account, "付款页", "失败 "+err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	subscribeLog(card.Id, card.Account, "付款页", "已打开 "+tier)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已取得付款页", "data": gin.H{"url": payURL}})
}

func cursorPaywallURL(workosID, sessionToken, tier string) (string, error) {
	httpClient := newSubscribeHTTPClient()
	setCursorSession(httpClient.Jar, workosID, sessionToken)
	args := map[string]string{
		"tier":       tier,
		"entrypoint": "plan_info",
		"host":       "web",
	}
	if sku := subscribeTierSKU[tier]; sku != "" {
		args["sku"] = sku
	}
	body, _ := json.Marshal(map[string]any{
		"action": "xaiCommercePaywall",
		"args":   args,
	})
	status, raw, err := postCursorJSON(httpClient, "https://cursor.com/api/dashboard/client-action", "https://cursor.com/dashboard", string(body))
	if err != nil {
		return "", fmt.Errorf("获取付款页失败: %w", err)
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("获取付款页失败: HTTP %d %s", status, clipText(string(raw)))
	}
	var wrap struct {
		OpenURL string `json:"openUrl"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil || wrap.OpenURL == "" {
		return "", fmt.Errorf("没有返回付款页: %s", clipText(string(raw)))
	}
	return wrap.OpenURL, nil
}

// OpenSubscribe 在服务端为该卡密选择档位、填写固定账单并提交卡支付。
func OpenSubscribe(c *gin.Context) {
	var req struct {
		Category   string `json:"category" binding:"required"`
		ID         int    `json:"id" binding:"required"`
		Tier       string `json:"tier" binding:"required"`
		CardNumber string `json:"card_number" binding:"required"`
		ExpMonth   string `json:"exp_month" binding:"required"`
		ExpYear    string `json:"exp_year" binding:"required"`
		CVC        string `json:"cvc" binding:"required"`
		Line1      string `json:"line1"`
		City       string `json:"city"`
		State      string `json:"state"`
		Postal     string `json:"postal"`
		Country    string `json:"country"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		subscribeLog(req.ID, "", "参数", "失败 "+err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	tier := strings.ToLower(strings.TrimSpace(req.Tier))
	planName, ok := subscribeTierName[tier]
	if !ok {
		subscribeLog(req.ID, "", "档位", "不支持 "+tier)
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "不支持的订阅档位"})
		return
	}
	if req.Category != "cursor" {
		subscribeLog(req.ID, "", "类别", "不支持 "+req.Category)
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "仅 cursor 卡密支持订阅"})
		return
	}
	number, month, year, cvc, err := normalizeCard(req.CardNumber, req.ExpMonth, req.ExpYear, req.CVC)
	if err != nil {
		subscribeLog(req.ID, "", "卡信息", "失败 "+err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	address, err := normalizeBilling(req.Line1, req.City, req.State, req.Postal, req.Country)
	if err != nil {
		subscribeLog(req.ID, "", "账单地址", "失败 "+err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	card, err := model.GetCardById(tableName, req.ID)
	if err != nil {
		subscribeLog(req.ID, "", "查询卡密", "失败 "+err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询卡密失败: " + err.Error()})
		return
	}
	if strings.TrimSpace(card.Token) == "" {
		subscribeLog(card.Id, card.Account, "Token", "为空")
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该卡密没有 Token"})
		return
	}
	subscribeLog(card.Id, card.Account, "开始", fmt.Sprintf("档位 %s 卡尾号 %s 有效期 %s/%s 地址 %s, %s, %s %s, %s", planName, cardLast4(number), month, year, address.Line1, address.City, address.State, address.Postal, address.Country))

	message, err := payCursorSubscription(card, planName, number, month, year, cvc, address)
	if err != nil {
		subscribeLog(card.Id, card.Account, "结束", "失败 "+err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": redactCard(err.Error())})
		return
	}
	subscribeLog(card.Id, card.Account, "结束", message)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": message,
		"data":    gin.H{"paid": true},
	})
}

func payCursorSubscription(card *model.AccountCard, planName, number, month, year, cvc string, address billingAddress) (string, error) {
	workosID, sessionToken, err := splitCursorSession(card.Token)
	if err != nil {
		subscribeLog(card.Id, card.Account, "Token", "解析失败 "+err.Error())
		return "", err
	}
	subscribeLog(card.Id, card.Account, "Token", "解析成功 "+workosID)
	httpClient := newSubscribeHTTPClient()
	setCursorSession(httpClient.Jar, workosID, sessionToken)
	bearer, err := cursorCommerceBearer(httpClient)
	if err != nil && strings.Contains(err.Error(), "not_linked") {
		subscribeLog(card.Id, card.Account, "Commerce凭证", "账号未关联 Grok，开始免费合并")
		email, already, mergeErr := ensureFreeUnification(httpClient)
		if mergeErr != nil {
			subscribeLog(card.Id, card.Account, "免费合并", "失败 "+mergeErr.Error())
			return "", fmt.Errorf("账号未关联 Grok，自动合并失败: %w", mergeErr)
		}
		if already {
			subscribeLog(card.Id, card.Account, "免费合并", "已经合并 "+email)
		} else {
			subscribeLog(card.Id, card.Account, "免费合并", "已合并 "+email)
		}
		bearer, err = cursorCommerceBearer(httpClient)
	}
	if err != nil {
		subscribeLog(card.Id, card.Account, "Commerce凭证", "失败 "+err.Error())
		return "", err
	}
	subscribeLog(card.Id, card.Account, "Commerce凭证", fmt.Sprintf("获取成功 长度 %d", len(bearer)))

	createBody := append(protoString(1, xpassPaywallName), protoString(2, address.Country)...)
	createBody = append(createBody, protoEmbed(3, []byte{billingPlatformStripe})...)
	raw, err := commerceCall(httpClient, bearer, "https://grok.com/commerce.prepurchase.v1.PrePurchaseService/CreateCheckoutSession", createBody)
	if err != nil {
		subscribeLog(card.Id, card.Account, "创建结账", "失败 "+err.Error())
		return "", fmt.Errorf("创建结账失败: %w", err)
	}
	payload := firstGRPCPayload(raw)
	session := checkoutSessionToken(payload)
	if session == "" {
		detail := subscribeExcerpt(raw)
		subscribeLog(card.Id, card.Account, "创建结账", "没有会话 "+detail)
		if msg := findErrorDetail(payload); msg != "" {
			return "", fmt.Errorf("创建结账失败: %s", msg)
		}
		return "", fmt.Errorf("创建结账失败")
	}
	subscribeLog(card.Id, card.Account, "创建结账", fmt.Sprintf("成功 会话长度 %d", len(session)))
	sku, err := pickMonthlySKU(payload, planName)
	if err != nil {
		subscribeLog(card.Id, card.Account, "选择月付", "失败 "+err.Error()+" 摘要 "+subscribeExcerpt(raw))
		return "", err
	}
	subscribeLog(card.Id, card.Account, "选择月付", sku)

	cartRaw, err := commerceCall(httpClient, bearer, "https://grok.com/commerce.prepurchase.v1.PrePurchaseService/GetCartForSession", cartRequest(session, sku))
	if err != nil {
		subscribeLog(card.Id, card.Account, "购物车", "失败 "+err.Error())
		return "", fmt.Errorf("生成购物车失败: %w", err)
	}
	cartPayload := firstGRPCPayload(cartRaw)
	cart := cartID(cartPayload)
	if cart == "" {
		detail := subscribeExcerpt(cartRaw)
		subscribeLog(card.Id, card.Account, "购物车", "没有 cart "+detail)
		if msg := findErrorDetail(cartPayload); msg != "" {
			return "", fmt.Errorf("生成购物车失败: %s", msg)
		}
		return "", fmt.Errorf("生成购物车失败")
	}
	publishableKey := stripeKeyPattern.FindString(string(cartRaw))
	if publishableKey == "" {
		subscribeLog(card.Id, card.Account, "购物车", "成功 "+cart+" 但没有支付公钥 摘要 "+subscribeExcerpt(cartRaw))
		return "", fmt.Errorf("没有拿到支付公钥")
	}
	keyHint := publishableKey
	if len(keyHint) > 20 {
		keyHint = keyHint[:20]
	}
	subscribeLog(card.Id, card.Account, "购物车", fmt.Sprintf("成功 %s 公钥 %s", cart, keyHint))

	paymentMethod, err := createStripePaymentMethod(httpClient, publishableKey, number, month, year, cvc, strings.TrimSpace(card.Account), address)
	if err != nil {
		subscribeLog(card.Id, card.Account, "Stripe", "失败 "+err.Error())
		return "", err
	}
	subscribeLog(card.Id, card.Account, "Stripe", "支付方式 "+paymentMethod)
	teamID := cartTeamID(cartPayload)
	if teamID == "" {
		subscribeLog(card.Id, card.Account, "绑定支付方式", "购物车没有 team_id，直接提交")
	} else {
		subscribeLog(card.Id, card.Account, "绑定支付方式", "team "+teamID)
		attachRaw, attachErr := commerceCall(httpClient, bearer, "https://grok.com/commerce.billing.v1.AcquisitionService/AttachPaymentMethod", attachPaymentMethodRequest(teamID, paymentMethod, address))
		if attachErr != nil {
			subscribeLog(card.Id, card.Account, "绑定支付方式", "失败 "+attachErr.Error())
			return "", fmt.Errorf("绑定支付方式失败: %w", attachErr)
		}
		attachPayload := firstGRPCPayload(attachRaw)
		if msg := messageError(attachPayload, 3); msg != "" && !strings.Contains(msg, "ALREADY") {
			if reason := declineReasonText(attachPayload, 4); reason != "" {
				msg = msg + "（" + reason + "）"
			}
			subscribeLog(card.Id, card.Account, "绑定支付方式", "业务失败 "+msg+" 摘要 "+subscribeExcerpt(attachRaw))
			return "", fmt.Errorf("绑定支付方式失败: %s", msg)
		}
		if secret := metadataString(attachPayload, 1); strings.HasPrefix(secret, "seti_") {
			subscribeLog(card.Id, card.Account, "绑定支付方式", "银行要求 3DS 验证")
			return "", fmt.Errorf("这张卡需要银行验证，服务端无法完成 3DS")
		}
		subscribeLog(card.Id, card.Account, "绑定支付方式", "成功")
	}
	acquireRaw, err := commerceCall(httpClient, bearer, "https://grok.com/commerce.billing.v1.AcquisitionService/Acquire", acquireRequest(cart, paymentMethod, strings.TrimSpace(card.Account), address))
	if err != nil {
		subscribeLog(card.Id, card.Account, "提交支付", "失败 "+err.Error())
		return "", fmt.Errorf("提交支付失败: %w", err)
	}
	acquirePayload := firstGRPCPayload(acquireRaw)
	if msg := acquireError(acquirePayload); msg != "" {
		subscribeLog(card.Id, card.Account, "提交支付", "业务失败 "+msg+" 摘要 "+subscribeExcerpt(acquireRaw))
		return "", fmt.Errorf("支付未完成: %s", msg)
	}
	purchaseID, subscriptionID := acquireIDs(acquirePayload)
	if purchaseID == "" && subscriptionID == "" {
		detail := subscribeExcerpt(acquireRaw)
		subscribeLog(card.Id, card.Account, "提交支付", "没有订单号 "+detail)
		if msg := findErrorDetail(acquirePayload); msg != "" {
			return "", fmt.Errorf("支付未完成: %s", msg)
		}
		return "", fmt.Errorf("支付未完成")
	}
	subscribeLog(card.Id, card.Account, "提交支付", fmt.Sprintf("成功 purchase=%s subscription=%s", purchaseID, subscriptionID))
	message := "已提交 " + planName + " 支付"
	if subscriptionID != "" {
		message += "，订阅 " + subscriptionID
	}
	return message, nil
}

func newSubscribeHTTPClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if proxyURL := model.GetCursorPayProxyURL(); proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
		logger.SysLog("[subscribe] 使用配置代理 " + proxyURL.Scheme + "://" + proxyURL.Host)
	} else {
		logger.SysLog("[subscribe] 未配置后台代理，走环境代理或全局客户端")
		if client.HTTPClient != nil {
			if current, ok := client.HTTPClient.Transport.(*http.Transport); ok && current != nil && current.Proxy != nil {
				transport.Proxy = current.Proxy
			}
		}
	}
	profiled, profile := withBrowserProfile(transport)
	logger.SysLog("[subscribe] 浏览器 " + profile.Name)
	return &http.Client{Jar: jar, Timeout: 40 * time.Second, Transport: profiled}
}

func subscribeLog(id int, account, step, detail string) {
	logger.SysLogf("[subscribe][ID:%d Account:%s] %s | %s", id, account, step, redactSubscribeLog(detail))
}

func redactSubscribeLog(value string) string {
	value = redactCard(value)
	value = subscribeSecretPattern.ReplaceAllString(value, "[redacted]")
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 700 {
		return value[:700]
	}
	return value
}

func subscribeExcerpt(raw []byte) string {
	var b strings.Builder
	for _, r := range string(raw) {
		if r >= 32 && r < 127 {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return redactSubscribeLog(b.String())
}

func stripeDeviceID() string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return newIDKey()
	}
	return newIDKey() + hex.EncodeToString(buf)
}

func cardLast4(number string) string {
	if len(number) < 4 {
		return "****"
	}
	return number[len(number)-4:]
}

func cartRequest(sessionToken, sku string) []byte {
	interval := append(protoVarintField(1, 1), protoVarintField(2, intervalUnitMonth)...)
	selection := append(protoString(1, sku), protoVarintField(2, 1)...)
	selection = append(selection, protoEmbed(3, interval)...)
	body := append(protoString(1, sessionToken), protoEmbed(2, selection)...)
	return append(body, protoVarintField(3, billingPlatformStripe)...)
}

func taxAddressBytes(billing billingAddress) []byte {
	address := protoString(1, billing.Country)
	address = append(address, protoString(2, billing.Postal)...)
	address = append(address, protoString(3, billing.State)...)
	address = append(address, protoString(4, billing.Line1)...)
	return append(address, protoString(5, billing.City)...)
}

func cartTeamID(payload []byte) string {
	for _, field := range protoFields(payload) {
		if field.num == 8 && field.wt == 2 && isPrintable(string(field.raw)) {
			return string(field.raw)
		}
	}
	return ""
}

func attachPaymentMethodRequest(teamID, paymentMethodID string, billing billingAddress) []byte {
	body := protoVarintField(1, billingPlatformStripe)
	body = append(body, protoString(2, paymentMethodID)...)
	body = append(body, protoString(3, newIDKey())...)
	body = append(body, protoString(5, teamID)...)
	body = append(body, protoString(6, "https://cursor.com/dashboard")...)
	return append(body, protoEmbed(7, taxAddressBytes(billing))...)
}

func declineReasonText(payload []byte, fieldNum int) string {
	for _, field := range protoFields(payload) {
		if field.num != fieldNum || field.wt != 0 {
			continue
		}
		switch field.numVal {
		case 1:
			return "银行拒绝"
		case 2:
			return "余额不足"
		case 3:
			return "超过限额"
		case 4:
			return "卡已过期"
		case 5:
			return "卡号不正确"
		case 6:
			return "CVV 不正确"
		case 7:
			return "邮编不正确"
		case 8:
			return "卡不支持"
		case 9:
			return "暂时拒绝，可稍后重试"
		case 10:
			return "银行验证失败"
		case 11:
			return "银行验证超时"
		case 12:
			return "授权已撤销"
		default:
			return ""
		}
	}
	return ""
}

func metadataString(payload []byte, fieldNum int) string {
	for _, field := range protoFields(payload) {
		if field.num != 2 || field.wt != 2 {
			continue
		}
		for _, inner := range protoFields(field.raw) {
			if inner.num == fieldNum && inner.wt == 2 && isPrintable(string(inner.raw)) {
				return string(inner.raw)
			}
		}
	}
	return ""
}

func acquireRequest(cartID, paymentMethodID, email string, billing billingAddress) []byte {
	body := protoString(1, cartID)
	body = append(body, protoEmbed(2, protoString(1, paymentMethodID))...)
	body = append(body, protoString(4, newIDKey())...)
	body = append(body, protoEmbed(7, taxAddressBytes(billing))...)
	if strings.Contains(email, "@") {
		body = append(body, protoString(9, email)...)
	}
	return body
}

func checkoutSessionToken(payload []byte) string {
	for _, field := range protoFields(payload) {
		if field.num != 2 || field.wt != 2 {
			continue
		}
		for _, inner := range protoFields(field.raw) {
			if inner.num == 1 && inner.wt == 2 && strings.HasPrefix(string(inner.raw), "pps-session.") {
				return string(inner.raw)
			}
		}
	}
	return ""
}

func pickMonthlySKU(payload []byte, planName string) (string, error) {
	for _, field := range protoFields(payload) {
		if field.num != 3 || field.wt != 2 {
			continue
		}
		sku := ""
		name := ""
		for _, inner := range protoFields(field.raw) {
			if inner.num == 1 && inner.wt == 2 {
				sku = string(inner.raw)
			}
			if inner.num == 9 && inner.wt == 2 {
				name = string(inner.raw)
			}
		}
		if strings.EqualFold(name, planName) && strings.HasPrefix(sku, "sku_") && intervalUnit(field.raw) == intervalUnitMonth {
			return sku, nil
		}
	}
	return "", fmt.Errorf("没有找到 %s 月付价格", planName)
}

func intervalUnit(buf []byte) int {
	fields := protoFields(buf)
	count, unit := 0, 0
	hasCount, hasUnit := false, false
	nested := false
	for _, field := range fields {
		if field.wt == 2 {
			nested = true
			continue
		}
		if field.num == 1 && field.wt == 0 {
			count = field.numVal
			hasCount = true
		}
		if field.num == 2 && field.wt == 0 {
			unit = field.numVal
			hasUnit = true
		}
	}
	if !nested && hasCount && hasUnit && count == 1 && unit >= 1 && unit <= 5 {
		return unit
	}
	for _, field := range fields {
		if field.wt == 2 {
			if unit := intervalUnit(field.raw); unit != 0 {
				return unit
			}
		}
	}
	return 0
}

func cartID(payload []byte) string {
	for _, field := range protoFields(payload) {
		if field.num == 1 && field.wt == 2 && strings.HasPrefix(string(field.raw), "cart_") {
			return string(field.raw)
		}
	}
	return ""
}

func acquireIDs(payload []byte) (string, string) {
	purchaseID := ""
	subscriptionID := ""
	for _, field := range protoFields(payload) {
		if field.wt != 2 {
			continue
		}
		value := string(field.raw)
		if field.num == 2 && strings.HasPrefix(value, "crec_") {
			purchaseID = value
		}
		if field.num == 3 && strings.HasPrefix(value, "csub_") {
			subscriptionID = value
		}
	}
	return purchaseID, subscriptionID
}

func acquireError(payload []byte) string {
	return messageError(payload, 5)
}

func messageError(payload []byte, fieldNum int) string {
	for _, field := range protoFields(payload) {
		if field.num == fieldNum && field.wt == 2 {
			if msg := errorDetailText(field.raw); msg != "" {
				return msg
			}
		}
	}
	return ""
}

func findErrorDetail(buf []byte) string {
	fields := protoFields(buf)
	code, message := "", ""
	nested := false
	for _, field := range fields {
		if field.wt != 2 {
			continue
		}
		value := string(field.raw)
		if field.num == 1 && isErrorCode(value) {
			code = value
			continue
		}
		if field.num == 2 && isPrintable(value) && !strings.HasPrefix(value, "sku_") && !strings.HasPrefix(value, "pps-") {
			message = value
			continue
		}
		nested = true
	}
	if code != "" && message != "" && !nested {
		return code + ": " + message
	}
	for _, field := range fields {
		if field.wt == 2 {
			if msg := findErrorDetail(field.raw); msg != "" {
				return msg
			}
		}
	}
	return ""
}

func errorDetailText(buf []byte) string {
	code, message := "", ""
	for _, field := range protoFields(buf) {
		if field.wt != 2 {
			continue
		}
		if field.num == 1 {
			code = string(field.raw)
		}
		if field.num == 2 {
			message = string(field.raw)
		}
	}
	if code == "" {
		return message
	}
	if message == "" {
		return code
	}
	return code + ": " + message
}

func isErrorCode(value string) bool {
	if len(value) < 6 || len(value) > 80 || !strings.Contains(value, "_") {
		return false
	}
	for _, r := range value {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func isPrintable(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < 32 || r > 126 {
			return false
		}
	}
	return true
}

func commerceCall(httpClient *http.Client, bearer, rawURL string, payload []byte) ([]byte, error) {
	status, raw, err := grokRequestBody(httpClient, http.MethodPost, rawURL, "https://cursor.com/promos/grok-upgrade", grpcFrame(payload), "application/grpc-web+proto", bearer)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", status)
	}
	return raw, nil
}

type billingAddress struct {
	Line1   string
	City    string
	State   string
	Postal  string
	Country string
}

func normalizeBilling(line1, city, state, postal, country string) (billingAddress, error) {
	address := billingAddress{
		Line1:   strings.TrimSpace(line1),
		City:    strings.TrimSpace(city),
		State:   strings.TrimSpace(state),
		Postal:  strings.TrimSpace(postal),
		Country: strings.ToUpper(strings.TrimSpace(country)),
	}
	if address.Line1 == "" && address.City == "" && address.State == "" && address.Postal == "" && address.Country == "" {
		address = billingAddress{
			Line1:   billingLine1,
			City:    billingCity,
			State:   billingState,
			Postal:  billingPostal,
			Country: billingCountry,
		}
	}
	if len(address.State) == 2 {
		address.State = strings.ToUpper(address.State)
	}
	if address.Line1 == "" || address.City == "" || address.State == "" || address.Postal == "" || len(address.Country) != 2 {
		return billingAddress{}, fmt.Errorf("账单地址不完整")
	}
	if len(address.Line1) > 120 || len(address.City) > 80 || len(address.State) > 40 || len(address.Postal) > 20 {
		return billingAddress{}, fmt.Errorf("账单地址过长")
	}
	return address, nil
}

func createStripePaymentMethod(httpClient *http.Client, publishableKey, number, month, year, cvc, email string, billing billingAddress) (string, error) {
	form := url.Values{}
	form.Set("type", "card")
	form.Set("card[number]", number)
	form.Set("card[exp_month]", month)
	form.Set("card[exp_year]", year)
	form.Set("card[cvc]", cvc)
	form.Set("billing_details[address][line1]", billing.Line1)
	form.Set("billing_details[address][city]", billing.City)
	form.Set("billing_details[address][state]", billing.State)
	form.Set("billing_details[address][postal_code]", billing.Postal)
	form.Set("billing_details[address][country]", billing.Country)
	if strings.Contains(email, "@") {
		form.Set("billing_details[email]", email)
	}
	form.Set("guid", stripeDeviceID())
	form.Set("muid", stripeDeviceID())
	form.Set("sid", stripeDeviceID())
	form.Set("payment_user_agent", stripePaymentElementUA)
	form.Set("referrer", "https://cursor.com/promos/grok-upgrade")
	form.Set("time_on_page", "30000")
	form.Set("client_attribution_metadata[client_session_id]", newIDKey())
	form.Set("client_attribution_metadata[merchant_integration_source]", "elements")
	form.Set("client_attribution_metadata[merchant_integration_subtype]", "payment-element")
	form.Set("client_attribution_metadata[merchant_integration_version]", "2022")
	form.Set("key", publishableKey)

	req, err := http.NewRequest(http.MethodPost, "https://api.stripe.com/v1/payment_methods", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("origin", "https://js.stripe.com")
	req.Header.Set("referer", "https://js.stripe.com/")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("创建支付方式失败: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("创建支付方式失败: %w", err)
	}
	var parsed struct {
		ID    string `json:"id"`
		Error struct {
			Type        string `json:"type"`
			Code        string `json:"code"`
			DeclineCode string `json:"decline_code"`
			Message     string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("创建支付方式失败: HTTP %d 响应无法解析", resp.StatusCode)
	}
	if parsed.ID == "" || !strings.HasPrefix(parsed.ID, "pm_") {
		parts := []string{parsed.Error.Type, parsed.Error.Code, parsed.Error.DeclineCode, redactCard(parsed.Error.Message)}
		detail := strings.TrimSpace(strings.Join(parts, " "))
		if detail == "" {
			detail = "没有返回支付方式"
		}
		return "", fmt.Errorf("银行卡未通过: HTTP %d %s", resp.StatusCode, detail)
	}
	return parsed.ID, nil
}

func normalizeCard(number, month, year, cvc string) (string, string, string, string, error) {
	number = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, number)
	if len(number) < 13 || len(number) > 19 {
		return "", "", "", "", fmt.Errorf("卡号长度不正确")
	}
	monthNumber, err := strconv.Atoi(strings.TrimSpace(month))
	if err != nil || monthNumber < 1 || monthNumber > 12 {
		return "", "", "", "", fmt.Errorf("月份不正确")
	}
	year = strings.TrimSpace(year)
	if len(year) == 2 {
		year = "20" + year
	}
	yearNumber, err := strconv.Atoi(year)
	if err != nil || len(year) != 4 || yearNumber < 2024 || yearNumber > 2100 {
		return "", "", "", "", fmt.Errorf("年份不正确")
	}
	cvc = strings.TrimSpace(cvc)
	if len(cvc) < 3 || len(cvc) > 4 {
		return "", "", "", "", fmt.Errorf("CVV 不正确")
	}
	for _, r := range cvc {
		if r < '0' || r > '9' {
			return "", "", "", "", fmt.Errorf("CVV 不正确")
		}
	}
	return number, fmt.Sprintf("%02d", monthNumber), year, cvc, nil
}

func redactCard(value string) string {
	return longDigitPattern.ReplaceAllString(value, "[redacted]")
}

func protoEmbed(field int, msg []byte) []byte {
	out := appendVarint(nil, field<<3|2)
	out = appendVarint(out, len(msg))
	return append(out, msg...)
}

func protoVarintField(field, value int) []byte {
	out := appendVarint(nil, field<<3)
	return appendVarint(out, value)
}
