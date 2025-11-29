package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/controller"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	{
		// 系统初始化相关（无需认证）
		apiRouter.GET("/system/init/status", controller.CheckInitStatus)
		apiRouter.POST("/system/init/test-db", controller.TestDBConnection)
		apiRouter.POST("/system/init", controller.InitializeSystem)

		// 管理员认证相关（无需认证）
		apiRouter.POST("/admin/login", controller.AdminLogin)

		// 管理员相关（需要认证）
		// TODO: 添加认证中间件
		apiRouter.GET("/admin/info", controller.GetAdminInfo)
		apiRouter.POST("/admin/logout", controller.Logout)
		apiRouter.POST("/admin/change-password", controller.ChangePassword)

		// 卡密相关
		apiRouter.GET("/cdk/:number", controller.GetCardDetail)
		apiRouter.POST("/cdk/check", controller.GetCardDetail)
		apiRouter.POST("/cdk/top-up", controller.GetCardDetail)
		apiRouter.GET("/cdk/thread/:id", controller.GetCardDetail)
	}
}
