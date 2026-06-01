package integration

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/iliaonishchenko/gophkeeper/internal/clientapp"
	"github.com/iliaonishchenko/gophkeeper/internal/grpcserver"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
	"github.com/iliaonishchenko/gophkeeper/internal/repository"
	"github.com/iliaonishchenko/gophkeeper/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type pgUniqueErr struct{}

func (pgUniqueErr) Error() string    { return "duplicate key" }
func (pgUniqueErr) SQLState() string { return "23505" }

type memUserStore struct {
	mu    sync.Mutex
	users map[string]*model.User
}

func (s *memUserStore) CreateUser(_ context.Context, u *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.Login]; ok {
		return pgUniqueErr{}
	}
	s.users[u.Login] = u
	return nil
}

func (s *memUserStore) GetUserByLogin(_ context.Context, login string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[login]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

type memItemStore struct {
	mu    sync.Mutex
	items map[string]*model.Item
	clock int64
}

func (s *memItemStore) UpsertItem(_ context.Context, it *model.Item) (*model.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock++
	it.UpdatedAt = s.clock
	if prev, ok := s.items[it.ID]; ok {
		it.Version = prev.Version + 1
	} else {
		it.Version = 1
	}
	cp := *it
	s.items[it.ID] = &cp
	return it, nil
}

func (s *memItemStore) GetItem(_ context.Context, userID, id string) (*model.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[id]
	if !ok || it.UserID != userID {
		return nil, repository.ErrItemNotFound
	}
	cp := *it
	return &cp, nil
}

func (s *memItemStore) ChangedSince(_ context.Context, userID string, since int64) ([]*model.Item, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cursor := since
	var out []*model.Item
	for _, it := range s.items {
		if it.UserID == userID && it.UpdatedAt > since {
			cp := *it
			out = append(out, &cp)
			if it.UpdatedAt > cursor {
				cursor = it.UpdatedAt
			}
		}
	}
	return out, cursor, nil
}

func startServer(t *testing.T) (*bufconn.Listener, *memItemStore) {
	t.Helper()
	const secret = "integration-secret"

	users := &memUserStore{users: map[string]*model.User{}}
	items := &memItemStore{items: map[string]*model.Item{}}
	authSvc := service.NewAuthService(users, secret, time.Hour)
	vaultSvc := service.NewVaultService(items)

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.AuthInterceptor(secret)))
	pb.RegisterKeeperServer(srv, grpcserver.NewKeeperServer(authSvc, vaultSvc))

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis, items
}

func dial(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestEndToEndZeroKnowledge(t *testing.T) {
	lis, items := startServer(t)
	ctx := context.Background()

	c1 := clientapp.NewClientWithConn(dial(t, lis), "")
	tok, encSalt, err := c1.Register(ctx, "alice", "master-pass")
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	s1 := &clientapp.Session{Login: "alice", Token: tok, EncSalt: encSalt}
	authed1 := clientapp.NewClientWithConn(dial(t, lis), tok)
	key1, err := s1.DeriveKey("master-pass")
	require.NoError(t, err)

	payload := clientapp.CredentialsPayload{Login: "alice@site", Password: "p@ss"}
	it := &model.Item{Type: model.TypeCredentials, Name: "сайт", Metadata: "example.com"}
	saved, err := clientapp.Add(ctx, authed1, s1, key1, it, payload)
	require.NoError(t, err)
	require.NotEmpty(t, saved.ID)

	stored := items.items[saved.ID]
	require.NotNil(t, stored)
	assert.NotContains(t, string(stored.Ciphertext), "p@ss")
	assert.NotContains(t, string(stored.Ciphertext), "alice@site")

	c2 := clientapp.NewClientWithConn(dial(t, lis), "")
	tok2, encSalt2, err := c2.Login(ctx, "alice", "master-pass")
	require.NoError(t, err)
	assert.Equal(t, encSalt, encSalt2)

	s2 := &clientapp.Session{Login: "alice", Token: tok2, EncSalt: encSalt2}
	authed2 := clientapp.NewClientWithConn(dial(t, lis), tok2)
	n, err := clientapp.Sync(ctx, authed2, s2)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	got := s2.FindItem(saved.ID)
	require.NotNil(t, got)
	key2, err := s2.DeriveKey("master-pass")
	require.NoError(t, err)
	plain, err := clientapp.Decrypt(key2, got)
	require.NoError(t, err)

	var decoded clientapp.CredentialsPayload
	require.NoError(t, json.Unmarshal(plain, &decoded))
	assert.Equal(t, payload, decoded)
}

func TestEndToEndAuthRequired(t *testing.T) {
	lis, _ := startServer(t)
	anon := clientapp.NewClientWithConn(dial(t, lis), "")
	_, _, err := anon.Sync(context.Background(), 0)
	assert.Error(t, err)
}

func TestEndToEndWrongPassword(t *testing.T) {
	lis, _ := startServer(t)
	c := clientapp.NewClientWithConn(dial(t, lis), "")
	_, _, err := c.Register(context.Background(), "bob", "right")
	require.NoError(t, err)

	_, _, err = c.Login(context.Background(), "bob", "wrong")
	assert.Error(t, err)
}
