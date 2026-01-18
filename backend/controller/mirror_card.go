package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"gorm.io/gorm"
)

// GetMirrorCardList 获取镜像卡密列表
func GetMirrorCardList(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.DefaultQuery("keyword", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取列表
	cards, total, err := model.GetMirrorCardList(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取列表失败: " + err.Error(),
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

// GetMirrorCardDetail 获取镜像卡密详情
func GetMirrorCardDetail(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	card, err := model.GetMirrorCardById(id)
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

// CreateMirrorCard 创建镜像卡密
func CreateMirrorCard(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	card := &model.MirrorCard{
		Username:   req.Username,
		Password:   req.Password,
		BindStatus: model.MirrorCardBindStatusUnbound,
		Status:     model.MirrorCardStatusEnabled,
	}

	if err := model.CreateMirrorCard(card); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    card,
	})
}

// UpdateMirrorCard 更新镜像卡密
func UpdateMirrorCard(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var req struct {
		Username   *string `json:"username"`
		Password   *string `json:"password"`
		Status     *int    `json:"status"`
		BindUserId *int    `json:"bind_user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取原有卡密
	card, err := model.GetMirrorCardById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "卡密不存在: " + err.Error(),
		})
		return
	}

	// 记录原有的绑定状态
	oldBindUserId := card.BindUserId

	// 更新字段
	if req.Username != nil {
		card.Username = *req.Username
	}
	if req.Password != nil {
		card.Password = *req.Password
	}
	if req.Status != nil {
		card.Status = *req.Status
	}

	// 处理绑定用户逻辑
	if req.BindUserId != nil {
		newBindUserId := *req.BindUserId

		// 如果要绑定到新用户
		if newBindUserId > 0 {
			// 检查用户是否存在
			_, err := model.GetUserById(newBindUserId, false)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"code":    400,
					"message": "绑定的用户不存在",
				})
				return
			}

			// 如果原来绑定了其他用户，需要清空那个用户的mirror_card_id
			if oldBindUserId > 0 && oldBindUserId != newBindUserId {
				db.Model(&model.User{}).Where("id = ?", oldBindUserId).Update("mirror_card_id", 0)
			}

			// 更新新用户的mirror_card_id
			db.Model(&model.User{}).Where("id = ?", newBindUserId).Update("mirror_card_id", id)

			// 更新卡密绑定信息
			now := time.Now()
			card.BindStatus = model.MirrorCardBindStatusBound
			card.BindUserId = newBindUserId
			card.BindTime = &now
		} else {
			// 解绑：newBindUserId == 0
			if oldBindUserId > 0 {
				// 清空原用户的mirror_card_id
				db.Model(&model.User{}).Where("id = ?", oldBindUserId).Update("mirror_card_id", 0)
			}

			// 更新卡密为未绑定状态
			card.BindStatus = model.MirrorCardBindStatusUnbound
			card.BindUserId = 0
			card.BindTime = nil
		}
	}

	// 更新卡密
	if err := model.UpdateMirrorCard(card); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    card,
	})
}

// DeleteMirrorCard 删除镜像卡密
func DeleteMirrorCard(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	// 删除卡密（会自动清空对应用户的mirror_card_id）
	if err := model.DeleteMirrorCard(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    404,
				"message": "卡密不存在",
			})
			return
		}
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

// BatchImportMirrorCards 批量导入镜像卡密
func BatchImportMirrorCards(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	var req struct {
		Data string `json:"data" binding:"required"` // 格式: 用户名----密码，每行一条
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 解析数据
	lines := strings.Split(strings.TrimSpace(req.Data), "\n")
	var cards []*model.MirrorCard
	var errors []string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 分割用户名和密码
		parts := strings.Split(line, "----")
		if len(parts) != 2 {
			errors = append(errors, "第"+strconv.Itoa(i+1)+"行格式错误")
			continue
		}

		username := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])

		if username == "" || password == "" {
			errors = append(errors, "第"+strconv.Itoa(i+1)+"行用户名或密码为空")
			continue
		}

		card := &model.MirrorCard{
			Username:   username,
			Password:   password,
			BindStatus: model.MirrorCardBindStatusUnbound,
			Status:     model.MirrorCardStatusEnabled,
		}
		cards = append(cards, card)
	}

	// 如果没有有效数据
	if len(cards) == 0 {
		message := "没有有效的数据可导入"
		if len(errors) > 0 {
			message += "，错误信息: " + strings.Join(errors, "; ")
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": message,
		})
		return
	}

	// 批量创建
	if err := model.BatchCreateMirrorCards(cards); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "批量导入失败: " + err.Error(),
		})
		return
	}

	message := "成功导入" + strconv.Itoa(len(cards)) + "条记录"
	if len(errors) > 0 {
		message += "，" + strconv.Itoa(len(errors)) + "条记录失败: " + strings.Join(errors, "; ")
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": message,
		"data": gin.H{
			"success": len(cards),
			"failed":  len(errors),
		},
	})
}
