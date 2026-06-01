package clientapp

import (
	"context"
	"errors"
	"testing"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakeKeeperClient struct {
	lastCtx context.Context
	err     error
}

func (f *fakeKeeperClient) Register(_ context.Context, _ *pb.RegisterRequest, _ ...grpc.CallOption) (*pb.RegisterResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &pb.RegisterResponse{Token: "tok", EncSalt: []byte("salt")}, nil
}

func (f *fakeKeeperClient) Login(_ context.Context, _ *pb.LoginRequest, _ ...grpc.CallOption) (*pb.LoginResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &pb.LoginResponse{Token: "tok", EncSalt: []byte("salt")}, nil
}

func (f *fakeKeeperClient) UpsertItem(ctx context.Context, _ *pb.UpsertItemRequest, _ ...grpc.CallOption) (*pb.UpsertItemResponse, error) {
	f.lastCtx = ctx
	if f.err != nil {
		return nil, f.err
	}
	return &pb.UpsertItemResponse{Id: "srv-id", Version: 3, UpdatedAt: 99}, nil
}

func (f *fakeKeeperClient) GetItem(ctx context.Context, _ *pb.GetItemRequest, _ ...grpc.CallOption) (*pb.GetItemResponse, error) {
	f.lastCtx = ctx
	if f.err != nil {
		return nil, f.err
	}
	return &pb.GetItemResponse{Item: &pb.Item{Id: "srv-id", Name: "n", Type: pb.ItemType_ITEM_TYPE_CARD}}, nil
}

func (f *fakeKeeperClient) DeleteItem(ctx context.Context, _ *pb.DeleteItemRequest, _ ...grpc.CallOption) (*pb.DeleteItemResponse, error) {
	f.lastCtx = ctx
	if f.err != nil {
		return nil, f.err
	}
	return &pb.DeleteItemResponse{Version: 2, UpdatedAt: 5}, nil
}

func (f *fakeKeeperClient) Sync(ctx context.Context, _ *pb.SyncRequest, _ ...grpc.CallOption) (*pb.SyncResponse, error) {
	f.lastCtx = ctx
	if f.err != nil {
		return nil, f.err
	}
	return &pb.SyncResponse{
		Items:  []*pb.Item{{Id: "a", UpdatedAt: 10}, {Id: "b", UpdatedAt: 20}},
		Cursor: 20,
	}, nil
}

func newTestClient(fake pb.KeeperClient, token string) *Client {
	return &Client{keeper: fake, token: token}
}

func TestClientRegisterLogin(t *testing.T) {
	c := newTestClient(&fakeKeeperClient{}, "")

	tok, salt, err := c.Register(context.Background(), "a", "b")
	require.NoError(t, err)
	assert.Equal(t, "tok", tok)
	assert.Equal(t, []byte("salt"), salt)

	tok, salt, err = c.Login(context.Background(), "a", "b")
	require.NoError(t, err)
	assert.Equal(t, "tok", tok)
	assert.Equal(t, []byte("salt"), salt)
}

func TestClientRegisterError(t *testing.T) {
	c := newTestClient(&fakeKeeperClient{err: errors.New("сбой")}, "")
	_, _, err := c.Register(context.Background(), "a", "b")
	assert.Error(t, err)
	_, _, err = c.Login(context.Background(), "a", "b")
	assert.Error(t, err)
}

func TestClientUpsert(t *testing.T) {
	fake := &fakeKeeperClient{}
	c := newTestClient(fake, "my-token")

	it := &model.Item{Type: model.TypeText, Name: "n", Ciphertext: []byte("ct")}
	saved, err := c.Upsert(context.Background(), it)
	require.NoError(t, err)
	assert.Equal(t, "srv-id", saved.ID)
	assert.Equal(t, int64(3), saved.Version)

	md, ok := metadata.FromOutgoingContext(fake.lastCtx)
	require.True(t, ok)
	assert.Equal(t, []string{"Bearer my-token"}, md.Get("authorization"))
}

func TestClientGet(t *testing.T) {
	c := newTestClient(&fakeKeeperClient{}, "tok")
	it, err := c.Get(context.Background(), "srv-id")
	require.NoError(t, err)
	assert.Equal(t, "srv-id", it.ID)
	assert.Equal(t, model.TypeCard, it.Type)
}

func TestClientDelete(t *testing.T) {
	c := newTestClient(&fakeKeeperClient{}, "tok")
	assert.NoError(t, c.Delete(context.Background(), "srv-id"))

	c = newTestClient(&fakeKeeperClient{err: errors.New("сбой")}, "tok")
	assert.Error(t, c.Delete(context.Background(), "srv-id"))
}

func TestClientSync(t *testing.T) {
	c := newTestClient(&fakeKeeperClient{}, "tok")
	items, cursor, err := c.Sync(context.Background(), 0)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(20), cursor)
}

func TestClientSyncError(t *testing.T) {
	c := newTestClient(&fakeKeeperClient{err: errors.New("сбой")}, "tok")
	_, _, err := c.Sync(context.Background(), 0)
	assert.Error(t, err)
}

func TestNewClientAndClose(t *testing.T) {
	c, err := NewClient("localhost:3200", "tok")
	require.NoError(t, err)
	assert.NoError(t, c.Close())

	empty := &Client{}
	ctx := empty.authCtx(context.Background())
	_, ok := metadata.FromOutgoingContext(ctx)
	assert.False(t, ok)
}

func TestProtoConversionNil(t *testing.T) {
	assert.Equal(t, &model.Item{}, protoToItem(nil))
}
