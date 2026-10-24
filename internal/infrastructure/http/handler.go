package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"app/internal/application/dto"
	"app/internal/application/usecase"
	"app/internal/infrastructure/http/response"
)

type TodoHandler struct {
	service *usecase.TodoService
}

func NewTodoHandler(service *usecase.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func (h *TodoHandler) Create(ctx *gin.Context) {
	request, err := bindJSON[createTodoRequest](ctx)
	if err != nil {
		handleBindingError(ctx, err)
		return
	}

	id, err := h.service.Create(ctx.Request.Context(), dto.CreateTodoInput{
		Title:   request.Title,
		Content: request.Content,
	})
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, createTodoResponse{Message: "ToDo Criado", ID: id})
}

func (h *TodoHandler) Get(ctx *gin.Context) {
	paramID := ctx.Param("id")
	if paramID == "" {
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", "id não fornecido", nil)
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", "id inválido", err.Error())
		return
	}

	todo, err := h.service.Get(ctx.Request.Context(), int64(id))
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, toTodoResponse(*todo))
}

func (h *TodoHandler) List(ctx *gin.Context) {
	todos, err := h.service.List(ctx.Request.Context())
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, toTodoResponseList(todos))
}

func (h *TodoHandler) Update(ctx *gin.Context) {
	paramID := ctx.Param("id")
	if paramID == "" {
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", "id não fornecido", nil)
		return
	}

	request, err := bindJSON[updateTodoRequest](ctx)
	if err != nil {
		handleBindingError(ctx, err)
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", "id inválido", err.Error())
		return
	}

	err = h.service.Update(ctx.Request.Context(), int64(id), dto.UpdateTodoInput{
		Title:   request.Title,
		Content: request.Content,
		Status:  request.Status,
	})
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, messageResponse{Message: "ToDo atualizado"})
}

func (h *TodoHandler) Delete(ctx *gin.Context) {
	paramID := ctx.Param("id")
	if paramID == "" {
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", "id não fornecido", nil)
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "INVALID_ID", "id inválido", err.Error())
		return
	}

	if err := h.service.Delete(ctx.Request.Context(), int64(id)); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, messageResponse{Message: "ToDo deletado"})
}

func bindJSON[T any](ctx *gin.Context) (*T, error) {
	var request T
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return &request, err
	}
	return &request, nil
}
