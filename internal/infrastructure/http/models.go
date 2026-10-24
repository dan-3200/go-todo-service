package http

import (
	"time"

	"app/internal/application/dto"
)

type createTodoRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Status  bool   `json:"status"`
}

type updateTodoRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Status  *bool   `json:"status"`
}

type todoResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type createTodoResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func toTodoResponse(todo dto.TodoOutput) todoResponse {
	return todoResponse{
		ID:        todo.ID,
		Title:     todo.Title,
		Content:   todo.Content,
		Status:    todo.Status,
		CreatedAt: todo.CreatedAt,
	}
}

func toTodoResponseList(todos []dto.TodoOutput) []todoResponse {
	response := make([]todoResponse, 0, len(todos))
	for _, todo := range todos {
		response = append(response, toTodoResponse(todo))
	}
	return response
}
