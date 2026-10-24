package postgres

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"app/internal/application/ports"
	"app/internal/domain/entity"
	"app/internal/domain/vo"
)

var statementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type TodoRepository struct {
	db *sql.DB
}

var _ ports.TodoRepository = (*TodoRepository)(nil)

func NewTodoRepository(connection *sql.DB) (*TodoRepository, error) {
	if err := connection.Ping(); err != nil {
		return &TodoRepository{}, err
	}

	return &TodoRepository{db: connection}, nil
}

func (r *TodoRepository) Create(ctx context.Context, todo *entity.Todo) (vo.TodoID, error) {
	query, args, err := statementBuilder.
		Insert("todo").
		Columns("title", "content", "status").
		Values(todo.Title, todo.Content, todo.Status).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, err
	}

	var id int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		return 0, err
	}

	return vo.TodoID(id), nil
}

func (r *TodoRepository) FindByID(ctx context.Context, id vo.TodoID) (*entity.Todo, error) {
	query, args, err := statementBuilder.
		Select("id", "title", "content", "status", "created_at").
		From("todo").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var todo entity.Todo
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Content,
		&todo.Status,
		&todo.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("Nenhum registro encontro.")
	}
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (r *TodoRepository) FindAll(ctx context.Context) ([]entity.Todo, error) {
	query, args, err := statementBuilder.
		Select("id", "title", "content", "status", "created_at").
		From("todo").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]entity.Todo, 0)
	for rows.Next() {
		var todo entity.Todo
		if err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Content,
			&todo.Status,
			&todo.CreatedAt,
		); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *entity.Todo) error {
	query, args, err := statementBuilder.
		Update("todo").
		Set("title", todo.Title).
		Set("content", todo.Content).
		Set("status", todo.Status).
		Where(sq.Eq{"id": todo.ID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *TodoRepository) Delete(ctx context.Context, id vo.TodoID) error {
	query, args, err := statementBuilder.
		Delete("todo").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Nenhum registro afetado")
	}

	return nil
}

func (r *TodoRepository) CreateSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS todo (
			id SERIAL PRIMARY KEY,
			title TEXT,
			content TEXT,
			status BOOLEAN,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}
