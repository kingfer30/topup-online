package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

// GetSalesTalkList 获取话术列表
func GetSalesTalkList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	tag := c.Query("tag")

	list, total, err := model.GetSalesTalkList(page, pageSize, keyword, tag)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetSalesTalkById 获取话术详情
func GetSalesTalkById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	item, err := model.GetSalesTalkById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "话术不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    item,
	})
}

// CreateSalesTalk 创建话术
func CreateSalesTalk(c *gin.Context) {
	var item model.SalesTalk
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := item.Create(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    item,
	})
}

// UpdateSalesTalk 更新话术
func UpdateSalesTalk(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var item model.SalesTalk
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	item.Id = id

	if err := item.Update(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    item,
	})
}

// BatchUpdateSalesTalkTag 批量更新标签
func BatchUpdateSalesTalkTag(c *gin.Context) {
	var req struct {
		Ids []int64 `json:"ids" binding:"required"`
		Tag string  `json:"tag"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if len(req.Ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请选择要更新的记录",
		})
		return
	}

	if err := model.BatchUpdateSalesTalkTag(req.Ids, req.Tag); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "批量更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量更新成功",
	})
}

// DeleteSalesTalk 删除话术
func DeleteSalesTalk(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	if err := model.DeleteSalesTalkById(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

