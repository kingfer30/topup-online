package router

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/constants"
)

func SetRouter(router *gin.Engine) {
	uploadsDir := filepath.Join(constants.GetDataDir(), "uploads")
	router.Static("/uploads", uploadsDir)
	SetApiRouter(router)
}
