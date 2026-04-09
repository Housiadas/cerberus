package config

type DB struct {
	Host                  string `validate:"required"`
	Port                  uint16
	Name                  string `validate:"required"`
	User                  string `validate:"required"`
	Password              string `validate:"required"`
	MinConns              int
	MaxConns              int
	DisableTLS            bool
	PostgresImage         string
	PostgresContainerName string
}
