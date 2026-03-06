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

			// 菜单管理接口（管理员专用）
			adminGroup.GET("/menus/tree", controller.GetMenuTree)            // 获取菜单树
			adminGroup.GET("/menus", controller.GetAllMenus)                 // 获取所有菜单
			adminGroup.GET("/menus/:id", controller.GetMenuById)             // 获取菜单详情
			adminGroup.POST("/menus", controller.CreateMenu)                 // 创建菜单
			adminGroup.PUT("/menus/:id", controller.UpdateMenu)              // 更新菜单
			adminGroup.DELETE("/menus/:id", controller.DeleteMenu)           // 删除菜单
			adminGroup.GET("/menus/children", controller.GetMenusByParentId) // 获取子菜单
			adminGroup.POST("/menus/card-menu", controller.CreateCardMenu)   // 创建卡密菜单

			// 控制台统计接口
			adminGroup.GET("/dashboard/stats", controller.GetDashboardStats) // 按卡密类型统计今日销售

			// 卡密管理接口（管理员专用）
			adminGroup.GET("/cards", controller.GetCardList)                                          // 获取卡密列表
			adminGroup.GET("/cards/:id", controller.GetCardById)                                      // 获取卡密详情
			adminGroup.POST("/cards", controller.CreateCard)                                          // 创建卡密
			adminGroup.PUT("/cards/:id", controller.UpdateCard)                                       // 更新卡密
			adminGroup.DELETE("/cards/:id", controller.DeleteCard)                                    // 删除卡密
			adminGroup.POST("/cards/batch-import", controller.BatchImportCards)                       // 批量导入卡密
			adminGroup.GET("/cards/unsold-subscription-types", controller.GetUnsoldSubscriptionTypes) // 获取未售订阅类型
			adminGroup.POST("/cards/pickup", controller.PickupCard)                                   // 取货
			adminGroup.POST("/cards/complete-pickup", controller.CompletePickup)                      // 完成取货
			adminGroup.POST("/cards/rollback-pickup", controller.RollbackPickup)                      // 回滚取货（发货中→未出售）
			adminGroup.POST("/cards/rollback-sold", controller.RollbackSoldCard)                      // 回滚已售（已出售→未出售）
			adminGroup.POST("/cards/batch-upgrade", controller.BatchUpgradeToProduct)                 // 批量升级为成品
			adminGroup.POST("/cards/batch-pickup", controller.BatchPickup)                            // 批量取货
			adminGroup.GET("/cards/export", controller.ExportCards)                                    // 导出卡密
			adminGroup.POST("/cards/batch-check", controller.BatchCheckCards)                          // 批量检查订阅状态

			// 话术管理接口（管理员专用）
			adminGroup.GET("/sales-talks", controller.GetSalesTalkList)                   // 获取话术列表
			adminGroup.GET("/sales-talks/:id", controller.GetSalesTalkById)               // 获取话术详情
			adminGroup.POST("/sales-talks", controller.CreateSalesTalk)                   // 创建话术
			adminGroup.PUT("/sales-talks/:id", controller.UpdateSalesTalk)                // 更新话术
			adminGroup.DELETE("/sales-talks/:id", controller.DeleteSalesTalk)             // 删除话术
			adminGroup.POST("/sales-talks/batch-tag", controller.BatchUpdateSalesTalkTag) // 批量更新标签

			// Digiseller 对接接口（管理员专用）
			adminGroup.GET("/digiseller/check-code/:unique_code", controller.CheckUniqueCode) // 查询唯一码支付信息
		}

		// 用户认证相关（无需认证）
		apiRouter.POST("/user/register", controller.Register)
		apiRouter.POST("/user/login", controller.Login)

		// 用户相关（需要认证）
		userGroup := apiRouter.Group("/user")
		userGroup.Use(middleware.UserAuth())
		{
			userGroup.GET("/info", controller.GetUserInfo)
			userGroup.POST("/logout", controller.UserLogout)

			// 房间列表接口
			userGroup.GET("/rooms", controller.RoomList)      // 获取房间列表
			userGroup.POST("/room/join", controller.JoinRoom) // 加入房间
		}

		// 房间反向代理（需要认证）
		// 处理 /rooms/:id 及其所有子路径
		roomGroup := apiRouter.Group("/rooms")
		roomGroup.Use(middleware.UserAuth())
		{
			roomGroup.Any("/:id", controller.RoomProxy)
			roomGroup.Any("/:id/*proxyPath", controller.RoomProxy)
		}

		// 卡密相关
		apiRouter.GET("/cdk/:number", controller.GetCardDetail)
		apiRouter.POST("/cdk/check", controller.GetCardDetail)
		apiRouter.POST("/cdk/top-up", controller.GetCardDetail)
		apiRouter.GET("/cdk/thread/:id", controller.GetCardDetail)
	}
}
