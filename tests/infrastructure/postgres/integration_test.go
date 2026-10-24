package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"app/internal/domain/entity"
	"app/internal/domain/vo"
	"app/internal/infrastructure/config"
	infrastructurepostgres "app/internal/infrastructure/postgres"
)

const integrationTimeout = time.Minute

var (
	integrationDatabase   *sql.DB
	integrationRepository *infrastructurepostgres.TodoRepository
	integrationContainer  *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	if err := startIntegrationDatabase(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start PostgreSQL test container:", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	if err := stopIntegrationDatabase(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to stop PostgreSQL test container:", err)
		if exitCode == 0 {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func startIntegrationDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("todo_test"),
		postgres.WithUsername("todo"),
		postgres.WithPassword("todo"),
		postgres.WithSQLDriver("pgx"),
		postgres.WithInitScripts(schemaPath()),
		postgres.BasicWaitStrategies(),
	)
	integrationContainer = container
	if err != nil {
		_ = stopIntegrationDatabase()
		return err
	}

	connectionString, err := integrationContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = stopIntegrationDatabase()
		return err
	}

	integrationDatabase, err = sql.Open("pgx", connectionString)
	if err != nil {
		_ = stopIntegrationDatabase()
		return err
	}
	if err := integrationDatabase.PingContext(ctx); err != nil {
		_ = stopIntegrationDatabase()
		return err
	}

	integrationRepository, err = infrastructurepostgres.NewTodoRepository(integrationDatabase)
	if err != nil {
		_ = stopIntegrationDatabase()
		return err
	}
	return nil
}

func stopIntegrationDatabase() error {
	var result error
	if integrationDatabase != nil {
		result = integrationDatabase.Close()
		integrationDatabase = nil
	}
	if integrationContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
		defer cancel()
		if err := integrationContainer.Terminate(ctx); err != nil && result == nil {
			result = err
		}
		integrationContainer = nil
	}
	return result
}

func schemaPath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to resolve integration test path")
	}
	return filepath.Join(filepath.Dir(currentFile), "../../../database/schema.sql")
}

func resetIntegrationDatabase(t *testing.T) {
	t.Helper()

	_, err := integrationDatabase.ExecContext(t.Context(), "TRUNCATE TABLE todo RESTART IDENTITY")
	require.NoError(t, err)
}

func createIntegrationTodo(t *testing.T, title, content string, status bool) *entity.Todo {
	t.Helper()

	todo := &entity.Todo{Title: title, Content: content, Status: status}
	id, err := integrationRepository.Create(t.Context(), todo)
	require.NoError(t, err)
	todo.ID = id
	return todo
}

func TestTodoRepositoryIntegrationCreate(t *testing.T) {
	resetIntegrationDatabase(t)

	id, err := integrationRepository.Create(t.Context(), &entity.Todo{
		Title: "title", Content: "content", Status: false,
	})

	require.NoError(t, err)
	require.Equal(t, vo.TodoID(1), id)

	todo, err := integrationRepository.FindByID(t.Context(), id)
	require.NoError(t, err)
	require.Equal(t, "title", todo.Title)
	require.Equal(t, "content", todo.Content)
	require.False(t, todo.Status)
	require.False(t, todo.CreatedAt.IsZero())
}

func TestOpenConfiguresConnectionPool(t *testing.T) {
	connectionString, err := integrationContainer.ConnectionString(t.Context(), "sslmode=disable")
	require.NoError(t, err)

	connectionURL, err := url.Parse(connectionString)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(connectionURL.Host)
	require.NoError(t, err)

	connection, err := infrastructurepostgres.Open(config.PostgresConfig{
		User:            "todo",
		Password:        "todo",
		Host:            host,
		Port:            port,
		Database:        "todo_test",
		MaxOpenConns:    3,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })

	require.Equal(t, 3, connection.Stats().MaxOpenConnections)
}

func TestTodoRepositoryIntegrationFindByIDNotFound(t *testing.T) {
	resetIntegrationDatabase(t)

	_, err := integrationRepository.FindByID(t.Context(), vo.TodoID(99))

	require.EqualError(t, err, "Nenhum registro encontro.")
}

func TestTodoRepositoryIntegrationFindAllReturnsEmptySlice(t *testing.T) {
	resetIntegrationDatabase(t)

	todos, err := integrationRepository.FindAll(t.Context())

	require.NoError(t, err)
	require.NotNil(t, todos)
	require.Empty(t, todos)
}

func TestTodoRepositoryIntegrationUpdate(t *testing.T) {
	resetIntegrationDatabase(t)
	todo := createIntegrationTodo(t, "old title", "old content", false)
	todo.Title = "new title"
	todo.Content = "new content"
	todo.Status = true

	err := integrationRepository.Update(t.Context(), todo)
	require.NoError(t, err)

	updated, err := integrationRepository.FindByID(t.Context(), todo.ID)
	require.NoError(t, err)
	require.Equal(t, "new title", updated.Title)
	require.Equal(t, "new content", updated.Content)
	require.True(t, updated.Status)
}

func TestTodoRepositoryIntegrationDelete(t *testing.T) {
	resetIntegrationDatabase(t)
	todo := createIntegrationTodo(t, "title", "content", false)

	err := integrationRepository.Delete(t.Context(), todo.ID)
	require.NoError(t, err)

	_, err = integrationRepository.FindByID(t.Context(), todo.ID)
	require.EqualError(t, err, "Nenhum registro encontro.")
}

func TestTodoRepositoryIntegrationDeleteNoRowsAffected(t *testing.T) {
	resetIntegrationDatabase(t)

	err := integrationRepository.Delete(t.Context(), vo.TodoID(20))

	require.EqualError(t, err, "Nenhum registro afetado")
}

func TestTodoRepositoryIntegrationRejectsNullableColumns(t *testing.T) {
	resetIntegrationDatabase(t)

	_, err := integrationDatabase.ExecContext(
		t.Context(),
		"INSERT INTO todo (title, content, status, created_at) VALUES (NULL, 'content', false, NOW())",
	)
	require.NoError(t, err)

	_, err = integrationRepository.FindByID(t.Context(), vo.TodoID(1))
	require.EqualError(t, err, "registro todo possui campo nulo")
}
