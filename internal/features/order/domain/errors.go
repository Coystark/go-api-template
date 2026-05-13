package orderdomain

import "github.com/caiohenrique/go-api-template/internal/platform/apperr"

// ErrNotFound indicates the order does not exist.
var ErrNotFound = apperr.NotFound("order not found")

// ErrServiceNotInOrder indicates an order_service does not belong to the target order.
var ErrServiceNotInOrder = apperr.BadRequest("order service does not belong to order")
