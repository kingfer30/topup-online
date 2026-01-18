package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/controller"
	"github.com/kingfer30/topup-online/middleware"
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
		adminGroup := apiRouter.Group("/admin")
		adminGroup.Use(middleware.AdminAuth())
		{
			adminGroup.GET("/info", controller.GetAdminInfo)
			adminGroup.POST("/logout", controller.Logout)
			adminGroup.POST("/change-password", controller.ChangePassword)

			// 用户管理接口（管理员专用）
			adminGroup.GET("/users", controller.GetUserList)       // 获取用户列表
			adminGroup.GET("/users/:id", controller.GetUserDetail) // 获取用户详情
			adminGroup.POST("/users", controller.CreateUser)       // 创建用户
			adminGroup.PUT("/users/:id", controller.UpdateUser)    // 更新用户
			adminGroup.DELETE("/users/:id", controller.DeleteUser) // 删除用户

			// 镜像卡密管理接口（管理员专用）
			adminGroup.GET("/mirror-cards", controller.GetMirrorCardList)                    // 获取卡密列表
			adminGroup.GET("/mirror-cards/:id", controller.GetMirrorCardDetail)              // 获取卡密详情
			adminGroup.POST("/mirror-cards", controller.CreateMirrorCard)                    // 创建卡密
			adminGroup.PUT("/mirror-cards/:id", controller.UpdateMirrorCard)                 // 更新卡密
			adminGroup.DELETE("/mirror-cards/:id", controller.DeleteMirrorCard)              // 删除卡密
			adminGroup.POST("/mirror-cards/batch-import", controller.BatchImportMirrorCards) // 批量导入
		}

		// 用户认证相关（无需认证）
		apiRouter.POST("/user/register", controller.Register)
		apiRouter.POST("/user/login", controller.Login)

		// 用户相关（需要认证）
		apiRouter.GET("/user/info", controller.GetUserInfo)
		apiRouter.POST("/user/logout", controller.UserLogout)

		// 卡密相关
		apiRouter.GET("/cdk/:number", controller.GetCardDetail)
		apiRouter.POST("/cdk/check", controller.GetCardDetail)
		apiRouter.POST("/cdk/top-up", controller.GetCardDetail)
		apiRouter.GET("/cdk/thread/:id", controller.GetCardDetail)
	}
}
