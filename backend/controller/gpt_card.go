package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/supplier"
	"github.com/kingfer30/topup-online/utils/logger"
)

// GetSuppliers 获取已注册的供应商列表
func GetSuppliers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    supplier.Names(),
	})
}

// GetGptCardList 获取GPT卡密列表
func GetGptCardList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	supplierName := c.Query("supplier")
	statusStr := c.DefaultQuery("status", "0")
	keyword := strings.TrimSpace(c.Query("keyword"))

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	status, _ := strconv.Atoi(statusStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	cards, total, err := model.GetGptCardList(page, pageSize, supplierName, status, keyword)
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
			"list":  cards,
			"total": total,
		},
	})
}

// BatchImportGptCardsRequest 批量导入请求
type BatchImportGptCardsRequest struct {
	Supplier string   `json:"supplier" binding:"required"`
	Keys     []string `json:"keys" binding:"required"`
}

// BatchImportGptCards 批量导入GPT卡密
func BatchImportGptCards(c *gin.Context) {
	var req BatchImportGptCardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if _, ok := supplier.Get(req.Supplier); !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的供应商: " + req.Supplier,
		})
		return
	}

	now := time.Now().Unix()
	var cards []model.GptCard
	for _, key := range req.Keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		cards = append(cards, model.GptCard{
			Supplier:   req.Supplier,
			Key:        key,
			Status:     1,
			ImportTime: &now,
		})
	}

	if len(cards) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "没有有效的卡密",
		})
		return
	}

	if err := model.BatchCreateGptCards(cards); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "导入失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    gin.H{"imported": len(cards)},
	})
}

// BatchCheckGptCardsRequest 批量检查请求
type BatchCheckGptCardsRequest struct {
	Ids []int `json:"ids" binding:"required"`
}

// CheckCardResult 单条检查结果
type CheckCardResult struct {
	Id        int    `json:"id"`
	Key       string `json:"key"`
	OldStatus int    `json:"old_status"`
	NewStatus int    `json:"new_status"`
	Changed   bool   `json:"changed"`
	Message   string `json:"message"`
}

// BatchCheckGptCards 批量检查卡密有效性并同步状态
// 规则：
//   - verify 成功（卡密有效/待使用）且 DB 状态为 2（已使用）→ 修正为 1（待使用）
//   - verify 失败（卡密已使用/无效）且 DB 状态为 1（待使用）→ 修正为 2（已使用）
func BatchCheckGptCards(c *gin.Context) {
	var req BatchCheckGptCardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	cards, err := model.GetGptCardsByIds(req.Ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	results := make([]CheckCardResult, 0, len(cards))
	updatedCount := 0

	for _, card := range cards {
		result := CheckCardResult{
			Id:        card.Id,
			Key:       card.Key,
			OldStatus: card.Status,
			NewStatus: card.Status,
		}

		// 跳过已作废的卡密
		if card.Status == -1 {
			result.Message = "已作废，跳过"
			results = append(results, result)
			continue
		}

		drv, ok := supplier.Get(card.Supplier)
		if !ok {
			result.Message = "未找到供应商: " + card.Supplier
			results = append(results, result)
			continue
		}

		verifyErr := drv.VerifyCard(card.Key)

		if verifyErr == nil {
			// 卡密有效（待使用），若 DB 标记为已使用则修正
			if card.Status == 2 {
				if updateErr := model.UpdateGptCard(card.Id, map[string]interface{}{"status": 1}); updateErr != nil {
					result.Message = "更新失败: " + updateErr.Error()
				} else {
					result.NewStatus = 1
					result.Changed = true
					result.Message = "卡密有效，已修正为待使用"
					updatedCount++
				}
			} else {
				result.Message = "卡密有效，状态正确"
			}
		} else {
			// 卡密已使用/无效，若 DB 标记为待使用则修正
			logger.SysLogf("[BatchCheckGptCards] card id=%d verify fail: %v", card.Id, verifyErr)
			if card.Status == 1 {
				if updateErr := model.UpdateGptCard(card.Id, map[string]interface{}{"status": 2}); updateErr != nil {
					result.Message = "更新失败: " + updateErr.Error()
				} else {
					result.NewStatus = 2
					result.Changed = true
					result.Message = "卡密已使用，已修正为已使用"
					updatedCount++
				}
			} else {
				result.Message = "卡密已使用，状态正确"
			}
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"checked": len(cards),
			"updated": updatedCount,
			"results": results,
		},
	})
}

// UpdateGptCardRequest 更新请求
type UpdateGptCardRequest struct {
	GptMail   string `json:"gpt_mail"`
	Buyer     string `json:"buyer"`
	Status    *int   `json:"status"`
	SubResult string `json:"sub_result"`
}

// UpdateGptCard 更新GPT卡密
func UpdateGptCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var req UpdateGptCardRequest
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
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.SubResult != "" {
		updates["sub_result"] = req.SubResult
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "没有需要更新的字段"})
		return
	}

	if err := model.UpdateGptCard(id, updates); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// DeleteGptCard 删除GPT卡密
func DeleteGptCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := model.DeleteGptCard(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// BatchDeleteGptCardsRequest 批量删除请求
type BatchDeleteGptCardsRequest struct {
	Ids []int `json:"ids" binding:"required"`
}

// BatchDeleteGptCards 批量删除GPT卡密
func BatchDeleteGptCards(c *gin.Context) {
	var req BatchDeleteGptCardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := model.BatchDeleteGptCards(req.Ids); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}
