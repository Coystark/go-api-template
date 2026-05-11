package order

import "errors"

// ErrNotFound indica que o pedido não existe.
var ErrNotFound = errors.New("order not found")

// ErrServiceNotInOrder indica que um order_service referenciado não pertence ao pedido alvo.
var ErrServiceNotInOrder = errors.New("order service does not belong to order")
