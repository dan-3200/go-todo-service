package dto

import "time"

type CreateTodoInput struct {
	Title   string
	Content string
}

type UpdateTodoInput struct {
	Title   *string
	Content *string
	Status  *bool
}

type TodoOutput struct {
	ID        int64
	Title     string
	Content   string
	Status    bool
	CreatedAt time.Time
}
