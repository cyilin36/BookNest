package common

import "github.com/gin-gonic/gin"

type ErrorPayload struct {
	Error     *APIError `json:"error"`
	Details   any       `json:"details,omitempty"`
	RequestID string    `json:"request_id"`
}

type SuccessPayload struct {
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type PaginatedPayload struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
	RequestID  string     `json:"request_id"`
}

func RespondError(c *gin.Context, requestID string, err error) {
	apiErr, ok := err.(*APIError)
	if !ok {
		apiErr = ErrInternal
	}
	c.JSON(apiErr.Status, ErrorPayload{Error: apiErr, RequestID: requestID})
}

// RespondErrorWithDetails 在错误响应中附带额外的结构化信息（如冲突的具体资源）。
func RespondErrorWithDetails(c *gin.Context, requestID string, err error, details any) {
	apiErr, ok := err.(*APIError)
	if !ok {
		apiErr = ErrInternal
	}
	c.JSON(apiErr.Status, ErrorPayload{Error: apiErr, Details: details, RequestID: requestID})
}

func RespondJSON(c *gin.Context, requestID string, data any) {
	c.JSON(200, SuccessPayload{Data: data, RequestID: requestID})
}

func RespondPage(c *gin.Context, requestID string, data any, page, pageSize int, total int64) {
	c.JSON(200, PaginatedPayload{
		Data:       data,
		Pagination: Pagination{Page: page, PageSize: pageSize, Total: total},
		RequestID:  requestID,
	})
}
