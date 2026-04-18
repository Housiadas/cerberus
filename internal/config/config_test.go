package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() Config {
	return Config{
		App:           App{Name: "cerberus"},
		Rest:          Rest{API: ":8080"},
		DB:            DB{Host: "localhost", User: "user", Password: "pass", Name: "cerberus"},
		Redis:         Redis{Host: "localhost:6379"},
		Vault:         Vault{Address: "http://vault:8200", Token: "root"},
		Elasticsearch: Elasticsearch{Hosts: "localhost:9200", Password: "secret"},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	err := cfg.validate()
	require.NoError(t, err)
}

func TestValidate_MissingFields(t *testing.T) {
	tests := []struct {
		name       string
		modify     func(*Config)
		wantFields []string
	}{
		{"App.Name", func(c *Config) { c.App.Name = "" }, []string{"Name"}},
		{"Rest.API", func(c *Config) { c.Rest.API = "" }, []string{"API"}},
		{"DB.Host", func(c *Config) { c.DB.Host = "" }, []string{"Host"}},
		{"DB.User", func(c *Config) { c.DB.User = "" }, []string{"User"}},
		{"DB.Password", func(c *Config) { c.DB.Password = "" }, []string{"Password"}},
		{"DB.Name", func(c *Config) { c.DB.Name = "" }, []string{"Name"}},
		{"Redis.Host", func(c *Config) { c.Redis.Host = "" }, []string{"Host"}},
		{"Vault.Address", func(c *Config) { c.Vault.Address = "" }, []string{"Address"}},
		{"Vault.Token", func(c *Config) { c.Vault.Token = "" }, []string{"Token"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modify(&cfg)
			err := cfg.validate()
			require.Error(t, err)
			for _, field := range tt.wantFields {
				assert.Contains(t, err.Error(), field)
			}
		})
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := Config{}
	err := cfg.validate()
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "Name")
	assert.Contains(t, msg, "Host")
	assert.Contains(t, msg, "Token")
}
