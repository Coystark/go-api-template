package id

import "github.com/google/uuid"

// New retorna um UUID versão 7 (ordenado por tempo), gerado na aplicação.
func New() (uuid.UUID, error) {
	return uuid.NewV7()
}
