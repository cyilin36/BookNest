package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"book-reader/backend/internal/auth"
	"book-reader/backend/internal/common"
	"book-reader/backend/internal/config"
	"book-reader/backend/internal/middleware"
	"book-reader/backend/internal/model"
	"book-reader/backend/internal/parser"
	"book-reader/backend/internal/storage"
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
	u.StorageUsedBytes = s.storageUsedBytes(u.ID)
	return &u, nil
}

func (s *Server) VerifyAccessToken(token string) (*model.AccessClaims, error) {
	return s.tokens.VerifyAccessToken(token)
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.GET("/health", s.health)
	api.GET("/system/info", s.systemInfo)
	api.GET("/reader/books/:bookId/resources", s.readerResource)

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

	authRoutes.POST("/bookshelf/upload", middleware.BodyLimit(int64(s.cfg.RequestBodyLimitMB)*1024*1024), s.uploadPrivateBook)
	authRoutes.GET("/bookshelf", s.bookshelfList)
	authRoutes.GET("/bookshelf/:id", s.bookshelfDetail)
	authRoutes.PATCH("/bookshelf/:id", s.updateBookshelf)
	authRoutes.DELETE("/bookshelf/:id", s.deleteBookshelf)
	authRoutes.POST("/bookshelf/from-library/:bookId", s.addLibraryBookToShelf)
	authRoutes.POST("/library/books/:id/add-to-bookshelf", s.addLibraryBookToShelfAlias)

	authRoutes.GET("/library/books", s.libraryList)
	authRoutes.GET("/library/books/:id", s.libraryDetail)
	authRoutes.POST("/library/books/upload", middleware.BodyLimit(int64(s.cfg.RequestBodyLimitMB)*1024*1024), s.uploadPublicBook)

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

	r.NoRoute(s.serveFrontend)
}

func (s *Server) health(c *gin.Context) {
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{"status": "ok"})
}

func (s *Server) systemInfo(c *gin.Context) {
	settings := s.loadSettings()
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{
		"site_name":                     settings.SiteName,
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
	u.StorageUsedBytes = s.storageUsedBytes(u.ID)
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
	u.StorageUsedBytes = s.storageUsedBytes(u.ID)
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
		SiteName: "Book Reader", AllowRegistration: s.cfg.AllowRegistration,
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

func (s *Server) storageUsedBytes(userID int64) int64 {
	var total int64
	_ = s.db.Model(&model.Book{}).Where("owner_user_id = ? AND deleted_at IS NULL", userID).Select("COALESCE(SUM(file_size), 0)").Scan(&total).Error
	return total
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
