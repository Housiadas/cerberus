// Package uuidgen is a wrapper for the uuid library
package uuidgen

import (
	"fmt"

	"github.com/google/uuid"
)

type Generator interface {
	Generate() (uuid.UUID, error)
}

type V7Generator struct{}

// NewV7 creates a new V7 UUID generator.
func NewV7() *V7Generator {
	return &V7Generator{}
}

// Generate creates a new UUID v7.
func (g *V7Generator) Generate() (uuid.UUID, error) {
	newUUID, err := uuid.NewV7()

	return newUUID, fmt.Errorf("uuid v7 error: %w", err)
}
