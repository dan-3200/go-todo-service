package apperror

import "errors"

var (
	ErrInvalidID        = errors.New("id inválido")
	ErrTitleRequired    = errors.New("título é obrigatório")
	ErrContentRequired  = errors.New("conteúdo é obrigatório")
	ErrNoFieldsProvided = errors.New("ao menos um campo deve ser informado")
	ErrTodoNotFound     = errors.New("todo não encontrado")
)
