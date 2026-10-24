package usecase

import (
	"context"
	"fmt"
	"strings"

	"app/internal/application/dto"
	"app/internal/application/ports"
	"app/internal/domain/entity"
	"app/internal/domain/vo"
)

type TodoService struct {
	repository ports.TodoRepository
}

func NewTodoService(repository ports.TodoRepository) *TodoService {
	return &TodoService{repository: repository}
}

func (s *TodoService) Create(ctx context.Context, input dto.CreateTodoInput) (int64, error) {
	if err := validateCreateInput(input); err != nil {
		return 0, err
	}

	todo, err := entity.NewTodo(input.Title, input.Content)
	if err != nil {
		return 0, err
	}

	id, err := s.repository.Create(ctx, todo)
	if err != nil {
		return 0, err
	}
	return id.Int64(), nil
}

func (s *TodoService) Get(ctx context.Context, rawID int64) (*dto.TodoOutput, error) {
	id, err := vo.NewTodoID(rawID)
	if err != nil {
		return nil, err
	}

	todo, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	output := toOutput(*todo)
	return &output, nil
}

func (s *TodoService) List(ctx context.Context) ([]dto.TodoOutput, error) {
	todos, err := s.repository.FindAll(ctx)
	if err != nil {
		return []dto.TodoOutput{}, err
	}

	output := make([]dto.TodoOutput, 0, len(todos))
	for _, todo := range todos {
		output = append(output, toOutput(todo))
	}

	return output, nil
}

func (s *TodoService) Update(ctx context.Context, rawID int64, input dto.UpdateTodoInput) error {
	if err := validateUpdateInput(input); err != nil {
		return err
	}

	id, err := vo.NewTodoID(rawID)
	if err != nil {
		return err
	}

	todo, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("não há elemento com tal id")
	}

	if input.Title != nil {
		todo.Title = *input.Title
	}
	if input.Content != nil {
		todo.Content = *input.Content
	}
	if input.Status != nil {
		todo.Status = *input.Status
	}

	return s.repository.Update(ctx, todo)
}

func (s *TodoService) Delete(ctx context.Context, rawID int64) error {
	id, err := vo.NewTodoID(rawID)
	if err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func toOutput(todo entity.Todo) dto.TodoOutput {
	return dto.TodoOutput{
		ID:        todo.ID.Int64(),
		Title:     todo.Title,
		Content:   todo.Content,
		Status:    todo.Status,
		CreatedAt: todo.CreatedAt,
	}
}

func validateCreateInput(input dto.CreateTodoInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("título é obrigatório")
	}
	if strings.TrimSpace(input.Content) == "" {
		return fmt.Errorf("conteúdo é obrigatório")
	}
	return nil
}

func validateUpdateInput(input dto.UpdateTodoInput) error {
	if input.Title == nil && input.Content == nil && input.Status == nil {
		return fmt.Errorf("ao menos um campo deve ser informado")
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return fmt.Errorf("título é obrigatório")
	}
	if input.Content != nil && strings.TrimSpace(*input.Content) == "" {
		return fmt.Errorf("conteúdo é obrigatório")
	}
	return nil
}
