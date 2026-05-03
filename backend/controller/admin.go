package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/middleware"
	"github.com/kingfer30/topup-online/model"
	crypto "github.com/kingfer30/topup-online/utils/cypto"
	"github.com/kingfer30/topup-online/utils/random"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	Admin model.Admin `json:"admin"`
}

var db *gorm.DB

// SetDB 设置数据库连接
func SetDB(database *gorm.DB) {
	db = database
}

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查数据库连接是否已初始化
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化，请先完成系统初始化",
		})
		return
	}

	// 查询管理员
	var admin model.Admin
	if err := db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "用户名或密码错误",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 检查状态
	if admin.Status != 1 {
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "账号已被禁用",
		})
		return
	}

	// 验证密码（前端已MD5加密，后端用bcrypt二次校验）
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 生成token（含版本号，版本号变更时已颁发的旧token自动失效）
	token := generateToken(admin.ID, admin.TokenVersion)

	// 记录登录 session（token 格式：admin_{id}_{version}_{uuid}_{timestamp}，UUID 在第4段）
	tokenParts := strings.Split(token, "_")
	if len(tokenParts) >= 4 && db != nil {
		session := model.AdminSession{
			AdminID:     admin.ID,
			SessionUUID: tokenParts[3],
			IPAddress:   c.ClientIP(),
			UserAgent:   c.GetHeader("User-Agent"),
			IsActive:    true,
		}
		db.Create(&session)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			Token: token,
			Admin: admin,
		},
	})
}

// GetAdminInfo 获取管理员信息（同时作为心跳检测接口）
func GetAdminInfo(c *gin.Context) {
	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "管理员不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    admin,
	})
}

// Logout 退出登录
func Logout(c *gin.Context) {
	// 使当前 session 失效
	if sessionUUID, exists := c.Get("session_uuid"); exists && db != nil {
		db.Model(&model.AdminSession{}).
			Where("session_uuid = ?", sessionUUID).
			Update("is_active", false)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "退出成功",
	})
}

// GetAdminSessions 获取当前管理员的活跃登录设备列表
func GetAdminSessions(c *gin.Context) {
	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "未登录"})
		return
	}
	currentUUID, _ := c.Get("session_uuid")

	var sessions []model.AdminSession
	if err := db.Where("admin_id = ? AND is_active = true", adminID).Order("created_at DESC").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	type SessionVO struct {
		ID          uint   `json:"id"`
		SessionUUID string `json:"session_uuid"`
		IPAddress   string `json:"ip_address"`
		UserAgent   string `json:"user_agent"`
		CreatedAt   int64  `json:"created_at"`
		IsCurrent   bool   `json:"is_current"`
	}
	var result []SessionVO
	for _, s := range sessions {
		result = append(result, SessionVO{
			ID:          s.ID,
			SessionUUID: s.SessionUUID,
			IPAddress:   s.IPAddress,
			UserAgent:   s.UserAgent,
			CreatedAt:   s.CreatedAt.Unix(),
			IsCurrent:   s.SessionUUID == currentUUID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": result})
}

// KickSession 踢出指定设备
func KickSession(c *gin.Context) {
	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "未登录"})
		return
	}
	sessionUUID := c.Param("uuid")
	if sessionUUID == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "session_uuid不能为空"})
		return
	}
	result := db.Model(&model.AdminSession{}).
		Where("session_uuid = ? AND admin_id = ?", sessionUUID, adminID).
		Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "操作失败: " + result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "session不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已踢出该设备"})
}

// KickAllSessions 踢出所有设备（含自己）
func KickAllSessions(c *gin.Context) {
	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "未登录"})
		return
	}
	// 所有 session 置为 inactive
	db.Model(&model.AdminSession{}).Where("admin_id = ?", adminID).Update("is_active", false)
	// 递增 token_version，使所有已颁发 token 直接失效（双保险）
	db.Exec("UPDATE admins SET token_version = token_version + 1 WHERE id = ?", adminID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已踢出所有设备"})
}

// ForceReLogin 强制所有已登录的管理员重新登录
// 通过递增 token_version 使所有已颁发的 token 立即失效
func ForceReLogin(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}
	if err := db.Exec("UPDATE admins SET token_version = token_version + 1").Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "操作失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已强制所有管理员重新登录",
	})
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "管理员不存在",
		})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "原密码错误",
		})
		return
	}

	// 新密码 bcrypt 哈希后存储
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}
	if err := db.Model(&admin).Update("password", string(newHashedPassword)).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "修改密码失败: " + err.Error(),
		})
		return
	}

	// 密码修改后使所有 session 失效，强制重新登录
	db.Exec("UPDATE admins SET token_version = token_version + 1 WHERE id = ?", adminID)
	db.Exec("UPDATE admin_sessions SET is_active = false WHERE admin_id = ?", adminID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功，请重新登录",
	})
}

// generateToken 生成token，格式：admin_{id}_{version}_{uuid}_{timestamp}
func generateToken(adminID uint, version int) string {
	return fmt.Sprintf("admin_%d_%d_%s_%d", adminID, version, uuid.New().String(), time.Now().Unix())
}

// hashPassword 对密码（前端已MD5）进行bcrypt哈希，用于安全存储
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ==================== 用户管理相关接口 ====================

// GetUserList 获取用户列表（带分页、搜索、排序）
func GetUserList(c *gin.Context) {
	// 检查数据库连接
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	// 获取查询参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	keyword := c.DefaultQuery("keyword", "")

	// 转换参数
	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)

	if pageNum < 1 {
		pageNum = 1
	}
	if pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 10
	}

	startIdx := (pageNum - 1) * pageSizeNum

	var users []*model.User
	var total int64
	var err error

	// 如果有搜索关键词，使用搜索功能
	if keyword != "" {
		users, err = model.SearchUsers(keyword)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "搜索用户失败: " + err.Error(),
			})
			return
		}
		total = int64(len(users))

		// 手动分页
		end := startIdx + pageSizeNum
		if end > len(users) {
			end = len(users)
		}
		if startIdx < len(users) {
			users = users[startIdx:end]
		} else {
			users = []*model.User{}
		}
	} else {
		// 获取总数
		db.Model(&model.User{}).Where("status != ?", constants.UserStatusDeleted).Count(&total)

		// 获取用户列表
		users, err = model.GetAllUsers(startIdx, pageSizeNum, "id")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "获取用户列表失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":      users,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSizeNum,
		},
	})
}

// GetUserDetail 获取用户详情
func GetUserDetail(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	userID := c.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "用户不存在: " + err.Error(),
		})
		return
	}

	// 检查用户是否已删除
	if user.Status == constants.UserStatusDeleted {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	var req struct {
		Username    string `json:"username" binding:"required,max=12"`
		Password    string `json:"password" binding:"required,min=8,max=20"`
		DisplayName string `json:"display_name" binding:"max=20"`
		Email       string `json:"email" binding:"omitempty,email,max=50"`
		Status      int    `json:"status"`
		Source      string `json:"source"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	if model.IsUsernameAlreadyTaken(req.Username) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	// 检查邮箱是否已存在
	if req.Email != "" && model.IsEmailAlreadyTaken(req.Email) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "邮箱已被使用",
		})
		return
	}

	// 密码加密：先SHA256，再bcrypt（与chat前端登录流程一致）
	// 1. SHA256加密
	sha256Hash := sha256.Sum256([]byte(req.Password))
	sha256Str := hex.EncodeToString(sha256Hash[:])

	// 2. bcrypt加密SHA256后的密码
	hashedPassword, err := crypto.Password2Hash(sha256Str)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "密码加密失败: " + err.Error(),
		})
		return
	}

	// 生成邀请码
	affCode := random.GetUUID()[:8]
	accessToken := random.GetUUID()

	// 设置默认值
	if req.Status == 0 {
		req.Status = constants.UserStatusEnabled
	}

	// 创建用户
	user := model.User{
		Username:    req.Username,
		Password:    hashedPassword,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Status:      req.Status,
		AffCode:     affCode,
		AccessToken: accessToken,
		Source:      req.Source,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建用户失败: " + err.Error(),
		})
		return
	}

	// 清空密码，不返回给前端
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建用户成功",
		"data":    user,
	})
}

// UpdateUser 更新用户信息
func UpdateUser(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	userID := c.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	// 获取当前管理员ID（如果使用AdminAuth中间件）
	// adminID, _ := c.Get("admin_id")

	var req struct {
		Username     *string `json:"username" binding:"omitempty,max=12"`
		Password     *string `json:"password" binding:"omitempty,min=8,max=20"`
		DisplayName  *string `json:"display_name" binding:"omitempty,max=20"`
		Email        *string `json:"email"`
		Status       *int    `json:"status"`
		Source       *string `json:"source"`
		MirrorCardId *int    `json:"mirror_card_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取要更新的用户
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "用户不存在: " + err.Error(),
		})
		return
	}

	// 检查用户是否已删除
	if user.Status == constants.UserStatusDeleted {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "用户不存在或已被删除",
		})
		return
	}

	// 更新字段
	updatePassword := false
	if req.Username != nil && *req.Username != user.Username {
		// 检查新用户名是否已被使用
		if model.IsUsernameAlreadyTaken(*req.Username) {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "用户名已存在",
			})
			return
		}
		user.Username = *req.Username
	}

	if req.Password != nil && *req.Password != "" {
		// 密码加密：先SHA256，再bcrypt（与chat前端登录流程一致）
		sha256Hash := sha256.Sum256([]byte(*req.Password))
		sha256Str := hex.EncodeToString(sha256Hash[:])
		user.Password = sha256Str
		updatePassword = true
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	if req.Email != nil && *req.Email != user.Email {
		// 如果邮箱不为空，检查格式和唯一性
		if *req.Email != "" {
			// 简单的邮箱格式验证
			if !strings.Contains(*req.Email, "@") || !strings.Contains(*req.Email, ".") {
				c.JSON(http.StatusOK, gin.H{
					"code":    400,
					"message": "邮箱格式不正确",
				})
				return
			}
			// 检查新邮箱是否已被使用
			if model.IsEmailAlreadyTaken(*req.Email) {
				c.JSON(http.StatusOK, gin.H{
					"code":    400,
					"message": "邮箱已被使用",
				})
				return
			}
		}
		user.Email = *req.Email
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	if req.Source != nil {
		user.Source = *req.Source
	}

	if req.MirrorCardId != nil {
		user.MirrorCardId = *req.MirrorCardId
	}

	// 更新用户
	if err := user.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新用户失败: " + err.Error(),
		})
		return
	}

	// 重新获取用户信息（不包含密码）
	user, _ = model.GetUserById(id, false)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新用户成功",
		"data":    user,
	})
}

// DeleteUser 删除用户（软删除）
func DeleteUser(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库未初始化",
		})
		return
	}

	userID := c.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	// 获取要删除的用户
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "用户不存在: " + err.Error(),
		})
		return
	}

	// 检查用户是否已删除
	if user.Status == constants.UserStatusDeleted {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "用户已被删除",
		})
		return
	}

	// 执行软删除
	if err := model.DeleteUserById(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "删除用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除用户成功",
	})
}
