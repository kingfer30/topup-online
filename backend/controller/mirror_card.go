package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/middleware"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/scheduler"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/random"
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

// RoomList 获取房间列表（根据平台类型）
func RoomList(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	platformType := c.DefaultQuery("platform_type", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 验证 platform_type
	validPlatforms := map[string]bool{
		"gpt-1":    true,
		"gpt-2":    true,
		"gpt-3":    true,
		"gemini-1": true,
		"claude-1": true,
		"mj-1":     true,
	}

	if !validPlatforms[platformType] {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的平台类型",
		})
		return
	}

	// 获取当前登录用户
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "未授权访问",
		})
		return
	}

	// 检查用户是否绑定了镜像卡密
	if user.MirrorCardId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "用户未绑定镜像卡密",
		})
		return
	}

	// 查询镜像卡密
	var mirrorCard model.MirrorCard
	if err := db.First(&mirrorCard, user.MirrorCardId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "镜像卡密不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询镜像卡密失败: " + err.Error(),
		})
		return
	}

	// 检查 Token 和 NodeURL 是否存在
	if mirrorCard.Token == "" || mirrorCard.NodeURL == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "镜像卡密配置不完整",
		})
		return
	}

	// 根据平台类型处理
	switch platformType {
	case "gpt-1":
		handleGPT1Request(c, mirrorCard, page, pageSize)
	case "gpt-2", "gpt-3", "gemini-1", "claude-1", "mj-1":
		// 其他平台暂未实现
		c.JSON(http.StatusOK, gin.H{
			"code":    501,
			"message": "该平台功能暂未实现",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "不支持的平台类型",
		})
	}
}

// JoinRoom 加入房间
func JoinRoom(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	// 接收参数
	var req struct {
		RoomId string `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前登录用户
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "未授权访问",
		})
		return
	}

	// 检查用户是否绑定了镜像卡密
	if user.MirrorCardId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "用户未绑定镜像卡密",
		})
		return
	}

	// 查询镜像卡密
	var mirrorCard model.MirrorCard
	if err := db.First(&mirrorCard, user.MirrorCardId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "镜像卡密不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询镜像卡密失败: " + err.Error(),
		})
		return
	}

	// 检查卡密状态是否启用
	if mirrorCard.Status != model.MirrorCardStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "镜像卡密未启用",
		})
		return
	}

	// 检查 Token 是否存在
	if mirrorCard.Token == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "镜像卡密 Token 不存在",
		})
		return
	}

	// 检查 NodeURL 是否存在
	if mirrorCard.NodeURL == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "镜像卡密配置不完整",
		})
		return
	}

	// 构建请求参数
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	requestBody := map[string]interface{}{
		"channel":   "xy",
		"car_id":    req.RoomId,
		"timestamp": timestamp,
		"sign":      "",
	}

	// 转换为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "构建请求参数失败: " + err.Error(),
		})
		return
	}

	// 构建请求 URL
	url := fmt.Sprintf("%s/share-login/v1/user/home/enter", mirrorCard.NodeURL)

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建请求失败: " + err.Error(),
		})
		return
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Cookie", "token="+mirrorCard.Token)

	logger.SysLog(fmt.Sprintf("请求 URL: %s, Request Body: %s, Token: %s", url, string(jsonData), mirrorCard.Token))
	// 发送请求
	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "请求外部接口失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode == 401 {
		// Token 已失效，清空数据库中的 token 字段
		if err := db.Model(&model.MirrorCard{}).Where("id = ?", mirrorCard.ID).Update("token", "").Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "Token已失效，清空失败: " + err.Error(),
			})
			return
		}

		// 手动触发一次 Token 获取
		if err := scheduler.TriggerFetchForCard(mirrorCard.ID); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "Token已失效，触发重新获取失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "Token已失效，正在重新获取，请稍后重试",
		})
		return
	}

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "读取响应失败: " + err.Error(),
		})
		return
	}
	logger.SysLog(fmt.Sprintf("响应 Body: %s", string(bodyBytes)))
	// 解析响应
	var apiResp struct {
		IsSuccess bool   `json:"isSuccess"`
		Msg       string `json:"msg"`
		RespData  string `json:"respData"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "解析响应失败: " + err.Error(),
		})
		return
	}

	// 检查接口是否成功
	if !apiResp.IsSuccess {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "外部接口返回失败: " + apiResp.Msg,
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "进入成功",
		"data": gin.H{
			"url":   apiResp.RespData,
			"token": mirrorCard.Token,
		},
	})
}

// handleGPT1Request 处理 gpt-1 平台的请求
func handleGPT1Request(c *gin.Context, mirrorCard model.MirrorCard, page, pageSize int) {
	// 构建请求 URL
	url := fmt.Sprintf("%s/share-login/v1/user/home/carpage?channel=xy&page=%d&pageSize=%d",
		mirrorCard.NodeURL, page, pageSize)

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建请求失败: " + err.Error(),
		})
		return
	}

	// 设置 Cookie
	req.Header.Set("Cookie", "token="+mirrorCard.Token)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "请求外部接口失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode == 401 {
		// Token 已失效，清空数据库中的 token 字段
		if err := db.Model(&model.MirrorCard{}).Where("id = ?", mirrorCard.ID).Update("token", "").Error; err != nil {
			// 记录错误但不影响返回
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "Token已失效，清空失败: " + err.Error(),
			})
			return
		}

		// 手动触发一次 Token 获取（异步执行，不阻塞当前请求）
		if err := scheduler.TriggerFetchForCard(mirrorCard.ID); err != nil {
			// 记录错误但不影响返回
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "Token已失效，触发重新获取失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "Token已失效，正在重新获取，请稍后重试",
		})
		return
	}

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "读取响应失败: " + err.Error(),
		})
		return
	}

	// 解析响应
	var apiResp struct {
		IsSuccess bool   `json:"isSuccess"`
		Msg       string `json:"msg"`
		RespData  struct {
			List []struct {
				Sort          int    `json:"sort"`
				CarID         string `json:"carid"`
				CarName       string `json:"carname"`
				Count         int    `json:"count"`
				Available     bool   `json:"available"`
				IsPaidAccount bool   `json:"is_paid_account"`
				PlanType      string `json:"plan_type"`
				HighIQStatus  bool   `json:"high_iq_status"`
			} `json:"list"`
			TotalCount int `json:"totalCount"`
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
		} `json:"respData"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "解析响应失败: " + err.Error(),
		})
		return
	}

	// 检查接口是否成功
	if !apiResp.IsSuccess {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "外部接口返回失败: " + apiResp.Msg,
		})
		return
	}

	// 整理返回数据
	var roomList []gin.H
	for _, room := range apiResp.RespData.List {
		// 根据规则修改数据
		count := room.Count
		carName := room.CarName

		// 当 high_iq_status=false 时，count 值固定=999
		if !room.HighIQStatus {
			count = 999
		}

		// 根据 plan_type 修改 carname
		if room.PlanType == "plus" {
			carName = random.GetRandomString(5) + "-Plus"
		} else if room.PlanType == "team" {
			carName = random.GetRandomString(5) + "-Team"
		} else {
			carName = random.GetRandomString(5)
		}

		roomList = append(roomList, gin.H{
			"sort":      room.Sort,
			"carid":     room.CarID,
			"carname":   carName,
			"count":     count,
			"available": room.Available,
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":       roomList,
			"totalCount": apiResp.RespData.TotalCount,
			"page":       apiResp.RespData.Page,
			"pageSize":   apiResp.RespData.PageSize,
		},
	})
}
