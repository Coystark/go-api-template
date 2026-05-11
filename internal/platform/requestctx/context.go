package requestctx

import (
	"context"

	"github.com/google/uuid"
)

type userIDKey struct{}

// WithUserID anexa o ID do usuário autenticado ao contexto da requisição.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserID retorna o ID do usuário autenticado, se existir.
func UserID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return v, v != uuid.Nil
}
