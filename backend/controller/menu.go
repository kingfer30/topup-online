package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

// MenuRequest 菜单请求结构
type MenuRequest struct {
	ParentId int    `json:"parent_id"`
	Title    string `json:"title" binding:"required"`
	Key      string `json:"key" binding:"required"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// CardMenuRequest 卡密菜单请求结构
type CardMenuRequest struct {
	Category string `json:"category" binding:"required"`
	MenuName string `json:"menuName" binding:"required"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
}

// GetMenuTree 获取菜单树结构
func GetMenuTree(c *gin.Context) {
	menus, err := model.GetMenuTree()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    menus,
	})
}

// GetAllMenus 获取所有菜单（扁平列表，用于管理）
func GetAllMenus(c *gin.Context) {
	menus, err := model.GetAllMenusForManagement()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    menus,
	})
}

// GetMenuById 根据ID获取菜单详情
func GetMenuById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的菜单ID",
		})
		return
	}

	menu, err := model.GetMenuById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "菜单不存在: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    menu,
	})
}

// CreateMenu 创建菜单
func CreateMenu(c *gin.Context) {
	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	menu := &model.Menu{
		ParentId: req.ParentId,
		Title:    req.Title,
		Key:      req.Key,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   req.Status,
	}

	// 如果没有设置状态，默认启用
	if menu.Status == 0 {
		menu.Status = 1
	}

	if err := menu.Create(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    menu,
	})
}

// UpdateMenu 更新菜单
func UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的菜单ID",
		})
		return
	}

	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	menu := &model.Menu{
		Id:       id,
		ParentId: req.ParentId,
		Title:    req.Title,
		Key:      req.Key,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   req.Status,
	}

	if err := menu.Update(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    menu,
	})
}

// DeleteMenu 删除菜单
func DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的菜单ID",
		})
		return
	}

	if err := model.DeleteMenuById(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "删除菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetMenusByParentId 获取指定父菜单的子菜单
func GetMenusByParentId(c *gin.Context) {
	parentIdStr := c.Query("parent_id")
	parentId := 0
	if parentIdStr != "" {
		var err error
		parentId, err = strconv.Atoi(parentIdStr)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "无效的父菜单ID",
			})
			return
		}
	}

	menus, err := model.GetMenusByParentId(parentId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    menus,
	})
}

// CreateCardMenu 创建卡密菜单（父菜单+3个子菜单）
func CreateCardMenu(c *gin.Context) {
	var req CardMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 调用 model 层创建卡密菜单和数据表
	if err := model.CreateCardMenuWithTable(req.Category, req.MenuName, req.Icon, req.Sort); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建卡密菜单失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
	})
}
