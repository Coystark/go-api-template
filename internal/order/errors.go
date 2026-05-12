package order

import "github.com/caiohenrique/go-api-template/internal/platform/apperr"

// ErrNotFound indica que o pedido não existe.
var ErrNotFound = apperr.NotFound("order not found")

// ErrServiceNotInOrder indica que um order_service referenciado não pertence ao pedido alvo.
var ErrServiceNotInOrder = apperr.BadRequest("order service does not belong to order")
