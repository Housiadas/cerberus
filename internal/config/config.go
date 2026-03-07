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
	App       App
	Rest      Rest
	DB        DB
	Redis     Redis
	Kafka     Kafka
	Vault     Vault
	Collector Collector
	Cors      CorsSettings
	Email     Email
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (Config, error) {
	var (
		config Config
		err    error
	)

	v := viper.New()
	v.SetConfigFile(filepath.Join(getConfigDir(), "config.yaml"))
	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("viper unable to read config file: %w", err)
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("viper unable to decode into struct: %w", err)
	}

	return config, nil
}

func getConfigDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}

	basepath := filepath.Dir(file)

	return filepath.Join(basepath, "../../")
}
