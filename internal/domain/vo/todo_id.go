package vo

import "app/internal/domain/apperror"

type TodoID int64

func NewTodoID(value int64) (TodoID, error) {
	if value <= 0 {
		return 0, apperror.ErrInvalidID
	}
	return TodoID(value), nil
}

func (id TodoID) Int64() int64 {
	return int64(id)
}
