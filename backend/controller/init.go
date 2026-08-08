package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kingfer30/topup-online/constants"
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
	dataDir := constants.GetDataDir()
	configPath := filepath.Join(dataDir, ".initialized")

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
	dataDir := constants.GetDataDir()
	configPath := filepath.Join(dataDir, ".initialized")
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
		&model.MirrorCard{},
		&model.Menu{},
		&model.SalesTalk{},
		&model.AdConfig{},
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建表结构失败: " + err.Error(),
		})
		return
	}

	// 对前端传来的MD5密码再做bcrypt哈希后存储
	hashedPass, err := hashPassword(req.AdminPass)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}
	admin := model.Admin{
		Username: req.AdminUser,
		Password: hashedPass,
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

	// 初始化默认菜单数据
	if err := seedDefaultMenus(db); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "初始化菜单数据失败: " + err.Error(),
		})
		return
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

	envPath := filepath.Join(dataDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
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

// seedDefaultMenus 初始化默认菜单数据
func seedDefaultMenus(db *gorm.DB) error {
	now := time.Now()

	// 辅助函数：创建菜单并返回插入后的 ID
	createMenu := func(menu *model.Menu) error {
		menu.Status = 1
		menu.IsDelete = -1
		menu.CreatedAt = now
		menu.UpdatedAt = now
		return db.Create(menu).Error
	}

	// 1. 控制台
	dashboard := &model.Menu{ParentId: 0, Title: "控制台", Key: "dashboard", Path: "/admin/dashboard", Icon: "📊", Sort: 1}
	if err := createMenu(dashboard); err != nil {
		return fmt.Errorf("创建控制台菜单失败: %w", err)
	}

	// 2. AI翻译
	aiTranslate := &model.Menu{ParentId: 0, Title: "AI翻译", Key: "ai-translate", Path: "/admin/ai-translate", Icon: "🌐", Sort: 2}
	if err := createMenu(aiTranslate); err != nil {
		return fmt.Errorf("创建AI翻译菜单失败: %w", err)
	}

	// 3. 用户管理
	userMenu := &model.Menu{ParentId: 0, Title: "用户管理", Key: "user", Icon: "👥", Sort: 3}
	if err := createMenu(userMenu); err != nil {
		return fmt.Errorf("创建用户管理菜单失败: %w", err)
	}
	// 用户管理 - 子菜单
	userChildren := []*model.Menu{
		{ParentId: userMenu.Id, Title: "用户列表", Key: "users", Path: "/admin/users", Sort: 1},
		{ParentId: userMenu.Id, Title: "角色管理", Key: "roles", Path: "/admin/roles", Sort: 2},
	}
	for _, child := range userChildren {
		if err := createMenu(child); err != nil {
			return fmt.Errorf("创建用户管理子菜单失败: %w", err)
		}
	}

	// 4. 订单管理
	orderMenu := &model.Menu{ParentId: 0, Title: "订单管理", Key: "order", Icon: "📦", Sort: 4}
	if err := createMenu(orderMenu); err != nil {
		return fmt.Errorf("创建订单管理菜单失败: %w", err)
	}
	// 订单管理 - 子菜单
	orderChildren := []*model.Menu{
		{ParentId: orderMenu.Id, Title: "订单列表", Key: "orders", Path: "/admin/orders", Sort: 1},
		{ParentId: orderMenu.Id, Title: "退款管理", Key: "refunds", Path: "/admin/refunds", Sort: 2},
	}
	for _, child := range orderChildren {
		if err := createMenu(child); err != nil {
			return fmt.Errorf("创建订单管理子菜单失败: %w", err)
		}
	}

	// 5. 话术管理
	salesTalkMenu := &model.Menu{ParentId: 0, Title: "话术管理", Key: "sales-talk", Path: "/admin/sales-talks", Icon: "💬", Sort: 5}
	if err := createMenu(salesTalkMenu); err != nil {
		return fmt.Errorf("创建话术管理菜单失败: %w", err)
	}

	// 6. 镜像管理
	mirrorMenu := &model.Menu{ParentId: 0, Title: "镜像管理", Key: "mirror", Icon: "🔐", Sort: 6}
	if err := createMenu(mirrorMenu); err != nil {
		return fmt.Errorf("创建镜像管理菜单失败: %w", err)
	}
	// 镜像管理 - 子菜单
	mirrorChild := &model.Menu{ParentId: mirrorMenu.Id, Title: "卡密管理", Key: "mirror-cards", Path: "/admin/mirror-cards", Sort: 1}
	if err := createMenu(mirrorChild); err != nil {
		return fmt.Errorf("创建镜像管理子菜单失败: %w", err)
	}

	// 7. GPT业务（父级 + 充值与链接 / GPT卡密 / GPT-CDK）
	gptRoot := &model.Menu{ParentId: 0, Title: "GPT业务", Key: "gpt-business-root", Icon: "🤖", Sort: 7}
	if err := createMenu(gptRoot); err != nil {
		return fmt.Errorf("创建GPT业务菜单失败: %w", err)
	}
	gptChildren := []*model.Menu{
		{ParentId: gptRoot.Id, Title: "充值与链接", Key: "gpt-business", Path: "/admin/gpt-business", Sort: 1},
		{ParentId: gptRoot.Id, Title: "GPT卡密", Key: "gpt-cards", Path: "/admin/gpt-cards", Icon: "🃏", Sort: 2},
		{ParentId: gptRoot.Id, Title: "GPT-CDK", Key: "gpt-cdk", Path: "/admin/gpt-cdk", Icon: "🔑", Sort: 3},
		{ParentId: gptRoot.Id, Title: "广告配置", Key: "ad-configs", Path: "/admin/ad-configs", Icon: "📢", Sort: 5},
	}
	for _, child := range gptChildren {
		if err := createMenu(child); err != nil {
			return fmt.Errorf("创建GPT业务子菜单失败: %w", err)
		}
	}

	// 8. 系统设置
	systemMenu := &model.Menu{ParentId: 0, Title: "系统设置", Key: "system", Icon: "⚙️", Sort: 8}
	if err := createMenu(systemMenu); err != nil {
		return fmt.Errorf("创建系统设置菜单失败: %w", err)
	}
	// 系统设置 - 子菜单
	systemChildren := []*model.Menu{
		{ParentId: systemMenu.Id, Title: "基础设置", Key: "settings", Path: "/admin/settings", Sort: 1},
		{ParentId: systemMenu.Id, Title: "操作日志", Key: "logs", Path: "/admin/logs", Sort: 2},
		{ParentId: systemMenu.Id, Title: "菜单管理", Key: "menu-management", Path: "/admin/menu-management", Sort: 3},
	}
	for _, child := range systemChildren {
		if err := createMenu(child); err != nil {
			return fmt.Errorf("创建系统设置子菜单失败: %w", err)
		}
	}

	return nil
}
