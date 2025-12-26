package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig               `yaml:"server"`
	Database  DatabaseConfig             `yaml:"database"`
	Redis     RedisConfig                `yaml:"redis"`
	Proxy     ProxyConfig                `yaml:"proxy"`
	Providers map[string]ProviderConfig  `yaml:"providers"`
}

type ProviderConfig struct {
	Enabled            bool              `yaml:"enabled"`
	AuthStrategy       string            `yaml:"auth_strategy"`
	BaseURL            string            `yaml:"base_url"`
	OAuthClientID      string            `yaml:"oauth_client_id"`
	OAuthClientSecret  string            `yaml:"oauth_client_secret"`
	BaseURLs           []string          `yaml:"base_urls"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type ProxyConfig struct {
	SelectionStrategy   string `yaml:"selection_strategy"`
	HealthCheckInterval int    `yaml:"health_check_interval"`
	MaxFailures         int    `yaml:"max_failures"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
