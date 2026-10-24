package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"app/internal/application/dto"
	"app/internal/application/usecase"
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
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	id, err := h.service.Create(ctx.Request.Context(), dto.CreateTodoInput{
		Title:   request.Title,
		Content: request.Content,
	})
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "ToDo Criado", "id": id})
}

func (h *TodoHandler) Get(ctx *gin.Context) {
	paramID := ctx.Param("id")
	if paramID == "" {
		ctx.String(http.StatusBadRequest, "Id não fornecido")
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	todo, err := h.service.Get(ctx.Request.Context(), int64(id))
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, toTodoResponse(*todo))
}

func (h *TodoHandler) List(ctx *gin.Context) {
	todos, err := h.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, toTodoResponseList(todos))
}

func (h *TodoHandler) Update(ctx *gin.Context) {
	paramID := ctx.Param("id")
	if paramID == "" {
		ctx.String(http.StatusBadRequest, "Id não fornecido")
		return
	}

	request, err := bindJSON[updateTodoRequest](ctx)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = h.service.Update(ctx.Request.Context(), int64(id), dto.UpdateTodoInput{
		Title:   request.Title,
		Content: request.Content,
		Status:  request.Status,
	})
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.String(http.StatusOK, "ToDo atualizado")
}

func (h *TodoHandler) Delete(ctx *gin.Context) {
	paramID := ctx.Param("id")
	if paramID == "" {
		ctx.String(http.StatusBadRequest, "Id não fornecido")
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.service.Delete(ctx.Request.Context(), int64(id)); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.String(http.StatusOK, "ToDo deletado")
}

func bindJSON[T any](ctx *gin.Context) (*T, error) {
	var request T
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return &request, err
	}
	return &request, nil
}
