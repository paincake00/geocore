package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPPort      string `envconfig:"HTTP_PORT" default:"8080"`
	DatabaseURL   string `envconfig:"DATABASE_URL" required:"true"`
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	MockServerURL string `envconfig:"MOCK_SERVER_URL" default:"http://localhost:9090"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process env vars: %w", err)
	}
	return &cfg, nil
}
