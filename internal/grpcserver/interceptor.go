package grpcserver

import (
	"context"
	"strings"

	"github.com/iliaonishchenko/gophkeeper/internal/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authMetadataKey = "authorization"

const bearerPrefix = "Bearer "

type userIDKey struct{}

var publicMethods = map[string]bool{
	"/keeper.Keeper/Register": true,
	"/keeper.Keeper/Login":    true,
}

func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "метаданные отсутствуют")
		}
		vals := md.Get(authMetadataKey)
		if len(vals) == 0 || vals[0] == "" {
			return nil, status.Error(codes.Unauthenticated, "токен отсутствует")
		}

		userID, err := token.Verify(secret, strings.TrimPrefix(vals[0], bearerPrefix))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "неверный токен")
		}

		return handler(context.WithValue(ctx, userIDKey{}, userID), req)
	}
}

func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok && userID != ""
}
