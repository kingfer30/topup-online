package main

import (
	"github.com/gin-gonic/gin"
	redis "github.com/kingfer30/topup-online/config/cache"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/middleware"
	"github.com/kingfer30/topup-online/router"
	"github.com/kingfer30/topup-online/utils/client"
	"github.com/kingfer30/topup-online/utils/logger"
)

func main() {
	logger.SetupLogger()
	// 创建 Gin 实例
	server := gin.Default()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())

	// Initialize Redis
	err := redis.InitRedisClient()
	if err != nil {
		logger.FatalLog("failed to initialize Redis: " + err.Error())
	}
	client.Init()
	middleware.SetUpLogger(server)
	router.SetRouter(server)
	// 启动服务
	server.Run(":" + constants.ServerPort)
}
