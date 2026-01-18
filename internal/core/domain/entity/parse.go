package entity

import "fmt"

// MustParse parses the string value and returns a role if one exists.
// If an error occurs, the function panics.
func MustParse(value string) Entity {
	entity, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return entity
}

// Parse parses the string value and returns a role if one exists.
func Parse(value string) (Entity, error) {
	entity, exists := getEntities()[value]
	if !exists {
		return Entity{}, fmt.Errorf("%w: %q", ErrInvalidEntity, value)
	}

	return entity, nil
}
