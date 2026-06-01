package main

import (
	"flag"

	"github.com/iliaonishchenko/gophkeeper/internal/config/server"
)

func parseFlags(cfg *server.Config) {
	var grpcAddress string
	var databaseDSN string
	var tokenSecret string
	var logLevel string

	flag.StringVar(&grpcAddress, "g", "", "адрес и порт gRPC-сервера")
	flag.StringVar(&databaseDSN, "d", "", "строка подключения к PostgreSQL")
	flag.StringVar(&tokenSecret, "s", "", "секретный ключ для подписи токенов")
	flag.StringVar(&logLevel, "l", "", "уровень логирования (debug, info, warn, error)")
	flag.Parse()

	const defaultGRPCAddress = "localhost:3200"

	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = grpcAddress
	}
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = defaultGRPCAddress
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = databaseDSN
	}
	if cfg.TokenSecret == "" {
		cfg.TokenSecret = tokenSecret
	}
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}
}
