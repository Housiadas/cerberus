// Package uuidgen is a wrapper for the uuid library
package uuidgen

import "github.com/google/uuid"

type Generator interface {
	Generate() (uuid.UUID, error)
}

type V7Generator struct{}

// NewV7 creates a new V7 UUID generator.
func NewV7() Generator {
	return &V7Generator{}
}

// Generate creates a new UUID v7.
func (g *V7Generator) Generate() (uuid.UUID, error) {
	return uuid.NewV7()
}
