package entity

import (
	"strings"
	"time"

	"app/internal/domain/apperror"
	"app/internal/domain/vo"
)

type Todo struct {
	ID        vo.TodoID
	Title     string
	Content   string
	Status    bool
	CreatedAt time.Time
}

func NewTodo(title, content string) (*Todo, error) {
	todo := &Todo{
		Title:     strings.ToLower(title),
		Content:   strings.ToLower(content),
		Status:    false,
		CreatedAt: time.Now(),
	}

	return todo, todo.Validate()
}

func (t *Todo) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return apperror.ErrTitleRequired
	}
	if strings.TrimSpace(t.Content) == "" {
		return apperror.ErrContentRequired
	}
	return nil
}
