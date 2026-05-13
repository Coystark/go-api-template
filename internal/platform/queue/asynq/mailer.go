package asynqq

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Mailer enqueues welcome-email tasks via asynq.
type Mailer struct {
	client *asynq.Client
}

// NewMailer wraps an asynq client as a welcome-email mailer.
func NewMailer(c *asynq.Client) *Mailer {
	return &Mailer{client: c}
}

// SendWelcome enqueues the welcome email task for the given address.
func (m *Mailer) SendWelcome(ctx context.Context, email string) error {
	payload, err := json.Marshal(WelcomeEmailPayload{Email: email})
	if err != nil {
		return err
	}
	_, err = m.client.EnqueueContext(ctx, asynq.NewTask(TaskTypeWelcomeEmail, payload))
	return err
}
