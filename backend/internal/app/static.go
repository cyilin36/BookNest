package app

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"booknest/backend/internal/common"
	"booknest/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (s *Server) serveFrontend(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	indexPath := filepath.Join(s.cfg.FrontendDistDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "<!doctype html><html><head><title>BookNest</title></head><body><div id=\"app\">BookNest backend is running.</div></body></html>")
		return
	}
	requestPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	if requestPath != "" {
		staticPath := filepath.Join(s.cfg.FrontendDistDir, filepath.Clean(requestPath))
		if isPathInside(s.cfg.FrontendDistDir, staticPath) {
			if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
				http.ServeFile(c.Writer, c.Request, staticPath)
				return
			}
		}
	}
	http.ServeFile(c.Writer, c.Request, indexPath)
}

func isPathInside(root, child string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	childAbs, err := filepath.Abs(child)
	if err != nil {
		return false
	}
	return childAbs == rootAbs || strings.HasPrefix(childAbs, rootAbs+string(os.PathSeparator))
}
