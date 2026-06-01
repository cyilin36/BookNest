package app

import (
	"encoding/json"
	"strings"
	"time"

	"book-reader/backend/internal/common"
	"book-reader/backend/internal/middleware"
	"book-reader/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (s *Server) bookmarks(c *gin.Context) {
	bookID, ok := parseID(c, "bookId")
	if !ok {
		return
	}
	u := middleware.CurrentUser(c)
	if _, err := s.readAllowed(u, bookID); err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	var rows []model.Bookmark
	if err := s.db.Where("user_id = ? AND book_id = ?", u.ID, bookID).Order("created_at desc").Find(&rows).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), rows)
}

func (s *Server) createBookmark(c *gin.Context) {
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
		Title         *string  `json:"title"`
		Note          *string  `json:"note"`
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
	if s.db.First(&shelf, "user_id = ? AND book_id = ? AND status = ?", u.ID, bookID, model.BookshelfStatusActive).Error == nil {
		shelfID = &shelf.ID
	}
	row := model.Bookmark{UserID: u.ID, BookID: bookID, BookshelfID: shelfID, Title: req.Title, Note: req.Note, PositionType: req.ProgressType, PositionValue: req.ProgressValue, Percentage: req.Percentage}
	if err := s.db.Create(&row).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), err)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), row)
}

func (s *Server) updateBookmark(c *gin.Context) {
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
	if v, exists := raw["title"]; exists {
		if isJSONNull(v) {
			updates["title"] = nil
		} else {
			var title string
			if err := json.Unmarshal(v, &title); err != nil {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			updates["title"] = strings.TrimSpace(title)
		}
	}
	if v, exists := raw["note"]; exists {
		if isJSONNull(v) {
			updates["note"] = nil
		} else {
			var note string
			if err := json.Unmarshal(v, &note); err != nil {
				common.RespondError(c, middleware.GetRequestID(c), common.ErrValidationFailed)
				return
			}
			updates["note"] = note
		}
	}
	u := middleware.CurrentUser(c)
	res := s.db.Model(&model.Bookmark{}).Where("id = ? AND user_id = ?", id, u.ID).Updates(updates)
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := s.db.Model(&model.Bookmark{}).Where("id = ? AND user_id = ?", id, u.ID).Count(&count).Error; err != nil {
			common.RespondError(c, middleware.GetRequestID(c), err)
			return
		}
		if count == 0 {
			common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
			return
		}
	}
	var row model.Bookmark
	if err := s.db.First(&row, "id = ? AND user_id = ?", id, u.ID).Error; err != nil {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), row)
}

func (s *Server) deleteBookmark(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	res := s.db.Where("id = ? AND user_id = ?", id, middleware.CurrentUser(c).ID).Delete(&model.Bookmark{})
	if res.Error != nil {
		common.RespondError(c, middleware.GetRequestID(c), res.Error)
		return
	}
	if res.RowsAffected == 0 {
		common.RespondError(c, middleware.GetRequestID(c), common.ErrNotFound)
		return
	}
	common.RespondJSON(c, middleware.GetRequestID(c), gin.H{})
}
