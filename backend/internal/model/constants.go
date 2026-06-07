package model

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"

	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"

	BookVisibilityPrivate = "private"
	BookVisibilityPublic  = "public"

	BookFormatEPUB = "epub"
	BookFormatPDF  = "pdf"
	BookFormatTXT  = "txt"

	LibraryStatusApproved = "approved"
	LibraryStatusHidden   = "hidden"
	LibraryStatusDeleted  = "deleted"

	BookshelfSourceUploaded    = "uploaded"
	BookshelfSourceLibrary     = "library"
	BookshelfStatusActive      = "active"
	BookshelfStatusRemoved     = "removed"
	BookshelfStatusUnavailable = "unavailable"

	ProgressTypeEpubCFI   = "epub_cfi"
	ProgressTypePDFPage   = "pdf_page"
	ProgressTypeTXTOffset = "txt_offset"
)
