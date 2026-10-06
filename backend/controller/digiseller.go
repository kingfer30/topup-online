package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
)

const (
	digisellerApiLoginURL     = "https://api.digiseller.com/api/apilogin"
	digisellerUniqueCodeURL   = "https://api.digiseller.com/api/purchases/unique-code/%s?token=%s"
	digisellerSellerSellsURL  = "https://api.digiseller.com/api/seller-sells/v2?token=%s"
	digisellerTokenValidMin   = 120 // token 有效期 2 小时（分钟）
	digisellerTokenPreExp     = 5   // 提前 5 分钟刷新（分钟）
	digisellerSyncLookbackDay = 30
	digisellerSyncPageRows    = 1000
	digisellerSyncMaxPages    = 20
)

// token 内存缓存，进程级别，重启后会重新获取
var (
	digiToken    string
	digiTokenExp time.Time
	digiTokenMu  sync.Mutex
)

// --- 请求/响应结构体 ---

type digisellerLoginRequest struct {
	SellerID  int    `json:"seller_id"`
	Timestamp int64  `json:"timestamp"`
	Sign      string `json:"sign"`
}

type digisellerLoginResponse struct {
	Retval    int    `json:"retval"`
	Desc      string `json:"desc"`
	Token     string `json:"token"`
	SellerID  int    `json:"seller_id"`
	ValidThru string `json:"valid_thru"`
}

type digisellerUniqueCodeState struct {
	State         int    `json:"state"`
	DateCheck     string `json:"date_check"`
	DateDelivery  string `json:"date_delivery"`
	DateConfirmed string `json:"date_confirmed"`
	DateRefuted   string `json:"date_refuted"`
}

type digisellerOption struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	VariantID *int   `json:"variant_id"`
}

type digisellerUniqueCodeResponse struct {
	Retval          int                       `json:"retval"`
	Retdesc         string                    `json:"retdesc"`
	Inv             int64                     `json:"inv"`
	IdGoods         int64                     `json:"id_goods"`
	Amount          float64                   `json:"amount"`
	TypeCurr        string                    `json:"type_curr"`
	Profit          string                    `json:"profit"`
	AmountUsd       float64                   `json:"amount_usd"`
	DatePay         string                    `json:"date_pay"`
	Email           string                    `json:"email"`
	AgentID         int                       `json:"agent_id"`
	AgentPercent    float64                   `json:"agent_percent"`
	UnitGoods       int                       `json:"unit_goods"`
	CntGoods        int                       `json:"cnt_goods"`
	PromoCode       string                    `json:"promo_code"`
	BonusCode       string                    `json:"bonus_code"`
	CartUID         string                    `json:"cart_uid"`
	UniqueCodeState digisellerUniqueCodeState `json:"unique_code_state"`
	Options         []digisellerOption        `json:"options"`
}

// --- 工具函数 ---

// getDigisellerConfig 从环境变量读取 API Key 和 Seller ID
func getDigisellerConfig() (apiKey string, sellerID int, err error) {
	apiKey = os.Getenv("DIGISELLER_API_KEY")
	if apiKey == "" {
		return "", 0, fmt.Errorf("环境变量 DIGISELLER_API_KEY 未配置")
	}
	sellerIDStr := os.Getenv("DIGISELLER_SELLER_ID")
	if sellerIDStr == "" {
		return "", 0, fmt.Errorf("环境变量 DIGISELLER_SELLER_ID 未配置")
	}
	sellerID, err = strconv.Atoi(sellerIDStr)
	if err != nil {
		return "", 0, fmt.Errorf("DIGISELLER_SELLER_ID 格式错误: %v", err)
	}
	return apiKey, sellerID, nil
}

// buildDigisellerSign 生成签名：SHA256(API_KEY + timestamp)
func buildDigisellerSign(apiKey string, timestamp int64) string {
	raw := fmt.Sprintf("%s%d", apiKey, timestamp)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

// parseDigisellerTime 解析 Digiseller 返回的时间字符串，格式不固定时返回 nil
func parseDigisellerTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t
		}
	}
	logger.SysLog(fmt.Sprintf("Digiseller 时间格式无法解析: %s", s))
	return nil
}

// --- Token 管理 ---

// doRefreshToken 实际执行 token 刷新请求，调用方必须已持有 digiTokenMu 锁
func doRefreshToken() (string, error) {
	apiKey, sellerID, err := getDigisellerConfig()
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Unix()
	reqBody := digisellerLoginRequest{
		SellerID:  sellerID,
		Timestamp: timestamp,
		Sign:      buildDigisellerSign(apiKey, timestamp),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, digisellerApiLoginURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求 Digiseller apilogin 失败: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var loginResp digisellerLoginResponse
	if err = json.Unmarshal(respBytes, &loginResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v, body: %s", err, string(respBytes))
	}

	if loginResp.Retval != 0 {
		return "", fmt.Errorf("Digiseller 登录失败，retval=%d, desc=%s", loginResp.Retval, loginResp.Desc)
	}
	if loginResp.Token == "" {
		return "", fmt.Errorf("Digiseller 返回 token 为空")
	}

	digiToken = loginResp.Token
	digiTokenExp = time.Now().Add(time.Duration(digisellerTokenValidMin) * time.Minute)
	logger.SysLog(fmt.Sprintf("Digiseller token 刷新成功，有效至 %s", digiTokenExp.Format(time.RFC3339)))
	return digiToken, nil
}

// getDigisellerToken 获取有效 token，不足 5 分钟过期则自动刷新
func getDigisellerToken() (string, error) {
	digiTokenMu.Lock()
	defer digiTokenMu.Unlock()

	if digiToken != "" && time.Now().Before(digiTokenExp.Add(-time.Duration(digisellerTokenPreExp)*time.Minute)) {
		return digiToken, nil
	}
	return doRefreshToken()
}

// forceRefreshDigisellerToken 强制丢弃缓存并重新获取 token（用于 401 场景）
func forceRefreshDigisellerToken() (string, error) {
	digiTokenMu.Lock()
	defer digiTokenMu.Unlock()
	digiToken = ""
	return doRefreshToken()
}

// --- API 调用 ---

// callUniqueCodeAPI 调用 Digiseller unique-code 接口，返回解析结果和 HTTP 状态码
func callUniqueCodeAPI(uniqueCode, token string) (*digisellerUniqueCodeResponse, int, error) {
	url := fmt.Sprintf(digisellerUniqueCodeURL, uniqueCode, token)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求 Digiseller unique-code 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, http.StatusUnauthorized, nil
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %v", err)
	}

	var result digisellerUniqueCodeResponse
	if err = json.Unmarshal(respBytes, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %v, body: %s", err, string(respBytes))
	}

	return &result, resp.StatusCode, nil
}

// --- Handler ---

// CheckUniqueCode 查询 Digiseller 唯一码支付信息
// GET /admin/digiseller/check-code/:unique_code
func CheckUniqueCode(c *gin.Context) {
	uniqueCode := c.Param("unique_code")
	if uniqueCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "唯一码不能为空"})
		return
	}

	// 1. 获取 token
	token, err := getDigisellerToken()
	if err != nil {
		logger.SysError("获取 Digiseller token 失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取 Digiseller token 失败: " + err.Error()})
		return
	}

	// 2. 调用 Digiseller API
	result, statusCode, err := callUniqueCodeAPI(uniqueCode, token)
	if err != nil {
		logger.SysError(fmt.Sprintf("调用 Digiseller unique-code API 失败: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "调用 Digiseller API 失败: " + err.Error()})
		return
	}

	// 3. 若返回 401，强制刷新 token 后重试一次
	if statusCode == http.StatusUnauthorized {
		logger.SysLog("Digiseller 返回 401，强制刷新 token 后重试")
		token, err = forceRefreshDigisellerToken()
		if err != nil {
			logger.SysError("强制刷新 Digiseller token 失败: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "刷新 Digiseller token 失败: " + err.Error()})
			return
		}
		result, statusCode, err = callUniqueCodeAPI(uniqueCode, token)
		if err != nil {
			logger.SysError(fmt.Sprintf("重试调用 Digiseller unique-code API 失败: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "调用 Digiseller API 失败: " + err.Error()})
			return
		}
		if statusCode == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Digiseller token 无效，请检查 API Key 和 Seller ID 配置"})
			return
		}
	}

	// 4. 检查业务状态码
	if result.Retval != 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    result.Retval,
			"message": result.Retdesc,
			"data":    nil,
		})
		return
	}

	// 5. 先入库，防止后续逻辑失败导致订单记录丢失
	optionsJSON := ""
	if len(result.Options) > 0 {
		if b, jsonErr := json.Marshal(result.Options); jsonErr == nil {
			optionsJSON = string(b)
		}
	}

	order := &model.DigisellerOrder{
		Inv:             result.Inv,
		UniqueCode:      uniqueCode,
		IdGoods:         result.IdGoods,
		Amount:          result.Amount,
		TypeCurr:        result.TypeCurr,
		AmountUsd:       result.AmountUsd,
		Profit:          result.Profit,
		DatePay:         parseDigisellerTime(result.DatePay),
		Email:           result.Email,
		AgentId:         result.AgentID,
		AgentPercent:    result.AgentPercent,
		CntGoods:        result.CntGoods,
		PromoCode:       result.PromoCode,
		BonusCode:       result.BonusCode,
		CartUid:         result.CartUID,
		UcState:         int8(result.UniqueCodeState.State),
		UcDateCheck:     parseDigisellerTime(result.UniqueCodeState.DateCheck),
		UcDateDelivery:  parseDigisellerTime(result.UniqueCodeState.DateDelivery),
		UcDateConfirmed: parseDigisellerTime(result.UniqueCodeState.DateConfirmed),
		UcDateRefuted:   parseDigisellerTime(result.UniqueCodeState.DateRefuted),
		OptionsJson:     optionsJSON,
	}

	if err = model.CreateOrUpdateDigisellerOrder(order); err != nil {
		// 入库失败仅记录日志，不阻断响应（可通过日志人工补录）
		logger.SysError(fmt.Sprintf("Digiseller 订单入库失败，inv=%d, unique_code=%s: %v", result.Inv, uniqueCode, err))
	} else {
		logger.SysLog(fmt.Sprintf("Digiseller 订单入库成功，inv=%d, unique_code=%s", result.Inv, uniqueCode))
	}

	// 6. 返回完整数据给前端
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    result,
	})
}

// GetDigisellerOrderList 分页查询 digiseller_orders
// GET /admin/digiseller/orders
func GetDigisellerOrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	filterState := false
	ucState := 0
	if raw := c.Query("uc_state"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "唯一码状态参数无效"})
			return
		}
		filterState = true
		ucState = parsed
	}

	list, total, err := model.ListDigisellerOrders(page, pageSize, keyword, ucState, filterState)
	if err != nil {
		logger.SysError("查询 Digiseller 订单失败: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

type sellerSellsRequest struct {
	DateStart  string `json:"date_start"`
	DateFinish string `json:"date_finish"`
	Returned   int    `json:"returned"`
	Page       int    `json:"page"`
	Rows       int    `json:"rows"`
}

type sellerSellRow struct {
	InvoiceID      int64       `json:"invoice_id"`
	ProductID      int64       `json:"product_id"`
	ProductName    string      `json:"product_name"`
	ProductEntryID int64       `json:"product_entry_id"`
	ProductEntry   string      `json:"product_entry"`
	DatePut        string      `json:"date_put"`
	DatePay        string      `json:"date_pay"`
	Email          string      `json:"email"`
	Wmid           string      `json:"wmid"`
	AmountIn       float64     `json:"amount_in"`
	AmountOut      float64     `json:"amount_out"`
	AmountCurrency string      `json:"amount_currency"`
	MethodPay      string      `json:"method_pay"`
	AggregatorPay  string      `json:"aggregator_pay"`
	IP             string      `json:"ip"`
	PartnerID      interface{} `json:"partner_id"`
	Lang           string      `json:"lang"`
	Returned       int         `json:"returned"`
	Owner          int         `json:"owner"`
}

type sellerSellsResponse struct {
	Retval    int             `json:"retval"`
	Retdesc   *string         `json:"retdesc"`
	TotalRows int             `json:"total_rows"`
	Pages     int             `json:"pages"`
	Page      int             `json:"page"`
	Rows      []sellerSellRow `json:"rows"`
}

func parsePartnerID(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(n))
		return i
	default:
		return 0
	}
}

func parseDigisellerMoscowTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	loc := time.FixedZone("MSK", 3*3600)
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc); err == nil {
		return &t
	}
	return parseDigisellerTime(s)
}

func truncateLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func callSellerSellsAPI(token, dateStart, dateFinish string, page, rows int) (*sellerSellsResponse, int, error) {
	reqBody := sellerSellsRequest{
		DateStart:  dateStart,
		DateFinish: dateFinish,
		Returned:   0,
		Page:       page,
		Rows:       rows,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化请求体失败: %v", err)
	}

	reqURL := fmt.Sprintf(digisellerSellerSellsURL, url.QueryEscape(token))
	logger.SysLog(fmt.Sprintf(
		"Digiseller seller-sells/v2 请求 start=%s finish=%s returned=0 page=%d rows=%d",
		dateStart, dateFinish, page, rows,
	))

	httpReq, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.SysError("Digiseller seller-sells/v2 请求失败: " + err.Error())
		return nil, 0, fmt.Errorf("请求 Digiseller seller-sells/v2 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		logger.SysError("Digiseller seller-sells/v2 返回 401")
		return nil, http.StatusUnauthorized, nil
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.SysError(fmt.Sprintf("Digiseller seller-sells/v2 读取响应失败 status=%d: %v", resp.StatusCode, err))
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %v", err)
	}
	bodyText := truncateLog(string(respBytes), 2000)

	var result sellerSellsResponse
	if err = json.Unmarshal(respBytes, &result); err != nil {
		logger.SysError(fmt.Sprintf("Digiseller seller-sells/v2 解析失败 status=%d body=%s err=%v", resp.StatusCode, bodyText, err))
		return nil, resp.StatusCode, fmt.Errorf("解析响应失败: %v, status=%d, body: %s", err, resp.StatusCode, bodyText)
	}
	if resp.StatusCode != http.StatusOK || result.Retval != 0 {
		logger.SysError(fmt.Sprintf(
			"Digiseller seller-sells/v2 失败 status=%d retval=%d body=%s",
			resp.StatusCode, result.Retval, bodyText,
		))
	}
	return &result, resp.StatusCode, nil
}

func sellerSellToOrder(row sellerSellRow) (*model.DigisellerOrder, error) {
	extra, err := json.Marshal(map[string]any{
		"product_name":     row.ProductName,
		"product_entry_id": row.ProductEntryID,
		"product_entry":    row.ProductEntry,
		"date_put":         row.DatePut,
		"wmid":             row.Wmid,
		"method_pay":       row.MethodPay,
		"aggregator_pay":   row.AggregatorPay,
		"ip":               row.IP,
		"lang":             row.Lang,
		"returned":         row.Returned,
		"owner":            row.Owner,
	})
	if err != nil {
		return nil, err
	}
	return &model.DigisellerOrder{
		Inv:         row.InvoiceID,
		IdGoods:     row.ProductID,
		Amount:      row.AmountIn,
		TypeCurr:    row.AmountCurrency,
		Profit:      strconv.FormatFloat(row.AmountOut, 'f', -1, 64),
		DatePay:     parseDigisellerMoscowTime(row.DatePay),
		Email:       row.Email,
		AgentId:     parsePartnerID(row.PartnerID),
		OptionsJson: string(extra),
	}, nil
}

// SyncDigisellerOrders 拉取近 30 天销售并写入 digiseller_orders
// POST /admin/digiseller/orders/sync
func SyncDigisellerOrders(c *gin.Context) {
	token, err := getDigisellerToken()
	if err != nil {
		logger.SysError("获取 Digiseller token 失败: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "获取 Digiseller token 失败: " + err.Error()})
		return
	}

	loc := time.FixedZone("MSK", 3*3600)
	now := time.Now().In(loc)
	dateFinish := now.Format("2006-01-02 15:04:05")
	dateStart := now.AddDate(0, 0, -digisellerSyncLookbackDay).Format("2006-01-02 15:04:05")

	saved := 0
	failed := 0
	totalRows := 0
	pages := 0

	for page := 1; page <= digisellerSyncMaxPages; page++ {
		if page > 1 {
			// 官方限制每分钟最多 4 次
			time.Sleep(16 * time.Second)
		}
		result, statusCode, callErr := callSellerSellsAPI(token, dateStart, dateFinish, page, digisellerSyncPageRows)
		if statusCode == http.StatusUnauthorized {
			logger.SysLog("Digiseller seller-sells/v2 返回 401，刷新 token 后重试")
			token, err = forceRefreshDigisellerToken()
			if err != nil {
				logger.SysError("强制刷新 Digiseller token 失败: " + err.Error())
				c.JSON(http.StatusOK, gin.H{"code": 500, "message": "刷新 Digiseller token 失败: " + err.Error()})
				return
			}
			result, statusCode, callErr = callSellerSellsAPI(token, dateStart, dateFinish, page, digisellerSyncPageRows)
			if statusCode == http.StatusUnauthorized {
				c.JSON(http.StatusOK, gin.H{"code": 401, "message": "Digiseller token 无效，请检查 API Key 和 Seller ID 配置"})
				return
			}
		}
		if callErr != nil {
			logger.SysError("拉取 Digiseller 销售失败: " + callErr.Error())
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "拉取 Digiseller 销售失败: " + callErr.Error()})
			return
		}
		if result.Retval != 0 {
			desc := ""
			if result.Retdesc != nil {
				desc = *result.Retdesc
			}
			logger.SysError(fmt.Sprintf("Digiseller 拉单失败 page=%d retval=%d desc=%s", page, result.Retval, desc))
			c.JSON(http.StatusOK, gin.H{
				"code":    result.Retval,
				"message": fmt.Sprintf("Digiseller 拉单失败，retval=%d, desc=%s", result.Retval, desc),
			})
			return
		}

		totalRows = result.TotalRows
		pages = result.Pages
		for _, row := range result.Rows {
			if row.InvoiceID <= 0 {
				failed++
				continue
			}
			order, convErr := sellerSellToOrder(row)
			if convErr != nil {
				failed++
				logger.SysError(fmt.Sprintf("转换 Digiseller 销售记录失败，inv=%d: %v", row.InvoiceID, convErr))
				continue
			}
			if saveErr := model.UpsertDigisellerOrderFromSale(order); saveErr != nil {
				failed++
				logger.SysError(fmt.Sprintf("保存 Digiseller 销售记录失败，inv=%d: %v", row.InvoiceID, saveErr))
				continue
			}
			saved++
		}

		if len(result.Rows) == 0 || (result.Pages > 0 && page >= result.Pages) {
			break
		}
	}

	logger.SysLog(fmt.Sprintf("Digiseller 拉单完成，saved=%d, failed=%d, total_rows=%d", saved, failed, totalRows))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"saved":       saved,
			"failed":      failed,
			"total_rows":  totalRows,
			"pages":       pages,
			"date_start":  dateStart,
			"date_finish": dateFinish,
		},
	})
}
