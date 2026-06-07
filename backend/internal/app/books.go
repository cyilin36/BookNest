package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"book-reader/backend/internal/common"
	"book-reader/backend/internal/middleware"
	"book-reader/backend/internal/model"
	"book-reader/backend/internal/parser"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const epubResourceURLTTL = 30 * time.Minute

func (s *Server) uploadPrivateBook(c *gin.Context) {
	s.uploadBook(c, model.BookVisibilityPrivate)
}

func (s *Server) uploadPublicBook(c *gin.Context) {
	s.uploadBook(c, model.BookVisibilityPublic)
}

func (s *Server) uploadBook(c *gin.Context, visibility string) {
	u := middleware.CurrentUser(c)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	defer file.Close()
	format, ok := bookFormat(header.Filename)
	if !ok {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFormatNotSup)
		return
	}
	stored, err := s.store.SaveUpload(file, header, visibility, format)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if stored.Size > int64(s.loadSettings().MaxUploadSizeMB)*1024*1024 {
		_ = os.Remove(stored.AbsolutePath)
		common.RespondError(c, middleware.GetRequestID(c), common.ErrPayloadTooLarge)
		return
	}
	if visibility == model.BookVisibilityPrivate {
		quota := int64(s.loadSettings().DefaultUserStorageQuotaMB) * 1024 * 1024
		if u.StorageQuotaBytes != nil {
			quota = *u.StorageQuotaBytes
		}
		if quota > 0 && s.storageUsedBytes(u.ID)+stored.Size > quota {
			_ = os.Remove(stored.AbsolutePath)
			common.RespondError(c, middleware.GetRequestID(c), common.ErrStorageQuotaExceeded)
			return
		}
	}
	title := c.PostForm("title")
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSuffix(header.Filename, "."+format)
	}
	categoryIDs, err := parseCSVInt64s(c.PostForm("category_ids"))
	if err != nil {
		_ = os.Remove(stored.AbsolutePath)
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	tagIDs, err := parseCSVInt64s(c.PostForm("tag_ids"))
	if err != nil {
		_ = os.Remove(stored.AbsolutePath)
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	categoryIDs, err = s.validSystemCategoryIDs(categoryIDs)
	if err != nil {
		_ = os.Remove(stored.AbsolutePath)
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	tagIDs, err = s.validSystemTagIDs(tagIDs)
	if err != nil {
		_ = os.Remove(stored.AbsolutePath)
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	book := model.Book{
		Title: title, Author: nullableString(c.PostForm("author")),
		Description:      nullableString(c.PostForm("description")),
		OriginalFilename: &header.Filename, Format: format, FileSize: stored.Size,
		FileHash: stored.Hash, FilePath: stored.RelativePath, Visibility: visibility,
		OwnerUserID: u.ID, ParseStatus: "partial",
	}
	if visibility == model.BookVisibilityPublic {
		status := model.LibraryStatusApproved
		book.LibraryStatus = &status
	}
	var shelf *model.Bookshelf
	var savedCoverAbs string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&book).Error; err != nil {
			return err
		}
		if format == model.BookFormatEPUB {
			if cover, err := parser.ExtractEPUBCover(stored.AbsolutePath); err == nil && cover != nil {
				coverRel, coverAbs, err := s.store.SaveCoverBytes(cover.Data, book.ID, filepath.Ext(cover.Name))
				if err == nil {
					savedCoverAbs = coverAbs
					book.CoverPath = &coverRel
					if err := tx.Model(&book).Updates(map[string]any{"cover_path": coverRel, "updated_at": time.Now()}).Error; err != nil {
						return err
					}
				}
			}
		}
		parsed, err := parser.Parse(stored.AbsolutePath, format, header.Filename, book.ID)
		if err == nil {
			book.ParseStatus = parsed.ParseStatus
			if strings.TrimSpace(c.PostForm("title")) == "" && parsed.Title != "" {
				book.Title = parsed.Title
			}
			if err := tx.Model(&book).Updates(map[string]any{"title": book.Title, "parse_status": book.ParseStatus, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
			for i := range parsed.Chapters {
				parsed.Chapters[i].BookID = book.ID
				parsed.Chapters[i].ChapterIndex = i
			}
			if len(parsed.Chapters) > 0 {
				if err := tx.Create(&parsed.Chapters).Error; err != nil {
					return err
				}
			}
		} else {
			msg := err.Error()
			_ = tx.Model(&book).Updates(map[string]any{"parse_status": "failed", "parse_error": msg}).Error
		}
		if visibility == model.BookVisibilityPrivate {
			var personalCategoryID *int64
			if len(categoryIDs) > 0 {
				personalCategoryID = &categoryIDs[0]
			}
			item := model.Bookshelf{UserID: u.ID, BookID: book.ID, SourceType: model.BookshelfSourceUploaded, Status: model.BookshelfStatusActive, AddedAt: time.Now(), PersonalCategoryID: personalCategoryID}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			if len(tagIDs) > 0 {
				rows := make([]model.BookshelfTag, 0, len(tagIDs))
				for _, tagID := range tagIDs {
					rows = append(rows, model.BookshelfTag{BookshelfID: item.ID, TagID: tagID})
				}
				if err := tx.Create(&rows).Error; err != nil {
					return err
				}
			}
			shelf = &item
		} else {
			if len(categoryIDs) > 0 {
				rows := make([]model.BookCategory, 0, len(categoryIDs))
				for _, categoryID := range categoryIDs {
					rows = append(rows, model.BookCategory{BookID: book.ID, CategoryID: categoryID})
				}
				if err := tx.Create(&rows).Error; err != nil {
					return err
				}
			}
			if len(tagIDs) > 0 {
				rows := make([]model.BookTag, 0, len(tagIDs))
				for _, tagID := range tagIDs {
					rows = append(rows, model.BookTag{BookID: book.ID, TagID: tagID})
				}
				if err := tx.Create(&rows).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		_ = os.Remove(stored.AbsolutePath)
		if savedCoverAbs != "" {
			_ = os.Remove(savedCoverAbs)
		}
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if visibility == model.BookVisibilityPrivate {
		common.RespondJSON(c, middleware.GetRequestID(c), s.bookshelfDTO(*shelf, book, nil))
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(book, u.Username, nil))
}

func (s *Server) bookshelfList(c *gin.Context) {
	u := middleware.CurrentUser(c)
	page, size := common.ParsePagination(c.Query("page"), c.Query("page_size"))
	sortBy, orderBy, ok := common.NormalizeSortOrder(c.Query("sort"), c.Query("order"), map[string]struct{}{
		"last_read_at": {},
		"added_at":     {},
		"title":        {},
	}, "")
	if !ok {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
		return
	}
	var total int64
	q := s.db.Table("bookshelves bs").Joins("JOIN books b ON b.id = bs.book_id").Where("bs.user_id = ? AND bs.status = ?", u.ID, model.BookshelfStatusActive)
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		q = q.Where("(b.title ILIKE ? OR b.author ILIKE ? OR bs.personal_title ILIKE ?)", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
	}
	if f := c.Query("format"); f != "" {
		if !validBookFormat(f) {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("b.format = ?", f)
	}
	if st := c.Query("source_type"); st != "" {
		if st != model.BookshelfSourceUploaded && st != model.BookshelfSourceLibrary {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("bs.source_type = ?", st)
	}
	if fav := c.Query("favorite"); fav != "" {
		if fav != "true" && fav != "false" {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("bs.favorite = ?", fav == "true")
	}
	if categoryID, exists, err := queryInt64Param(c, "category_id"); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	} else if exists {
		q = q.Where("EXISTS (SELECT 1 FROM bookshelves bs2 WHERE bs2.id = bs.id AND bs2.personal_category_id = ?)", categoryID)
	}
	if tagID, exists, err := queryInt64Param(c, "tag_id"); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	} else if exists {
		q = q.Where("EXISTS (SELECT 1 FROM bookshelf_tags bt WHERE bt.bookshelf_id = bs.id AND bt.tag_id = ?)", tagID)
	}
	q.Count(&total)
	type row struct {
		model.Bookshelf
		BookID2            int64 `gorm:"column:book_id2"`
		Title              string
		Author             *string
		Description        *string
		Format             string
		FilePath           string
		CoverPath          *string
		Visibility         string
		LibraryStatus      *string
		DeletedAt          *time.Time
		ProgressPercentage *float64
	}
	var rows []row
	order := bookshelfOrder(sortBy, orderBy)
	err := q.Select("bs.*, b.id AS book_id2, b.title, b.author, b.description, b.format, b.file_path, b.cover_path, b.visibility, b.library_status, b.deleted_at, rp.percentage AS progress_percentage").
		Joins("LEFT JOIN reading_progress rp ON rp.user_id = bs.user_id AND rp.book_id = bs.book_id").
		Order(order).Offset((page - 1) * size).Limit(size).Scan(&rows).Error
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	out := make([]BookshelfItemDTO, 0, len(rows))
	for _, r := range rows {
		b := model.Book{ID: r.BookID2, Title: r.Title, Author: r.Author, Description: r.Description, Format: r.Format, FilePath: r.FilePath, CoverPath: r.CoverPath, Visibility: r.Visibility, LibraryStatus: r.LibraryStatus, DeletedAt: r.DeletedAt}
		out = append(out, s.bookshelfDTO(r.Bookshelf, b, r.ProgressPercentage))
	}
	common.RespondPage(c, middleware.GetRequestID(c), out, page, size, total)
}

func (s *Server) bookshelfDetail(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	item, book, err := s.getBookshelfWithBook(middleware.CurrentUser(c).ID, id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.bookshelfDTO(item, book, s.bookshelfProgressFor(middleware.CurrentUser(c).ID, book.ID)))
}

func (s *Server) downloadBookshelfBook(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	item, book, err := s.getBookshelfWithBook(u.ID, id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	if book.Visibility == model.BookVisibilityPublic && (book.LibraryStatus == nil || *book.LibraryStatus != model.LibraryStatusApproved || book.DeletedAt != nil) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotAccessible)
		return
	}
	displayTitle := book.Title
	if item.SourceType == model.BookshelfSourceLibrary && item.PersonalTitle != nil && strings.TrimSpace(*item.PersonalTitle) != "" {
		displayTitle = *item.PersonalTitle
	}
	s.downloadBookFile(c, book, displayTitle)
}

func (s *Server) updateBookshelf(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	u := middleware.CurrentUser(c)
	item, book, err := s.getBookshelfWithBook(u.ID, id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	updates := map[string]any{}
	bookUpdates := map[string]any{}
	if v, exists := bookInfoRaw(raw, "title"); exists {
		title, err := requiredString(v)
		if err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if item.SourceType == model.BookshelfSourceUploaded && book.Visibility == model.BookVisibilityPrivate {
			bookUpdates["title"] = title
		} else {
			updates["personal_title"] = title
		}
	}
	if v, exists := raw["author"]; exists {
		author, err := nullableTrimmedString(v)
		if err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if item.SourceType == model.BookshelfSourceUploaded && book.Visibility == model.BookVisibilityPrivate {
			bookUpdates["author"] = author
		} else {
			updates["personal_author"] = author
		}
	}
	if v, exists := raw["description"]; exists {
		description, err := nullableTrimmedString(v)
		if err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if item.SourceType == model.BookshelfSourceUploaded && book.Visibility == model.BookVisibilityPrivate {
			bookUpdates["description"] = description
		} else {
			updates["personal_description"] = description
		}
	}
	_, updatesPersonalAuthor := updates["personal_author"]
	_, updatesPersonalDescription := updates["personal_description"]
	if item.SourceType == model.BookshelfSourceLibrary && item.PersonalTitle == nil && (updatesPersonalAuthor || updatesPersonalDescription) {
		updates["personal_title"] = book.Title
	}
	if v, exists := raw["personal_category_id"]; exists {
		if isJSONNull(v) {
			updates["personal_category_id"] = nil
		} else {
			var categoryID int64
			if err := json.Unmarshal(v, &categoryID); err != nil || categoryID <= 0 {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			ids, err := s.validSystemCategoryIDs([]int64{categoryID})
			if err != nil {
				common.RespondError(c, middleware.GetRequestID(c), err)
				return
			}
			updates["personal_category_id"] = ids[0]
		}
	}
	if v, exists := raw["favorite"]; exists {
		var favorite bool
		if err := json.Unmarshal(v, &favorite); err != nil {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		updates["favorite"] = favorite
	}
	if v, exists := raw["pinned"]; exists {
		var pinned bool
		if err := json.Unmarshal(v, &pinned); err != nil {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		updates["pinned"] = pinned
	}
	var tagIDs []int64
	replaceTags := false
	if v, exists := raw["tag_ids"]; exists {
		replaceTags = true
		if isJSONNull(v) {
			tagIDs = nil
		} else if err := json.Unmarshal(v, &tagIDs); err != nil {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		var err error
		tagIDs, err = s.validSystemTagIDs(normalizeInt64s(tagIDs))
		if err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			res := tx.Model(&model.Bookshelf{}).Where("id = ? AND user_id = ? AND status = ?", id, u.ID, model.BookshelfStatusActive).Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return common.ErrNotFound
			}
		} else {
			var count int64
			if err := tx.Model(&model.Bookshelf{}).Where("id = ? AND user_id = ? AND status = ?", id, u.ID, model.BookshelfStatusActive).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return common.ErrNotFound
			}
		}
		if len(bookUpdates) > 0 {
			bookUpdates["updated_at"] = time.Now()
			res := tx.Model(&model.Book{}).Where("id = ? AND owner_user_id = ? AND visibility = ? AND deleted_at IS NULL", book.ID, u.ID, model.BookVisibilityPrivate).Updates(bookUpdates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return common.ErrNotFound
			}
		}
		if replaceTags {
			if err := tx.Where("bookshelf_id = ?", id).Delete(&model.BookshelfTag{}).Error; err != nil {
				return err
			}
			if len(tagIDs) > 0 {
				rows := make([]model.BookshelfTag, 0, len(tagIDs))
				for _, tagID := range tagIDs {
					rows = append(rows, model.BookshelfTag{BookshelfID: id, TagID: tagID})
				}
				if err := tx.Create(&rows).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	item, book, err = s.getBookshelfWithBook(u.ID, id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.bookshelfDTO(item, book, s.bookshelfProgressFor(u.ID, book.ID)))
}

func (s *Server) deleteBookshelf(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	item, book, err := s.getBookshelfWithBook(u.ID, id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	now := time.Now()
	if item.SourceType == model.BookshelfSourceUploaded && book.Visibility == model.BookVisibilityPrivate {
		if err := s.db.Model(&model.Book{}).Where("id = ? AND owner_user_id = ? AND visibility = ?", book.ID, u.ID, model.BookVisibilityPrivate).
			Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if err := s.removeBookFiles(book); err != nil && !os.IsNotExist(err) {
			_ = s.db.Model(&model.Book{}).Where("id = ? AND owner_user_id = ? AND visibility = ?", book.ID, u.ID, model.BookVisibilityPrivate).
				Updates(map[string]any{"deleted_at": nil, "updated_at": time.Now()}).Error
			common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileDeleteFailed)
			return
		}
		if err := s.hardDeleteBookData(book.ID, model.BookVisibilityPrivate); err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
		return
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Bookshelf{}).Where("id = ? AND user_id = ?", id, u.ID).Updates(map[string]any{"status": model.BookshelfStatusRemoved, "removed_at": now}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}

func (s *Server) hardDeleteBookData(bookID int64, visibility string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var shelfIDs []int64
		if err := tx.Model(&model.Bookshelf{}).Where("book_id = ?", bookID).Pluck("id", &shelfIDs).Error; err != nil {
			return err
		}
		if len(shelfIDs) > 0 {
			if err := tx.Where("bookshelf_id IN ?", shelfIDs).Delete(&model.BookshelfTag{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("book_id = ?", bookID).Delete(&model.Bookmark{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", bookID).Delete(&model.ReadingProgress{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", bookID).Delete(&model.Bookshelf{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", bookID).Delete(&model.BookCategory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", bookID).Delete(&model.BookTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", bookID).Delete(&model.BookChapter{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ? AND visibility = ?", bookID, visibility).Delete(&model.Book{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrBookNotFound
		}
		return nil
	})
}

func (s *Server) addLibraryBookToShelf(c *gin.Context) {
	bookID, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	s.addLibrary(c, bookID)
}

func (s *Server) addLibraryBookToShelfAlias(c *gin.Context) {
	bookID, ok := parseID(c, "id")
	if !ok {
		return
	}
	s.addLibrary(c, bookID)
}

func (s *Server) addLibrary(c *gin.Context, bookID int64) {
	u := middleware.CurrentUser(c)
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ? AND deleted_at IS NULL", bookID, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	if b.LibraryStatus == nil || *b.LibraryStatus != model.LibraryStatusApproved {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrLibraryNotApproved)
		return
	}
	var existing model.Bookshelf
	if err := s.db.First(&existing, "user_id = ? AND book_id = ?", u.ID, b.ID).Error; err == nil {
		if existing.Status == model.BookshelfStatusActive {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrBookAlreadyInShelf)
			return
		}
		now := time.Now()
		res := s.db.Model(&model.Bookshelf{}).Where("id = ? AND user_id = ?", existing.ID, u.ID).Updates(map[string]any{
			"source_type": model.BookshelfSourceLibrary,
			"status":      model.BookshelfStatusActive,
			"removed_at":  nil,
			"added_at":    now,
		})
		if res.Error != nil {
			common.RespondError(c, middleware.GetRequestID(c), res.Error)
			return
		}
		existing.SourceType = model.BookshelfSourceLibrary
		existing.Status = model.BookshelfStatusActive
		existing.RemovedAt = nil
		existing.AddedAt = now
		common.RespondJSON(c, middleware.GetRequestID(c), s.bookshelfDTO(existing, b, s.bookshelfProgressFor(u.ID, b.ID)))
		return
	}
	now := time.Now()
	item := model.Bookshelf{
		UserID: u.ID, BookID: b.ID, SourceType: model.BookshelfSourceLibrary, Status: model.BookshelfStatusActive, AddedAt: now,
		PersonalTitle: &b.Title, PersonalAuthor: b.Author, PersonalDescription: b.Description,
	}
	if err := s.db.Create(&item).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.bookshelfDTO(item, b, nil))
}

func (s *Server) libraryList(c *gin.Context)      { s.libraryListBase(c, false) }
func (s *Server) adminLibraryList(c *gin.Context) { s.libraryListBase(c, true) }

func (s *Server) libraryListBase(c *gin.Context, admin bool) {
	u := middleware.CurrentUser(c)
	page, size := common.ParsePagination(c.Query("page"), c.Query("page_size"))
	sortBy, orderBy, ok := common.NormalizeSortOrder(c.Query("sort"), c.Query("order"), map[string]struct{}{
		"created_at": {},
		"title":      {},
	}, "created_at")
	if !ok {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
		return
	}
	q := s.db.Model(&model.Book{}).Where("books.visibility = ?", model.BookVisibilityPublic)
	if !admin || c.Query("status") != model.LibraryStatusDeleted {
		q = q.Where("books.deleted_at IS NULL")
	}
	if !admin {
		if c.Query("mine") == "true" {
			q = q.Where("books.owner_user_id = ?", u.ID)
			if st := c.Query("status"); st != "" {
				if !validLibraryShelfStatus(st) {
					common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
					return
				}
				q = q.Where("books.library_status = ?", st)
			} else {
				q = q.Where("books.library_status IN ?", []string{model.LibraryStatusApproved, model.LibraryStatusHidden})
			}
		} else {
			if c.Query("mine") != "" && c.Query("mine") != "false" {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			if c.Query("status") != "" {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			q = q.Where("books.library_status = ?", model.LibraryStatusApproved)
		}
	} else if st := c.Query("status"); st != "" {
		if !validLibraryStatus(st) {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("books.library_status = ?", st)
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		q = q.Where("(books.title ILIKE ? OR books.author ILIKE ?)", "%"+kw+"%", "%"+kw+"%")
	}
	if f := c.Query("format"); f != "" {
		if !validBookFormat(f) {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
			return
		}
		q = q.Where("books.format = ?", f)
	}
	if categoryID, exists, err := queryInt64Param(c, "category_id"); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	} else if exists {
		q = q.Where("EXISTS (SELECT 1 FROM book_categories bc WHERE bc.book_id = books.id AND bc.category_id = ?)", categoryID)
	}
	if tagID, exists, err := queryInt64Param(c, "tag_id"); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	} else if exists {
		q = q.Where("EXISTS (SELECT 1 FROM book_tags bt WHERE bt.book_id = books.id AND bt.tag_id = ?)", tagID)
	}
	var total int64
	q.Count(&total)
	type libraryRow struct {
		model.Book
		OwnerUsername *string `gorm:"column:owner_username"`
	}
	var books []libraryRow
	err := q.Select("books.*, users.username AS owner_username").
		Joins("LEFT JOIN users ON users.id = books.owner_user_id").
		Order(libraryOrder(sortBy, orderBy)).Offset((page - 1) * size).Limit(size).Scan(&books).Error
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	out := make([]LibraryBookDTO, 0, len(books))
	for _, b := range books {
		ownerUsername := ""
		if b.OwnerUsername != nil {
			ownerUsername = *b.OwnerUsername
		}
		dto := s.libraryDTO(b.Book, ownerUsername, nil)
		var shelf model.Bookshelf
		if s.db.First(&shelf, "user_id = ? AND book_id = ? AND status = ?", u.ID, b.ID, model.BookshelfStatusActive).Error == nil {
			dto.InBookshelf = true
			dto.BookshelfID = &shelf.ID
		}
		out = append(out, dto)
	}
	common.RespondPage(c, middleware.GetRequestID(c), out, page, size, total)
}

func (s *Server) libraryDetail(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	var b model.Book
	query := s.db
	if u.Role != model.UserRoleAdmin {
		query = query.Where("deleted_at IS NULL")
	}
	if err := query.First(&b, "id = ? AND visibility = ?", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	if u.Role != model.UserRoleAdmin && (b.LibraryStatus == nil || (*b.LibraryStatus != model.LibraryStatusApproved && !(b.OwnerUserID == u.ID && *b.LibraryStatus == model.LibraryStatusHidden))) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrForbidden)
		return
	}
	ownerUsername := ""
	var owner model.User
	if err := s.db.First(&owner, "id = ?", b.OwnerUserID).Error; err == nil {
		ownerUsername = owner.Username
	}
	var bookshelfID *int64
	var shelf model.Bookshelf
	if s.db.First(&shelf, "user_id = ? AND book_id = ? AND status = ?", u.ID, b.ID, model.BookshelfStatusActive).Error == nil {
		bookshelfID = &shelf.ID
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(b, ownerUsername, bookshelfID))
}

func (s *Server) downloadLibraryBook(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ? AND deleted_at IS NULL", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	if b.LibraryStatus == nil || *b.LibraryStatus != model.LibraryStatusApproved {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotAccessible)
		return
	}
	s.downloadBookFile(c, b, b.Title)
}

func (s *Server) updateOwnLibraryBook(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	updates, err := bookBodyUpdates(raw)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if len(updates) == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	u := middleware.CurrentUser(c)
	updates["updated_at"] = time.Now()
	res := s.db.Model(&model.Book{}).Where("id = ? AND visibility = ? AND owner_user_id = ? AND deleted_at IS NULL", id, model.BookVisibilityPublic, u.ID).Updates(updates)
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := s.db.Model(&model.Book{}).Where("id = ? AND visibility = ? AND owner_user_id = ? AND deleted_at IS NULL", id, model.BookVisibilityPublic, u.ID).Count(&count).Error; err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if count == 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
			return
		}
	}
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ?", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(b, u.Username, s.activeBookshelfID(u.ID, b.ID)))
}

func (s *Server) hideOwnLibraryBook(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ? AND deleted_at IS NULL", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	if b.OwnerUserID != u.ID {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrForbidden)
		return
	}
	if b.LibraryStatus != nil && *b.LibraryStatus == model.LibraryStatusDeleted {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrConflict)
		return
	}
	if b.LibraryStatus != nil && *b.LibraryStatus == model.LibraryStatusHidden {
		common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(b, u.Username, s.activeBookshelfID(u.ID, b.ID)))
		return
	}
	hidden := model.LibraryStatusHidden
	now := time.Now()
	res := s.db.Model(&model.Book{}).Where("id = ? AND visibility = ? AND owner_user_id = ? AND deleted_at IS NULL", id, model.BookVisibilityPublic, u.ID).
		Updates(map[string]any{"library_status": hidden, "updated_at": now})
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	b.LibraryStatus = &hidden
	b.UpdatedAt = now
	common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(b, u.Username, s.activeBookshelfID(u.ID, b.ID)))
}

func (s *Server) showOwnLibraryBook(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	var b model.Book
	if err := s.db.First(&b, "id = ? AND visibility = ? AND deleted_at IS NULL", id, model.BookVisibilityPublic).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookNotFound)
		return
	}
	if b.OwnerUserID != u.ID {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrForbidden)
		return
	}
	if b.LibraryStatus == nil || *b.LibraryStatus != model.LibraryStatusHidden {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrConflict)
		return
	}
	approved := model.LibraryStatusApproved
	now := time.Now()
	res := s.db.Model(&model.Book{}).Where("id = ? AND visibility = ? AND owner_user_id = ? AND deleted_at IS NULL AND library_status = ?", id, model.BookVisibilityPublic, u.ID, model.LibraryStatusHidden).
		Updates(map[string]any{"library_status": approved, "updated_at": now})
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrConflict)
		return
	}
	b.LibraryStatus = &approved
	b.UpdatedAt = now
	common.RespondJSON(c, middleware.GetRequestID(c), s.libraryDTO(b, u.Username, s.activeBookshelfID(u.ID, b.ID)))
}

func (s *Server) activeBookshelfID(userID, bookID int64) *int64 {
	var shelf model.Bookshelf
	if s.db.First(&shelf, "user_id = ? AND book_id = ? AND status = ?", userID, bookID, model.BookshelfStatusActive).Error != nil {
		return nil
	}
	return &shelf.ID
}

func (s *Server) categories(c *gin.Context) {
	var rows []model.Category
	if err := s.db.Where("scope = ?", "system").Order("name asc").Find(&rows).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), rows)
}

func (s *Server) tags(c *gin.Context) {
	var rows []model.Tag
	if err := s.db.Where("scope = ?", "system").Order("name asc").Find(&rows).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), rows)
}

func parseCSVInt64s(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, common.ErrValidationFailed
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, common.ErrValidationFailed
		}
		ids = append(ids, id)
	}
	return normalizeInt64s(ids), nil
}

func normalizeInt64s(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (s *Server) validSystemCategoryIDs(ids []int64) ([]int64, error) {
	ids = normalizeInt64s(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var found []int64
	if err := s.db.Model(&model.Category{}).Where("scope = ? AND id IN ?", "system", ids).Pluck("id", &found).Error; err != nil {
		return nil, err
	}
	found = normalizeInt64s(found)
	if len(found) != len(ids) {
		return nil, common.ErrCategoryNotFound
	}
	return found, nil
}

func (s *Server) validSystemTagIDs(ids []int64) ([]int64, error) {
	ids = normalizeInt64s(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var found []int64
	if err := s.db.Model(&model.Tag{}).Where("scope = ? AND id IN ?", "system", ids).Pluck("id", &found).Error; err != nil {
		return nil, err
	}
	found = normalizeInt64s(found)
	if len(found) != len(ids) {
		return nil, common.ErrTagNotFound
	}
	return found, nil
}

func queryInt64Param(c *gin.Context, key string) (int64, bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, false, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, true, common.ErrValidationFailed
	}
	return id, true, nil
}

func isJSONNull(raw json.RawMessage) bool {
	return strings.EqualFold(strings.TrimSpace(string(raw)), "null")
}

func bookInfoRaw(raw map[string]json.RawMessage, key string) (json.RawMessage, bool) {
	if v, exists := raw[key]; exists {
		return v, true
	}
	if key == "title" {
		v, exists := raw["personal_title"]
		return v, exists
	}
	return nil, false
}

func requiredString(raw json.RawMessage) (string, error) {
	if isJSONNull(raw) {
		return "", common.ErrValidationFailed
	}
	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", common.ErrValidationFailed
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", common.ErrValidationFailed
	}
	return v, nil
}

func nullableTrimmedString(raw json.RawMessage) (*string, error) {
	if isJSONNull(raw) {
		return nil, nil
	}
	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, common.ErrValidationFailed
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	return &v, nil
}

func bookBodyUpdates(raw map[string]json.RawMessage) (map[string]any, error) {
	updates := map[string]any{}
	if v, exists := bookInfoRaw(raw, "title"); exists {
		title, err := requiredString(v)
		if err != nil {
			return nil, err
		}
		updates["title"] = title
	}
	if v, exists := raw["author"]; exists {
		author, err := nullableTrimmedString(v)
		if err != nil {
			return nil, err
		}
		updates["author"] = author
	}
	if v, exists := raw["description"]; exists {
		description, err := nullableTrimmedString(v)
		if err != nil {
			return nil, err
		}
		updates["description"] = description
	}
	return updates, nil
}

func (s *Server) getBookshelfWithBook(userID, id int64) (model.Bookshelf, model.Book, error) {
	var item model.Bookshelf
	if err := s.db.First(&item, "id = ? AND user_id = ? AND status = ?", id, userID, model.BookshelfStatusActive).Error; err != nil {
		return item, model.Book{}, err
	}
	var b model.Book
	if err := s.db.First(&b, "id = ?", item.BookID).Error; err != nil {
		return item, b, err
	}
	return item, b, nil
}

func (s *Server) bookshelfDTO(item model.Bookshelf, b model.Book, progress *float64) BookshelfItemDTO {
	title := b.Title
	author := b.Author
	description := b.Description
	if item.SourceType == model.BookshelfSourceLibrary && item.PersonalTitle != nil {
		title = *item.PersonalTitle
		author = item.PersonalAuthor
		description = item.PersonalDescription
	}
	readable := true
	var reason *string
	if !s.bookFileExists(b) {
		readable = false
		r := "file_missing"
		reason = &r
	} else if b.DeletedAt != nil || (b.Visibility == model.BookVisibilityPublic && b.LibraryStatus != nil && *b.LibraryStatus == model.LibraryStatusDeleted) {
		readable = false
		r := "library_deleted"
		reason = &r
	} else if b.Visibility == model.BookVisibilityPublic && b.LibraryStatus != nil && *b.LibraryStatus != model.LibraryStatusApproved {
		readable = false
		switch *b.LibraryStatus {
		case model.LibraryStatusHidden:
			r := "library_hidden"
			reason = &r
		default:
			r := "permission_denied"
			reason = &r
		}
	}
	return BookshelfItemDTO{ID: item.ID, BookID: b.ID, Title: title, Author: author, Description: description, Format: b.Format, CoverURL: coverURL(b.ID, b.CoverPath), SourceType: item.SourceType, Visibility: b.Visibility, LibraryStatus: b.LibraryStatus, Favorite: item.Favorite, Pinned: item.Pinned, LastReadAt: item.LastReadAt, AddedAt: item.AddedAt, Readable: readable, UnreadableReason: reason, ProgressPercentage: progress}
}

func (s *Server) bookshelfProgressFor(userID, bookID int64) *float64 {
	var progress model.ReadingProgress
	if err := s.db.First(&progress, "user_id = ? AND book_id = ?", userID, bookID).Error; err == nil {
		return progress.Percentage
	}
	return nil
}

func (s *Server) libraryDTO(b model.Book, ownerUsername string, bookshelfID *int64) LibraryBookDTO {
	status := ""
	if b.LibraryStatus != nil {
		status = *b.LibraryStatus
	}
	var owner *string
	if ownerUsername != "" {
		owner = &ownerUsername
	}
	return LibraryBookDTO{ID: b.ID, Title: b.Title, Author: b.Author, Description: b.Description, Format: b.Format, CoverURL: coverURL(b.ID, b.CoverPath), FileSize: b.FileSize, LibraryStatus: status, OwnerUserID: b.OwnerUserID, OwnerUsername: owner, BookshelfID: bookshelfID, InBookshelf: bookshelfID != nil, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt}
}

func orderDir(v string) string {
	if strings.ToLower(v) == "asc" {
		return "ASC"
	}
	return "DESC"
}

func bookshelfOrder(sortBy, order string) string {
	if sortBy == "" {
		return "bs.pinned DESC, bs.last_read_at DESC NULLS LAST, bs.added_at DESC"
	}
	dir := orderDir(order)
	switch sortBy {
	case "last_read_at":
		return "bs.last_read_at " + dir + " NULLS LAST, bs.added_at DESC"
	case "added_at":
		return "bs.added_at " + dir
	case "title":
		return "b.title " + dir + ", bs.added_at DESC"
	default:
		return "bs.pinned DESC, bs.last_read_at DESC NULLS LAST, bs.added_at DESC"
	}
}

func libraryOrder(sortBy, order string) string {
	dir := orderDir(order)
	switch sortBy {
	case "title":
		return "books.title " + dir + ", books.created_at DESC"
	case "created_at", "":
		return "books.created_at " + dir
	default:
		return "books.created_at DESC"
	}
}

func validBookFormat(format string) bool {
	return format == model.BookFormatEPUB || format == model.BookFormatPDF || format == model.BookFormatTXT
}

func (s *Server) removeBookFiles(book model.Book) error {
	bookPath, err := s.store.BookPath(book.FilePath)
	if err != nil {
		return err
	}
	if err := os.Remove(bookPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if book.CoverPath != nil {
		coverPath, err := s.store.CoverPath(*book.CoverPath)
		if err != nil {
			return err
		}
		if err := os.Remove(coverPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (s *Server) bookFileExists(book model.Book) bool {
	bookPath, err := s.store.BookPath(book.FilePath)
	if err != nil {
		return false
	}
	st, err := os.Stat(bookPath)
	return err == nil && !st.IsDir()
}

func (s *Server) downloadBookFile(c *gin.Context, book model.Book, displayTitle string) {
	path, err := s.store.BookPath(book.FilePath)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	filename := downloadFilename(displayTitle, book.Format)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", contentDispositionAttachment(filename))
	http.ServeContent(c.Writer, c.Request, filename, st.ModTime(), f)
}

func downloadFilename(title, format string) string {
	name := strings.TrimSpace(title)
	if name == "" {
		name = "book"
	}
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			continue
		}
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	name = strings.TrimSpace(b.String())
	if name == "" {
		name = "book"
	}
	ext := "." + strings.TrimPrefix(strings.ToLower(format), ".")
	if !strings.HasSuffix(strings.ToLower(name), ext) {
		name += ext
	}
	return name
}

func contentDispositionAttachment(filename string) string {
	escaped := strings.ReplaceAll(filename, `"`, `'`)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, escaped, url.PathEscape(filename))
}

func (s *Server) readAllowed(u *model.User, bookID int64) (model.Book, error) {
	var b model.Book
	query := s.db
	if u.Role != model.UserRoleAdmin {
		query = query.Where("deleted_at IS NULL")
	}
	if err := query.First(&b, "id = ?", bookID).Error; err != nil {
		return b, common.ErrBookNotFound
	}
	if u.Role == model.UserRoleAdmin {
		return b, nil
	}
	if b.Visibility == model.BookVisibilityPrivate {
		var count int64
		s.db.Model(&model.Bookshelf{}).Where("user_id = ? AND book_id = ? AND status = ?", u.ID, b.ID, model.BookshelfStatusActive).Count(&count)
		if count > 0 {
			return b, nil
		}
		return b, common.ErrBookNotAccessible
	}
	if b.LibraryStatus != nil && *b.LibraryStatus == model.LibraryStatusApproved {
		return b, nil
	}
	return b, common.ErrBookNotAccessible
}

func (s *Server) readerMeta(c *gin.Context) {
	id, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	b, err := s.readAllowed(middleware.CurrentUser(c), id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), bookMetaDTO(b))
}

func (s *Server) readerFile(c *gin.Context) {
	id, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	b, err := s.readAllowed(middleware.CurrentUser(c), id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	path, err := s.store.BookPath(b.FilePath)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	c.Header("Accept-Ranges", "bytes")
	http.ServeContent(c.Writer, c.Request, b.Title+"."+b.Format, st.ModTime(), f)
}

func (s *Server) readerCover(c *gin.Context) {
	id, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	b, err := s.readAllowed(middleware.CurrentUser(c), id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if b.CoverPath == nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	path, err := s.store.CoverPath(*b.CoverPath)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	if _, err := os.Stat(path); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	c.File(path)
}

func (s *Server) readerText(c *gin.Context) {
	id, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	b, err := s.readAllowed(middleware.CurrentUser(c), id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if b.Format != model.BookFormatTXT {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrUnsupportedMedia)
		return
	}
	path, err := s.store.BookPath(b.FilePath)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	offsetRaw := strings.TrimSpace(c.Query("offset"))
	offset := int64(0)
	if offsetRaw != "" {
		var err error
		offset, err = strconv.ParseInt(offsetRaw, 10, 64)
		if err != nil || offset < 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
			return
		}
	}
	limitRaw := strings.TrimSpace(c.Query("limit"))
	limit := int64(s.cfg.TXTChunkSize)
	if limitRaw != "" {
		var err error
		limit, err = strconv.ParseInt(limitRaw, 10, 64)
		if err != nil || limit <= 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
			return
		}
	}
	if limit <= 0 || limit > 1024*1024 {
		limit = 1024 * 1024
	}
	f, err := os.Open(path)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	if offset < 0 {
		offset = 0
	}
	if offset > st.Size() {
		offset = st.Size()
	}
	buf := make([]byte, min64(limit, st.Size()-offset))
	_, _ = f.ReadAt(buf, offset)
	next := offset + int64(len(buf))
	var nextPtr *int64
	if next < st.Size() {
		nextPtr = &next
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{"offset": offset, "limit": limit, "next_offset": nextPtr, "content": string(buf)})
}

func (s *Server) readerChapters(c *gin.Context) {
	id, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	_, err := s.readAllowed(middleware.CurrentUser(c), id)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	var rows []model.BookChapter
	if err := s.db.Where("book_id = ?", id).Order("chapter_index asc").Find(&rows).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), rows)
}

func (s *Server) readerChapterContent(c *gin.Context) {
	bookID, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	chapterID, ok := parseID(c, "chapterId")
	if !ok {
		return
	}
	b, err := s.readAllowed(middleware.CurrentUser(c), bookID)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	var ch model.BookChapter
	if err := s.db.First(&ch, "id = ? AND book_id = ?", chapterID, bookID).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	contentType, content, err := s.chapterContentFromBook(b, ch, middleware.CurrentUser(c).ID)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{"id": ch.ID, "chapter_index": ch.ChapterIndex, "title": ch.Title, "content_type": contentType, "content": content})
}

func (s *Server) readerResource(c *gin.Context) {
	bookID, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	href := parser.CleanResourceHref(c.Query("href"))
	if strings.TrimSpace(href) == "" {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrInvalidRequest)
		return
	}
	b, err := s.readerResourceBook(c, bookID, href)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	if b.Format != model.BookFormatEPUB {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrUnsupportedMedia)
		return
	}
	path, err := s.store.BookPath(b.FilePath)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrBookFileMissing)
		return
	}
	resource, err := parser.ReadEPUBResource(path, href)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	defer resource.Body.Close()
	c.Header("Content-Type", resource.ContentType)
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, resource.Name))
	_, _ = io.Copy(c.Writer, resource.Body)
}

func (s *Server) readerResourceBook(c *gin.Context, bookID int64, href string) (model.Book, error) {
	if u, err := s.userFromAuthorization(c); err != nil {
		return model.Book{}, common.ErrUnauthorized
	} else if u != nil {
		return s.readAllowed(u, bookID)
	}
	userID, ok := s.verifyEPUBResourceSignature(c, bookID, href)
	if !ok {
		return model.Book{}, common.ErrUnauthorized
	}
	u, err := s.FindUserByID(userID)
	if err != nil || u.Status != model.UserStatusActive {
		return model.Book{}, common.ErrUnauthorized
	}
	return s.readAllowed(u, bookID)
}

func (s *Server) userFromAuthorization(c *gin.Context) (*model.User, error) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, nil
	}
	claims, err := s.VerifyAccessToken(strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")))
	if err != nil {
		return nil, err
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, err
	}
	u, err := s.FindUserByID(userID)
	if err != nil || u.Status != model.UserStatusActive {
		return nil, common.ErrUnauthorized
	}
	return u, nil
}

func (s *Server) epubResourceURL(bookID, userID int64, href string) string {
	cleanHref := parser.CleanResourceHref(href)
	expires := time.Now().Add(epubResourceURLTTL).Unix()
	q := url.Values{}
	q.Set("href", cleanHref)
	q.Set("uid", strconv.FormatInt(userID, 10))
	q.Set("expires", strconv.FormatInt(expires, 10))
	q.Set("sig", s.signEPUBResource(bookID, userID, cleanHref, expires))
	return fmt.Sprintf("/api/v1/reader/books/%d/resources?%s", bookID, q.Encode())
}

func (s *Server) verifyEPUBResourceSignature(c *gin.Context, bookID int64, href string) (int64, bool) {
	userID, err := strconv.ParseInt(c.Query("uid"), 10, 64)
	if err != nil || userID <= 0 {
		return 0, false
	}
	expires, err := strconv.ParseInt(c.Query("expires"), 10, 64)
	if err != nil || expires < time.Now().Unix() {
		return 0, false
	}
	expected := s.signEPUBResource(bookID, userID, href, expires)
	if !hmac.Equal([]byte(expected), []byte(c.Query("sig"))) {
		return 0, false
	}
	return userID, true
}

func (s *Server) signEPUBResource(bookID, userID int64, href string, expires int64) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.JWTSecret))
	_, _ = fmt.Fprintf(mac, "%d\n%d\n%s\n%d", bookID, userID, href, expires)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Server) readerProgress(c *gin.Context) {
	bookID, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	if _, err := s.readAllowed(middleware.CurrentUser(c), bookID); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	var p model.ReadingProgress
	err := s.db.First(&p, "user_id = ? AND book_id = ?", middleware.CurrentUser(c).ID, bookID).Error
	if err != nil {
		common.RespondJSON(c, middleware.GetRequestID(c), nil)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), p)
}

func (s *Server) saveReaderProgress(c *gin.Context) {
	bookID, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	b, err := s.readAllowed(u, bookID)
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	var req struct {
		ProgressType  string   `json:"progress_type"`
		ProgressValue string   `json:"progress_value"`
		Percentage    *float64 `json:"percentage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProgressValue == "" || !validProgressType(b.Format, req.ProgressType) || (req.Percentage != nil && (*req.Percentage < 0 || *req.Percentage > 100)) {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
		return
	}
	var shelf model.Bookshelf
	var shelfID *int64
	if s.db.First(&shelf, "user_id = ? AND book_id = ? AND status = ?", u.ID, b.ID, model.BookshelfStatusActive).Error == nil {
		shelfID = &shelf.ID
	}
	p := model.ReadingProgress{UserID: u.ID, BookID: b.ID, BookshelfID: shelfID, Format: b.Format, ProgressType: req.ProgressType, ProgressValue: req.ProgressValue, Percentage: req.Percentage, UpdatedAt: time.Now()}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "book_id"}}, DoUpdates: clause.Assignments(map[string]any{"bookshelf_id": shelfID, "progress_type": req.ProgressType, "progress_value": req.ProgressValue, "percentage": req.Percentage, "updated_at": time.Now()})}).Create(&p).Error; err != nil {
			return err
		}
		return tx.Model(&model.Bookshelf{}).Where("user_id = ? AND book_id = ? AND status = ?", u.ID, b.ID, model.BookshelfStatusActive).Update("last_read_at", time.Now()).Error
	})
	if err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	s.db.First(&p, "user_id = ? AND book_id = ?", u.ID, b.ID)
	common.RespondJSON(c, middleware.GetRequestID(c), p)
}

func validProgressType(format, pt string) bool {
	return (format == model.BookFormatEPUB && pt == model.ProgressTypeEpubCFI) || (format == model.BookFormatPDF && pt == model.ProgressTypePDFPage) || (format == model.BookFormatTXT && pt == model.ProgressTypeTXTOffset)
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
