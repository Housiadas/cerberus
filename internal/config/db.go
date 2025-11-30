package config

type DB struct {
	User                  string
	Password              string
	Name                  string
	Port                  string
	Host                  string
	MaxOpenConns          int
	MaxIdleConns          int
	ConnectionIdleTime    string
	DisableTLS            bool
	PostgresImage         string
	PostgresContainerName string
}
