package auth

// ErrorResponse representa erro JSON da API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// LoginDTO representa credenciais de login.
type LoginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenResponseDTO retorna o token JWT ao cliente.
type TokenResponseDTO struct {
	AccessToken string `json:"access_token"`
}
