package config

import (
	"fmt"
	"path/filepath"
	"runtime"

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
func LoadConfig() (config Config, err error) {
	viper.SetConfigFile(filepath.Join(getConfigDir(), "config.yaml"))
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("viper unable to read config file: %w", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("viper unable to decode into struct: %w", err)
	}

	return config, nil
}

func getConfigDir() string {
	_, file, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(file)

	return filepath.Join(basepath, "../../")
}
