package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

// NewBcrypt creates a new bcrypt hasher with default cost.
func NewBcrypt() *BcryptHasher {
	return &BcryptHasher{
		cost: bcrypt.DefaultCost,
	}
}

// NewBcryptWithCost creates a bcrypt hasher with custom cost.
func NewBcryptWithCost(cost int) *BcryptHasher {
	return &BcryptHasher{
		cost: cost,
	}
}

// Hash generates a bcrypt hash from a password.
func (h *BcryptHasher) Hash(password string) ([]byte, error) {
	fromPassword, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	return fromPassword, nil
}

// Compare verifies a password against a hash.
func (h *BcryptHasher) Compare(hashedPassword []byte, password string) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}

	return nil
}
