package hasher

import "golang.org/x/crypto/bcrypt"

// Hasher defines the interface for password hashing operations
type Hasher interface {
	Hash(password string) ([]byte, error)
	Compare(hashedPassword []byte, password string) error
}

// BcryptHasher implements Hasher using bcrypt
type BcryptHasher struct {
	cost int
}

// NewBcrypt creates a new bcrypt hasher with default cost
func NewBcrypt() Hasher {
	return &BcryptHasher{
		cost: bcrypt.DefaultCost,
	}
}

// NewBcryptWithCost creates a bcrypt hasher with custom cost
func NewBcryptWithCost(cost int) Hasher {
	return &BcryptHasher{
		cost: cost,
	}
}

// Hash generates a bcrypt hash from a password
func (h *BcryptHasher) Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), h.cost)
}

// Compare verifies a password against a hash
func (h *BcryptHasher) Compare(hashedPassword []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}
