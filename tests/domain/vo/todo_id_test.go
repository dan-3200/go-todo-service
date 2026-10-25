package vo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"app/internal/domain/apperror"
	"app/internal/domain/vo"
)

func TestNewTodoID(t *testing.T) {
	id, err := vo.NewTodoID(7)

	require.NoError(t, err)
	require.Equal(t, int64(7), id.Int64())
}

func TestNewTodoIDRejectsNonPositiveValues(t *testing.T) {
	for _, value := range []int64{0, -1} {
		_, err := vo.NewTodoID(value)
		require.ErrorIs(t, err, apperror.ErrInvalidID)
	}
}
