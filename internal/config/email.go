package config

// Email holds SMTP configuration.
type Email struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      bool
}
