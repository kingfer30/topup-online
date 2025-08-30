package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/utils/helper"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := helper.GenRequestID()
		c.Set(constants.RequestIdKey, id)
		ctx := context.WithValue(c.Request.Context(), constants.RequestIdKey, id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(constants.RequestIdKey, id)
		c.Next()
	}
}
