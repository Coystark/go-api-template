package bcryptadapter

import "golang.org/x/crypto/bcrypt"

// Hasher implements password hashing with bcrypt (shared platform adapter).
type Hasher struct{}

// New returns a bcrypt-backed hasher.
func New() *Hasher { return &Hasher{} }

// Hash returns a bcrypt hash of plain.
func (Hasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// Compare checks plain against a bcrypt hash.
func (Hasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
