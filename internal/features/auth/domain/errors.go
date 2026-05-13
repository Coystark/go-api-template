package authdomain

import "github.com/caiohenrique/go-api-template/internal/platform/apperr"

// ErrInvalidCredentials indicates failed login (invalid credentials).
var ErrInvalidCredentials = apperr.Unauthorized("invalid credentials")
