package config

// Redis stores the Redis configuration.
type Redis struct {
	Host               string `validate:"required"`
	Password           string
	DB                 int
	RedisImage         string
	RedisContainerName string
}
