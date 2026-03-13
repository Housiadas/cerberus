package config

type DB struct {
	User                  string `validate:"required"`
	Password              string `validate:"required"`
	Name                  string `validate:"required"`
	Port                  string
	Host                  string `validate:"required"`
	MaxOpenConns          int
	MaxIdleConns          int
	ConnectionIdleTime    string
	DisableTLS            bool
	PostgresImage         string
	PostgresContainerName string
}
