package app

import (
	"log/slog"

	"book-reader/backend/internal/config"
	"book-reader/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, server *Server) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recover(), middleware.CORS(), middleware.Logger(logger))
	server.RegisterRoutes(r)
	return r
}
