package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

// GetDigisellerPrices 获取所有Digiseller订阅类型售价配置
func GetDigisellerPrices(c *gin.Context) {
	prices, err := model.GetAllDigisellerPrices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取价格配置失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    prices,
	})
}

// UpsertDigisellerPriceRequest 设置售价请求
type UpsertDigisellerPriceRequest struct {
	SubscriptionType string  `json:"subscription_type" binding:"required"`
	Price            float64 `json:"price" binding:"required,min=0"`
}

// UpsertDigisellerPrice 新增或更新某订阅类型的今日售价
func UpsertDigisellerPrice(c *gin.Context) {
	var req UpsertDigisellerPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := model.UpsertDigisellerPrice(req.SubscriptionType, req.Price); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存价格配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
	})
}
