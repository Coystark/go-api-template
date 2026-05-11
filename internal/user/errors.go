package user

import "errors"

var (
	// ErrNotFound indica que o usuário não existe.
	ErrNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists indica email duplicado.
	ErrEmailAlreadyExists = errors.New("email already exists")
)
