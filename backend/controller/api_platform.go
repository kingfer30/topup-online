package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/service/apiplatform"
	"gorm.io/gorm"
)

func platformID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "台子 ID 无效"})
		return 0, false
	}
	return id, true
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

func statusLabel(status int) string {
	switch status {
	case 1:
		return "启用"
	case 2:
		return "手动禁用"
	case 3:
		return "自动禁用"
	default:
		return strconv.Itoa(status)
	}
}

func billStatusLabel(status int) string {
	switch status {
	case model.ApiBillPending:
		return "待支付"
	case model.ApiBillPaid:
		return "已支付"
	case model.ApiBillVoid:
		return "已作废"
	default:
		return strconv.Itoa(status)
	}
}

type platformView struct {
	model.ApiPlatform
	APIKeyMask   string `json:"api_key_mask"`
	PasswordMask string `json:"password_mask"`
	APIKey       string `json:"api_key,omitempty"`
	Password     string `json:"password,omitempty"`
}

func ListApiPlatforms(c *gin.Context) {
	raw, err := model.ListApiPlatforms()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取台子失败: " + err.Error()})
		return
	}
	views := make([]platformView, 0, len(raw))
	for _, row := range raw {
		secretKey := row.APIKey
		secretPass := row.Password
		row.APIKey = ""
		row.Password = ""
		views = append(views, platformView{
			ApiPlatform:  row,
			APIKeyMask:   maskSecret(secretKey),
			PasswordMask: maskSecret(secretPass),
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": views})
}

func GetApiPlatform(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	row, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": row})
}

type platformRequest struct {
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Account  string `json:"account"`
	Password string `json:"password"`
	Remark   string `json:"remark"`
}

func CreateApiPlatform(c *gin.Context) {
	var req platformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.APIKey = strings.TrimSpace(req.APIKey)
	if req.BaseURL == "" || req.APIKey == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "台子地址和 API Key 必填"})
		return
	}
	if _, err := apiplatform.NormalizeBase(req.BaseURL); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	row := model.ApiPlatform{
		BaseURL:  req.BaseURL,
		APIKey:   req.APIKey,
		Account:  strings.TrimSpace(req.Account),
		Password: req.Password,
		Remark:   strings.TrimSpace(req.Remark),
	}
	if err := model.CreateApiPlatform(&row); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": row})
}

func UpdateApiPlatform(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	current, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	var req platformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	if req.BaseURL == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "台子地址必填"})
		return
	}
	if _, err := apiplatform.NormalizeBase(req.BaseURL); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	current.BaseURL = req.BaseURL
	if strings.TrimSpace(req.APIKey) != "" {
		current.APIKey = strings.TrimSpace(req.APIKey)
	}
	current.Account = strings.TrimSpace(req.Account)
	current.Password = req.Password
	current.Remark = strings.TrimSpace(req.Remark)
	if err := model.UpdateApiPlatform(current); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

func DeleteApiPlatform(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	if err := model.DeleteApiPlatform(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

type channelView struct {
	model.ApiPlatformChannel
	StatusLabel   string `json:"status_label"`
	InPendingBill bool   `json:"in_pending_bill"`
}

func ListApiPlatformChannels(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	rows, err := model.ListApiPlatformChannels(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取渠道失败: " + err.Error()})
		return
	}
	pending, err := model.PendingUpstreamChannelIDs(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取账单失败: " + err.Error()})
		return
	}
	views := make([]channelView, 0, len(rows))
	for _, row := range rows {
		_, locked := pending[row.ChannelId]
		views = append(views, channelView{
			ApiPlatformChannel: row,
			StatusLabel:        statusLabel(row.Status),
			InPendingBill:      locked,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": views})
}

type pullRequest struct {
	MinID   int    `json:"min_id"`
	MaxID   int    `json:"max_id"`
	Keyword string `json:"keyword"`
	Status  string `json:"status"`
}

func PullApiPlatformChannels(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	var req pullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	rows, perUnit, err := apiplatform.PullChannels(platform.BaseURL, platform.APIKey, strings.TrimSpace(req.Keyword), strings.TrimSpace(req.Status), req.MinID, req.MaxID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	type item struct {
		ChannelId   int     `json:"channel_id"`
		Name        string  `json:"name"`
		Models      string  `json:"models"`
		Status      int     `json:"status"`
		StatusLabel string  `json:"status_label"`
		UsedQuota   int64   `json:"used_quota"`
		UsedAmount  float64 `json:"used_amount"`
	}
	out := make([]item, 0, len(rows))
	for _, row := range rows {
		out = append(out, item{
			ChannelId:   row.ID,
			Name:        row.Name,
			Models:      row.Models,
			Status:      row.Status,
			StatusLabel: statusLabel(row.Status),
			UsedQuota:   row.UsedQuota,
			UsedAmount:  apiplatform.QuotaToUSD(row.UsedQuota, perUnit),
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": out})
}

type saveChannelItem struct {
	ChannelId int    `json:"channel_id"`
	Name      string `json:"name"`
	Models    string `json:"models"`
	Status    int    `json:"status"`
	UsedQuota int64  `json:"used_quota"`
}

type saveChannelsRequest struct {
	FilterName string            `json:"filter_name"`
	CostPrice  *float64          `json:"cost_price"`
	SalePrice  *float64          `json:"sale_price"`
	Channels   []saveChannelItem `json:"channels"`
}

func SaveApiPlatformChannels(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	var req saveChannelsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Channels) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请勾选渠道"})
		return
	}
	perUnit, err := apiplatform.QuotaPerUnit(platform.BaseURL, platform.APIKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	filterName := strings.TrimSpace(req.FilterName)
	for _, item := range req.Channels {
		if item.ChannelId <= 0 {
			continue
		}
		used := apiplatform.QuotaToUSD(item.UsedQuota, perUnit)
		incoming := model.ApiPlatformChannel{
			ChannelId:       item.ChannelId,
			Name:            item.Name,
			FilterName:      filterName,
			Models:          item.Models,
			Status:          item.Status,
			UsedQuota:       item.UsedQuota,
			UsedAmount:      used,
			UnsettledAmount: used,
			CostPrice:       req.CostPrice,
			SalePrice:       req.SalePrice,
		}
		if err := model.SavePulledChannel(id, incoming); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存渠道失败: " + err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

func SyncApiPlatformChannels(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	updated, failed, err := apiplatform.SyncPlatform(platform)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": gin.H{"updated": updated, "failed": failed}})
}

type channelMoneyRequest struct {
	CostPrice     *float64 `json:"cost_price"`
	SalePrice     *float64 `json:"sale_price"`
	SettledAmount *float64 `json:"settled_amount"`
}

func UpdateApiPlatformChannelMoney(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	rowID, err := strconv.Atoi(c.Param("channelRowId"))
	if err != nil || rowID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "渠道 ID 无效"})
		return
	}
	var req channelMoneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := model.UpdateApiPlatformChannelMoney(id, rowID, req.CostPrice, req.SalePrice, req.SettledAmount); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "渠道不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

type batchChannelPriceRequest struct {
	ChannelRowIDs []int    `json:"channel_row_ids"`
	CostPrice     *float64 `json:"cost_price"`
	SalePrice     *float64 `json:"sale_price"`
}

func BatchUpdateApiPlatformChannelPrices(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	var req batchChannelPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	updated, err := model.BatchUpdateApiPlatformChannelPrices(id, req.ChannelRowIDs, req.CostPrice, req.SalePrice)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": gin.H{"updated": updated}})
}

type testChannelRequest struct {
	Model string `json:"model"`
}

func TestApiPlatformChannel(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	rowID, err := strconv.Atoi(c.Param("channelRowId"))
	if err != nil || rowID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "渠道 ID 无效"})
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	channel, err := model.GetApiPlatformChannel(id, rowID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "渠道不存在"})
		return
	}
	var req testChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Model) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择模型"})
		return
	}
	code, message, err := apiplatform.TestSavedChannel(platform, channel, strings.TrimSpace(req.Model))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": gin.H{
		"status_code": code,
		"message":     message,
		"model":       strings.TrimSpace(req.Model),
	}})
}

type updateModelsRequest struct {
	Models []string `json:"models"`
}

func UpdateApiPlatformChannelModels(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	rowID, err := strconv.Atoi(c.Param("channelRowId"))
	if err != nil || rowID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "渠道 ID 无效"})
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	channel, err := model.GetApiPlatformChannel(id, rowID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "渠道不存在"})
		return
	}
	var req updateModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	joined, err := apiplatform.ReplaceChannelModels(platform, channel, req.Models)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": gin.H{"models": joined}})
}

type channelStatusRequest struct {
	Status int `json:"status"`
}

func UpdateApiPlatformChannelStatus(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	rowID, err := strconv.Atoi(c.Param("channelRowId"))
	if err != nil || rowID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "渠道 ID 无效"})
		return
	}
	var req channelStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != 1 && req.Status != 2) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "状态无效"})
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	channel, err := model.GetApiPlatformChannel(id, rowID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "渠道不存在"})
		return
	}
	if err := apiplatform.SetSavedChannelStatus(platform, channel, req.Status); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

type billView struct {
	model.ApiPlatformBill
	StatusLabel string `json:"status_label"`
}

func ListApiPlatformBills(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	rows, err := model.ListApiPlatformBills(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取账单失败: " + err.Error()})
		return
	}
	views := make([]billView, 0, len(rows))
	for _, row := range rows {
		views = append(views, billView{ApiPlatformBill: row, StatusLabel: billStatusLabel(row.Status)})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": views})
}

func GetApiPlatformBill(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	billID, err := strconv.Atoi(c.Param("billId"))
	if err != nil || billID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "账单 ID 无效"})
		return
	}
	row, err := model.GetApiPlatformBill(id, billID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "账单不存在"})
		return
	}
	platform, err := model.GetApiPlatform(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "台子不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": gin.H{
		"bill":         row,
		"status_label": billStatusLabel(row.Status),
		"platform": gin.H{
			"id":       platform.Id,
			"base_url": platform.BaseURL,
			"account":  platform.Account,
			"remark":   platform.Remark,
		},
	}})
}

type createBillRequest struct {
	ChannelRowIDs []int `json:"channel_row_ids"`
}

func CreateApiPlatformBill(c *gin.Context) {
	id, ok := platformID(c)
	if !ok {
		return
	}
	var req createBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	bill, err := model.CreateApiPlatformBill(id, req.ChannelRowIDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": bill})
}

func PayApiPlatformBill(c *gin.Context) {
	id, billID, ok := billPath(c)
	if !ok {
		return
	}
	if err := model.PayApiPlatformBill(id, billID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

func VoidApiPlatformBill(c *gin.Context) {
	id, billID, ok := billPath(c)
	if !ok {
		return
	}
	if err := model.VoidApiPlatformBill(id, billID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

func billPath(c *gin.Context) (int, int, bool) {
	id, ok := platformID(c)
	if !ok {
		return 0, 0, false
	}
	billID, err := strconv.Atoi(c.Param("billId"))
	if err != nil || billID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "账单 ID 无效"})
		return 0, 0, false
	}
	return id, billID, true
}

func GetAdminAPIPlatformSettings(c *gin.Context) {
	cfg, err := model.GetAPIPlatformSettings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取配置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": cfg})
}

func UpdateAdminAPIPlatformSettings(c *gin.Context) {
	var cfg model.APIPlatformSettings
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := model.SaveAPIPlatformSettings(cfg); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存失败: " + err.Error()})
		return
	}
	saved, err := model.GetAPIPlatformSettings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功", "data": saved})
}
