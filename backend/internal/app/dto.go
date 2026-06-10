package app

import "time"

type AuthSession struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        any    `json:"user"`
}

type BookMetaDTO struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Author        *string   `json:"author"`
	Description   *string   `json:"description"`
	Format        string    `json:"format"`
	CoverURL      *string   `json:"cover_url"`
	FileSize      int64     `json:"file_size"`
	Visibility    string    `json:"visibility"`
	LibraryStatus *string   `json:"library_status"`
	ParseStatus   string    `json:"parse_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BookshelfItemDTO struct {
	ID                 int64      `json:"id"`
	BookID             int64      `json:"book_id"`
	Title              string     `json:"title"`
	Author             *string    `json:"author"`
	Description        *string    `json:"description"`
	Format             string     `json:"format"`
	CoverURL           *string    `json:"cover_url"`
	SourceType         string     `json:"source_type"`
	Visibility         string     `json:"visibility"`
	LibraryStatus      *string    `json:"library_status"`
	Favorite           bool       `json:"favorite"`
	Pinned             bool       `json:"pinned"`
	LastReadAt         *time.Time `json:"last_read_at"`
	AddedAt            time.Time  `json:"added_at"`
	Readable           bool       `json:"readable"`
	UnreadableReason   *string    `json:"unreadable_reason"`
	ProgressPercentage *float64   `json:"progress_percentage"`
}

type LibraryBookDTO struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Author        *string   `json:"author"`
	Description   *string   `json:"description"`
	Format        string    `json:"format"`
	CoverURL      *string   `json:"cover_url"`
	FileSize      int64     `json:"file_size"`
	LibraryStatus string    `json:"library_status"`
	OwnerUserID   int64     `json:"owner_user_id"`
	OwnerUsername *string   `json:"owner_username"`
	CategoryIDs   []int64   `json:"category_ids"`
	TagIDs        []int64   `json:"tag_ids"`
	InBookshelf   bool      `json:"in_bookshelf"`
	BookshelfID   *int64    `json:"bookshelf_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SystemSettings struct {
	SiteName                  string  `json:"site_name"`
	SiteIconURL               *string `json:"site_icon_url"`
	AllowRegistration         bool    `json:"allow_registration"`
	LibraryReviewRequired     bool    `json:"library_review_required"`
	MaxUploadSizeMB           int     `json:"max_upload_size_mb"`
	DefaultUserStorageQuotaMB int     `json:"default_user_storage_quota_mb"`
}
