package grpcserver

import (
	"context"
	"errors"

	"github.com/iliaonishchenko/gophkeeper/internal/logger"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
	"github.com/iliaonishchenko/gophkeeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (*service.AuthResult, error)
	Login(ctx context.Context, login, password string) (*service.AuthResult, error)
}

type VaultService interface {
	Upsert(ctx context.Context, userID string, it *model.Item) (*model.Item, error)
	Get(ctx context.Context, userID, id string) (*model.Item, error)
	Delete(ctx context.Context, userID, id string) (*model.Item, error)
	Sync(ctx context.Context, userID string, since int64) ([]*model.Item, int64, error)
}

type KeeperServer struct {
	pb.UnimplementedKeeperServer
	auth  AuthService
	vault VaultService
}

func NewKeeperServer(auth AuthService, vault VaultService) *KeeperServer {
	return &KeeperServer{auth: auth, vault: vault}
}

func (s *KeeperServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	res, err := s.auth.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, authError(err)
	}
	return &pb.RegisterResponse{Token: res.Token, EncSalt: res.EncSalt}, nil
}

func (s *KeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	res, err := s.auth.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, authError(err)
	}
	return &pb.LoginResponse{Token: res.Token, EncSalt: res.EncSalt}, nil
}

func (s *KeeperServer) UpsertItem(ctx context.Context, req *pb.UpsertItemRequest) (*pb.UpsertItemResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "пользователь не определён")
	}

	saved, err := s.vault.Upsert(ctx, userID, protoToModel(req.GetItem()))
	if err != nil {
		logger.Log.Error("ошибка сохранения элемента", logger.Err(err))
		return nil, status.Error(codes.Internal, "не удалось сохранить элемент")
	}
	return &pb.UpsertItemResponse{Id: saved.ID, Version: saved.Version, UpdatedAt: saved.UpdatedAt}, nil
}

func (s *KeeperServer) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "пользователь не определён")
	}

	it, err := s.vault.Get(ctx, userID, req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "элемент не найден")
		}
		logger.Log.Error("ошибка получения элемента", logger.Err(err))
		return nil, status.Error(codes.Internal, "не удалось получить элемент")
	}
	return &pb.GetItemResponse{Item: modelToProto(it)}, nil
}

func (s *KeeperServer) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "пользователь не определён")
	}

	saved, err := s.vault.Delete(ctx, userID, req.GetId())
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "элемент не найден")
		}
		logger.Log.Error("ошибка удаления элемента", logger.Err(err))
		return nil, status.Error(codes.Internal, "не удалось удалить элемент")
	}
	return &pb.DeleteItemResponse{Version: saved.Version, UpdatedAt: saved.UpdatedAt}, nil
}

func (s *KeeperServer) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "пользователь не определён")
	}

	items, cursor, err := s.vault.Sync(ctx, userID, req.GetSince())
	if err != nil {
		logger.Log.Error("ошибка синхронизации", logger.Err(err))
		return nil, status.Error(codes.Internal, "не удалось синхронизировать данные")
	}

	out := make([]*pb.Item, 0, len(items))
	for _, it := range items {
		out = append(out, modelToProto(it))
	}
	return &pb.SyncResponse{Items: out, Cursor: cursor}, nil
}

func authError(err error) error {
	switch {
	case errors.Is(err, service.ErrUserExists):
		return status.Error(codes.AlreadyExists, "пользователь уже существует")
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "неверный логин или пароль")
	case errors.Is(err, service.ErrEmptyCredentials):
		return status.Error(codes.InvalidArgument, "логин и пароль обязательны")
	default:
		logger.Log.Error("ошибка аутентификации", logger.Err(err))
		return status.Error(codes.Internal, "внутренняя ошибка")
	}
}
