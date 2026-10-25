package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/domain/apperror"
	"app/internal/infrastructure/http/response"
)

func handleServiceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidID):
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", apperror.ErrInvalidID.Error(), nil)
	case errors.Is(err, apperror.ErrTitleRequired):
		response.Error(ctx, http.StatusBadRequest, "VALIDATION_ERROR", apperror.ErrTitleRequired.Error(), nil)
	case errors.Is(err, apperror.ErrContentRequired):
		response.Error(ctx, http.StatusBadRequest, "VALIDATION_ERROR", apperror.ErrContentRequired.Error(), nil)
	case errors.Is(err, apperror.ErrNoFieldsProvided):
		response.Error(ctx, http.StatusBadRequest, "VALIDATION_ERROR", apperror.ErrNoFieldsProvided.Error(), nil)
	case errors.Is(err, apperror.ErrTodoNotFound):
		response.Error(ctx, http.StatusNotFound, "TODO_NOT_FOUND", "todo não encontrado", nil)
	default:
		response.InternalServerError(ctx)
	}
}

func handleBindingError(ctx *gin.Context, err error) {
	response.Error(ctx, http.StatusBadRequest, "INVALID_REQUEST", "corpo da requisição inválido", err.Error())
}

func handlePanic(ctx *gin.Context, recovered any) {
	response.InternalServerError(ctx)
}
