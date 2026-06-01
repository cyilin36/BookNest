package middleware

import (
	"book-reader/backend/internal/common"
	"book-reader/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil || u.Role != model.UserRoleAdmin {
			common.RespondError(c, GetRequestID(c), common.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
