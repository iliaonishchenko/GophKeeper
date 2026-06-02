package main

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/iliaonishchenko/gophkeeper"
	"github.com/iliaonishchenko/gophkeeper/internal/config/server"
	"github.com/iliaonishchenko/gophkeeper/internal/grpcserver"
	"github.com/iliaonishchenko/gophkeeper/internal/logger"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
	"github.com/iliaonishchenko/gophkeeper/internal/repository"
	"github.com/iliaonishchenko/gophkeeper/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Println("Build version: " + cmp.Or(buildVersion, "N/A"))
	fmt.Println("Build date: " + cmp.Or(buildDate, "N/A"))
	fmt.Println("Build commit: " + cmp.Or(buildCommit, "N/A"))

	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("ошибка загрузки конфигурации: %v", err)
	}
	parseFlags(cfg)

	if cfg.DatabaseDSN == "" {
		log.Fatal("не задана строка подключения к базе данных (DATABASE_DSN или -d)")
	}
	if cfg.TokenSecret == "" {
		log.Fatal("не задан секрет для подписи токенов (TOKEN_SECRET или -s)")
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("ошибка инициализации логера: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := run(ctx, cfg); err != nil {
		logger.Log.Fatal("ошибка работы сервера", logger.Err(err))
	}
}

func run(ctx context.Context, cfg *server.Config) error {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("подключение к базе данных: %w", err)
	}
	defer db.Close()

	if err := gophkeeper.RunMigrations(db); err != nil {
		return fmt.Errorf("выполнение миграций: %w", err)
	}

	classifier := repository.NewPostgresErrorClassifier()
	userRepo := repository.NewUsersRepository(db, classifier)
	itemRepo := repository.NewItemsRepository(db, classifier)

	authSvc := service.NewAuthService(userRepo, cfg.TokenSecret, cfg.TokenTTL)
	vaultSvc := service.NewVaultService(itemRepo)

	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return fmt.Errorf("прослушивание gRPC-адреса %q: %w", cfg.GRPCAddress, err)
	}

	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.AuthInterceptor(cfg.TokenSecret)))
	pb.RegisterKeeperServer(grpcSrv, grpcserver.NewKeeperServer(authSvc, vaultSvc))

	errCh := make(chan error, 1)
	go func() {
		logger.Log.Info("запуск gRPC-сервера", zap.String("address", cfg.GRPCAddress))
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("остановка сервера...")
		grpcSrv.GracefulStop()
		return nil
	case err := <-errCh:
		grpcSrv.GracefulStop()
		return err
	}
}
