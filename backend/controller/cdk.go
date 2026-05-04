package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	redis "github.com/kingfer30/topup-online/config/cache"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/supplier"
)

// 验证卡密（委托给供应商驱动，默认使用三川）
func VerifyCard(c *gin.Context) {
	param := constants.CDKVerifyRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}

	drv, ok := supplier.Get("三川")
	if !ok {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "supplier not found"})
		return
	}

	if err := drv.VerifyCard(param.CardInfo); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "success",
	})
}

// 验证卡密并充值（委托给供应商驱动）
func TopUp(c *gin.Context) {
	param := constants.CDKTopupRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}

	// 防重复提交：以 userEmail 为维度限流
	lockKey := fmt.Sprintf("Thread:%s", param.UserEmail)
	if count, serr := redis.Exists(lockKey); serr != nil || count == 0 {
		if ok, err := redis.SetNx(lockKey, "1", time.Duration(constants.GetCacheFrequency())*time.Second); ok || err == nil {
			drv, ok := supplier.Get("三川")
			if !ok {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": "supplier not found"})
				return
			}
			result, err := drv.TopUp(supplier.TopupParam{
				CardInfo:     param.CardInfo,
				UserEmail:    param.UserEmail,
				UserGptToken: param.UserGptToken,
				FullAuthData: param.FullAuthData,
				ProductId:    param.ProductId,
			})
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
					"data":    "",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "success",
				"data": gin.H{
					"taskId":       result.TaskId,
					"processing":   result.Processing,
					"needsPolling": result.NeedsPolling,
					"message":      result.Message,
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "Please do not submit repeated requests",
	})
}

// 查询充值任务状态（委托给供应商驱动）
func QueryTaskStatus(c *gin.Context) {
	param := constants.CDKQueryTaskRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}

	drv, ok := supplier.Get("三川")
	if !ok {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "supplier not found"})
		return
	}

	result, err := drv.QueryTaskStatus(param.TaskId, param.ProductId, param.CardInfo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    "",
		})
		return
	}

	switch result.Status {
	case "success":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": result.Message,
			"data":    gin.H{"status": "success"},
		})
	case "processing":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": result.Message,
			"data":    gin.H{"status": "processing"},
		})
	case "failed":
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": result.Message,
			"data":    gin.H{"status": "failed"},
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "unknown task status: " + result.Status,
			"data":    "",
		})
	}
}
