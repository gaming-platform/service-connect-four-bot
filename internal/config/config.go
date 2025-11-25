package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username    string `env:"APP_USERNAME,required"`
	Level       int    `env:"APP_LEVEL,required"`
	JoinAfter   int    `env:"APP_JOIN_AFTER,required"`
	RabbitMqDsn string `env:"APP_RABBIT_MQ_DSN,required"`
	NchanSubUrl string `env:"APP_NCHAN_SUB_URL,required"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %w", err)
	}

	return cfg, nil
}
