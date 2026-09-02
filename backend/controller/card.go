package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/scheduler"
	"github.com/kingfer30/topup-online/utils/logger"
)

// CardRequest 卡密请求结构
type CardRequest struct {
	Category                string   `json:"category"`
	Account                 string   `json:"account" binding:"required"`
	Password                string   `json:"password"`
	MailPassword            string   `json:"mail_password"`
	SubscriptionStatus      int      `json:"subscription_status"`
	SubscriptionType        string   `json:"subscription_type"`
	SubscriptionTime        *int64   `json:"subscription_time"`
	SubscriptionExpiredTime *int64   `json:"subscription_expired_time"`
	PurchaseDate            *int64   `json:"purchase_date"`
	PurchasePrice           *float64 `json:"purchase_price"`
	PurchaseFrom            string   `json:"purchase_from"`
	PurchaseBy              string   `json:"purchase_by"`
	SellPrice               *float64 `json:"sell_price"`
	SellDate                *int64   `json:"sell_date"`
	SellTo                  string   `json:"sell_to"`
	SellOrderNo             string   `json:"sell_order_no"`
	SellStatus              int      `json:"sell_status"`
	AccountType             int      `json:"account_type"`
	Status                  int      `json:"status"`
	ApiKey                  string   `json:"api_key"`
	Token                   string   `json:"token"`
	TwoFA                   string   `json:"2fa"`
	ClientId                string   `json:"client_id"`
	MailUrl                 string   `json:"mail_url"`
	Remark                  string   `json:"remark"`
	CodeLink                string   `json:"code_link"`
	Phone                   string   `json:"phone"`
	PhoneLink               string   `json:"phone_link"`
	SubscriptionCredits     *float64 `json:"subscription_credits"`
}

// parseAccountsFromQuery 解析账号搜索参数（accounts 优先，兼容 keyword）
func parseAccountsFromQuery(c *gin.Context) []string {
	if raw := strings.TrimSpace(c.Query("accounts")); raw != "" {
		return model.ParseAccountSearchList(raw)
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		return model.ParseAccountSearchList(kw)
	}
	return nil
}

// GetCardList 获取卡密列表
func GetCardList(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少卡密类别参数",
		})
		return
	}

	cardType := c.DefaultQuery("type", "all")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	accounts := parseAccountsFromQuery(c)
	subscriptionType := c.Query("subscription_type")
	subscriptionStatusStr := c.DefaultQuery("subscription_status", "0")
	isCheckStr := c.DefaultQuery("is_check", "0")
	purchaseDateStr := c.Query("purchase_date")
	freezeStatusStr := c.DefaultQuery("freeze_status", "0")
	freezeTimeStr := c.Query("freeze_time")
	sellTo := strings.TrimSpace(c.Query("sell_to"))
	purchaseBy := strings.TrimSpace(c.Query("purchase_by"))
	promo50OffStr := c.DefaultQuery("promo_50off", "0")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	subscriptionStatus, _ := strconv.Atoi(subscriptionStatusStr)
	isCheck, _ := strconv.Atoi(isCheckStr)
	freezeStatus, _ := strconv.Atoi(freezeStatusStr)
	promo50Off, _ := strconv.Atoi(promo50OffStr)

	// 将日期输入转换为当天零点 Unix 时间戳（UTC+8）
	cst := time.FixedZone("CST", 8*3600)
	parseDayStart := func(dateStr string) int64 {
		if dateStr == "" {
			return 0
		}
		if ts, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
			if ts > 1e12 {
				ts = ts / 1000
			}
			t := time.Unix(ts, 0).In(cst)
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
		}
		if t, err := time.ParseInLocation("2006-01-02", dateStr, cst); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, cst); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
		}
		return 0
	}
	purchaseDate := parseDayStart(purchaseDateStr)
	freezeTime := parseDayStart(freezeTimeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	// 查询卡密列表
	cards, total, err := model.GetCardList(tableName, cardType, page, pageSize, accounts, subscriptionType, subscriptionStatus, isCheck, purchaseDate, freezeStatus, freezeTime, sellTo, purchaseBy, promo50Off)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取卡密列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":      cards,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// BatchDeleteCards 批量删除卡密（status=-1 软删）
func BatchDeleteCards(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "未选择任何记录"})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	affected, err := model.BatchDeleteByStatus(tableName, req.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功删除 %d 条记录", affected),
		"data":    affected,
	})
}

// GetCardById 根据ID获取卡密详情
func GetCardById(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少卡密类别参数",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的卡密ID",
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	card, err := model.GetCardById(tableName, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "卡密不存在: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    card,
	})
}

// CreateCard 创建卡密
func CreateCard(c *gin.Context) {
	var req CardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if req.Category == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少卡密类别参数",
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(req.Category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	card := &model.AccountCard{
		Account:                 req.Account,
		Password:                req.Password,
		MailPassword:            req.MailPassword,
		SubscriptionStatus:      req.SubscriptionStatus,
		SubscriptionType:        req.SubscriptionType,
		SubscriptionTime:        req.SubscriptionTime,
		SubscriptionExpiredTime: req.SubscriptionExpiredTime,
		PurchaseDate:            req.PurchaseDate,
		PurchasePrice:           req.PurchasePrice,
		PurchaseFrom:            req.PurchaseFrom,
		PurchaseBy:              req.PurchaseBy,
		SellPrice:               req.SellPrice,
		SellDate:                req.SellDate,
		SellTo:                  req.SellTo,
		SellOrderNo:             req.SellOrderNo,
		SellStatus:              req.SellStatus,
		AccountType:             req.AccountType,
		Status:                  req.Status,
		ApiKey:                  req.ApiKey,
		Token:                   req.Token,
		TwoFA:                   req.TwoFA,
		ClientId:                req.ClientId,
		MailUrl:                 req.MailUrl,
		Remark:                  req.Remark,
		CodeLink:                req.CodeLink,
		Phone:                   req.Phone,
		PhoneLink:               req.PhoneLink,
		SubscriptionCredits:     req.SubscriptionCredits,
	}

	// 设置默认值
	if card.SubscriptionStatus == 0 {
		card.SubscriptionStatus = 1
	}
	if card.SellStatus == 0 {
		card.SellStatus = 1
	}
	if card.AccountType == 0 {
		card.AccountType = 1
	}
	if card.Status == 0 {
		card.Status = 1
	}

	if err := model.CreateCard(tableName, card); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建卡密失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    card,
	})
}

// UpdateCard 更新卡密
func UpdateCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的卡密ID",
		})
		return
	}

	var req CardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if req.Category == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少卡密类别参数",
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(req.Category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	card := &model.AccountCard{
		Account:                 req.Account,
		Password:                req.Password,
		MailPassword:            req.MailPassword,
		SubscriptionStatus:      req.SubscriptionStatus,
		SubscriptionType:        req.SubscriptionType,
		SubscriptionTime:        req.SubscriptionTime,
		SubscriptionExpiredTime: req.SubscriptionExpiredTime,
		PurchaseDate:            req.PurchaseDate,
		PurchasePrice:           req.PurchasePrice,
		PurchaseFrom:            req.PurchaseFrom,
		PurchaseBy:              req.PurchaseBy,
		SellPrice:               req.SellPrice,
		SellDate:                req.SellDate,
		SellTo:                  req.SellTo,
		SellOrderNo:             req.SellOrderNo,
		SellStatus:              req.SellStatus,
		AccountType:             req.AccountType,
		Status:                  req.Status,
		ApiKey:                  req.ApiKey,
		Token:                   req.Token,
		TwoFA:                   req.TwoFA,
		ClientId:                req.ClientId,
		MailUrl:                 req.MailUrl,
		Remark:                  req.Remark,
		CodeLink:                req.CodeLink,
		Phone:                   req.Phone,
		PhoneLink:               req.PhoneLink,
		SubscriptionCredits:     req.SubscriptionCredits,
	}

	if err := model.UpdateCard(tableName, id, card); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新卡密失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    card,
	})
}

// DeleteCard 删除卡密
func DeleteCard(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少卡密类别参数",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的卡密ID",
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	if err := model.DeleteCard(tableName, id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "删除卡密失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// BatchImportCards 批量导入卡密
func BatchImportCards(c *gin.Context) {
	var req struct {
		Category string        `json:"category" binding:"required"`
		Cards    []CardRequest `json:"cards" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(req.Category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	// 转换为 AccountCard 模型
	cards := make([]model.AccountCard, 0, len(req.Cards))
	for _, cardReq := range req.Cards {
		card := model.AccountCard{
			Account:                 cardReq.Account,
			Password:                cardReq.Password,
			MailPassword:            cardReq.MailPassword,
			SubscriptionStatus:      cardReq.SubscriptionStatus,
			SubscriptionType:        cardReq.SubscriptionType,
			SubscriptionTime:        cardReq.SubscriptionTime,
			SubscriptionExpiredTime: cardReq.SubscriptionExpiredTime,
			PurchaseDate:            cardReq.PurchaseDate,
			PurchasePrice:           cardReq.PurchasePrice,
			PurchaseFrom:            cardReq.PurchaseFrom,
			PurchaseBy:              cardReq.PurchaseBy,
			SellPrice:               cardReq.SellPrice,
			SellDate:                cardReq.SellDate,
			SellTo:                  cardReq.SellTo,
			SellOrderNo:             cardReq.SellOrderNo,
			SellStatus:              cardReq.SellStatus,
			AccountType:             cardReq.AccountType,
			Status:                  cardReq.Status,
			ApiKey:                  cardReq.ApiKey,
			Token:                   cardReq.Token,
			TwoFA:                   cardReq.TwoFA,
			ClientId:                cardReq.ClientId,
			MailUrl:                 cardReq.MailUrl,
			Remark:                  cardReq.Remark,
			CodeLink:                cardReq.CodeLink,
			Phone:                   cardReq.Phone,
			PhoneLink:               cardReq.PhoneLink,
			SubscriptionCredits:     cardReq.SubscriptionCredits,
		}

		// 设置默认值
		if card.SubscriptionStatus == 0 {
			card.SubscriptionStatus = 1
		}
		if card.SellStatus == 0 {
			card.SellStatus = 1
		}
		if card.AccountType == 0 {
			card.AccountType = 1
		}
		if card.Status == 0 {
			card.Status = 1
		}

		cards = append(cards, card)
	}

	if err := model.BatchCreateCards(tableName, cards); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "批量导入失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量导入成功",
	})
}

// CardTableInfo 卡密表信息（表名 + 类别名）
type CardTableInfo struct {
	Table    string `json:"table"`
	Category string `json:"category"`
}

// GetCardTableNames 获取所有 cards_* 表名（供关联下拉选择使用）
func GetCardTableNames(c *gin.Context) {
	tables, err := model.GetAllCardTableNames()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取表列表失败: " + err.Error(),
		})
		return
	}

	list := make([]CardTableInfo, 0, len(tables))
	for _, t := range tables {
		list = append(list, CardTableInfo{
			Table:    t,
			Category: strings.TrimPrefix(t, "cards_"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    list,
	})
}

// GetUnsoldSubscriptionTypes 获取未售出的订阅类型列表
func GetUnsoldSubscriptionTypes(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少卡密类别参数",
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	types, err := model.GetUnsoldSubscriptionTypes(tableName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取订阅类型失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    types,
	})
}

// PickupCard 取货
func PickupCard(c *gin.Context) {
	var req struct {
		Category         string `json:"category" binding:"required"`
		SubscriptionType string `json:"subscription_type" binding:"required"`
		Format           string `json:"format"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(req.Category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	card, err := model.PickupCard(tableName, req.SubscriptionType, req.Format)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "取货成功",
		"data":    card,
	})
}

// CompletePickup 完成取货
func CompletePickup(c *gin.Context) {
	var req struct {
		Category  string   `json:"category" binding:"required"`
		Id        int      `json:"id" binding:"required"`
		SellPrice *float64 `json:"sell_price"`
		SellTo    string   `json:"sell_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(req.Category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	err := model.CompletePickup(tableName, req.Id, req.SellPrice, req.SellTo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "完成取货失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "完成取货成功",
	})
}

// RollbackSoldCard 回滚已售：将已出售状态重置为未出售
func RollbackSoldCard(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		Id       int    `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)

	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	err := model.RollbackSoldCard(tableName, req.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "回滚失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "回滚成功",
	})
}

// BatchCheckCards 批量检查卡密订阅状态（逻辑同定时任务）
func BatchCheckCards(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择要检查的卡密"})
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

	go func() {
		successCount := 0
		failCount := 0
		for _, card := range cards {
			if err := scheduler.CheckSingleCard(tableName, card); err != nil {
				logger.SysError(fmt.Sprintf("批量检查卡密 [ID:%d, Account:%s] 失败: %v", card.Id, card.Account, err))
				failCount++
			} else {
				successCount++
			}
			time.Sleep(1 * time.Second)
		}
		logger.SysLog(fmt.Sprintf("批量检查完成，成功: %d, 失败: %d", successCount, failCount))
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("已提交 %d 张卡密检查任务，正在后台执行", len(cards)),
		"data":    len(cards),
	})
}

// PollCardSubscription 只查询当前订阅类型，不改卡密状态
func PollCardSubscription(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		ID       int    `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	card, err := model.GetCardById(tableName, req.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询卡密失败: " + err.Error()})
		return
	}
	if card.Token == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该卡密没有 Token"})
		return
	}

	membership, err := scheduler.PeekCursorMembership(card.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询订阅失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"subscription_type": membership,
			"subscribed":        scheduler.IsPaidMembership(membership),
		},
	})
}

// GetDashboardStats 获取控制台统计数据（按卡密类型统计销售，支持按日期查询）
func GetDashboardStats(c *gin.Context) {
	// date 参数格式 YYYY-MM-DD，不传则默认今天
	dateStr := c.Query("date")

	stats, err := model.GetDashboardStats(dateStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "获取统计数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// BatchPickup 批量取货：将勾选的未售卡密直接标记为已出售
func BatchPickup(c *gin.Context) {
	var req struct {
		Category  string   `json:"category" binding:"required"`
		IDs       []int    `json:"ids" binding:"required"`
		SellPrice *float64 `json:"sell_price"`
		SellTo    string   `json:"sell_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	affected, err := model.BatchPickup(tableName, req.IDs, req.SellPrice, req.SellTo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "批量取货失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功完成 %d 条取货", affected),
		"data":    affected,
	})
}

// ExportCards 导出全部符合条件的卡密（不分页）
func ExportCards(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "缺少卡密类别参数"})
		return
	}

	cardType := c.DefaultQuery("type", "all")
	accounts := parseAccountsFromQuery(c)
	subscriptionType := c.Query("subscription_type")
	subscriptionStatusStr := c.DefaultQuery("subscription_status", "0")
	isCheckStr := c.DefaultQuery("is_check", "0")
	purchaseDateStr := c.Query("purchase_date")
	exportFreezeStatus, _ := strconv.Atoi(c.DefaultQuery("freeze_status", "0"))
	freezeTimeStr := c.Query("freeze_time")
	sellTo := strings.TrimSpace(c.Query("sell_to"))
	purchaseBy := strings.TrimSpace(c.Query("purchase_by"))
	subscriptionStatus, _ := strconv.Atoi(subscriptionStatusStr)
	isCheck, _ := strconv.Atoi(isCheckStr)
	promo50Off, _ := strconv.Atoi(c.DefaultQuery("promo_50off", "0"))

	// purchase_date / freeze_time：统一解析到当天零点（UTC+8）
	cst := time.FixedZone("CST", 8*3600)
	parseDayStart := func(dateStr string) int64 {
		if dateStr == "" {
			return 0
		}
		if ts, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
			if ts > 1e12 {
				ts = ts / 1000
			}
			t := time.Unix(ts, 0).In(cst)
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
		}
		if t, err := time.ParseInLocation("2006-01-02", dateStr, cst); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, cst); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
		}
		return 0
	}
	purchaseDate := parseDayStart(purchaseDateStr)
	freezeTime := parseDayStart(freezeTimeStr)

	tableName := model.GetTableNameByCategory(category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	cards, err := model.GetAllCardsForExport(tableName, cardType, accounts, subscriptionType, subscriptionStatus, isCheck, purchaseDate, exportFreezeStatus, freezeTime, sellTo, purchaseBy, promo50Off)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "导出失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    cards,
	})
}

// BatchUpgradeToProduct 批量将普号升级为成品
func BatchUpgradeToProduct(c *gin.Context) {
	var req struct {
		Category                  string   `json:"category" binding:"required"`
		IDs                       []int    `json:"ids" binding:"required"`
		SubscriptionType          string   `json:"subscription_type"`
		SubscriptionTime          *int64   `json:"subscription_time"`
		SubscriptionRemainingDays *int     `json:"subscription_remaining_days"`
		PurchasePrice             *float64 `json:"purchase_price"`
		PurchaseFrom              string   `json:"purchase_from"`
		PurchaseDate              *int64   `json:"purchase_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)

	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	subscriptionType := strings.TrimSpace(req.SubscriptionType)
	if subscriptionType == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请选择订阅类型",
		})
		return
	}

	affected, err := model.BatchUpgradeToProduct(tableName, model.BatchUpgradeRequest{
		IDs:                       req.IDs,
		SubscriptionType:          subscriptionType,
		SubscriptionTime:          req.SubscriptionTime,
		SubscriptionRemainingDays: req.SubscriptionRemainingDays,
		PurchasePrice:             req.PurchasePrice,
		PurchaseFrom:              req.PurchaseFrom,
		PurchaseDate:              req.PurchaseDate,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功更新 %d 条记录", affected),
		"data":    affected,
	})
}

// RollbackPickup 回滚取货：将发货中状态重置为未出售
func RollbackPickup(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		Id       int    `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取表名
	tableName := model.GetTableNameByCategory(req.Category)

	// 检查表是否存在
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	err := model.RollbackPickup(tableName, req.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "回滚失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "回滚成功",
	})
}

// BatchDashboardGotoResolve 批量处理提链结果为 cursor dashboard 的卡密：仅掉订阅(-1)回滚已售并标记订阅状态 -2、检查成功
func BatchDashboardGotoResolve(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}
	success, errMsgs := model.BatchDashboardGotoAfterResolve(tableName, req.IDs)
	msg := fmt.Sprintf("已处理 %d 条（dashboard 回滚+标记）", success)
	if len(errMsgs) > 0 {
		msg += "；" + strings.Join(errMsgs, "；")
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": msg, "data": success})
}

// EnableOnDemandSpendHandler 为指定卡密开启按需付费
func EnableOnDemandSpendHandler(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		Id       int    `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	card, err := model.GetCardById(tableName, req.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "卡密不存在"})
		return
	}

	if card.Token == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该卡密没有 Token，无法开启按需付费"})
		return
	}

	workosID, err := scheduler.ExtractWorkosID(card.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "提取 workos_id 失败: " + err.Error()})
		return
	}

	if err := scheduler.EnableOnDemandSpend(workosID, card.Token); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "开启按需付费失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "按需付费已成功开启",
	})
}

// UpdateCardRemarkRequest 更新备注请求
type UpdateCardRemarkRequest struct {
	Category string `json:"category" binding:"required"`
	Id       int    `json:"id" binding:"required"`
	Remark   string `json:"remark"`
}

// UpdateCardRemark 单独更新卡密备注字段
func UpdateCardRemark(c *gin.Context) {
	var req UpdateCardRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "该卡密类别不存在",
		})
		return
	}

	if err := model.DB.Table(tableName).Where("id = ?", req.Id).Update("remark", req.Remark).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新备注失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "备注已更新",
	})
}

// BatchEnableOnDemandSpendHandler 批量为卡密开启按需付费
func BatchEnableOnDemandSpendHandler(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	successCount := 0
	failMessages := []string{}

	for _, id := range req.IDs {
		card, err := model.GetCardById(tableName, id)
		if err != nil || card.Token == "" {
			failMessages = append(failMessages, fmt.Sprintf("ID %d: token 为空或不存在", id))
			continue
		}
		workosID, err := scheduler.ExtractWorkosID(card.Token)
		if err != nil {
			failMessages = append(failMessages, fmt.Sprintf("ID %d: 提取 workos_id 失败", id))
			continue
		}
		if err := scheduler.EnableOnDemandSpend(workosID, card.Token); err != nil {
			failMessages = append(failMessages, fmt.Sprintf("ID %d: %v", id, err))
			continue
		}
		successCount++
	}

	msg := fmt.Sprintf("成功开启 %d 条", successCount)
	if len(failMessages) > 0 {
		msg += fmt.Sprintf("，失败 %d 条", len(failMessages))
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": msg, "data": successCount})
}

// BatchFreezeCards 批量冻结/解冻普号卡密
func BatchFreezeCards(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		IDs      []int  `json:"ids" binding:"required"`
		Freeze   int    `json:"freeze"` // 1=冻结 -1=解冻
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tableName := model.GetTableNameByCategory(req.Category)
	if !model.CheckTableExists(tableName) {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "该卡密类别不存在"})
		return
	}

	affected, err := model.BatchFreezeCards(tableName, req.IDs, req.Freeze, req.Remark)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "操作失败: " + err.Error()})
		return
	}

	action := "冻结"
	if req.Freeze == -1 {
		action = "解冻"
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功%s %d 条记录", action, affected),
		"data":    affected,
	})
}
