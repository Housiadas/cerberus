package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	App     App
	Version Version
	Rest    Rest
	DB      DB
	Tempo   Tempo
	Cors    CorsSettings
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigFile(filepath.Join(path, "config.yml"))
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
