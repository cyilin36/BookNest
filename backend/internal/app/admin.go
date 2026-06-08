package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"booknest/backend/internal/common"
	"booknest/backend/internal/middleware"
	"booknest/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) adminUsers(c *gin.Context) {
	page, size := common.ParsePagination(c.Query("page"), c.Query("page_size"))
	sortBy, order, ok := common.NormalizeSortOrder(c.Query("sort"), c.Query("order"), map[string]struct{}{
		"created_at":    {},
		"last_login_at": {},
		"username":      {},
	}, "created_at")
	if !ok {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
		return
	}
	q := s.db.Model(&model.User{})
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		q = q.Where("username ILIKE ? OR email ILIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	if st := c.Query("status"); st != "" {
		if st != model.UserStatusActive && st != model.UserStatusDisabled {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("status = ?", st)
	}
	if role := c.Query("role"); role != "" {
		if role != model.UserRoleAdmin && role != model.UserRoleUser {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("role = ?", role)
	}
	var total int64
	q.Count(&total)
	var users []model.User
	orderExpr := map[string]string{
		"created_at":    "created_at " + strings.ToUpper(order),
		"last_login_at": "last_login_at " + strings.ToUpper(order) + " NULLS LAST",
		"username":      "username " + strings.ToUpper(order),
	}[sortBy]
	if err := q.Order(orderExpr).Offset((page - 1) * size).Limit(size).Find(&users).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	for i := range users {
		users[i].StorageUsedBytes = s.storageUsedBytes(users[i].ID)
	}
	common.RespondPage(c, middleware.GetRequestID(c), users, page, size, total)
}

func (s *Server) adminUserDetail(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u, err := s.FindUserByID(id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), u)
}

func (s *Server) adminUpdateUser(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
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
	if v, exists := raw["storage_quota_bytes"]; exists {
		if isJSONNull(v) {
			updates["storage_quota_bytes"] = nil
		} else {
			var quota int64
			if err := json.Unmarshal(v, &quota); err != nil || quota < 0 {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			updates["storage_quota_bytes"] = quota
		}
	}
	res := s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		if isUniqueViolation(res.Error) {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrEmailExists)
			return
		}
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := s.db.Model(&model.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if count == 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
			return
		}
	}
	u, err := s.FindUserByID(id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), u)
}

func (s *Server) adminUpdateUserStatus(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if id == middleware.CurrentUser(c).ID {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrForbidden)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != model.UserStatusActive && req.Status != model.UserStatusDisabled) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	now := time.Now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{"status": req.Status, "updated_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrNotFound
		}
		if req.Status == model.UserStatusDisabled {
			return tx.Model(&model.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
		}
		return nil
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	u, err := s.FindUserByID(id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), u)
}

func (s *Server) adminUpdateUserRole(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if id == middleware.CurrentUser(c).ID {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrForbidden)
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Role != model.UserRoleAdmin && req.Role != model.UserRoleUser) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	res := s.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{"role": req.Role, "updated_at": time.Now()})
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	u, err := s.FindUserByID(id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), u)
}

func (s *Server) adminDeleteUser(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if id == middleware.CurrentUser(c).ID {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrForbidden)
		return
	}
	var target model.User
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	var ownedBooks []model.Book
	if err := s.db.Find(&ownedBooks, "owner_user_id = ?", id).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	for _, book := range ownedBooks {
		if err := s.removeBookFiles(book); err != nil && !os.IsNotExist(err) {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ownedBookIDs := make([]int64, 0, len(ownedBooks))
		for _, book := range ownedBooks {
			ownedBookIDs = append(ownedBookIDs, book.ID)
		}
		shelfIDs, err := userDeletionBookshelfIDs(tx, id, ownedBookIDs)
		if err != nil {
			return err
		}
		if len(shelfIDs) > 0 {
			if err := tx.Where("bookshelf_id IN ?", shelfIDs).Delete(&model.BookshelfTag{}).Error; err != nil {
				return err
			}
		}
		bookmarkQuery := tx.Where("user_id = ?", id)
		progressQuery := tx.Where("user_id = ?", id)
		if len(ownedBookIDs) > 0 {
			bookmarkQuery = bookmarkQuery.Or("book_id IN ?", ownedBookIDs)
			progressQuery = progressQuery.Or("book_id IN ?", ownedBookIDs)
		}
		if len(shelfIDs) > 0 {
			bookmarkQuery = bookmarkQuery.Or("bookshelf_id IN ?", shelfIDs)
			progressQuery = progressQuery.Or("bookshelf_id IN ?", shelfIDs)
		}
		if err := bookmarkQuery.Delete(&model.Bookmark{}).Error; err != nil {
			return err
		}
		if err := progressQuery.Delete(&model.ReadingProgress{}).Error; err != nil {
			return err
		}
		shelfQuery := tx.Where("user_id = ?", id)
		if len(ownedBookIDs) > 0 {
			shelfQuery = shelfQuery.Or("book_id IN ?", ownedBookIDs)
		}
		if err := shelfQuery.Delete(&model.Bookshelf{}).Error; err != nil {
			return err
		}
		if len(ownedBookIDs) > 0 {
			if err := tx.Where("book_id IN ?", ownedBookIDs).Delete(&model.BookCategory{}).Error; err != nil {
				return err
			}
			if err := tx.Where("book_id IN ?", ownedBookIDs).Delete(&model.BookTag{}).Error; err != nil {
				return err
			}
			if err := tx.Where("book_id IN ?", ownedBookIDs).Delete(&model.BookChapter{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", ownedBookIDs).Delete(&model.Book{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.RefreshToken{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.User{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrNotFound
		}
		return nil
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func userDeletionBookshelfIDs(tx *gorm.DB, userID int64, ownedBookIDs []int64) ([]int64, error) {
	q := tx.Model(&model.Bookshelf{}).Where("user_id = ?", userID)
	if len(ownedBookIDs) > 0 {
		q = q.Or("book_id IN ?", ownedBookIDs)
	}
	var ids []int64
	if err := q.Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Server) adminUpdateLibraryStatus(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status string  `json:"status"`
		Reason *string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !validLibraryShelfStatus(req.Status) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ?", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	if !validLibraryTransition(b.LibraryStatus, req.Status) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrConflict)
		return
	}
	now := time.Now()
	if err := s.db.Model(&model.Book{}).Where("id = ? AND visibility = ?", id, model.BookVisibilityPublic).Updates(map[string]any{"library_status": req.Status, "updated_at": now}).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	s.db.First(&b, "id = ?", id)
	common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(b, "", nil))
}

func (s *Server) adminDeleteLibraryBook(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ?", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	now := time.Now()
	deleted := model.LibraryStatusDeleted
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Book{}).Where("id = ? AND visibility = ?", id, model.BookVisibilityPublic).
			Updates(map[string]any{"library_status": deleted, "deleted_at": now, "updated_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrBookNotFound
		}
		return nil
	}); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if err := s.removeBookFiles(b); err != nil && !os.IsNotExist(err) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	if err := s.hardDeleteBookData(id, model.BookVisibilityPublic); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func validLibraryStatus(v string) bool {
	switch v {
	case model.LibraryStatusApproved, model.LibraryStatusHidden, model.LibraryStatusDeleted:
		return true
	}
	return false
}

func validLibraryShelfStatus(v string) bool {
	switch v {
	case model.LibraryStatusApproved, model.LibraryStatusHidden:
		return true
	}
	return false
}

func (s *Server) adminCreateCategory(c *gin.Context) { s.upsertCategory(c, 0) }
func (s *Server) adminUpdateCategory(c *gin.Context) {
	id, ok := parseID(c, "id")
	if ok {
		s.upsertCategory(c, id)
	}
}
func (s *Server) adminDeleteCategory(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if s.hasCategoryAssociations(id) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrConflict)
		return
	}
	res := s.db.Delete(&model.Category{}, id)
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrCategoryNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func (s *Server) upsertCategory(c *gin.Context, id int64) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	cat := model.Category{Name: strings.TrimSpace(req.Name), Slug: slug(req.Name), Description: req.Description, Scope: "system"}
	if id == 0 {
		if err := s.db.Create(&cat).Error; err != nil {
			common.RespondError(c, middleware.GetRequestID(c), conflictForUnique(err))
			return
		}
	} else {
		res := s.db.Model(&model.Category{}).Where("id = ?", id).Updates(map[string]any{"name": cat.Name, "slug": cat.Slug, "description": cat.Description, "updated_at": time.Now()})
		if res.Error != nil {
			common.RespondError(c, middleware.GetRequestID(c), conflictForUnique(res.Error))
			return
		}
		if res.RowsAffected == 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrCategoryNotFound)
			return
		}
		s.db.First(&cat, id)
	}
	common.RespondJSON(c, middleware.GetRequestID(c), cat)
}

func (s *Server) adminCreateTag(c *gin.Context) { s.upsertTag(c, 0) }
func (s *Server) adminUpdateTag(c *gin.Context) {
	id, ok := parseID(c, "id")
	if ok {
		s.upsertTag(c, id)
	}
}
func (s *Server) adminDeleteTag(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if s.hasTagAssociations(id) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrConflict)
		return
	}
	res := s.db.Delete(&model.Tag{}, id)
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrTagNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func (s *Server) upsertTag(c *gin.Context, id int64) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	tag := model.Tag{Name: strings.TrimSpace(req.Name), Slug: slug(req.Name), Description: req.Description, Scope: "system"}
	if id == 0 {
		if err := s.db.Create(&tag).Error; err != nil {
			common.RespondError(c, middleware.GetRequestID(c), conflictForUnique(err))
			return
		}
	} else {
		res := s.db.Model(&model.Tag{}).Where("id = ?", id).Updates(map[string]any{"name": tag.Name, "slug": tag.Slug, "description": tag.Description, "updated_at": time.Now()})
		if res.Error != nil {
			common.RespondError(c, middleware.GetRequestID(c), conflictForUnique(res.Error))
			return
		}
		if res.RowsAffected == 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrTagNotFound)
			return
		}
		s.db.First(&tag, id)
	}
	common.RespondJSON(c, middleware.GetRequestID(c), tag)
}

func slug(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "-")
	if v == "" {
		return "item"
	}
	return v
}

func conflictForUnique(err error) error {
	if isUniqueViolation(err) {
		return common.ErrConflict
	}
	return err
}

func dirSize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

func (s *Server) adminStorage(c *gin.Context) {
	var privateBytes, publicBytes int64
	s.db.Model(&model.Book{}).Where("visibility = ? AND deleted_at IS NULL", model.BookVisibilityPrivate).Select("COALESCE(SUM(file_size),0)").Scan(&privateBytes)
	s.db.Model(&model.Book{}).Where("visibility = ? AND deleted_at IS NULL", model.BookVisibilityPublic).Select("COALESCE(SUM(file_size),0)").Scan(&publicBytes)
	var total, priv, pub int64
	s.db.Model(&model.Book{}).Where("deleted_at IS NULL").Count(&total)
	s.db.Model(&model.Book{}).Where("visibility = ? AND deleted_at IS NULL", model.BookVisibilityPrivate).Count(&priv)
	s.db.Model(&model.Book{}).Where("visibility = ? AND deleted_at IS NULL", model.BookVisibilityPublic).Count(&pub)
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{"private_books_bytes": privateBytes, "public_books_bytes": publicBytes, "covers_bytes": dirSize(s.cfg.CoversDir), "total_books": total, "private_books": priv, "public_books": pub})
}

func (s *Server) adminSettings(c *gin.Context) {
	common.RespondJSON(c, middleware.GetRequestID(c), s.loadSettings())
}

func (s *Server) adminUpdateSettings(c *gin.Context) {
	var req SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil ||
		req.MaxUploadSizeMB <= 0 ||
		req.DefaultUserStorageQuotaMB < 0 ||
		req.MaxUploadSizeMB > s.cfg.RequestBodyLimitMB {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	if strings.TrimSpace(req.SiteName) == "" {
		req.SiteName = "BookNest"
	}
	if err := s.saveSettings(req); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.loadSettings())
}

func (s *Server) hasCategoryAssociations(categoryID int64) bool {
	var count int64
	_ = s.db.Table("book_categories").Where("category_id = ?", categoryID).Count(&count).Error
	if count > 0 {
		return true
	}
	_ = s.db.Table("bookshelves").Where("personal_category_id = ?", categoryID).Count(&count).Error
	return count > 0
}

func (s *Server) hasTagAssociations(tagID int64) bool {
	var count int64
	_ = s.db.Table("book_tags").Where("tag_id = ?", tagID).Count(&count).Error
	if count > 0 {
		return true
	}
	_ = s.db.Table("bookshelf_tags").Where("tag_id = ?", tagID).Count(&count).Error
	return count > 0
}

func validLibraryTransition(from *string, to string) bool {
	if from == nil {
		return to == model.LibraryStatusApproved || to == model.LibraryStatusHidden
	}
	switch *from {
	case model.LibraryStatusApproved:
		return to == model.LibraryStatusHidden
	case model.LibraryStatusHidden:
		return to == model.LibraryStatusApproved
	case model.LibraryStatusDeleted:
		return false
	default:
		return false
	}
}
