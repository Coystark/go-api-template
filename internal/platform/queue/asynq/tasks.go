package asynqq

// Task types consumidos pelo worker.
const (
	TaskTypeWelcomeEmail = "user:welcome_email"
)

// WelcomeEmailPayload é o corpo JSON da task de boas-vindas.
type WelcomeEmailPayload struct {
	Email string `json:"email"`
}
