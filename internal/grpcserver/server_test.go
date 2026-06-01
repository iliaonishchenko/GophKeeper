package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
	"github.com/iliaonishchenko/gophkeeper/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeAuth struct {
	res *service.AuthResult
	err error
}

func (f *fakeAuth) Register(context.Context, string, string) (*service.AuthResult, error) {
	return f.res, f.err
}
func (f *fakeAuth) Login(context.Context, string, string) (*service.AuthResult, error) {
	return f.res, f.err
}

type fakeVault struct {
	item   *model.Item
	items  []*model.Item
	cursor int64
	err    error
}

func (f *fakeVault) Upsert(_ context.Context, _ string, it *model.Item) (*model.Item, error) {
	if f.err != nil {
		return nil, f.err
	}
	return it, nil
}
func (f *fakeVault) Get(context.Context, string, string) (*model.Item, error) {
	return f.item, f.err
}
func (f *fakeVault) Delete(context.Context, string, string) (*model.Item, error) {
	return f.item, f.err
}
func (f *fakeVault) Sync(context.Context, string, int64) ([]*model.Item, int64, error) {
	return f.items, f.cursor, f.err
}

func authCtx() context.Context {
	return context.WithValue(context.Background(), userIDKey{}, "user-1")
}

func TestRegisterHandler(t *testing.T) {
	srv := NewKeeperServer(&fakeAuth{res: &service.AuthResult{Token: "tok", EncSalt: []byte("salt")}}, nil)
	resp, err := srv.Register(context.Background(), &pb.RegisterRequest{Login: "a", Password: "b"})
	require.NoError(t, err)
	assert.Equal(t, "tok", resp.GetToken())
}

func TestRegisterHandlerDuplicate(t *testing.T) {
	srv := NewKeeperServer(&fakeAuth{err: service.ErrUserExists}, nil)
	_, err := srv.Register(context.Background(), &pb.RegisterRequest{Login: "a", Password: "b"})
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestLoginHandlerInvalid(t *testing.T) {
	srv := NewKeeperServer(&fakeAuth{err: service.ErrInvalidCredentials}, nil)
	_, err := srv.Login(context.Background(), &pb.LoginRequest{Login: "a", Password: "b"})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUpsertHandlerUnauthenticated(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{})
	_, err := srv.UpsertItem(context.Background(), &pb.UpsertItemRequest{Item: &pb.Item{}})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUpsertHandlerSuccess(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{})
	resp, err := srv.UpsertItem(authCtx(), &pb.UpsertItemRequest{
		Item: &pb.Item{Id: "item-1", Type: pb.ItemType_ITEM_TYPE_TEXT, Name: "n"},
	})
	require.NoError(t, err)
	assert.Equal(t, "item-1", resp.GetId())
}

func TestGetHandlerNotFound(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{err: service.ErrItemNotFound})
	_, err := srv.GetItem(authCtx(), &pb.GetItemRequest{Id: "x"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetHandlerSuccess(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{item: &model.Item{ID: "item-1", Name: "n", Type: model.TypeCard}})
	resp, err := srv.GetItem(authCtx(), &pb.GetItemRequest{Id: "item-1"})
	require.NoError(t, err)
	assert.Equal(t, "item-1", resp.GetItem().GetId())
	assert.Equal(t, pb.ItemType_ITEM_TYPE_CARD, resp.GetItem().GetType())
}

func TestDeleteHandlerSuccess(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{item: &model.Item{Version: 2, UpdatedAt: 5}})
	resp, err := srv.DeleteItem(authCtx(), &pb.DeleteItemRequest{Id: "item-1"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), resp.GetVersion())
}

func TestDeleteHandlerError(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{err: errors.New("сбой")})
	_, err := srv.DeleteItem(authCtx(), &pb.DeleteItemRequest{Id: "item-1"})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestSyncHandlerSuccess(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{
		items:  []*model.Item{{ID: "a", UpdatedAt: 10}, {ID: "b", UpdatedAt: 20}},
		cursor: 20,
	})
	resp, err := srv.Sync(authCtx(), &pb.SyncRequest{Since: 0})
	require.NoError(t, err)
	assert.Len(t, resp.GetItems(), 2)
	assert.Equal(t, int64(20), resp.GetCursor())
}

func TestSyncHandlerUnauthenticated(t *testing.T) {
	srv := NewKeeperServer(nil, &fakeVault{})
	_, err := srv.Sync(context.Background(), &pb.SyncRequest{})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}
