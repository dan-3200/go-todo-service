package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"app/internal/application/dto"
	"app/internal/application/usecase"
	"app/internal/domain/entity"
	"app/internal/domain/vo"
)

type fakeTodoRepository struct {
	todo        *entity.Todo
	todos       []entity.Todo
	createdTodo *entity.Todo
	updatedTodo *entity.Todo
	deletedID   vo.TodoID
	createID    vo.TodoID
	findErr     error
	listErr     error
	createErr   error
	updateErr   error
	deleteErr   error
}

func (r *fakeTodoRepository) Create(_ context.Context, todo *entity.Todo) (vo.TodoID, error) {
	r.createdTodo = todo
	return r.createID, r.createErr
}

func (r *fakeTodoRepository) FindByID(_ context.Context, _ vo.TodoID) (*entity.Todo, error) {
	return r.todo, r.findErr
}

func (r *fakeTodoRepository) FindAll(context.Context) ([]entity.Todo, error) {
	return r.todos, r.listErr
}

func (r *fakeTodoRepository) Update(_ context.Context, todo *entity.Todo) error {
	r.updatedTodo = todo
	return r.updateErr
}

func (r *fakeTodoRepository) Delete(_ context.Context, id vo.TodoID) error {
	r.deletedID = id
	return r.deleteErr
}

func TestTodoServiceCreate(t *testing.T) {
	repository := &fakeTodoRepository{createID: vo.TodoID(12)}
	service := usecase.NewTodoService(repository)

	id, err := service.Create(context.Background(), dto.CreateTodoInput{Title: "TITLE", Content: "CONTENT"})

	require.NoError(t, err)
	require.Equal(t, int64(12), id)
	require.Equal(t, "title", repository.createdTodo.Title)
	require.Equal(t, "content", repository.createdTodo.Content)
}

func TestTodoServiceGet(t *testing.T) {
	createdAt := time.Now()
	repository := &fakeTodoRepository{todo: &entity.Todo{
		ID: 1, Title: "title", Content: "content", CreatedAt: createdAt,
	}}
	service := usecase.NewTodoService(repository)

	output, err := service.Get(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, &dto.TodoOutput{
		ID: 1, Title: "title", Content: "content", CreatedAt: createdAt,
	}, output)
}

func TestTodoServiceRejectsInvalidIDs(t *testing.T) {
	service := usecase.NewTodoService(&fakeTodoRepository{})

	_, err := service.Get(context.Background(), 0)
	require.EqualError(t, err, "id inválido")
	require.EqualError(t, service.Update(context.Background(), 0, dto.UpdateTodoInput{}), "id inválido")
	require.EqualError(t, service.Delete(context.Background(), 0), "id inválido")
}

func TestTodoServiceListReturnsEmptySlice(t *testing.T) {
	service := usecase.NewTodoService(&fakeTodoRepository{})

	output, err := service.List(context.Background())

	require.NoError(t, err)
	require.NotNil(t, output)
	require.Empty(t, output)
}

func TestTodoServiceListPropagatesErrorWithEmptySlice(t *testing.T) {
	repositoryError := errors.New("database error")
	service := usecase.NewTodoService(&fakeTodoRepository{listErr: repositoryError})

	output, err := service.List(context.Background())

	require.ErrorIs(t, err, repositoryError)
	require.NotNil(t, output)
	require.Empty(t, output)
}

func TestTodoServiceUpdatePartialFields(t *testing.T) {
	todo := &entity.Todo{ID: 1, Title: "old", Content: "content", Status: false}
	repository := &fakeTodoRepository{todo: todo}
	service := usecase.NewTodoService(repository)
	title := "New Title"
	status := true

	err := service.Update(context.Background(), 1, dto.UpdateTodoInput{Title: &title, Status: &status})

	require.NoError(t, err)
	require.Same(t, todo, repository.updatedTodo)
	require.Equal(t, "New Title", repository.updatedTodo.Title)
	require.Equal(t, "content", repository.updatedTodo.Content)
	require.True(t, repository.updatedTodo.Status)
}

func TestTodoServiceUpdateHidesRepositoryLookupError(t *testing.T) {
	service := usecase.NewTodoService(&fakeTodoRepository{findErr: errors.New("database error")})

	err := service.Update(context.Background(), 1, dto.UpdateTodoInput{})

	require.EqualError(t, err, "não há elemento com tal id")
}

func TestTodoServiceDelete(t *testing.T) {
	repository := &fakeTodoRepository{}
	service := usecase.NewTodoService(repository)

	err := service.Delete(context.Background(), 7)

	require.NoError(t, err)
	require.Equal(t, vo.TodoID(7), repository.deletedID)
}
