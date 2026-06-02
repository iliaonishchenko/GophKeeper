package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/iliaonishchenko/gophkeeper/internal/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const secret = "interceptor-secret"

func invoke(t *testing.T, ctx context.Context, method string) (any, error) {
	t.Helper()
	interceptor := AuthInterceptor(secret)
	info := &grpc.UnaryServerInfo{FullMethod: method}
	handler := func(ctx context.Context, _ any) (any, error) {
		uid, _ := userIDFromContext(ctx)
		return uid, nil
	}
	return interceptor(ctx, nil, info, handler)
}

func TestInterceptorSkipsPublic(t *testing.T) {
	_, err := invoke(t, context.Background(), "/keeper.Keeper/Register")
	assert.NoError(t, err)
	_, err = invoke(t, context.Background(), "/keeper.Keeper/Login")
	assert.NoError(t, err)
}

func TestInterceptorNoMetadata(t *testing.T) {
	_, err := invoke(t, context.Background(), "/keeper.Keeper/Sync")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestInterceptorNoToken(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	_, err := invoke(t, ctx, "/keeper.Keeper/Sync")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestInterceptorBadToken(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer invalid"))
	_, err := invoke(t, ctx, "/keeper.Keeper/Sync")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestInterceptorValidToken(t *testing.T) {
	tok := token.Issue(secret, "user-99", time.Hour)
	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer "+tok))
	uid, err := invoke(t, ctx, "/keeper.Keeper/Sync")
	require.NoError(t, err)
	assert.Equal(t, "user-99", uid)
}
