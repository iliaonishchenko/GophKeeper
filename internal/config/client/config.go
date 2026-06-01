package client

import "github.com/caarlos0/env/v11"

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
