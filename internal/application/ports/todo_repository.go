package ports

import (
	"context"

	"app/internal/domain/entity"
	"app/internal/domain/vo"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *entity.Todo) (vo.TodoID, error)
	FindByID(ctx context.Context, id vo.TodoID) (*entity.Todo, error)
	FindAll(ctx context.Context) ([]entity.Todo, error)
	Update(ctx context.Context, todo *entity.Todo) error
	Delete(ctx context.Context, id vo.TodoID) error
}
