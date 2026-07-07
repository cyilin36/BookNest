package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"booknest/backend/internal/auth"
	"booknest/backend/internal/common"
	"booknest/backend/internal/config"
	"booknest/backend/internal/middleware"
	"booknest/backend/internal/model"
	"booknest/backend/internal/parser"
	"booknest/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Server struct {
	cfg    *config.Config
	db     *gorm.DB
	tokens *auth.TokenService
	store  *storage.Local
}

func NewServer(cfg *config.Config, db *gorm.DB, store *storage.Local) *Server {
	return &Server{cfg: cfg, db: db, tokens: auth.NewTokenService(cfg), store: store}
}

func (s *Server) FindUserByID(id int64) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.fillUserComputed(&u)
	return &u, nil
}

func (s *Server) VerifyAccessToken(token string) (*model.AccessClaims, error) {
	return s.tokens.VerifyAccessToken(token)
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.GET("/health", s.health)
	api.GET("/system/info", s.systemInfo)
	api.GET("/system/icon", s.systemIcon)
	api.HEAD("/system/icon", s.systemIcon)
	api.GET("/system/login-background", s.loginBackground)
	api.HEAD("/system/login-background", s.loginBackground)
	api.GET("/reader/books/:bookId/resources", s.readerResource)
	api.GET("/users/:userId/avatar", s.userAvatar)
	api.HEAD("/users/:userId/avatar", s.userAvatar)

	api.POST("/auth/register", s.register)
	api.POST("/auth/login", s.login)
	api.POST("/auth/refresh", s.refresh)
	api.POST("/auth/logout", s.logout)

	authRoutes := api.Group("")
	authRoutes.Use(middleware.AuthRequired(s, s))
	authRoutes.GET("/auth/me", s.me)
	authRoutes.GET("/users/me", s.me)
	authRoutes.PATCH("/users/me", s.updateMe)
	authRoutes.PATCH("/users/me/password", s.changePassword)
	authRoutes.POST("/users/me/avatar/upload", middleware.BodyLimit(5*1024*1024), s.uploadAvatar)
	authRoutes.POST("/users/me/avatar/default", s.setDefaultAvatar)

	authRoutes.POST("/bookshelf/upload", middleware.BodyLimit(int64(s.cfg.RequestBodyLimitMB)*1024*1024), s.uploadPrivateBook)
	authRoutes.GET("/bookshelf", s.bookshelfList)
	authRoutes.GET("/bookshelf/:id", s.bookshelfDetail)
	authRoutes.GET("/bookshelf/:id/cover", s.bookshelfCover)
	authRoutes.HEAD("/bookshelf/:id/cover", s.bookshelfCover)
	authRoutes.GET("/bookshelf/:id/download", s.downloadBookshelfBook)
	authRoutes.PATCH("/bookshelf/:id", s.updateBookshelf)
	authRoutes.PUT("/bookshelf/:id/cover", middleware.BodyLimit(6*1024*1024), s.updateBookshelfCover)
	authRoutes.DELETE("/bookshelf/:id", s.deleteBookshelf)
	authRoutes.POST("/bookshelf/from-library/:bookId", s.addLibraryBookToShelf)
	authRoutes.POST("/library/books/:id/add-to-bookshelf", s.addLibraryBookToShelfAlias)

	authRoutes.GET("/library/books", s.libraryList)
	authRoutes.GET("/library/books/:id", s.libraryDetail)
	authRoutes.GET("/library/books/:id/download", s.downloadLibraryBook)
	authRoutes.PATCH("/library/books/:id", s.updateOwnLibraryBook)
	authRoutes.PUT("/library/books/:id/cover", middleware.BodyLimit(6*1024*1024), s.updateOwnLibraryBookCover)
	authRoutes.POST("/library/books/upload", middleware.BodyLimit(int64(s.cfg.RequestBodyLimitMB)*1024*1024), s.uploadPublicBook)
	authRoutes.POST("/library/books/:id/hide", s.hideOwnLibraryBook)
	authRoutes.POST("/library/books/:id/show", s.showOwnLibraryBook)

	authRoutes.GET("/categories", s.categories)
	authRoutes.GET("/tags", s.tags)

	authRoutes.GET("/reader/books/:bookId/meta", s.readerMeta)
	authRoutes.GET("/reader/books/:bookId/file", s.readerFile)
	authRoutes.GET("/reader/books/:bookId/cover", s.readerCover)
	authRoutes.GET("/reader/books/:bookId/text", s.readerText)
	authRoutes.GET("/reader/books/:bookId/chapters", s.readerChapters)
	authRoutes.GET("/reader/books/:bookId/chapters/:chapterId/content", s.readerChapterContent)
	authRoutes.GET("/reader/books/:bookId/progress", s.readerProgress)
	authRoutes.PUT("/reader/books/:bookId/progress", s.saveReaderProgress)

	authRoutes.GET("/books/:bookId/bookmarks", s.bookmarks)
	authRoutes.POST("/books/:bookId/bookmarks", s.createBookmark)
	authRoutes.PATCH("/bookmarks/:id", s.updateBookmark)
	authRoutes.DELETE("/bookmarks/:id", s.deleteBookmark)

	admin := authRoutes.Group("/admin")
	admin.Use(middleware.AdminRequired())
	admin.GET("/users", s.adminUsers)
	admin.GET("/users/:id", s.adminUserDetail)
	admin.PATCH("/users/:id", s.adminUpdateUser)
	admin.PATCH("/users/:id/status", s.adminUpdateUserStatus)
	admin.PATCH("/users/:id/role", s.adminUpdateUserRole)
	admin.DELETE("/users/:id", s.adminDeleteUser)
	admin.GET("/library/books", s.adminLibraryList)
	admin.PATCH("/library/books/:id/status", s.adminUpdateLibraryStatus)
	admin.DELETE("/library/books/:id", s.adminDeleteLibraryBook)
	admin.POST("/categories", s.adminCreateCategory)
	admin.PATCH("/categories/:id", s.adminUpdateCategory)
	admin.DELETE("/categories/:id", s.adminDeleteCategory)
	admin.POST("/tags", s.adminCreateTag)
	admin.PATCH("/tags/:id", s.adminUpdateTag)
	admin.DELETE("/tags/:id", s.adminDeleteTag)
	admin.GET("/system/storage", s.adminStorage)
	admin.GET("/system/settings", s.adminSettings)
	admin.PUT("/system/settings", s.adminUpdateSettings)
	admin.POST("/system/icon", middleware.BodyLimit(3*1024*1024), s.adminUploadSystemIcon)
	admin.DELETE("/system/icon", s.adminDeleteSystemIcon)
	admin.POST("/system/login-background", middleware.BodyLimit(10*1024*1024), s.adminUploadLoginBackground)
	admin.DELETE("/system/login-background", s.adminDeleteLoginBackground)

	r.NoRoute(s.serveFrontend)
}

func (s *Server) health(c *gin.Context) {
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{"status": "ok"})
}

func (s *Server) systemIcon(c *gin.Context) {
	rel := s.systemSettingValue("site_icon_path")
	if rel == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	path, err := s.store.AssetPath(rel)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	c.File(path)
}

func (s *Server) loginBackground(c *gin.Context) {
	rel := s.systemSettingValue("login_background_path")
	if rel == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	path, err := s.store.AssetPath(rel)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	c.File(path)
}

func (s *Server) systemInfo(c *gin.Context) {
	settings := s.loadSettings()
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{
		"site_name":                     settings.SiteName,
		"site_icon_url":                 settings.SiteIconURL,
		"login_background_url":          settings.LoginBackgroundURL,
		"allow_registration":            settings.AllowRegistration,
		"library_review_required":       settings.LibraryReviewRequired,
		"supported_formats":             []string{model.BookFormatEPUB, model.BookFormatPDF, model.BookFormatTXT},
		"max_upload_size_mb":            settings.MaxUploadSizeMB,
		"default_user_storage_quota_mb": settings.DefaultUserStorageQuotaMB,
	})
}

type registerReq struct {
	Username string  `json:"username"`
	Email    *string `json:"email"`
	Password string  `json:"password"`
	Nickname *string `json:"nickname"`
}

func (s *Server) register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Username) == "" || len(req.Password) < 6 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	var created *model.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("LOCK TABLE users IN EXCLUSIVE MODE").Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.User{}).Count(&count).Error; err != nil {
			return err
		}
		settings := s.loadSettingsTx(tx)
		if count > 0 && !settings.AllowRegistration {
			return common.ErrRegistrationDisable
		}
		var exists int64
		if err := tx.Model(&model.User{}).Where("username = ?", req.Username).Count(&exists).Error; err != nil {
			return err
		}
		if exists > 0 {
			return common.ErrUsernameExists
		}
		if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
			email := strings.TrimSpace(*req.Email)
			req.Email = &email
			if err := tx.Model(&model.User{}).Where("email = ?", email).Count(&exists).Error; err != nil {
				return err
			}
			if exists > 0 {
				return common.ErrEmailExists
			}
		} else {
			req.Email = nil
		}
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return err
		}
		role := model.UserRoleUser
		if count == 0 {
			role = model.UserRoleAdmin
		}
		u := model.User{Username: req.Username, Email: req.Email, PasswordHash: hash, Nickname: req.Nickname, Role: role, Status: model.UserStatusActive}
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		created = &u
		return nil
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	s.respondSession(c, created)
}

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (s *Server) login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	var u model.User
	if err := s.db.Where("username = ? OR email = ?", req.Login, req.Login).First(&u).Error; err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidCredentials)
		return
	}
	if u.Status != model.UserStatusActive {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrUserDisabled)
		return
	}
	now := time.Now()
	_ = s.db.Model(&u).Updates(map[string]any{"last_login_at": now, "updated_at": now}).Error
	u.LastLoginAt = &now
	s.respondSession(c, &u)
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) refresh(c *gin.Context) {
	token := s.refreshTokenFromRequest(c)
	if token == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrRefreshInvalid)
		return
	}
	hash := auth.HashRefreshToken(token)
	var rt model.RefreshToken
	if err := s.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).First(&rt).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrRefreshInvalid)
		return
	}
	u, err := s.FindUserByID(rt.UserID)
	if err != nil || u.Status != model.UserStatusActive {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrUserDisabled)
		return
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&model.RefreshToken{}).Where("id = ? AND revoked_at IS NULL", rt.ID).Update("revoked_at", now).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	s.respondSession(c, u)
}

func (s *Server) logout(c *gin.Context) {
	token := s.refreshTokenFromRequest(c)
	if token != "" {
		now := time.Now()
		_ = s.db.Model(&model.RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", auth.HashRefreshToken(token)).Update("revoked_at", now).Error
	}
	clearRefreshCookie(c)
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func (s *Server) me(c *gin.Context) {
	u := middleware.CurrentUser(c)
	s.fillUserComputed(u)
	common.RespondJSON(c, middleware.GetRequestID(c), u)
}

func (s *Server) updateMe(c *gin.Context) {
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	u := middleware.CurrentUser(c)
	updates := map[string]any{"updated_at": time.Now()}
	if v, exists := raw["email"]; exists {
		if isJSONNull(v) {
			updates["email"] = nil
		} else {
			var email string
			if err := json.Unmarshal(v, &email); err != nil {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			email = strings.TrimSpace(email)
			if email == "" {
				updates["email"] = nil
			} else {
				updates["email"] = email
			}
		}
	}
	if v, exists := raw["nickname"]; exists {
		if isJSONNull(v) {
			updates["nickname"] = nil
		} else {
			var nickname string
			if err := json.Unmarshal(v, &nickname); err != nil {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			updates["nickname"] = strings.TrimSpace(nickname)
		}
	}
	if err := s.db.Model(u).Updates(updates).Error; err != nil {
		if isUniqueViolation(err) {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrEmailExists)
			return
		}
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	fresh, _ := s.FindUserByID(u.ID)
	common.RespondJSON(c, middleware.GetRequestID(c), fresh)
}

func (s *Server) changePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.NewPassword) < 6 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	u := middleware.CurrentUser(c)
	if !auth.CheckPassword(u.PasswordHash, req.OldPassword) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidCredentials)
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	now := time.Now()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", u.ID).Updates(map[string]any{"password_hash": hash, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&model.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", u.ID).Update("revoked_at", now).Error
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	clearRefreshCookie(c)
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func (s *Server) uploadAvatar(c *gin.Context) {
	u := middleware.CurrentUser(c)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	defer file.Close()

	ext, contentType, ok := avatarFormat(header.Filename, header.Header.Get("Content-Type"))
	if !ok {
		common.RespondError(c, middleware.GetRequestID(c), &common.APIError{
			Code:    "avatar_format_not_supported",
			Message: "Only PNG, JPEG, WebP, GIF images are supported",
			Status:  400,
		})
		return
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	if n > 0 {
		file.Seek(0, 0)
	}
	if !validAvatarBytes(buf[:n], contentType) {
		common.RespondError(c, middleware.GetRequestID(c), &common.APIError{
			Code:    "avatar_content_invalid",
			Message: "File content does not match expected image format",
			Status:  400,
		})
		return
	}

	if header.Size > 5*1024*1024 {
		common.RespondError(c, middleware.GetRequestID(c), &common.APIError{
			Code:    "avatar_too_large",
			Message: "Avatar size must not exceed 5 MB",
			Status:  400,
		})
		return
	}

	relPath, err := s.store.SaveUserAvatar(file, u.ID, ext)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}

	oldPath := u.AvatarPath
	if err := s.db.Model(&model.User{}).Where("id = ?", u.ID).Updates(map[string]any{
		"avatar_path": relPath,
		"updated_at":  time.Now(),
	}).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}

	if oldPath != nil && !strings.HasPrefix(*oldPath, "default/") {
		if absPath, err := s.store.AssetPath(*oldPath); err == nil {
			_ = os.Remove(absPath)
		}
	}

	fresh, _ := s.FindUserByID(u.ID)
	common.RespondJSON(c, middleware.GetRequestID(c), fresh)
}

func (s *Server) setDefaultAvatar(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var req struct {
		AvatarName string `json:"avatar_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AvatarName == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}

	allowed := []string{"default1", "default2", "default3", "default4", "default5", "default6"}
	valid := false
	for _, name := range allowed {
		if req.AvatarName == name {
			valid = true
			break
		}
	}
	if !valid {
		common.RespondError(c, middleware.GetRequestID(c), &common.APIError{
			Code:    "invalid_avatar_name",
			Message: "Invalid default avatar name",
			Status:  400,
		})
		return
	}

	defaultPath := "default/" + req.AvatarName + ".svg"
	oldPath := u.AvatarPath
	if err := s.db.Model(&model.User{}).Where("id = ?", u.ID).Updates(map[string]any{
		"avatar_path": defaultPath,
		"updated_at":  time.Now(),
	}).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}

	if oldPath != nil && !strings.HasPrefix(*oldPath, "default/") {
		if absPath, err := s.store.AssetPath(*oldPath); err == nil {
			_ = os.Remove(absPath)
		}
	}

	fresh, _ := s.FindUserByID(u.ID)
	common.RespondJSON(c, middleware.GetRequestID(c), fresh)
}

func (s *Server) userAvatar(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}

	var avatarPath *string
	if err := s.db.Model(&model.User{}).Select("avatar_path").Where("id = ?", userID).Scan(&avatarPath).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}

	if avatarPath == nil || strings.TrimSpace(*avatarPath) == "" {
		c.Status(404)
		return
	}

	absPath, err := s.store.AssetPath(*avatarPath)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		c.Status(404)
		return
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	contentType := "application/octet-stream"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".webp":
		contentType = "image/webp"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600")
	c.File(absPath)
}

func avatarFormat(filename, contentTypeValue string) (string, string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	normalizedContentType := strings.ToLower(strings.TrimSpace(strings.Split(contentTypeValue, ";")[0]))
	formats := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".webp": "image/webp",
		".gif":  "image/gif",
	}
	if ct, ok := formats[ext]; ok {
		return ext, ct, true
	}
	for allowedExt, allowedContentType := range formats {
		if normalizedContentType == allowedContentType {
			return allowedExt, allowedContentType, true
		}
	}
	return "", "", false
}

func validAvatarBytes(data []byte, expectedContentType string) bool {
	if len(data) < 4 {
		return false
	}
	switch expectedContentType {
	case "image/png":
		return len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47
	case "image/jpeg":
		return data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
	case "image/webp":
		return len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
			data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50
	case "image/gif":
		return len(data) >= 6 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 &&
			data[3] == 0x38 && (data[4] == 0x37 || data[4] == 0x39) && data[5] == 0x61
	}
	return false
}

func (s *Server) respondSession(c *gin.Context, u *model.User) {
	access, expires, err := s.tokens.CreateAccessToken(u)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	refreshToken, refreshHash, err := auth.NewRefreshToken()
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	ua := c.Request.UserAgent()
	ip := c.ClientIP()
	rt := model.RefreshToken{
		UserID: u.ID, TokenHash: refreshHash, UserAgent: &ua, IPAddress: &ip,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}
	if err := s.db.Create(&rt).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	setRefreshCookie(c, refreshToken, s.cfg.RefreshTokenTTL, s.cfg.AppEnv == "production")
	s.fillUserComputed(u)
	common.RespondJSON(c, middleware.GetRequestID(c), AuthSession{AccessToken: access, TokenType: "Bearer", ExpiresIn: expires, User: u})
}

func (s *Server) chapterContentFromBook(book model.Book, ch model.BookChapter, userID int64) (string, string, error) {
	path, err := s.store.BookPath(book.FilePath)
	if err != nil {
		return "", "", err
	}
	switch book.Format {
	case model.BookFormatTXT:
		content, err := parser.ReadTXTContent(path, ch)
		return "text", content, err
	case model.BookFormatPDF:
		return "html", parser.PDFChapterHTML(ch), nil
	case model.BookFormatEPUB:
		content, err := parser.ReadEPUBContentWithResourceURL(path, ch, func(href string) string {
			return s.epubResourceURL(book.ID, userID, href)
		})
		return "html", content, err
	default:
		return "text", "", common.ErrUnsupportedMedia
	}
}

func (s *Server) refreshTokenFromRequest(c *gin.Context) string {
	if cookie, err := c.Cookie("refresh_token"); err == nil && cookie != "" {
		return cookie
	}
	var req refreshReq
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	return req.RefreshToken
}

func setRefreshCookie(c *gin.Context, token string, ttl time.Duration, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "refresh_token", Value: token, Path: "/api/v1/auth",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
		MaxAge: int(ttl.Seconds()),
	})
}

func clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: "refresh_token", Value: "", Path: "/api/v1/auth", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

func (s *Server) loadSettings() SystemSettings {
	return s.loadSettingsTx(s.db)
}

func (s *Server) loadSettingsTx(tx *gorm.DB) SystemSettings {
	settings := SystemSettings{
		SiteName: "BookNest", AllowRegistration: s.cfg.AllowRegistration,
		LibraryReviewRequired:     s.cfg.LibraryReviewRequired,
		MaxUploadSizeMB:           s.cfg.MaxUploadSizeMB,
		DefaultUserStorageQuotaMB: s.cfg.DefaultUserStorageQuotaMB,
	}
	var rows []model.SystemSetting
	if err := tx.Find(&rows).Error; err != nil {
		return settings
	}
	for _, row := range rows {
		switch row.Key {
		case "site_name":
			settings.SiteName = row.Value
		case "site_icon_path":
			settings.SiteIconURL = siteIconURL(row.Value)
		case "login_background_path":
			settings.LoginBackgroundURL = loginBackgroundURL(row.Value)
		case "allow_registration":
			settings.AllowRegistration = row.Value == "true"
		case "library_review_required":
			settings.LibraryReviewRequired = row.Value == "true"
		case "max_upload_size_mb":
			if n, err := strconv.Atoi(row.Value); err == nil {
				settings.MaxUploadSizeMB = n
			}
		case "default_user_storage_quota_mb":
			if n, err := strconv.Atoi(row.Value); err == nil {
				settings.DefaultUserStorageQuotaMB = n
			}
		}
	}
	return settings
}

func (s *Server) saveSettings(settings SystemSettings) error {
	values := map[string]string{
		"site_name":                     settings.SiteName,
		"allow_registration":            strconv.FormatBool(settings.AllowRegistration),
		"library_review_required":       strconv.FormatBool(settings.LibraryReviewRequired),
		"max_upload_size_mb":            strconv.Itoa(settings.MaxUploadSizeMB),
		"default_user_storage_quota_mb": strconv.Itoa(settings.DefaultUserStorageQuotaMB),
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for k, v := range values {
			row := model.SystemSetting{Key: k, Value: v, UpdatedAt: time.Now()}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.Assignments(map[string]any{"value": v, "updated_at": time.Now()}),
			}).Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Server) systemSettingValue(key string) string {
	var row model.SystemSetting
	if err := s.db.First(&row, "key = ?", key).Error; err != nil {
		return ""
	}
	return row.Value
}

func (s *Server) saveSystemSetting(key, value string) error {
	row := model.SystemSetting{Key: key, Value: value, UpdatedAt: time.Now()}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{"value": value, "updated_at": time.Now()}),
	}).Create(&row).Error
}

func (s *Server) removeAssetFile(relative string) error {
	path, err := s.store.AssetPath(relative)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// storageUsedBytes 返回用户已用的存储空间（字节）。
// 只统计私人书籍（visibility = private），公共图书不计入用户配额。
func (s *Server) storageUsedBytes(userID int64) int64 {
	var total int64
	_ = s.db.Model(&model.Book{}).
		Where("owner_user_id = ? AND visibility = ? AND deleted_at IS NULL", userID, model.BookVisibilityPrivate).
		Select("COALESCE(SUM(file_size), 0)").Scan(&total).Error
	return total
}

// storageQuotaEnforced 判断是否对该用户执行存储配额限制。
// 管理员不受配额限制。
func (s *Server) storageQuotaEnforced(u *model.User) bool {
	if u == nil {
		return false
	}
	return u.Role != model.UserRoleAdmin
}

// effectiveStorageQuotaBytes 返回该用户实际生效的存储配额（字节）。
// 用户设置了专属配额时使用专属值，否则回退到全局默认配额。
// 返回值 <= 0 表示不限制。
func (s *Server) effectiveStorageQuotaBytes(u *model.User) int64 {
	if u != nil && u.StorageQuotaBytes != nil {
		return *u.StorageQuotaBytes
	}
	return int64(s.loadSettings().DefaultUserStorageQuotaMB) * 1024 * 1024
}

// fillUserComputed 填充 User 的计算字段：已用存储、生效配额、头像 URL。
// EffectiveStorageQuotaBytes 为 nil 表示不限制（管理员，或生效配额 <= 0）。
func (s *Server) fillUserComputed(u *model.User) {
	if u == nil {
		return
	}
	u.StorageUsedBytes = s.storageUsedBytes(u.ID)
	u.AvatarURL = avatarURL(u.ID, u.AvatarPath)
	if !s.storageQuotaEnforced(u) {
		u.EffectiveStorageQuotaBytes = nil
		return
	}
	quota := s.effectiveStorageQuotaBytes(u)
	if quota <= 0 {
		u.EffectiveStorageQuotaBytes = nil
		return
	}
	u.EffectiveStorageQuotaBytes = &quota
}

func bookFormat(filename string) (string, bool) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".epub":
		return model.BookFormatEPUB, true
	case ".pdf":
		return model.BookFormatPDF, true
	case ".txt":
		return model.BookFormatTXT, true
	default:
		return "", false
	}
}

func siteIconFormat(filename, contentTypeValue string) (string, string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	normalizedContentType := strings.ToLower(strings.TrimSpace(strings.Split(contentTypeValue, ";")[0]))
	formats := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
	}
	if contentTypeValue, ok := formats[ext]; ok {
		return ext, contentTypeValue, true
	}
	for allowedExt, allowedContentType := range formats {
		if normalizedContentType == allowedContentType {
			return allowedExt, allowedContentType, true
		}
	}
	if normalizedContentType == "image/vnd.microsoft.icon" {
		return ".ico", "image/x-icon", true
	}
	return "", "", false
}

func validSiteIconBytes(ext string, data []byte) bool {
	if len(data) == 0 {
		return false
	}
	switch ext {
	case ".png":
		return http.DetectContentType(data) == "image/png"
	case ".jpg", ".jpeg":
		return http.DetectContentType(data) == "image/jpeg"
	case ".webp":
		return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	case ".ico":
		return len(data) >= 4 && data[0] == 0 && data[1] == 0 && data[2] == 1 && data[3] == 0
	case ".svg":
		text := strings.ToLower(strings.TrimSpace(string(data)))
		return strings.HasPrefix(text, "<svg") || (strings.HasPrefix(text, "<?xml") && strings.Contains(text, "<svg"))
	default:
		return false
	}
}

func nullableString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "idx_users_email_unique_not_null") ||
		strings.Contains(msg, "users_username_key") ||
		strings.Contains(msg, "idx_categories_system_slug") ||
		strings.Contains(msg, "idx_tags_system_slug")
}

func parseID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
		return 0, false
	}
	return id, true
}

func coverURL(bookID int64, coverPath *string) *string {
	if coverPath == nil || *coverPath == "" {
		return nil
	}
	v := fmt.Sprintf("/api/v1/reader/books/%d/cover", bookID)
	return &v
}

func bookshelfCoverURL(bookshelfID int64, coverPath *string) *string {
	if coverPath == nil || *coverPath == "" {
		return nil
	}
	v := fmt.Sprintf("/api/v1/bookshelf/%d/cover", bookshelfID)
	return &v
}

func siteIconURL(iconPath string) *string {
	if strings.TrimSpace(iconPath) == "" {
		return nil
	}
	v := "/api/v1/system/icon"
	return &v
}

func loginBackgroundURL(backgroundPath string) *string {
	if strings.TrimSpace(backgroundPath) == "" {
		return nil
	}
	v := "/api/v1/system/login-background"
	return &v
}

func avatarURL(userID int64, avatarPath *string) *string {
	if avatarPath == nil || strings.TrimSpace(*avatarPath) == "" {
		return nil
	}
	v := fmt.Sprintf("/api/v1/users/%d/avatar", userID)
	return &v
}

func bookMetaDTO(b model.Book) BookMetaDTO {
	return BookMetaDTO{
		ID: b.ID, Title: b.Title, Author: b.Author, Description: b.Description, Format: b.Format,
		CoverURL: coverURL(b.ID, b.CoverPath), FileSize: b.FileSize, Visibility: b.Visibility,
		LibraryStatus: b.LibraryStatus, ParseStatus: b.ParseStatus, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

func apiErrOrInternal(err error) error {
	var api *common.APIError
	if errors.As(err, &api) {
		return api
	}
	return common.ErrInternal
}

func contentType(path string) string {
	if ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func copyRange(dst io.Writer, src io.Reader, n int64) error {
	_, err := io.CopyN(dst, src, n)
	if err == io.EOF {
		return nil
	}
	return err
}
