package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/infrastructure/http/response"
)

func handleServiceError(ctx *gin.Context, err error) {
	message := err.Error()

	switch message {
	case "id inválido":
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", message, nil)
	case "título é obrigatório", "conteúdo é obrigatório", "ao menos um campo deve ser informado":
		response.Error(ctx, http.StatusBadRequest, "VALIDATION_ERROR", message, nil)
	case "Nenhum registro encontro.", "não há elemento com tal id", "Nenhum registro afetado":
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
