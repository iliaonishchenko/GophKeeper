package server

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	GRPCAddress string        `env:"GRPC_ADDRESS"`
	DatabaseDSN string        `env:"DATABASE_DSN"`
	TokenSecret string        `env:"TOKEN_SECRET"`
	TokenTTL    time.Duration `env:"TOKEN_TTL" envDefault:"24h"`
	LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
