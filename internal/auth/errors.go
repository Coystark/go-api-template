package auth

import "errors"

// ErrInvalidCredentials indica falha de login (credenciais inválidas).
var ErrInvalidCredentials = errors.New("invalid credentials")
