package user

import "github.com/caiohenrique/go-api-template/internal/platform/apperr"

var (
	// ErrNotFound indica que o usuário não existe.
	ErrNotFound = apperr.NotFound("user not found")
	// ErrEmailAlreadyExists indica email duplicado.
	ErrEmailAlreadyExists = apperr.Conflict("email already exists")
)
