package auth

import "github.com/caiohenrique/go-api-template/internal/platform/apperr"

// ErrInvalidCredentials indica falha de login (credenciais inválidas).
var ErrInvalidCredentials = apperr.Unauthorized("invalid credentials")
