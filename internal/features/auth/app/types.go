package authapp

// LoginInput is input for login.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenView is the login response with JWT.
type TokenView struct {
	AccessToken string `json:"access_token"`
}
