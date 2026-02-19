package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
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
	PurchaseOrderNo         string   `json:"purchase_order_no"`
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
	MailUrl                 string   `json:"mail_url"`
	Remark                  string   `json:"remark"`
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
	keyword := c.Query("keyword")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

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
	cards, total, err := model.GetCardList(tableName, cardType, page, pageSize, keyword)
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
		PurchaseOrderNo:         req.PurchaseOrderNo,
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
		MailUrl:                 req.MailUrl,
		Remark:                  req.Remark,
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
		PurchaseOrderNo:         req.PurchaseOrderNo,
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
		MailUrl:                 req.MailUrl,
		Remark:                  req.Remark,
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
			PurchaseOrderNo:         cardReq.PurchaseOrderNo,
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
			MailUrl:                 cardReq.MailUrl,
			Remark:                  cardReq.Remark,
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

	card, err := model.PickupCard(tableName, req.SubscriptionType)
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
