package config

// Redis stores the Redis configuration.
type Redis struct {
	Host               string
	Password           string
	DB                 int
	TTL                string
	RedisImage         string
	RedisContainerName string
}
