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
		apiRouter.POST("/admin/force-relogin", controller.ForceReLogin) // 强制所有 token 失效，要求重新登录

		// GPT RT 许可证验证/消耗/预占（公开，无需认证）
		apiRouter.POST("/gpt-rt-license/verify", controller.VerifyGptRtLicense)
		apiRouter.POST("/gpt-rt-license/consume", controller.ConsumeGptRtLicense)
		apiRouter.POST("/gpt-rt-license/reserve", controller.ReserveGptRtLicense)
		apiRouter.POST("/gpt-rt-license/confirm", controller.ConfirmGptRtLicense)
		apiRouter.POST("/gpt-rt-license/release", controller.ReleaseGptRtLicense)

		// 广告（公开，无需认证）
		apiRouter.GET("/ads/active", controller.GetActiveAds)
		apiRouter.POST("/ads/click", controller.ClickAd)

		// Cursor 短信验证码查询（公开，无需认证，供独立取码页 /sms/cursor 使用）
		apiRouter.GET("/sms/cursor/query", controller.GetCursorSmsCode)

		// 管理员相关（需要认证）
		adminGroup := apiRouter.Group("/admin")
		adminGroup.Use(middleware.AdminAuth())
		{
			adminGroup.GET("/info", controller.GetAdminInfo)
			adminGroup.POST("/logout", controller.Logout)
			adminGroup.POST("/change-password", controller.ChangePassword)

			// 登录设备管理
			adminGroup.GET("/sessions", controller.GetAdminSessions)     // 获取已登录设备列表
			adminGroup.DELETE("/sessions/:uuid", controller.KickSession) // 踢出指定设备
			adminGroup.DELETE("/sessions", controller.KickAllSessions)   // 踢出所有设备（含自己）

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
			adminGroup.GET("/cards", controller.GetCardList)                                             // 获取卡密列表
			adminGroup.GET("/cards/:id", controller.GetCardById)                                         // 获取卡密详情
			adminGroup.POST("/cards", controller.CreateCard)                                             // 创建卡密
			adminGroup.PUT("/cards/:id", controller.UpdateCard)                                          // 更新卡密
			adminGroup.DELETE("/cards/:id", controller.DeleteCard)                                       // 删除卡密
			adminGroup.POST("/cards/batch-import", controller.BatchImportCards)                          // 批量导入卡密
			adminGroup.GET("/cards/unsold-subscription-types", controller.GetUnsoldSubscriptionTypes)    // 获取未售订阅类型
			adminGroup.POST("/cards/pickup", controller.PickupCard)                                      // 取货
			adminGroup.POST("/cards/complete-pickup", controller.CompletePickup)                         // 完成取货
			adminGroup.POST("/cards/rollback-pickup", controller.RollbackPickup)                         // 回滚取货（发货中→未出售）
			adminGroup.POST("/cards/rollback-sold", controller.RollbackSoldCard)                         // 回滚已售（已出售→未出售）
			adminGroup.POST("/cards/batch-dashboard-goto-resolve", controller.BatchDashboardGotoResolve) // 提链 dashboard 批量回滚并标记 -2
			adminGroup.POST("/cards/batch-upgrade", controller.BatchUpgradeToProduct)                    // 批量升级为成品
			adminGroup.POST("/cards/batch-pickup", controller.BatchPickup)                               // 批量取货
			adminGroup.GET("/cards/export", controller.ExportCards)                                      // 导出卡密
			adminGroup.POST("/cards/batch-check", controller.BatchCheckCards)                            // 批量检查订阅状态
			adminGroup.POST("/cards/enable-on-demand", controller.EnableOnDemandSpendHandler)            // 开启按需付费
			adminGroup.POST("/cards/batch-enable-on-demand", controller.BatchEnableOnDemandSpendHandler) // 批量开启按需付费
			adminGroup.POST("/cards/update-remark", controller.UpdateCardRemark)                         // 单独更新备注
			adminGroup.POST("/cards/goto-pro", controller.GotoPro)                                       // 提链：获取 Cursor Pro 付款链接
			adminGroup.POST("/cards/stripe-alipay", controller.SubmitStripeAlipay)                       // 自动提交 Stripe Alipay 账单并返回付款页
			adminGroup.POST("/cards/poll-subscription", controller.PollCardSubscription)                 // 轮询卡密当前订阅类型
			adminGroup.POST("/cards/half-price-checkout", controller.HalfPriceCheckout)                  // 半价提链：活动页预检+开单
			adminGroup.GET("/cards/half-price-quota", controller.GetHalfPriceQuota)                      // 半价提链：抓取活动页余量
			adminGroup.POST("/cards/batch-freeze", controller.BatchFreezeCards)                          // 批量冻结/解冻普号
			adminGroup.POST("/cards/batch-delete", controller.BatchDeleteCards)                          // 批量删除（status=-1）
			adminGroup.GET("/cards/table-names", controller.GetCardTableNames)                           // 获取所有 cards_* 表名

			// 话术管理接口（管理员专用）
			adminGroup.GET("/sales-talks", controller.GetSalesTalkList)                   // 获取话术列表
			adminGroup.GET("/sales-talks/:id", controller.GetSalesTalkById)               // 获取话术详情
			adminGroup.POST("/sales-talks", controller.CreateSalesTalk)                   // 创建话术
			adminGroup.PUT("/sales-talks/:id", controller.UpdateSalesTalk)                // 更新话术
			adminGroup.DELETE("/sales-talks/:id", controller.DeleteSalesTalk)             // 删除话术
			adminGroup.POST("/sales-talks/batch-tag", controller.BatchUpdateSalesTalkTag) // 批量更新标签

			// Digiseller 对接接口（管理员专用）
			adminGroup.GET("/digiseller/check-code/:unique_code", controller.CheckUniqueCode) // 查询唯一码支付信息
			adminGroup.GET("/digiseller/prices", controller.GetDigisellerPrices)              // 获取订阅类型售价配置
			adminGroup.POST("/digiseller/prices", controller.UpsertDigisellerPrice)           // 新增或更新订阅类型售价

			// AI 模型设置与翻译（管理员专用）
			adminGroup.GET("/settings/ai", controller.GetAdminAISettings)
			adminGroup.PUT("/settings/ai", controller.UpdateAdminAISettings)
			adminGroup.GET("/settings/cursor-pay", controller.GetAdminCursorPaySettings)
			adminGroup.PUT("/settings/cursor-pay", controller.UpdateAdminCursorPaySettings)
			adminGroup.POST("/ai/translate", controller.AdminAITranslate)

			// GPT卡密管理（供应商卡密）
			adminGroup.GET("/gpt-cards/suppliers", controller.GetSuppliers)            // 获取供应商列表
			adminGroup.GET("/gpt-cards", controller.GetGptCardList)                    // 获取卡密列表
			adminGroup.POST("/gpt-cards/batch-import", controller.BatchImportGptCards) // 批量导入
			adminGroup.POST("/gpt-cards/batch-check", controller.BatchCheckGptCards)   // 批量检查
			adminGroup.PUT("/gpt-cards/:id", controller.UpdateGptCard)                 // 更新卡密
			adminGroup.DELETE("/gpt-cards/:id", controller.DeleteGptCard)              // 删除卡密
			adminGroup.POST("/gpt-cards/batch-delete", controller.BatchDeleteGptCards) // 批量删除

			// GPT-CDK管理（自建CDK）
			adminGroup.GET("/gpt-cdk", controller.GetGptCdkList)                       // 获取CDK列表
			adminGroup.POST("/gpt-cdk/batch-generate", controller.BatchGenerateGptCdk) // 批量生成
			adminGroup.PUT("/gpt-cdk/:id", controller.UpdateGptCdk)                    // 更新CDK
			adminGroup.DELETE("/gpt-cdk/:id", controller.DeleteGptCdkSingle)           // 删除CDK
			adminGroup.POST("/gpt-cdk/batch-delete", controller.BatchDeleteGptCdks)    // 批量删除

			// Outlook OAuth 取件
			adminGroup.POST("/outlook-oauth/fetch", controller.OutlookOauthFetch)
			adminGroup.POST("/outlook-oauth/detail", controller.OutlookFetchDetail)

			// 微软邮箱库存
			adminGroup.GET("/microsoft-mails", controller.GetMicrosoftMailList)
			adminGroup.GET("/microsoft-mails/export", controller.ExportMicrosoftMails)
			adminGroup.GET("/microsoft-mails/by-card", controller.GetMicrosoftMailByCard)
			adminGroup.GET("/microsoft-mails/:id", controller.GetMicrosoftMailById)
			adminGroup.POST("/microsoft-mails", controller.CreateMicrosoftMail)
			adminGroup.PUT("/microsoft-mails/:id", controller.UpdateMicrosoftMail)
			adminGroup.DELETE("/microsoft-mails/:id", controller.DeleteMicrosoftMail)
			adminGroup.POST("/microsoft-mails/batch-import", controller.BatchImportMicrosoftMails)
			adminGroup.POST("/microsoft-mails/pickup", controller.PickupMicrosoftMail)
			adminGroup.POST("/microsoft-mails/complete-pickup", controller.CompleteMicrosoftMailPickup)
			adminGroup.POST("/microsoft-mails/rollback-pickup", controller.RollbackMicrosoftMailPickup)
			adminGroup.POST("/microsoft-mails/rollback-sold", controller.RollbackMicrosoftMailSold)
			adminGroup.POST("/microsoft-mails/batch-pickup", controller.BatchPickupMicrosoftMails)
			adminGroup.POST("/microsoft-mails/batch-check", controller.BatchCheckMicrosoftMails)
			adminGroup.POST("/microsoft-mails/batch-delete", controller.BatchDeleteMicrosoftMails)
			adminGroup.POST("/microsoft-mails/update-remark", controller.UpdateMicrosoftMailRemark)

			// Web 邮箱取件
			adminGroup.POST("/webmail/lqqq/fetch", controller.LqqqFetch)
			adminGroup.POST("/webmail/lqqq/detail", controller.LqqqDetail)
			adminGroup.POST("/webmail/toolsvip/fetch", controller.ToolsvipFetch)

			// GPT RT 许可证管理
			adminGroup.GET("/gpt-rt-licenses", controller.ListGptRtLicenses)
			adminGroup.GET("/gpt-rt-licenses/:id/devices", controller.ListGptRtLicenseDevices)
			adminGroup.POST("/gpt-rt-licenses", controller.CreateGptRtLicense)
			adminGroup.PUT("/gpt-rt-licenses/:id", controller.UpdateGptRtLicense)
			adminGroup.DELETE("/gpt-rt-licenses/:id", controller.DeleteGptRtLicense)

			// 广告配置
			adminGroup.GET("/ad-configs", controller.GetAdConfigList)
			adminGroup.GET("/ad-configs/:id", controller.GetAdConfigById)
			adminGroup.POST("/ad-configs", controller.CreateAdConfig)
			adminGroup.PUT("/ad-configs/:id", controller.UpdateAdConfig)
			adminGroup.DELETE("/ad-configs/:id", controller.DeleteAdConfig)
			adminGroup.POST("/upload/ad-image", controller.UploadAdImage)
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
		apiRouter.POST("/cdk/verify", controller.VerifyCard)
		apiRouter.POST("/cdk/top-up", controller.TopUp)
		apiRouter.POST("/cdk/query-task-status", controller.QueryTaskStatus)

		// GPT充值流程（公开，无需认证）
		apiRouter.POST("/topup/verify-cdk", controller.VerifyCdk)
		apiRouter.POST("/topup/start", controller.StartTopup)
		apiRouter.GET("/topup/task/:task_id", controller.GetTopupTaskStatus)
	}
}
