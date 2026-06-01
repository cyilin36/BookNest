package common

import "github.com/gin-gonic/gin"

type ErrorPayload struct {
	Error     *APIError `json:"error"`
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
