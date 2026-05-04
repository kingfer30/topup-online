package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/random"
)

// GetGptCdkList 获取GPT-CDK列表
func GetGptCdkList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	sellStatusStr := c.DefaultQuery("sell_status", "0")
	useStatusStr := c.DefaultQuery("use_status", "0")
	keyword := strings.TrimSpace(c.Query("keyword"))

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	sellStatus, _ := strconv.Atoi(sellStatusStr)
	useStatus, _ := strconv.Atoi(useStatusStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := model.GetGptCdkList(page, pageSize, sellStatus, useStatus, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  list,
			"total": total,
		},
	})
}

// BatchGenerateGptCdkRequest 批量生成CDK请求
type BatchGenerateGptCdkRequest struct {
	Count      int    `json:"count" binding:"required,min=1,max=500"`
	ExpireTime *int64 `json:"expire_time"`
}

// BatchGenerateGptCdk 批量生成GPT-CDK
func BatchGenerateGptCdk(c *gin.Context) {
	var req BatchGenerateGptCdkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	cdks := make([]model.GptCdk, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		key := time.Now().Format("20060102") + "-" + random.GetRandomString(24)
		cdk := model.GptCdk{
			Key:        key,
			SellStatus: 1,
			UseStatus:  1,
			ExpireTime: req.ExpireTime,
		}
		cdks = append(cdks, cdk)
	}

	if err := model.BatchCreateGptCdks(cdks); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "生成失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    gin.H{"generated": req.Count},
	})
}

// UpdateGptCdkRequest 更新CDK请求
type UpdateGptCdkRequest struct {
	GptMail    string `json:"gpt_mail"`
	Buyer      string `json:"buyer"`
	SellStatus *int   `json:"sell_status"`
	UseStatus  *int   `json:"use_status"`
	CardId     *int   `json:"card_id"`
	SubResult  string `json:"sub_result"`
	IpAddr     string `json:"ip_addr"`
	DeviceInfo string `json:"device_info"`
	ExpireTime *int64 `json:"expire_time"`
}

// UpdateGptCdk 更新GPT-CDK
func UpdateGptCdk(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var req UpdateGptCdkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.GptMail != "" {
		updates["gpt_mail"] = req.GptMail
	}
	if req.Buyer != "" {
		updates["buyer"] = req.Buyer
	}
	if req.SellStatus != nil {
		updates["sell_status"] = *req.SellStatus
	}
	if req.UseStatus != nil {
		updates["use_status"] = *req.UseStatus
	}
	if req.CardId != nil {
		updates["card_id"] = *req.CardId
	}
	if req.SubResult != "" {
		updates["sub_result"] = req.SubResult
	}
	if req.IpAddr != "" {
		updates["ip_addr"] = req.IpAddr
	}
	if req.DeviceInfo != "" {
		updates["device_info"] = req.DeviceInfo
	}
	if req.ExpireTime != nil {
		updates["expire_time"] = *req.ExpireTime
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateGptCdk(id, updates); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteGptCdkSingle 删除单条CDK
func DeleteGptCdkSingle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := model.DeleteGptCdk(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// BatchDeleteGptCdksRequest 批量删除CDK请求
type BatchDeleteGptCdksRequest struct {
	Ids []int `json:"ids" binding:"required"`
}

// BatchDeleteGptCdks 批量删除CDK
func BatchDeleteGptCdks(c *gin.Context) {
	var req BatchDeleteGptCdksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := model.BatchDeleteGptCdks(req.Ids); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}
