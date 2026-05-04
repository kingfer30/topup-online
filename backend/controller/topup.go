package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/supplier"
	"github.com/kingfer30/topup-online/utils/logger"
	"gorm.io/gorm"
)

type verifyCdkReq struct {
	CdkKey string `json:"cdk_key" binding:"required"`
}

// VerifyCdk 验证CDK有效性
func VerifyCdk(c *gin.Context) {
	var req verifyCdkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	cdk, err := model.GetGptCdkByKey(req.CdkKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "CDK不存在"})
		return
	}

	if cdk.UseStatus != 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "CDK已被使用或占用中"})
		return
	}

	now := time.Now().UnixMilli()
	if cdk.ExpireTime != nil && *cdk.ExpireTime > 0 && *cdk.ExpireTime < now {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "CDK已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "CDK有效",
		"data": gin.H{
			"valid":       true,
			"cdk_id":      cdk.Id,
			"expire_time": cdk.ExpireTime,
		},
	})
}

type startTopupReq struct {
	CdkKey       string `json:"cdk_key" binding:"required"`
	UserEmail    string `json:"user_email" binding:"required"`
	UserGptToken string `json:"user_gpt_token" binding:"required"`
	FullAuthData string `json:"full_auth_data"`
	Supplier     string `json:"supplier"`
}

// StartTopup 开始充值主流程
func StartTopup(c *gin.Context) {
	var req startTopupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	supplierName := req.Supplier
	if supplierName == "" {
		names := supplier.Names()
		if len(names) == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "暂无可用供应商"})
			return
		}
		supplierName = names[0]
	}

	drv, ok := supplier.Get(supplierName)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "供应商不存在: " + supplierName})
		return
	}

	var (
		task model.GptTopupTask
		card *model.GptCard
		cdk  *model.GptCdk
	)

	// 数据库事务：锁CDK、随机取卡密、创建任务
	txErr := model.DB.Transaction(func(tx *gorm.DB) error {
		var err error

		// 加锁查询CDK
		cdk, err = model.GetGptCdkByKeyForUpdate(tx, req.CdkKey)
		if err != nil {
			return err
		}
		if cdk.UseStatus != 1 {
			return gorm.ErrRecordNotFound
		}

		// 随机取一张待使用的卡密并加锁
		card, err = model.PickRandomAvailableGptCard(tx)
		if err != nil {
			return err
		}

		// 占用CDK
		if err = model.OccupyCdk(tx, cdk.Id); err != nil {
			return err
		}

		// 占用卡密
		if err = model.OccupyGptCard(tx, card.Id); err != nil {
			return err
		}

		// 创建任务记录
		task = model.GptTopupTask{
			CdkId:     cdk.Id,
			CdkKey:    cdk.Key,
			CardId:    &card.Id,
			UserEmail: req.UserEmail,
			Status:    1,
		}
		return tx.Create(&task).Error
	})

	if txErr != nil {
		if cdk != nil && cdk.UseStatus != 1 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "CDK已被使用或占用中，请勿重复提交"})
			return
		}
		if card == nil {
			c.JSON(http.StatusOK, gin.H{"code": 503, "message": "暂无可用卡密，请稍后重试"})
			return
		}
		logger.SysLogf("[topup] StartTopup 事务失败: %v", txErr)
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "服务异常，请稍后重试"})
		return
	}

	// 调用供应商API发起充值
	productId := 3
	topupResult, topupErr := drv.TopUp(supplier.TopupParam{
		CardInfo:     card.Key,
		UserEmail:    req.UserEmail,
		UserGptToken: req.UserGptToken,
		FullAuthData: req.FullAuthData,
		ProductId:    &productId,
	})
	if topupErr != nil {
		logger.SysLogf("[topup] StartTopup 调用供应商失败 taskId=%d err=%v", task.Id, topupErr)
		model.ReleaseCdk(cdk.Id)
		model.ReleaseGptCard(card.Id)
		model.UpdateTopupTask(task.Id, map[string]interface{}{
			"status":  3,
			"message": topupErr.Error(),
		})
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "充值请求失败: " + topupErr.Error()})
		return
	}

	// 更新任务的供应商任务ID
	model.UpdateTopupTask(task.Id, map[string]interface{}{
		"supplier_task_id": topupResult.TaskId,
	})

	// 启动后台goroutine轮询供应商任务状态
	go pollUntilDone(task.Id, cdk.Id, card.Id, card.Key, topupResult.TaskId, supplierName, productId)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "充值任务已创建",
		"data": gin.H{
			"task_id": task.Id,
		},
	})
}

// pollUntilDone 后台goroutine，轮询供应商任务直到成功或超时
func pollUntilDone(taskId, cdkId, cardId int, cardKey, supplierTaskId, supplierName string, productId int) {
	maxRetry := 30
	interval := 5 * time.Second

	drv, ok := supplier.Get(supplierName)
	if !ok {
		logger.SysLogf("[topup] pollUntilDone 供应商不存在: %s", supplierName)
		rollbackTopup(taskId, cdkId, cardId, "供应商不存在")
		return
	}

	for i := 0; i < maxRetry; i++ {
		time.Sleep(interval)

		result, err := drv.QueryTaskStatus(supplierTaskId, productId, cardKey)
		if err != nil {
			logger.SysLogf("[topup] pollUntilDone taskId=%d 查询状态失败 err=%v", taskId, err)
			continue
		}

		logger.SysLogf("[topup] pollUntilDone taskId=%d status=%s message=%s", taskId, result.Status, result.Message)

		switch result.Status {
		case "success":
			now := time.Now().UnixMilli()
			nowTime := time.Now()
			model.CompleteCdk(cdkId, cardId, now)
			model.CompleteGptCard(cardId, now, result.Message)
			model.UpdateTopupTask(taskId, map[string]interface{}{
				"status":       2,
				"message":      result.Message,
				"completed_at": nowTime,
			})
			logger.SysLogf("[topup] pollUntilDone taskId=%d 充值成功", taskId)
			return
		case "failed":
			rollbackTopup(taskId, cdkId, cardId, result.Message)
			return
		}
	}

	// 超时回滚
	logger.SysLogf("[topup] pollUntilDone taskId=%d 轮询超时，自动回滚", taskId)
	rollbackTopup(taskId, cdkId, cardId, "充值超时，请联系客服")
}

// rollbackTopup 回滚CDK和卡密状态，将任务标记为失败
func rollbackTopup(taskId, cdkId, cardId int, message string) {
	model.ReleaseCdk(cdkId)
	model.ReleaseGptCard(cardId)
	now := time.Now()
	model.UpdateTopupTask(taskId, map[string]interface{}{
		"status":       3,
		"message":      message,
		"completed_at": now,
	})
}

// GetTopupTaskStatus 查询充值任务状态（前端轮询接口）
func GetTopupTaskStatus(c *gin.Context) {
	taskIdStr := c.Param("task_id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil || taskId <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "任务ID无效"})
		return
	}

	task, err := model.GetTopupTaskById(taskId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"task_id":    task.Id,
			"status":     task.Status,
			"message":    task.Message,
			"created_at": task.CreatedAt,
		},
	})
}
