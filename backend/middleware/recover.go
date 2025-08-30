package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/request"
)

func RelayPanicRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				logger.Errorf(ctx, fmt.Sprintf("panic detected: %v", err))
				logger.Errorf(ctx, fmt.Sprintf("stacktrace from panic: %s", string(debug.Stack())))
				logger.Errorf(ctx, fmt.Sprintf("request: %s %s", c.Request.Method, c.Request.URL.Path))
				body, _ := request.GetRequestBody(c)
				logger.Errorf(ctx, fmt.Sprintf("request body: %s", string(body)))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("Panic detected, error: %v. Please concat us at https://t.me/aiguoguo199", err),
						"type":    "guoguo_api_panic",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
