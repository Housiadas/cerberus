package dbtest

import (
	"path/filepath"
	"runtime"
	"testing"

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
	cfg, err := config.LoadConfig(getConfigDir())
	if err != nil {
		t.Fatalf("[TEST]: error creating config %v", err)
	}

	return Config{
		DBUser:                cfg.DB.User,
		DBPassword:            cfg.DB.Password,
		DBName:                cfg.DB.Name,
		DBPort:                cfg.DB.Port,
		PostgresImage:         cfg.DB.PostgresImage,
		PostgresContainerName: cfg.DB.PostgresContainerName,
	}
}

func getConfigDir() string {
	_, file, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(file)
	return filepath.Join(basepath, "../../../")
}
