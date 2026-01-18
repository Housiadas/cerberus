package entity

const (
	UserEntity = "USER"
	RoleEntity = "ROLE"
)

// Entity represents a domain in the system.
type Entity struct {
	value string
}

func New(entity string) Entity {
	return Entity{entity}
}

// String returns the name of the role.
func (e Entity) String() string {
	return e.value
}

// Equal provides support for the go-cmp package and testing.
func (e Entity) Equal(d2 Entity) bool {
	return e.value == d2.value
}

// MarshalText provides support for logging and any marshal needs.
func (e Entity) MarshalText() ([]byte, error) {
	return []byte(e.value), nil
}

func getEntities() map[string]Entity {
	return map[string]Entity{
		UserEntity: New("USER"),
		RoleEntity: New("ROLE"),
	}
}
