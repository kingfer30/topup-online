package controller

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kingfer30/topup-online/constants"
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

	// // 验证密码（前端已MD5加密）
	// if admin.Password != req.Password {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code":    401,
	// 		"message": "用户名或密码错误",
	// 	})
	// 	return
	// }

	// 生成token
	token := generateToken(admin.ID)

	// 保存token到数据库或Redis（这里简化处理，实际应该用Redis）
	// TODO: 将token保存到Redis，设置过期时间

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			Token: token,
			Admin: admin,
		},
	})
}

// GetAdminInfo 获取管理员信息
func GetAdminInfo(c *gin.Context) {
	// 从header获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// TODO: 从Redis验证token并获取admin_id
	// 这里简化处理，直接从token中解析（实际应该用JWT或Redis）

	// 模拟返回管理员信息
	adminID := 1 // 这里应该从token中解析出来

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
	// TODO: 从Redis删除token

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "退出成功",
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

	// TODO: 从token获取当前管理员ID
	adminID := 1

	var admin model.Admin
	if err := db.First(&admin, adminID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    404,
			"message": "管理员不存在",
		})
		return
	}

	// 验证旧密码
	if admin.Password != req.OldPassword {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "原密码错误",
		})
		return
	}

	// 更新密码
	if err := db.Model(&admin).Update("password", req.NewPassword).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "修改密码失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
	})
}

// generateToken 生成token
func generateToken(adminID uint) string {
	// 简单的token生成（生产环境应该使用JWT）
	return fmt.Sprintf("admin_%d_%s_%d", adminID, uuid.New().String(), time.Now().Unix())
}

// hashPassword MD5加密密码
func hashPassword(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
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
