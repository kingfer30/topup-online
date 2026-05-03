package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/model"
)

// AdminAuth 管理员认证中间件
// 验证token并将admin_id存入context
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "未登录或token已过期",
			})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// token格式：admin_{id}_{version}_{uuid}_{timestamp}
		parts := strings.Split(token, "_")
		if len(parts) < 3 {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "无效的token",
			})
			c.Abort()
			return
		}

		adminID, err := strconv.Atoi(parts[1])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "无效的token格式",
			})
			c.Abort()
			return
		}

		tokenVersion, err := strconv.Atoi(parts[2])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "无效的token格式",
			})
			c.Abort()
			return
		}

		// 验证管理员是否存在且状态正常
		var admin model.Admin
		if err := model.DB.First(&admin, adminID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "管理员不存在或已被删除",
			})
			c.Abort()
			return
		}

		if admin.Status != 1 {
			c.JSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "账号已被禁用",
			})
			c.Abort()
			return
		}

		// 校验token版本号，不匹配说明已被强制登出
		if tokenVersion != admin.TokenVersion {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "token已失效，请重新登录",
			})
			c.Abort()
			return
		}

		// 校验 session 是否有效（是否被单独踢出）
		// token格式：admin_{id}_{version}_{uuid}_{timestamp}，UUID 在第4段
		if len(parts) >= 4 {
			sessionUUID := parts[3]
			var session model.AdminSession
			err := model.DB.Where("session_uuid = ? AND admin_id = ?", sessionUUID, adminID).First(&session).Error
			if err != nil {
				// session 记录不存在（可能是 backend 重启前颁发的旧 token），
				// token_version 已通过校验，补建 session 记录并放行
				newSession := model.AdminSession{
					AdminID:     uint(adminID),
					SessionUUID: sessionUUID,
					IPAddress:   c.ClientIP(),
					UserAgent:   c.GetHeader("User-Agent"),
					IsActive:    true,
				}
				model.DB.Create(&newSession)
			} else if !session.IsActive {
				// session 存在但已被踢出
				c.JSON(http.StatusOK, gin.H{
					"code":    401,
					"message": "您的登录已在其他位置被踢出，请重新登录",
				})
				c.Abort()
				return
			}
			// 将 session_uuid 存入 context，供 KickSession/Logout 等接口使用
			c.Set("session_uuid", sessionUUID)
		}

		// 将admin_id存入context，供后续handler使用
		c.Set("admin_id", adminID)
		c.Next()
	}
}

// UserAuth 用户认证中间件
// 验证token并将user_id存入context
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从 Authorization 头获取 token（用于 API 请求）
		token := c.GetHeader("Authorization")

		// 如果 Authorization 头为空，尝试从 Cookie 中获取（用于页面访问）
		if token == "" {
			token, _ = c.Cookie("access_token")
		}

		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "未登录或token已过期",
			})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 验证access_token
		user := model.ValidateAccessToken(token)
		if user == nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "无效的token或token已过期",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if user.Status != constants.UserStatusEnabled {
			c.JSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "账号已被禁用或删除",
			})
			c.Abort()
			return
		}

		// 将user_id存入context
		c.Set("user_id", user.Id)
		c.Set("user", user)
		c.Next()
	}
}

// GetAdminID 从context获取admin_id
func GetAdminID(c *gin.Context) (uint, bool) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return 0, false
	}
	return uint(adminID.(int)), true
}

// GetUserID 从context获取user_id
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

// GetUser 从context获取user对象
func GetUser(c *gin.Context) (*model.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	return user.(*model.User), true
}
