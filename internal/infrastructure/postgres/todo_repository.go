package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"app/internal/application/ports"
	"app/internal/domain/entity"
	"app/internal/domain/vo"
	postgresql "app/internal/infrastructure/postgres/sqlc"
)

type TodoRepository struct {
	queries *postgresql.Queries
}

var _ ports.TodoRepository = (*TodoRepository)(nil)

func NewTodoRepository(connection *sql.DB) (*TodoRepository, error) {
	if err := connection.Ping(); err != nil {
		return &TodoRepository{}, err
	}

	return &TodoRepository{queries: postgresql.New(connection)}, nil
}

func (r *TodoRepository) Create(ctx context.Context, todo *entity.Todo) (vo.TodoID, error) {
	id, err := r.queries.CreateTodo(ctx, postgresql.CreateTodoParams{
		Title:   sql.NullString{String: todo.Title, Valid: true},
		Content: sql.NullString{String: todo.Content, Valid: true},
		Status:  sql.NullBool{Bool: todo.Status, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	return vo.NewTodoID(id)
}

func (r *TodoRepository) FindByID(ctx context.Context, id vo.TodoID) (*entity.Todo, error) {
	todo, err := r.queries.GetTodo(ctx, id.Int64())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("Nenhum registro encontro.")
	}
	if err != nil {
		return nil, err
	}

	result, err := toEntity(todo)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *TodoRepository) FindAll(ctx context.Context) ([]entity.Todo, error) {
	todos, err := r.queries.ListTodos(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]entity.Todo, 0, len(todos))
	for _, todo := range todos {
		item, err := toEntity(todo)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *entity.Todo) error {
	return r.queries.UpdateTodo(ctx, postgresql.UpdateTodoParams{
		Title:   sql.NullString{String: todo.Title, Valid: true},
		Content: sql.NullString{String: todo.Content, Valid: true},
		Status:  sql.NullBool{Bool: todo.Status, Valid: true},
		ID:      todo.ID.Int64(),
	})
}

func (r *TodoRepository) Delete(ctx context.Context, id vo.TodoID) error {
	rowsAffected, err := r.queries.DeleteTodo(ctx, id.Int64())
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Nenhum registro afetado")
	}

	return nil
}

func toEntity(todo postgresql.Todo) (entity.Todo, error) {
	id, err := vo.NewTodoID(todo.ID)
	if err != nil {
		return entity.Todo{}, err
	}
	if !todo.Title.Valid || !todo.Content.Valid || !todo.Status.Valid || !todo.CreatedAt.Valid {
		return entity.Todo{}, fmt.Errorf("registro todo possui campo nulo")
	}

	return entity.Todo{
		ID:        id,
		Title:     todo.Title.String,
		Content:   todo.Content.String,
		Status:    todo.Status.Bool,
		CreatedAt: todo.CreatedAt.Time,
	}, nil
}
