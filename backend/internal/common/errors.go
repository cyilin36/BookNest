package common

import "net/http"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Code + ": " + e.Message
}

func NewError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

var (
	ErrInvalidRequest       = NewError(http.StatusBadRequest, "invalid_request", "请求参数不合法")
	ErrUnauthorized         = NewError(http.StatusUnauthorized, "unauthorized", "未登录或凭证无效")
	ErrForbidden            = NewError(http.StatusForbidden, "forbidden", "无权限访问")
	ErrNotFound             = NewError(http.StatusNotFound, "not_found", "资源不存在")
	ErrConflict             = NewError(http.StatusConflict, "conflict", "资源冲突")
	ErrValidationFailed     = NewError(http.StatusBadRequest, "validation_failed", "校验失败")
	ErrPayloadTooLarge      = NewError(http.StatusRequestEntityTooLarge, "payload_too_large", "请求体过大")
	ErrUnsupportedMedia     = NewError(http.StatusUnsupportedMediaType, "unsupported_media_type", "不支持的媒体类型")
	ErrInvalidImageFormat   = NewError(http.StatusBadRequest, "invalid_image_format", "不支持的图片格式")
	ErrInternal             = NewError(http.StatusInternalServerError, "internal_error", "服务内部错误")
	ErrRegistrationDisable  = NewError(http.StatusForbidden, "registration_disabled", "当前关闭注册")
	ErrUsernameExists       = NewError(http.StatusConflict, "username_exists", "用户名已存在")
	ErrEmailExists          = NewError(http.StatusConflict, "email_exists", "邮箱已存在")
	ErrInvalidCredentials   = NewError(http.StatusUnauthorized, "invalid_credentials", "用户名或密码错误")
	ErrUserDisabled         = NewError(http.StatusForbidden, "user_disabled", "用户已被禁用")
	ErrRefreshInvalid       = NewError(http.StatusUnauthorized, "refresh_token_invalid", "刷新令牌无效")
	ErrBookNotFound         = NewError(http.StatusNotFound, "book_not_found", "图书不存在")
	ErrBookNotAccessible    = NewError(http.StatusForbidden, "book_not_accessible", "图书不可访问")
	ErrBookAlreadyInShelf   = NewError(http.StatusConflict, "book_already_in_bookshelf", "图书已在书架中")
	ErrBookFormatNotSup     = NewError(http.StatusBadRequest, "book_format_not_supported", "不支持的图书格式")
	ErrBookFileMissing      = NewError(http.StatusNotFound, "book_file_missing", "图书文件不存在")
	ErrBookFileDeleteFailed = NewError(http.StatusInternalServerError, "book_file_delete_failed", "图书文件删除失败，已取消删除标记，请稍后重试")
	ErrLibraryNotApproved   = NewError(http.StatusForbidden, "library_book_not_approved", "公共图书未上架")
	ErrCategoryNotFound     = NewError(http.StatusNotFound, "category_not_found", "分类不存在")
	ErrTagNotFound          = NewError(http.StatusNotFound, "tag_not_found", "标签不存在")
	ErrStorageQuotaExceeded = NewError(http.StatusForbidden, "storage_quota_exceeded", "存储配额不足")
)
