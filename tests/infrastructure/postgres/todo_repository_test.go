package postgres_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"app/internal/domain/entity"
	"app/internal/domain/vo"
	"app/internal/infrastructure/postgres"
)

func newMockRepository(t *testing.T) (*postgres.TodoRepository, sqlmock.Sqlmock) {
	t.Helper()

	database, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	mock.ExpectPing()
	repository, err := postgres.NewTodoRepository(database)
	require.NoError(t, err)

	return repository, mock
}

func TestTodoRepositoryCreate(t *testing.T) {
	repository, mock := newMockRepository(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO todo (title,content,status) VALUES ($1,$2,$3) RETURNING id",
	)).WithArgs("title", "content", false).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))

	id, err := repository.Create(context.Background(), &entity.Todo{
		Title: "title", Content: "content", Status: false,
	})

	require.NoError(t, err)
	require.Equal(t, vo.TodoID(4), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTodoRepositoryFindByID(t *testing.T) {
	repository, mock := newMockRepository(t)
	createdAt := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, title, content, status, created_at FROM todo WHERE id = $1",
	)).WithArgs(vo.TodoID(1)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "title", "content", "status", "created_at"}).
			AddRow(1, "title", "content", false, createdAt),
	)

	todo, err := repository.FindByID(context.Background(), vo.TodoID(1))

	require.NoError(t, err)
	require.Equal(t, &entity.Todo{
		ID: 1, Title: "title", Content: "content", Status: false, CreatedAt: createdAt,
	}, todo)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTodoRepositoryFindByIDNotFound(t *testing.T) {
	repository, mock := newMockRepository(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, title, content, status, created_at FROM todo WHERE id = $1",
	)).WithArgs(vo.TodoID(99)).WillReturnError(sql.ErrNoRows)

	_, err := repository.FindByID(context.Background(), vo.TodoID(99))

	require.EqualError(t, err, "Nenhum registro encontro.")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTodoRepositoryFindAllReturnsEmptySlice(t *testing.T) {
	repository, mock := newMockRepository(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, title, content, status, created_at FROM todo",
	)).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "status", "created_at"}))

	todos, err := repository.FindAll(context.Background())

	require.NoError(t, err)
	require.NotNil(t, todos)
	require.Empty(t, todos)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTodoRepositoryUpdate(t *testing.T) {
	repository, mock := newMockRepository(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE todo SET title = $1, content = $2, status = $3 WHERE id = $4",
	)).WithArgs("title", "content", true, vo.TodoID(2)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Update(context.Background(), &entity.Todo{
		ID: 2, Title: "title", Content: "content", Status: true,
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTodoRepositoryDeleteNoRowsAffected(t *testing.T) {
	repository, mock := newMockRepository(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM todo WHERE id = $1",
	)).WithArgs(vo.TodoID(20)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Delete(context.Background(), vo.TodoID(20))

	require.EqualError(t, err, "Nenhum registro afetado")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTodoRepositoryCreateSchema(t *testing.T) {
	repository, mock := newMockRepository(t)
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS todo").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.CreateSchema(context.Background())

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
