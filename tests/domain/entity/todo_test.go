package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"app/internal/domain/entity"
)

func TestNewTodo(t *testing.T) {
	todo, err := entity.NewTodo("TITLE", "Content")

	require.NoError(t, err)
	require.Equal(t, "title", todo.Title)
	require.Equal(t, "content", todo.Content)
	require.False(t, todo.Status)
	require.False(t, todo.CreatedAt.IsZero())
}

func TestNewTodoValidatesRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		message string
	}{
		{name: "title", title: "  ", content: "content", message: "título é obrigatório"},
		{name: "content", title: "title", content: "  ", message: "conteúdo é obrigatório"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := entity.NewTodo(test.title, test.content)
			require.EqualError(t, err, test.message)
		})
	}
}
