package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"app/internal/application/usecase"
	"app/internal/domain/apperror"
	"app/internal/domain/entity"
	"app/internal/domain/vo"
	httpadapter "app/internal/infrastructure/http"
	"app/internal/infrastructure/http/middleware"
)

const testJWTSecret = "test-secret-with-at-least-32-bytes"

type fakeRepository struct {
	todo      *entity.Todo
	todos     []entity.Todo
	created   *entity.Todo
	updated   *entity.Todo
	deletedID vo.TodoID
	findErr   error
	updateErr error
	deleteErr error
}

func (r *fakeRepository) Create(_ context.Context, todo *entity.Todo) (vo.TodoID, error) {
	r.created = todo
	return vo.TodoID(10), nil
}

func (r *fakeRepository) FindByID(_ context.Context, _ vo.TodoID) (*entity.Todo, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.todo == nil {
		return nil, apperror.ErrTodoNotFound
	}
	return r.todo, nil
}

func (r *fakeRepository) FindAll(context.Context) ([]entity.Todo, error) {
	return r.todos, nil
}

func (r *fakeRepository) Update(_ context.Context, todo *entity.Todo) error {
	r.updated = todo
	return r.updateErr
}

func (r *fakeRepository) Delete(_ context.Context, id vo.TodoID) error {
	r.deletedID = id
	return r.deleteErr
}

type fakePinger struct{ err error }

func (p fakePinger) Ping() error { return p.err }

func newTestServer(repository *fakeRepository, pinger fakePinger) http.Handler {
	gin.SetMode(gin.TestMode)
	service := usecase.NewTodoService(repository)
	handler := httpadapter.NewTodoHandler(service)
	authMiddleware, err := middleware.NewJWT(testJWTSecret)
	if err != nil {
		panic(err)
	}
	return httpadapter.NewServer(handler, pinger, authMiddleware)
}

func performRequest(server http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequestWithContext(context.Background(), method, target, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func performAuthenticatedRequest(t *testing.T, server http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "test-user",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	request := httptest.NewRequestWithContext(t.Context(), method, target, bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+signed)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func requireEnvelope(t *testing.T, response *httptest.ResponseRecorder, status int, success bool) map[string]any {
	t.Helper()
	require.Equal(t, status, response.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, success, payload["success"])
	require.EqualValues(t, status, payload["status"])

	timestamp, ok := payload["timestamp"].(string)
	require.True(t, ok)
	_, err := time.Parse(time.RFC3339Nano, timestamp)
	require.NoError(t, err)

	return payload
}

func requireError(t *testing.T, payload map[string]any, code, message string) {
	t.Helper()
	errorPayload, ok := payload["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, code, errorPayload["code"])
	require.Equal(t, message, errorPayload["message"])
}

func TestCreateTodoContract(t *testing.T) {
	repository := &fakeRepository{}
	server := newTestServer(repository, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodPost, "/todo", []byte(`{
		"title":"TITLE","content":"CONTENT","status":true
	}`))

	payload := requireEnvelope(t, response, http.StatusOK, true)
	data := payload["data"].(map[string]any)
	require.EqualValues(t, 10, data["id"])
	require.Equal(t, "ToDo Criado", data["message"])
	require.Equal(t, "title", repository.created.Title)
	require.Equal(t, "content", repository.created.Content)
	require.False(t, repository.created.Status)
}

func TestCreateTodoBindingErrorUsesErrorEnvelope(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodPost, "/todo", []byte(`{"title":"only title"}`))

	payload := requireEnvelope(t, response, http.StatusBadRequest, false)
	requireError(t, payload, "INVALID_REQUEST", "corpo da requisição inválido")
	require.NotEmpty(t, payload["error"].(map[string]any)["details"])
}

func TestGetTodoContract(t *testing.T) {
	createdAt := time.Date(2026, time.October, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{todo: &entity.Todo{
		ID: 1, Title: "title", Content: "content", Status: true, CreatedAt: createdAt,
	}}
	server := newTestServer(repository, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodGet, "/todo/1", nil)

	payload := requireEnvelope(t, response, http.StatusOK, true)
	data := payload["data"].(map[string]any)
	require.EqualValues(t, 1, data["id"])
	require.Equal(t, "title", data["title"])
	require.Equal(t, "content", data["content"])
	require.Equal(t, true, data["status"])
	require.Equal(t, "2026-10-24T12:00:00Z", data["created_at"])
}

func TestInvalidPathIDUsesErrorEnvelope(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodGet, "/todo/not-a-number", nil)

	payload := requireEnvelope(t, response, http.StatusBadRequest, false)
	requireError(t, payload, "INVALID_ID", "id inválido")
}

func TestSemanticValidationErrorsUseErrorEnvelope(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		target  string
		body    []byte
		code    string
		message string
	}{
		{
			name: "invalid ID", method: http.MethodGet, target: "/todo/0",
			code: "INVALID_ID", message: "id inválido",
		},
		{
			name: "title required", method: http.MethodPost, target: "/todo",
			body: []byte(`{"title":" ","content":"content"}`),
			code: "VALIDATION_ERROR", message: "título é obrigatório",
		},
		{
			name: "content required", method: http.MethodPost, target: "/todo",
			body: []byte(`{"title":"title","content":" "}`),
			code: "VALIDATION_ERROR", message: "conteúdo é obrigatório",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newTestServer(&fakeRepository{}, fakePinger{})

			response := performAuthenticatedRequest(t, server, test.method, test.target, test.body)

			payload := requireEnvelope(t, response, http.StatusBadRequest, false)
			requireError(t, payload, test.code, test.message)
		})
	}
}

func TestListTodoContractUsesEmptyArray(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodGet, "/todo", nil)

	payload := requireEnvelope(t, response, http.StatusOK, true)
	data, ok := payload["data"].([]any)
	require.True(t, ok)
	require.Empty(t, data)
}

func TestUpdateTodoContract(t *testing.T) {
	repository := &fakeRepository{todo: &entity.Todo{
		ID: 3, Title: "old", Content: "content", Status: false,
	}}
	server := newTestServer(repository, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodPatch, "/todo/3", []byte(`{
		"title":"New Title","status":true
	}`))

	payload := requireEnvelope(t, response, http.StatusOK, true)
	require.Equal(t, "ToDo atualizado", payload["data"].(map[string]any)["message"])
	require.Equal(t, "New Title", repository.updated.Title)
	require.Equal(t, "content", repository.updated.Content)
	require.True(t, repository.updated.Status)
}

func TestUpdateEmptyInputUsesValidationError(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodPatch, "/todo/1", []byte(`{}`))

	payload := requireEnvelope(t, response, http.StatusBadRequest, false)
	requireError(t, payload, "VALIDATION_ERROR", "ao menos um campo deve ser informado")
}

func TestWrappedTodoNotFoundUsesNotFoundEnvelope(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		repository := &fakeRepository{findErr: fmt.Errorf("lookup failed: %w", apperror.ErrTodoNotFound)}
		server := newTestServer(repository, fakePinger{})

		response := performAuthenticatedRequest(t, server, http.MethodGet, "/todo/1", nil)

		payload := requireEnvelope(t, response, http.StatusNotFound, false)
		requireError(t, payload, "TODO_NOT_FOUND", "todo não encontrado")
	})

	t.Run("delete", func(t *testing.T) {
		repository := &fakeRepository{deleteErr: fmt.Errorf("delete failed: %w", apperror.ErrTodoNotFound)}
		server := newTestServer(repository, fakePinger{})

		response := performAuthenticatedRequest(t, server, http.MethodDelete, "/todo/1", nil)

		payload := requireEnvelope(t, response, http.StatusNotFound, false)
		requireError(t, payload, "TODO_NOT_FOUND", "todo não encontrado")
	})

	t.Run("update", func(t *testing.T) {
		repository := &fakeRepository{findErr: fmt.Errorf("lookup failed: %w", apperror.ErrTodoNotFound)}
		server := newTestServer(repository, fakePinger{})

		response := performAuthenticatedRequest(t, server, http.MethodPatch, "/todo/1", []byte(`{"status":true}`))

		payload := requireEnvelope(t, response, http.StatusNotFound, false)
		requireError(t, payload, "TODO_NOT_FOUND", "todo não encontrado")
	})
}

func TestUpdateUnexpectedRepositoryErrorUsesInternalError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeRepository{findErr: repositoryError}
	server := newTestServer(repository, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodPatch, "/todo/1", []byte(`{"status":true}`))

	payload := requireEnvelope(t, response, http.StatusInternalServerError, false)
	requireError(t, payload, "INTERNAL_ERROR", "erro interno do servidor")
	require.NotContains(t, payload["error"].(map[string]any), "details")
}

func TestDeleteTodoContract(t *testing.T) {
	repository := &fakeRepository{}
	server := newTestServer(repository, fakePinger{})

	response := performAuthenticatedRequest(t, server, http.MethodDelete, "/todo/8", nil)

	payload := requireEnvelope(t, response, http.StatusOK, true)
	require.Equal(t, "ToDo deletado", payload["data"].(map[string]any)["message"])
	require.Equal(t, vo.TodoID(8), repository.deletedID)
}

func TestTodoRoutesRequireAuthentication(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	tests := []struct {
		method string
		target string
		body   []byte
	}{
		{method: http.MethodPost, target: "/todo", body: []byte(`{"title":"title","content":"content"}`)},
		{method: http.MethodGet, target: "/todo"},
		{method: http.MethodGet, target: "/todo/1"},
		{method: http.MethodPatch, target: "/todo/1", body: []byte(`{"status":true}`)},
		{method: http.MethodDelete, target: "/todo/1"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.target, func(t *testing.T) {
			response := performRequest(server, test.method, test.target, test.body)
			payload := requireEnvelope(t, response, http.StatusUnauthorized, false)
			requireError(t, payload, "UNAUTHORIZED", "token de autenticação ausente ou inválido")
			require.Equal(t, "Bearer", response.Header().Get("WWW-Authenticate"))
		})
	}
}

func TestHealthAndReadinessContracts(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := newTestServer(&fakeRepository{}, fakePinger{})

		health := requireEnvelope(t, performRequest(server, http.MethodGet, "/healthz", nil), http.StatusOK, true)
		ready := requireEnvelope(t, performRequest(server, http.MethodGet, "/readiness", nil), http.StatusOK, true)

		require.Equal(t, "healthy", health["data"].(map[string]any)["status"])
		require.Equal(t, "ready", ready["data"].(map[string]any)["status"])
		require.Equal(t, true, ready["data"].(map[string]any)["postgres"])
	})

	t.Run("database unavailable", func(t *testing.T) {
		server := newTestServer(&fakeRepository{}, fakePinger{err: errors.New("offline")})

		payload := requireEnvelope(t, performRequest(server, http.MethodGet, "/readiness", nil), http.StatusServiceUnavailable, false)
		requireError(t, payload, "DATABASE_UNAVAILABLE", "banco de dados indisponível")
	})
}

func TestUnknownRouteUsesErrorEnvelope(t *testing.T) {
	server := newTestServer(&fakeRepository{}, fakePinger{})

	payload := requireEnvelope(t, performRequest(server, http.MethodGet, "/unknown", nil), http.StatusNotFound, false)
	requireError(t, payload, "ROUTE_NOT_FOUND", "rota não encontrada")
}
