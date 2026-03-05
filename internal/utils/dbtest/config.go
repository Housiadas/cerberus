package dbtest

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/config"
)

// Config holds test database configuration.
type Config struct {
	DBUser                string
	DBPassword            string
	DBName                string
	DBPort                string
	PostgresImage         string
	PostgresContainerName string
}

func newConfig() (Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	return Config{
		DBUser:                cfg.DB.User,
		DBPassword:            cfg.DB.Password,
		DBName:                cfg.DB.Name,
		DBPort:                cfg.DB.Port,
		PostgresImage:         cfg.DB.PostgresImage,
		PostgresContainerName: cfg.DB.PostgresContainerName,
	}, nil
}
