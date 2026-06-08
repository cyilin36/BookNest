package middleware

import (
	"strconv"
	"strings"

	"booknest/backend/internal/common"
	"booknest/backend/internal/model"
	"github.com/gin-gonic/gin"
)

type TokenVerifier interface {
	VerifyAccessToken(token string) (*model.AccessClaims, error)
}

type UserFinder interface {
	FindUserByID(id int64) (*model.User, error)
}

func AuthRequired(v TokenVerifier, f UserFinder) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			common.RespondError(c, GetRequestID(c), common.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := v.VerifyAccessToken(strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")))
		if err != nil {
			common.RespondError(c, GetRequestID(c), common.ErrUnauthorized)
			c.Abort()
			return
		}
		id, _ := strconv.ParseInt(claims.Subject, 10, 64)
		user, err := f.FindUserByID(id)
		if err != nil || user.Status != model.UserStatusActive {
			common.RespondError(c, GetRequestID(c), common.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(UserIDKey, user.ID)
		c.Set("user", user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *model.User {
	if v, ok := c.Get("user"); ok {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return nil
}
