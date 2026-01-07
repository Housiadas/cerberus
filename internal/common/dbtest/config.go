package dbtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/config"
)

type Config struct {
	DBUser                string
	DBPassword            string
	DBName                string
	DBPort                string
	PostgresImage         string
	PostgresContainerName string
}

func newConfig(t *testing.T) Config {
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	return Config{
		DBUser:                cfg.DB.User,
		DBPassword:            cfg.DB.Password,
		DBName:                cfg.DB.Name,
		DBPort:                cfg.DB.Port,
		PostgresImage:         cfg.DB.PostgresImage,
		PostgresContainerName: cfg.DB.PostgresContainerName,
	}
}
