package controller

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

const defaultGptRtLicenseAppID = "gpt-rt-register"

var (
	gptRtLicensePrivOnce sync.Once
	gptRtLicensePrivKey  ed25519.PrivateKey
	gptRtLicensePrivErr  error
)

func loadGptRtLicensePrivateKey() (ed25519.PrivateKey, error) {
	gptRtLicensePrivOnce.Do(func() {
		b64 := strings.TrimSpace(os.Getenv("GPT_RT_LICENSE_ED25519_PRIVATE_KEY"))
		if b64 == "" {
			gptRtLicensePrivErr = errors.New("未配置 GPT_RT_LICENSE_ED25519_PRIVATE_KEY")
			return
		}
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			gptRtLicensePrivErr = fmt.Errorf("私钥 Base64 解码失败: %w", err)
			return
		}
		if len(raw) != ed25519.PrivateKeySize {
			gptRtLicensePrivErr = fmt.Errorf("私钥长度无效: %d", len(raw))
			return
		}
		gptRtLicensePrivKey = ed25519.PrivateKey(raw)
	})
	return gptRtLicensePrivKey, gptRtLicensePrivErr
}

func gptRtLicenseSignPayload(licenseKey, machineID string, expiresAt time.Time) (string, error) {
	priv, err := loadGptRtLicensePrivateKey()
	if err != nil {
		return "", err
	}
	payload := licenseKey + "|" + machineID + "|" + strconv.FormatInt(expiresAt.Unix(), 10)
	sig := ed25519.Sign(priv, []byte(payload))
	return base64.StdEncoding.EncodeToString(sig), nil
}

func generateGptRtLicenseKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hexStr := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("%s-%s-%s-%s",
		hexStr[0:4], hexStr[4:8], hexStr[8:12], hexStr[12:16]), nil
}

type verifyGptRtLicenseReq struct {
	LicenseKey string `json:"license_key" binding:"required"`
	MachineId  string `json:"machine_id" binding:"required"`
	AppId      string `json:"app_id" binding:"required"`
}

func gptRtLicenseClientMeta(c *gin.Context) (ip, ua string) {
	ip = c.ClientIP()
	ua = strings.TrimSpace(c.GetHeader("User-Agent"))
	if len(ua) > 512 {
		ua = ua[:512]
	}
	return
}

// bindGptRtLicenseDevice 绑定设备；已绑定则更新 IP/UA 后放行，超限则返回 reason
func bindGptRtLicenseDevice(lic *model.GptRtLicense, machineID, ip, ua string) (reason string, err error) {
	_, err = model.GetGptRtLicenseDeviceByMachine(lic.Id, machineID)
	if err == nil {
		if uerr := model.UpdateGptRtLicenseDeviceMeta(lic.Id, machineID, ip, ua); uerr != nil {
			return "", uerr
		}
		return "", nil
	}
	if !model.IsGptRtLicenseNotFound(err) {
		return "", err
	}

	count, err := model.CountGptRtLicenseDevices(lic.Id)
	if err != nil {
		return "", err
	}
	maxDevices := lic.MaxDevices
	if maxDevices < 1 {
		maxDevices = 1
	}
	if int(count) >= maxDevices {
		return "超过绑定设备数", nil
	}

	now := time.Now()
	dev := &model.GptRtLicenseDevice{
		LicenseId: lic.Id,
		MachineId: machineID,
		LoginIP:   ip,
		UserAgent: ua,
		BoundAt:   now,
	}
	if err := model.CreateGptRtLicenseDevice(dev); err != nil {
		return "", err
	}
	if lic.ActivatedAt == nil {
		lic.ActivatedAt = &now
	}
	return "", nil
}

// VerifyGptRtLicense 公开接口：验证/激活 GPT RT 许可证
func VerifyGptRtLicense(c *gin.Context) {
	var req verifyGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.MachineId = strings.TrimSpace(req.MachineId)
	req.AppId = strings.TrimSpace(req.AppId)

	if IsRevoked(req.LicenseKey) {
		c.JSON(http.StatusOK, gin.H{
			"code": 200, "message": "success",
			"data": gin.H{"valid": false, "reason": "许可证已禁用"},
		})
		return
	}

	lic, err := model.GetGptRtLicenseByKey(req.LicenseKey)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{
				"code": 200, "message": "success",
				"data": gin.H{"valid": false, "reason": "激活码不存在"},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	if lic.AppId != req.AppId {
		c.JSON(http.StatusOK, gin.H{
			"code": 200, "message": "success",
			"data": gin.H{"valid": false, "reason": "应用标识不匹配"},
		})
		return
	}
	if lic.Status != 1 {
		SetRevoked(lic.LicenseKey)
		c.JSON(http.StatusOK, gin.H{
			"code": 200, "message": "success",
			"data": gin.H{"valid": false, "reason": "许可证已禁用"},
		})
		return
	}
	if time.Now().After(lic.ExpiresAt) {
		c.JSON(http.StatusOK, gin.H{
			"code": 200, "message": "success",
			"data": gin.H{"valid": false, "reason": "许可证已过期"},
		})
		return
	}

	ip, ua := gptRtLicenseClientMeta(c)
	activatedWasNil := lic.ActivatedAt == nil
	reason, err := bindGptRtLicenseDevice(lic, req.MachineId, ip, ua)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "绑定设备失败: " + err.Error()})
		return
	}
	if reason != "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 200, "message": "success",
			"data": gin.H{"valid": false, "reason": reason},
		})
		return
	}

	now := time.Now()
	// 节流：仅当 LastVerifiedAt 为空或超过 5 分钟时更新 last_verified_at/last_using_ip
	needSaveVerify := lic.LastVerifiedAt == nil || now.Sub(*lic.LastVerifiedAt) >= 5*time.Minute
	if needSaveVerify {
		lic.LastVerifiedAt = &now
		lic.LastUsingIP = ip
	}
	// 首次激活写入 ActivatedAt，或节流命中时落库
	if needSaveVerify || (activatedWasNil && lic.ActivatedAt != nil) {
		if err := model.SaveGptRtLicense(lic); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
			return
		}
	}

	sig, err := gptRtLicenseSignPayload(lic.LicenseKey, req.MachineId, lic.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "签名失败: " + err.Error()})
		return
	}

	used, available, held := lic.UsedCount, lic.AvailableCount, lic.HeldCount
	if cu, ca, ch, ok := GetQuotaCache(lic.Id); ok {
		used, available, held = cu, ca, ch
	} else {
		SetQuotaCache(lic.Id, used, available, held)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{
			"valid":           true,
			"expires_at":      lic.ExpiresAt.UTC().Format(time.RFC3339),
			"signature":       sig,
			"customer":        lic.Customer,
			"name":            lic.Customer,
			"used_count":      used,
			"available_count": available,
			"held_count":      held,
			"max_devices":     lic.MaxDevices,
		},
	})
}

type consumeGptRtLicenseReq struct {
	LicenseKey string `json:"license_key" binding:"required"`
	MachineId  string `json:"machine_id" binding:"required"`
	AppId      string `json:"app_id" binding:"required"`
}

// ConsumeGptRtLicense 公开接口：提取 RT 成功后扣减可用次数
func ConsumeGptRtLicense(c *gin.Context) {
	var req consumeGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.MachineId = strings.TrimSpace(req.MachineId)
	req.AppId = strings.TrimSpace(req.AppId)

	lic, err := model.GetGptRtLicenseByKey(req.LicenseKey)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "激活码不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	if lic.AppId != req.AppId {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "应用标识不匹配"})
		return
	}
	if lic.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "许可证已禁用"})
		return
	}
	if time.Now().After(lic.ExpiresAt) {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "许可证已过期"})
		return
	}

	_, err = model.GetGptRtLicenseDeviceByMachine(lic.Id, req.MachineId)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 403, "message": "设备未绑定"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询设备失败: " + err.Error()})
		return
	}

	if lic.AvailableCount <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "可用提取次数已用尽"})
		return
	}

	if err := model.ConsumeGptRtLicenseQuota(lic.Id); err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 403, "message": "可用提取次数已用尽"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "扣减失败: " + err.Error()})
		return
	}
	InvalidateQuotaCache(lic.Id)

	ip, _ := gptRtLicenseClientMeta(c)
	_ = model.UpdateGptRtLicense(lic.Id, map[string]interface{}{"last_using_ip": ip})

	updated, _ := model.GetGptRtLicenseByID(lic.Id)
	used, available := lic.UsedCount+1, lic.AvailableCount-1
	if updated != nil {
		used = updated.UsedCount
		available = updated.AvailableCount
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{
			"used_count":      used,
			"available_count": available,
		},
	})
}

const defaultGptRtHoldTTL = 10 * time.Minute

type reserveGptRtLicenseReq struct {
	LicenseKey string `json:"license_key" binding:"required"`
	MachineId  string `json:"machine_id" binding:"required"`
	AppId      string `json:"app_id" binding:"required"`
	Email      string `json:"email"`
}

type confirmReleaseGptRtLicenseReq struct {
	LicenseKey string `json:"license_key" binding:"required"`
	MachineId  string `json:"machine_id" binding:"required"`
	AppId      string `json:"app_id" binding:"required"`
	HoldId     string `json:"hold_id" binding:"required"`
}

// loadGptRtLicenseForQuotaOp 校验 key/app/状态/过期/设备绑定（与 consume 一致）
// requireActive=false 时跳过 status 校验（confirm/release 需允许禁用许可证走释放逻辑）
func loadGptRtLicenseForQuotaOp(licenseKey, machineID, appID string, requireActive bool) (*model.GptRtLicense, int, string) {
	lic, err := model.GetGptRtLicenseByKey(licenseKey)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			return nil, 404, "激活码不存在"
		}
		return nil, 500, "查询失败: " + err.Error()
	}
	if lic.AppId != appID {
		return nil, 400, "应用标识不匹配"
	}
	if requireActive && lic.Status != 1 {
		return nil, 403, "许可证已禁用"
	}
	if requireActive && time.Now().After(lic.ExpiresAt) {
		return nil, 403, "许可证已过期"
	}
	_, err = model.GetGptRtLicenseDeviceByMachine(lic.Id, machineID)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			return nil, 403, "设备未绑定"
		}
		return nil, 500, "查询设备失败: " + err.Error()
	}
	return lic, 0, ""
}

// ReserveGptRtLicense 公开接口：预占一次配额
func ReserveGptRtLicense(c *gin.Context) {
	var req reserveGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.MachineId = strings.TrimSpace(req.MachineId)
	req.AppId = strings.TrimSpace(req.AppId)
	req.Email = strings.TrimSpace(req.Email)

	if IsRevoked(req.LicenseKey) {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "许可证已禁用"})
		return
	}

	_, _ = model.ExpireGptRtLicenseHolds(50)

	lic, code, msg := loadGptRtLicenseForQuotaOp(req.LicenseKey, req.MachineId, req.AppId, true)
	if lic == nil {
		if code == 403 && strings.Contains(msg, "已禁用") {
			SetRevoked(req.LicenseKey)
		}
		c.JSON(http.StatusOK, gin.H{"code": code, "message": msg})
		return
	}
	if lic.AvailableCount <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "可用提取次数已用尽"})
		return
	}

	holdID, updated, err := model.ReserveGptRtLicenseQuota(lic.Id, req.MachineId, req.Email, defaultGptRtHoldTTL)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 403, "message": "可用提取次数已用尽"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "预占失败: " + err.Error()})
		return
	}
	InvalidateQuotaCache(lic.Id)

	used, available, held := updated.UsedCount, updated.AvailableCount, updated.HeldCount
	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{
			"hold_id":         holdID,
			"used_count":      used,
			"available_count": available,
			"held_count":      held,
		},
	})
}

// ConfirmGptRtLicense 公开接口：确认预占，计入已用
func ConfirmGptRtLicense(c *gin.Context) {
	var req confirmReleaseGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.MachineId = strings.TrimSpace(req.MachineId)
	req.AppId = strings.TrimSpace(req.AppId)
	req.HoldId = strings.TrimSpace(req.HoldId)

	lic, code, msg := loadGptRtLicenseForQuotaOp(req.LicenseKey, req.MachineId, req.AppId, false)
	if lic == nil {
		c.JSON(http.StatusOK, gin.H{"code": code, "message": msg})
		return
	}

	updated, err := model.ConfirmGptRtLicenseHold(req.HoldId)
	if err != nil {
		if errors.Is(err, model.ErrGptRtLicenseDisabled) {
			InvalidateQuotaCache(lic.Id)
			c.JSON(http.StatusOK, gin.H{"code": 403, "message": "许可证已禁用"})
			return
		}
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "预占不存在或已处理"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "确认失败: " + err.Error()})
		return
	}
	InvalidateQuotaCache(lic.Id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{
			"used_count":      updated.UsedCount,
			"available_count": updated.AvailableCount,
			"held_count":      updated.HeldCount,
		},
	})
}

// ReleaseGptRtLicense 公开接口：释放预占，恢复可用
func ReleaseGptRtLicense(c *gin.Context) {
	var req confirmReleaseGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.MachineId = strings.TrimSpace(req.MachineId)
	req.AppId = strings.TrimSpace(req.AppId)
	req.HoldId = strings.TrimSpace(req.HoldId)

	lic, code, msg := loadGptRtLicenseForQuotaOp(req.LicenseKey, req.MachineId, req.AppId, false)
	if lic == nil {
		c.JSON(http.StatusOK, gin.H{"code": code, "message": msg})
		return
	}

	updated, err := model.ReleaseGptRtLicenseHold(req.HoldId)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "预占不存在或已处理"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "释放失败: " + err.Error()})
		return
	}
	InvalidateQuotaCache(lic.Id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{
			"used_count":      updated.UsedCount,
			"available_count": updated.AvailableCount,
			"held_count":      updated.HeldCount,
		},
	})
}

// ListGptRtLicenses 列表
func ListGptRtLicenses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	statusFilter, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	// 0=全部 1=正常 2=禁用
	dbStatus := -1
	switch statusFilter {
	case 1:
		dbStatus = 1
	case 2:
		dbStatus = 0
	}

	list, total, err := model.GetGptRtLicenseList(page, pageSize, dbStatus, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{"list": list, "total": total},
	})
}

// ListGptRtLicenseDevices 绑定设备明细
func ListGptRtLicenseDevices(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的 ID"})
		return
	}
	if _, err := model.GetGptRtLicenseByID(id); err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "许可证不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	list, err := model.GetGptRtLicenseDevices(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": gin.H{"list": list},
	})
}

type createGptRtLicenseReq struct {
	Customer       string `json:"customer"`
	AppId          string `json:"app_id"`
	Months         int    `json:"months" binding:"required"`
	AvailableCount int    `json:"available_count" binding:"required"`
	MaxDevices     int    `json:"max_devices"`
}

// CreateGptRtLicense 颁发许可证
func CreateGptRtLicense(c *gin.Context) {
	var req createGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if req.Months < 1 || req.Months > 120 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "有效期月数须在 1~120 之间"})
		return
	}
	if req.AvailableCount < 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "可用数量必须为正整数"})
		return
	}
	maxDevices := req.MaxDevices
	if maxDevices <= 0 {
		maxDevices = 1
	}
	appID := strings.TrimSpace(req.AppId)
	if appID == "" {
		appID = defaultGptRtLicenseAppID
	}

	key, err := generateGptRtLicenseKey()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "生成激活码失败: " + err.Error()})
		return
	}

	lic := &model.GptRtLicense{
		LicenseKey:     key,
		AppId:          appID,
		Customer:       strings.TrimSpace(req.Customer),
		Status:         1,
		ExpiresAt:      time.Now().AddDate(0, req.Months, 0),
		AvailableCount: req.AvailableCount,
		MaxDevices:     maxDevices,
	}
	if err := model.CreateGptRtLicense(lic); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "success",
		"data": lic,
	})
}

type updateGptRtLicenseReq struct {
	Customer          *string `json:"customer"`
	Status            *int    `json:"status"`
	ExpiresAt         *string `json:"expires_at"`
	Months            *int    `json:"months"`
	AvailableCount    *int    `json:"available_count"`
	AddAvailableCount *int    `json:"add_available_count"`
	MaxDevices        *int    `json:"max_devices"`
}

// UpdateGptRtLicense 更新（续期/禁用/配额）
func UpdateGptRtLicense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的 ID"})
		return
	}
	var req updateGptRtLicenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	lic, err := model.GetGptRtLicenseByID(id)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "许可证不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Customer != nil {
		updates["customer"] = strings.TrimSpace(*req.Customer)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.ExpiresAt))
		if parseErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "expires_at 格式无效"})
			return
		}
		updates["expires_at"] = t
		lic.ExpiresAt = t
	}
	if req.Months != nil && *req.Months > 0 {
		base := lic.ExpiresAt
		if base.Before(time.Now()) {
			base = time.Now()
		}
		updates["expires_at"] = base.AddDate(0, *req.Months, 0)
	}
	if req.AvailableCount != nil {
		if *req.AvailableCount < 1 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "可用数量必须为正整数"})
			return
		}
		updates["available_count"] = *req.AvailableCount
	}
	if req.AddAvailableCount != nil {
		if *req.AddAvailableCount < 1 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "增加数量必须为正整数"})
			return
		}
		if req.AvailableCount != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "不能同时设置可用数量与增加数量"})
			return
		}
		updates["available_count"] = model.ExprAddAvailableCount(*req.AddAvailableCount)
	}
	if req.MaxDevices != nil {
		if *req.MaxDevices < 1 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "可绑定设备数必须为正整数"})
			return
		}
		updates["max_devices"] = *req.MaxDevices
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "没有可更新的字段"})
		return
	}
	if err := model.UpdateGptRtLicense(id, updates); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}

	if req.Status != nil {
		if *req.Status == 0 {
			_ = model.RevokeGptRtLicenseHolds(id)
			SetRevoked(lic.LicenseKey)
		} else if *req.Status == 1 {
			ClearRevoked(lic.LicenseKey)
		}
	}
	if req.Status != nil || req.AvailableCount != nil || req.AddAvailableCount != nil || req.Months != nil {
		InvalidateQuotaCache(id)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteGptRtLicense 删除
func DeleteGptRtLicense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的 ID"})
		return
	}
	lic, err := model.GetGptRtLicenseByID(id)
	if err != nil {
		if model.IsGptRtLicenseNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "许可证不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	_ = model.RevokeGptRtLicenseHolds(id)
	SetRevoked(lic.LicenseKey)
	InvalidateQuotaCache(id)
	if err := model.DeleteGptRtLicense(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}
