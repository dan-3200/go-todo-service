package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type PageInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type Response[T any] struct {
	Success    bool      `json:"success"`
	Status     int       `json:"status"`
	Data       *T        `json:"data,omitempty"`
	Pagination *PageInfo `json:"pagination,omitempty"`
	Error      *APIError `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

func Success[T any](ctx *gin.Context, status int, data T) {
	ctx.JSON(status, Response[T]{
		Success:   true,
		Status:    status,
		Data:      &data,
		Timestamp: time.Now().UTC(),
	})
}

func Error(ctx *gin.Context, status int, code, message string, details any) {
	ctx.JSON(status, Response[any]{
		Success: false,
		Status:  status,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	})
}

func InternalServerError(ctx *gin.Context) {
	Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "erro interno do servidor", nil)
}
