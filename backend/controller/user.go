package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/model"
	crypto "github.com/kingfer30/topup-online/utils/cypto"
	"github.com/kingfer30/topup-online/utils/random"
)

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,len=64"` // SHA256加密后的密码固定64位
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

// Register 用户注册
func Register(c *gin.Context) {
	var req UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名是否已被使用
	if model.IsUsernameAlreadyTaken(req.Username) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "用户名已被使用",
		})
		return
	}

	// 检查邮箱是否已被使用
	if model.IsEmailAlreadyTaken(req.Email) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "邮箱已被使用",
		})
		return
	}

	// 前端已经用SHA256加密了密码（64位十六进制字符串）
	// 后端再用bcrypt加密，实现双重加密：bcrypt(SHA256(原始密码))
	hashedPassword, err := crypto.Password2Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 生成access_token
	accessToken := random.GetUUID()

	// 生成邀请码
	affCode := random.GetRandomString(8)

	// 创建用户
	user := model.User{
		Username:     req.Username,
		Password:     hashedPassword,
		Email:        req.Email,
		DisplayName:  req.Username,
		Role:         constants.RoleCommonUser,
		Status:       constants.UserStatusEnabled,
		AccessToken:  accessToken,
		AffCode:      affCode,
		Quota:        0,
		UsedQuota:    0,
		RequestCount: 0,
		Group:        "default",
	}

	// 保存到数据库
	if err := model.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "注册失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"data":    nil,
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证用户
	// 前端发送的是SHA256加密后的密码
	// ValidateAndFill会用bcrypt验证SHA256密码和数据库中的bcrypt(SHA256(原始密码))
	user := model.User{
		Username: req.Username,
		Password: req.Password, // SHA256加密后的密码
	}

	if err := user.ValidateAndFill(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	// 如果access_token为空，生成新的token
	if user.AccessToken == "" {
		user.AccessToken = random.GetUUID()
		if err := model.DB.Model(&user).Update("access_token", user.AccessToken).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "生成token失败",
			})
			return
		}
	}

	// 清空密码字段，不返回给前端
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": UserLoginResponse{
			Token: user.AccessToken,
			User:  user,
		},
	})
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	// 从请求头获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	// 验证token
	user := model.ValidateAccessToken(token)
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "无效的token",
		})
		return
	}

	// 清空密码字段
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// Logout 用户登出
func UserLogout(c *gin.Context) {
	// 从请求头获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	// 验证token
	user := model.ValidateAccessToken(token)
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "无效的token",
		})
		return
	}

	// 可选：清除token（如果需要强制登出）
	// model.DB.Model(&user).Update("access_token", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}
