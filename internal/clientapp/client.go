package clientapp

import (
	"context"
	"fmt"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn   *grpc.ClientConn
	keeper pb.KeeperClient
	token  string
}

func NewClient(target, token string) (*Client, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("создание gRPC-клиента: %w", err)
	}
	return &Client{conn: conn, keeper: pb.NewKeeperClient(conn), token: token}, nil
}

func NewClientWithConn(conn *grpc.ClientConn, token string) *Client {
	return &Client{conn: conn, keeper: pb.NewKeeperClient(conn), token: token}
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) authCtx(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}

func (c *Client) Register(ctx context.Context, login, password string) (string, []byte, error) {
	resp, err := c.keeper.Register(ctx, &pb.RegisterRequest{Login: login, Password: password})
	if err != nil {
		return "", nil, err
	}
	return resp.GetToken(), resp.GetEncSalt(), nil
}

func (c *Client) Login(ctx context.Context, login, password string) (string, []byte, error) {
	resp, err := c.keeper.Login(ctx, &pb.LoginRequest{Login: login, Password: password})
	if err != nil {
		return "", nil, err
	}
	return resp.GetToken(), resp.GetEncSalt(), nil
}

func (c *Client) Upsert(ctx context.Context, it *model.Item) (*model.Item, error) {
	resp, err := c.keeper.UpsertItem(c.authCtx(ctx), &pb.UpsertItemRequest{Item: itemToProto(it)})
	if err != nil {
		return nil, err
	}
	it.ID = resp.GetId()
	it.Version = resp.GetVersion()
	it.UpdatedAt = resp.GetUpdatedAt()
	return it, nil
}

func (c *Client) Get(ctx context.Context, id string) (*model.Item, error) {
	resp, err := c.keeper.GetItem(c.authCtx(ctx), &pb.GetItemRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToItem(resp.GetItem()), nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := c.keeper.DeleteItem(c.authCtx(ctx), &pb.DeleteItemRequest{Id: id})
	return err
}

func (c *Client) Sync(ctx context.Context, since int64) ([]*model.Item, int64, error) {
	resp, err := c.keeper.Sync(c.authCtx(ctx), &pb.SyncRequest{Since: since})
	if err != nil {
		return nil, since, err
	}
	items := make([]*model.Item, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, protoToItem(it))
	}
	return items, resp.GetCursor(), nil
}

func itemToProto(it *model.Item) *pb.Item {
	return &pb.Item{
		Id:         it.ID,
		Type:       pb.ItemType(it.Type),
		Name:       it.Name,
		Ciphertext: it.Ciphertext,
		Metadata:   it.Metadata,
		Version:    it.Version,
		UpdatedAt:  it.UpdatedAt,
		Deleted:    it.Deleted,
	}
}

func protoToItem(it *pb.Item) *model.Item {
	if it == nil {
		return &model.Item{}
	}
	return &model.Item{
		ID:         it.GetId(),
		Type:       model.ItemType(it.GetType()),
		Name:       it.GetName(),
		Ciphertext: it.GetCiphertext(),
		Metadata:   it.GetMetadata(),
		Version:    it.GetVersion(),
		UpdatedAt:  it.GetUpdatedAt(),
		Deleted:    it.GetDeleted(),
	}
}
