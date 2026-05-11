package order

import "errors"

// ErrNotFound indica que o pedido não existe.
var ErrNotFound = errors.New("order not found")
