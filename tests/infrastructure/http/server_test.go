package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"app/internal/application/usecase"
	"app/internal/domain/entity"
	"app/internal/domain/vo"
	httpadapter "app/internal/infrastructure/http"
)

type fakeRepository struct {
	todo      *entity.Todo
	todos     []entity.Todo
	created   *entity.Todo
	updated   *entity.Todo
	deletedID vo.TodoID
}

func (r *fakeRepository) Create(_ context.Context, todo *entity.Todo) (vo.TodoID, error) {
	r.created = todo
	return vo.TodoID(10), nil
}

func (r *fakeRepository) FindByID(_ context.Context, _ vo.TodoID) (*entity.Todo, error) {
	if r.todo == nil {
		return nil, errors.New("Nenhum registro encontro.")
	}
	return r.todo, nil
}

func (r *fakeRepository) FindAll(context.Context) ([]entity.Todo, error) {
	return r.todos, nil
}

func (r *fakeRepository) Update(_ context.Context, todo *entity.Todo) error {
	r.updated = todo
	return nil
}

func (r *fakeRepository) Delete(_ context.Context, id vo.TodoID) error {
	r.deletedID = id
	return nil
}

type fakePinger struct{ err error }

func (p fakePinger) Ping() error { return p.err }

func newTestServer(repository *fakeRepository, pinger fakePinger) http.Handler {
	gin.SetMode(gin.TestMode)
	service := usecase.NewTodoService(repository)
	handler := httpadapter.NewTodoHandler(service)
	return httpadapter.NewServer(handler, pinger)
}

func performRequest(server http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func TestCreateTodoContract(t *testing.T) {
	repository := &fakeRepository{}
	server := newTestServer(repository, fakePinger{})

	response := performRequest(server, http.MethodPost, "/todo", []byte(`{
		"title":"TITLE","content":"CONTENT","status":true
	}`))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"id":10,"msg":"ToDo Criado"}`, response.Body.String())
	require.Equal(t, "title", repository.created.Title)
	require.Equal(t, "content", repository.created.Content)
	require.False(t, repository.created.Status)
}

func TestCreateTodoBindingErrorReturnsInternalServerError(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performRequest(server, http.MethodPost, "/todo", []byte(`{"title":"only title"}`))

	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetTodoContract(t *testing.T) {
	createdAt := time.Date(2026, time.October, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{todo: &entity.Todo{
		ID: 1, Title: "title", Content: "content", Status: true, CreatedAt: createdAt,
	}}
	server := newTestServer(repository, fakePinger{})

	response := performRequest(server, http.MethodGet, "/todo/1", nil)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
		"id":1,"title":"title","content":"content","status":true,
		"created_at":"2026-10-24T12:00:00Z"
	}`, response.Body.String())
}

func TestInvalidPathIDReturnsInternalServerError(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performRequest(server, http.MethodGet, "/todo/not-a-number", nil)

	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestListTodoContractUsesEmptyArray(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performRequest(server, http.MethodGet, "/todo", nil)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `[]`, response.Body.String())
}

func TestUpdateTodoContract(t *testing.T) {
	repository := &fakeRepository{todo: &entity.Todo{
		ID: 3, Title: "old", Content: "content", Status: false,
	}}
	server := newTestServer(repository, fakePinger{})

	response := performRequest(server, http.MethodPatch, "/todo/3", []byte(`{
		"title":"New Title","status":true
	}`))

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ToDo atualizado", response.Body.String())
	require.Equal(t, "New Title", repository.updated.Title)
	require.Equal(t, "content", repository.updated.Content)
	require.True(t, repository.updated.Status)
}

func TestDeleteTodoContract(t *testing.T) {
	repository := &fakeRepository{}
	server := newTestServer(repository, fakePinger{})

	response := performRequest(server, http.MethodDelete, "/todo/8", nil)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ToDo deletado", response.Body.String())
	require.Equal(t, vo.TodoID(8), repository.deletedID)
}

func TestHealthAndReadinessContracts(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := newTestServer(&fakeRepository{}, fakePinger{})

		health := performRequest(server, http.MethodGet, "/healthz", nil)
		ready := performRequest(server, http.MethodGet, "/readiness", nil)

		require.Equal(t, http.StatusOK, health.Code)
		require.JSONEq(t, `{"status":"healthy"}`, health.Body.String())
		require.Equal(t, http.StatusOK, ready.Code)
		require.JSONEq(t, `{"status":"ready","postgres":true}`, ready.Body.String())
	})

	t.Run("database unavailable", func(t *testing.T) {
		server := newTestServer(&fakeRepository{}, fakePinger{err: errors.New("offline")})

		response := performRequest(server, http.MethodGet, "/readiness", nil)

		require.Equal(t, http.StatusServiceUnavailable, response.Code)
		require.JSONEq(t, `{"postgres":false}`, response.Body.String())
	})
}

func TestTodoResponseHasExpectedJSONFields(t *testing.T) {
	server := newTestServer(&fakeRepository{todo: &entity.Todo{ID: 1}}, fakePinger{})
	response := performRequest(server, http.MethodGet, "/todo/1", nil)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.ElementsMatch(t,
		[]string{"id", "title", "content", "status", "created_at"},
		mapKeys(payload),
	)
}

func mapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
