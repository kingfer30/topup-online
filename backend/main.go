package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	redis "github.com/kingfer30/topup-online/config/cache"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/controller"
	"github.com/kingfer30/topup-online/middleware"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/router"
	"github.com/kingfer30/topup-online/scheduler"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

// init 函数在包初始化时执行，先于 main 函数
// 必须在这里加载 .env 文件，确保后续读取环境变量时能获取到正确的值
func init() {
	// 加载.env文件（如果存在），使用Overload强制覆盖已存在的环境变量
	// 这样可以确保.env文件中的配置优先级最高
	_ = godotenv.Overload()
}

func main() {
	logger.SetupLogger()

	// 检查是否已初始化，如果已初始化则连接数据库
	if _, err := os.Stat(".initialized"); err == nil {
		logger.SysLog("System is initialized, connecting to database...")
		model.InitDB()
		// 传递数据库连接给controller
		controller.SetDB(model.DB)
		logger.SysLog("Database connection established")

		// 在main.go中也执行一次数据库迁移，确保所有表结构都是最新的
		logger.SysLog("Running database migration...")
		if err := model.MigrateDB(); err != nil {
			logger.FatalLog("failed to migrate database in main: " + err.Error())
		}
		logger.SysLog("Database migration completed")

		// 启动镜像卡密 Token 定时获取任务（每 30 分钟执行一次）
		logger.SysLog("Starting MirrorCard Token Scheduler...")
		tokenScheduler := scheduler.NewMirrorCardTokenScheduler(30)
		tokenScheduler.Start()
		logger.SysLog("MirrorCard Token Scheduler started successfully")

		// 启动镜像卡密用户信息同步任务（每 10 分钟执行一次）
		logger.SysLog("Starting MirrorCard Sync Scheduler...")
		syncScheduler := scheduler.NewMirrorCardSyncScheduler(10)
		syncScheduler.Start()
		logger.SysLog("MirrorCard Sync Scheduler started successfully")
	} else {
		logger.SysLog("System not initialized yet, waiting for initialization...")
	}

	// 创建 Gin 实例
	server := gin.Default()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	server.Use(middleware.CORS()) // 添加CORS支持

	// Initialize Redis
	err := redis.InitRedisClient()
	if err != nil {
		logger.FatalLog("failed to initialize Redis: " + err.Error())
	}
	client.Init()
	middleware.SetUpLogger(server)
	router.SetRouter(server)
	// 启动服务
	server.Run(":" + constants.GetServerPort())
}
