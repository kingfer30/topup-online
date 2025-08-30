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

		apiRouter.GET("/cdk/:number", controller.GetCardDetail)
		apiRouter.POST("/cdk/check", controller.GetCardDetail)
		apiRouter.POST("/cdk/top-up", controller.GetCardDetail)
		apiRouter.GET("/cdk/thread/:id", controller.GetCardDetail)
	}
}
