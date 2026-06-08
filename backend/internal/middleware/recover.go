package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"booknest/backend/internal/common"
	"github.com/gin-gonic/gin"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("panic request_id=%s panic=%v stack=%s\n", GetRequestID(c), r, string(debug.Stack()))
				common.RespondError(c, GetRequestID(c), common.ErrInternal)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
