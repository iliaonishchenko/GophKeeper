package clientapp

import (
	"context"
	"github.com/iliaonishchenko/gophkeeper/internal/crypto"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"iter"
)

type Upserter interface {
	Upsert(ctx context.Context, it *model.Item) (*model.Item, error)
}

type Syncer interface {
	Sync(ctx context.Context, since int64) ([]*model.Item, int64, error)
}

func (s *Session) DeriveKey(password string) ([]byte, error) {
	return crypto.DeriveKey(password, s.EncSalt)
}

func Add(ctx context.Context, c Upserter, s *Session, key []byte, it *model.Item, payload any) (*model.Item, error) {
	ciphertext, err := EncryptPayload(key, it.Type, payload)
	if err != nil {
		return nil, err
	}
	it.Ciphertext = ciphertext

	saved, err := c.Upsert(ctx, it)
	if err != nil {
		return nil, err
	}
	s.mergeItem(saved)
	if saved.UpdatedAt > s.Cursor {
		s.Cursor = saved.UpdatedAt
	}
	return saved, nil
}

func Sync(ctx context.Context, c Syncer, s *Session) (int, error) {
	items, cursor, err := c.Sync(ctx, s.Cursor)
	if err != nil {
		return 0, err
	}
	for _, it := range items {
		s.mergeItem(it)
	}
	s.Cursor = cursor
	return len(items), nil
}

func Decrypt(key []byte, it *model.Item) ([]byte, error) {
	return DecryptPayload(key, it.Ciphertext)
}

func (s *Session) ActiveItems() iter.Seq[*model.Item] {
	return func(yield func(*model.Item) bool) {
		for _, it := range s.Items {
			if it.Deleted {
				continue
			}
			if !yield(it) {
				return
			}
		}
	}
}

func (s *Session) FindItem(id string) *model.Item {
	for _, it := range s.Items {
		if it.ID == id {
			return it
		}
	}
	return nil
}
