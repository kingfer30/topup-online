package controller

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kingfer30/topup-online/model"
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

	// 验证密码（前端已MD5加密）
	if admin.Password != req.Password {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

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
func generateToken(adminID int) string {
	// 简单的token生成（生产环境应该使用JWT）
	return fmt.Sprintf("admin_%d_%s_%d", adminID, uuid.New().String(), time.Now().Unix())
}

// hashPassword MD5加密密码
func hashPassword(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}
