package config

// Elasticsearch holds settings for the Elasticsearch connection.
type Elasticsearch struct {
	Hosts    string `validate:"required"`
	Username string
	Password string `validate:"required"`
}
