package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/iliaonishchenko/gophkeeper/internal/repository"
)

type VaultService struct {
	items ItemStore
}

func NewVaultService(items ItemStore) *VaultService {
	return &VaultService{items: items}
}

func (s *VaultService) Upsert(ctx context.Context, userID string, it *model.Item) (*model.Item, error) {
	if it.ID == "" {
		id, err := model.NewID()
		if err != nil {
			return nil, err
		}
		it.ID = id
	}
	it.UserID = userID
	it.UpdatedAt = time.Now().UnixMilli()

	saved, err := s.items.UpsertItem(ctx, it)
	if err != nil {
		return nil, fmt.Errorf("сохранение элемента: %w", err)
	}
	return saved, nil
}

func (s *VaultService) Get(ctx context.Context, userID, id string) (*model.Item, error) {
	it, err := s.items.GetItem(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("получение элемента: %w", err)
	}
	return it, nil
}

func (s *VaultService) Delete(ctx context.Context, userID, id string) (*model.Item, error) {
	existing, err := s.items.GetItem(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("получение элемента: %w", err)
	}

	existing.Deleted = true
	existing.UpdatedAt = time.Now().UnixMilli()
	saved, err := s.items.UpsertItem(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("удаление элемента: %w", err)
	}
	return saved, nil
}

func (s *VaultService) Sync(ctx context.Context, userID string, since int64) ([]*model.Item, int64, error) {
	items, cursor, err := s.items.ChangedSince(ctx, userID, since)
	if err != nil {
		return nil, since, fmt.Errorf("синхронизация элементов: %w", err)
	}
	return items, cursor, nil
}
