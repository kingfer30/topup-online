package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kingfer30/topup-online/model"
)

// InitRequest 初始化请求
type InitRequest struct {
	DBHost     string `json:"db_host" binding:"required"`
	DBPort     string `json:"db_port" binding:"required"`
	DBName     string `json:"db_name" binding:"required"`
	DBUser     string `json:"db_user" binding:"required"`
	DBPassword string `json:"db_password"`
	AdminUser  string `json:"admin_user" binding:"required"`
	AdminPass  string `json:"admin_pass" binding:"required"`
	AdminEmail string `json:"admin_email" binding:"required,email"`
}

// CheckInitStatus 检查系统是否已初始化
func CheckInitStatus(c *gin.Context) {
	configPath := filepath.Join(".", ".initialized")

	// 检查是否存在初始化标记文件
	if _, err := os.Stat(configPath); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "系统已初始化",
			"data": gin.H{
				"initialized": true,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "系统未初始化",
		"data": gin.H{
			"initialized": false,
		},
	})
}

// TestDBConnection 测试数据库连接
func TestDBConnection(c *gin.Context) {
	var req struct {
		DBHost     string `json:"db_host" binding:"required"`
		DBPort     string `json:"db_port" binding:"required"`
		DBName     string `json:"db_name" binding:"required"`
		DBUser     string `json:"db_user" binding:"required"`
		DBPassword string `json:"db_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 构建DSN，添加超时参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=10s&writeTimeout=10s",
		req.DBUser,
		req.DBPassword,
		req.DBHost,
		req.DBPort,
		req.DBName,
	)

	// 尝试连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库连接失败: " + err.Error(),
		})
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取数据库连接失败: " + err.Error(),
		})
		return
	}
	defer sqlDB.Close()

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库连接测试失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "数据库连接成功",
		"data": gin.H{
			"success": true,
		},
	})
}

// InitializeSystem 初始化系统
func InitializeSystem(c *gin.Context) {
	var req InitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查是否已初始化
	configPath := filepath.Join(".", ".initialized")
	if _, err := os.Stat(configPath); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "系统已初始化，无需重复初始化",
		})
		return
	}

	// 构建DSN，添加超时参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=10s&writeTimeout=10s",
		req.DBUser,
		req.DBPassword,
		req.DBHost,
		req.DBPort,
		req.DBName,
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库连接失败: " + err.Error(),
		})
		return
	}

	// 获取底层SQL DB并设置连接池
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.SystemConfig{},
		&model.Admin{},
		&model.User{},
		&model.Order{},
		&model.Card{},
		&model.MirrorCard{},
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建表结构失败: " + err.Error(),
		})
		return
	}

	// 创建管理员账号（前端已MD5加密）
	admin := model.Admin{
		Username: req.AdminUser,
		Password: req.AdminPass, // 前端传过来已经是MD5加密的
		Email:    req.AdminEmail,
		Status:   1,
	}

	if err := db.Create(&admin).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建管理员账号失败: " + err.Error(),
		})
		return
	}

	// 保存系统配置
	configs := []model.SystemConfig{
		{Key: "db_host", Value: req.DBHost},
		{Key: "db_port", Value: req.DBPort},
		{Key: "db_name", Value: req.DBName},
		{Key: "db_user", Value: req.DBUser},
		{Key: "system_initialized", Value: "true"},
		{Key: "site_name", Value: "ChatGPT充值平台"},
	}

	for _, config := range configs {
		if err := db.Create(&config).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "保存系统配置失败: " + err.Error(),
			})
			return
		}
	}

	// 创建环境配置文件
	sqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		req.DBUser,
		req.DBPassword,
		req.DBHost,
		req.DBPort,
		req.DBName,
	)

	envContent := fmt.Sprintf(`# 数据库连接DSN
SQL_DSN=%s

# Redis连接字符串
REDIS_CONN_STRING=redis://localhost:6379

# 日志级别
LOG_LEVEL=info

# 数据库连接池配置
SQL_MAX_IDLE_CONNS=100
SQL_MAX_OPEN_CONNS=1000
SQL_MAX_LIFETIME=60
`, sqlDSN)

	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建配置文件失败: " + err.Error(),
		})
		return
	}

	// 立即设置环境变量，确保当前进程也使用新配置
	os.Setenv("SQL_DSN", sqlDSN)

	// 创建初始化标记文件
	if err := os.WriteFile(configPath, []byte("initialized"), 0644); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建初始化标记失败: " + err.Error(),
		})
		return
	}

	// 重要：将新创建的数据库连接设置为全局DB对象
	// 这样初始化完成后，后续的登录等操作就会使用正确的数据库
	model.DB = db
	SetDB(db)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "系统初始化成功",
		"data": gin.H{
			"success": true,
		},
	})
}
