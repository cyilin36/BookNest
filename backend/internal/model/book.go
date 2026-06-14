package model

import "time"

type Book struct {
	ID               int64      `gorm:"column:id;primaryKey" json:"id"`
	Title            string     `gorm:"column:title" json:"title"`
	Author           *string    `gorm:"column:author" json:"author"`
	Description      *string    `gorm:"column:description" json:"description"`
	OriginalFilename *string    `gorm:"column:original_filename" json:"-"`
	Format           string     `gorm:"column:format" json:"format"`
	FileSize         int64      `gorm:"column:file_size" json:"file_size"`
	FileHash         string     `gorm:"column:file_hash" json:"-"`
	FilePath         string     `gorm:"column:file_path" json:"-"`
	CoverPath        *string    `gorm:"column:cover_path" json:"-"`
	Charset          *string    `gorm:"column:charset" json:"-"`
	Visibility       string     `gorm:"column:visibility" json:"visibility"`
	OwnerUserID      int64      `gorm:"column:owner_user_id" json:"-"`
	LibraryStatus    *string    `gorm:"column:library_status" json:"library_status"`
	ParseStatus      string     `gorm:"column:parse_status" json:"parse_status"`
	ParseError       *string    `gorm:"column:parse_error" json:"-"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Book) TableName() string { return "books" }

type Bookshelf struct {
	ID                  int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID              int64      `gorm:"column:user_id" json:"-"`
	BookID              int64      `gorm:"column:book_id" json:"book_id"`
	SourceType          string     `gorm:"column:source_type" json:"source_type"`
	Status              string     `gorm:"column:status" json:"-"`
	Favorite            bool       `gorm:"column:favorite" json:"favorite"`
	Pinned              bool       `gorm:"column:pinned" json:"pinned"`
	PersonalTitle       *string    `gorm:"column:personal_title" json:"-"`
	PersonalAuthor      *string    `gorm:"column:personal_author" json:"-"`
	PersonalDescription *string    `gorm:"column:personal_description" json:"-"`
	PersonalCoverPath   *string    `gorm:"column:personal_cover_path" json:"-"`
	PersonalCategoryID  *int64     `gorm:"column:personal_category_id" json:"-"`
	AddedAt             time.Time  `gorm:"column:added_at" json:"added_at"`
	LastReadAt          *time.Time `gorm:"column:last_read_at" json:"last_read_at"`
	RemovedAt           *time.Time `gorm:"column:removed_at" json:"-"`
}

func (Bookshelf) TableName() string { return "bookshelves" }

type Category struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Slug        string    `gorm:"column:slug" json:"-"`
	Description *string   `gorm:"column:description" json:"description"`
	Scope       string    `gorm:"column:scope" json:"scope"`
	OwnerUserID *int64    `gorm:"column:owner_user_id" json:"-"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"-"`
}

func (Category) TableName() string { return "categories" }

type Tag struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Slug        string    `gorm:"column:slug" json:"-"`
	Description *string   `gorm:"column:description" json:"description"`
	Scope       string    `gorm:"column:scope" json:"scope"`
	OwnerUserID *int64    `gorm:"column:owner_user_id" json:"-"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"-"`
}

func (Tag) TableName() string { return "tags" }

type BookCategory struct {
	BookID     int64 `gorm:"column:book_id"`
	CategoryID int64 `gorm:"column:category_id"`
}

func (BookCategory) TableName() string { return "book_categories" }

type BookTag struct {
	BookID int64 `gorm:"column:book_id"`
	TagID  int64 `gorm:"column:tag_id"`
}

func (BookTag) TableName() string { return "book_tags" }

type BookshelfTag struct {
	BookshelfID int64 `gorm:"column:bookshelf_id"`
	TagID       int64 `gorm:"column:tag_id"`
}

func (BookshelfTag) TableName() string { return "bookshelf_tags" }

type BookChapter struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	BookID       int64     `gorm:"column:book_id" json:"-"`
	ChapterIndex int       `gorm:"column:chapter_index" json:"chapter_index"`
	Title        string    `gorm:"column:title" json:"title"`
	Locator      string    `gorm:"column:locator" json:"-"`
	StartOffset  *int64    `gorm:"column:start_offset" json:"-"`
	EndOffset    *int64    `gorm:"column:end_offset" json:"-"`
	Href         *string   `gorm:"column:href" json:"-"`
	IsVolume     bool      `gorm:"column:is_volume" json:"is_volume"`
	WordCount    *int64    `gorm:"column:word_count" json:"word_count"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"-"`
}

func (BookChapter) TableName() string { return "book_chapters" }

type ReadingProgress struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"-"`
	UserID        int64     `gorm:"column:user_id" json:"-"`
	BookID        int64     `gorm:"column:book_id" json:"-"`
	BookshelfID   *int64    `gorm:"column:bookshelf_id" json:"-"`
	Format        string    `gorm:"column:format" json:"-"`
	ProgressType  string    `gorm:"column:progress_type" json:"progress_type"`
	ProgressValue string    `gorm:"column:progress_value" json:"progress_value"`
	Percentage    *float64  `gorm:"column:percentage" json:"percentage"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ReadingProgress) TableName() string { return "reading_progress" }

type Bookmark struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64     `gorm:"column:user_id" json:"-"`
	BookID        int64     `gorm:"column:book_id" json:"book_id"`
	BookshelfID   *int64    `gorm:"column:bookshelf_id" json:"bookshelf_id"`
	Title         *string   `gorm:"column:title" json:"title"`
	PositionType  string    `gorm:"column:position_type" json:"progress_type"`
	PositionValue string    `gorm:"column:position_value" json:"progress_value"`
	Percentage    *float64  `gorm:"column:percentage" json:"percentage"`
	Note          *string   `gorm:"column:note" json:"note"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Bookmark) TableName() string { return "bookmarks" }

type SystemSetting struct {
	Key       string    `gorm:"column:key;primaryKey"`
	Value     string    `gorm:"column:value"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SystemSetting) TableName() string { return "system_settings" }
