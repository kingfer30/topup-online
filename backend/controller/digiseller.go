package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
)

const (
	digisellerApiLoginURL   = "https://api.digiseller.com/api/apilogin"
	digisellerUniqueCodeURL = "https://api.digiseller.com/api/purchases/unique-code/%s?token=%s"
	digisellerTokenValidMin = 120 // token 有效期 2 小时（分钟）
	digisellerTokenPreExp   = 5   // 提前 5 分钟刷新（分钟）
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
