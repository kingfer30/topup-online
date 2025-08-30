package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/constants"
)

func SetUpLogger(server *gin.Engine) {
	server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var requestID = param.Keys[constants.RequestIdKey].(string)
		var log string
		log = fmt.Sprintf("[GIN] %s | %s | %3d | %13v | %15s | %7s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			requestID,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
		return log
	}))
}
