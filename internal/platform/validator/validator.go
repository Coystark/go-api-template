package validator

import (
	playground "github.com/go-playground/validator/v10"
)

// Validator encapsula validator/v10 para injeção de dependência.
type Validator struct {
	v *playground.Validate
}

// New cria uma instância pronta para uso.
func New() *Validator {
	return &Validator{v: playground.New()}
}

// Struct valida uma struct com tags `validate`.
func (val *Validator) Struct(s interface{}) error {
	return val.v.Struct(s)
}
