package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/iliaonishchenko/gophkeeper/internal/repository"
	"github.com/iliaonishchenko/gophkeeper/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertCreatesID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().UpsertItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, it *model.Item) (*model.Item, error) {
			assert.NotEmpty(t, it.ID)
			assert.Equal(t, "user-1", it.UserID)
			assert.NotZero(t, it.UpdatedAt)
			it.Version = 1
			return it, nil
		})

	svc := NewVaultService(store)
	saved, err := svc.Upsert(context.Background(), "user-1", &model.Item{
		Type: model.TypeText,
		Name: "заметка",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), saved.Version)
}

func TestUpsertKeepsID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().UpsertItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, it *model.Item) (*model.Item, error) {
			assert.Equal(t, "item-7", it.ID)
			return it, nil
		})

	svc := NewVaultService(store)
	_, err := svc.Upsert(context.Background(), "user-1", &model.Item{ID: "item-7"})
	require.NoError(t, err)
}

func TestUpsertStoreError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().UpsertItem(gomock.Any(), gomock.Any()).Return(nil, errors.New("сбой"))

	svc := NewVaultService(store)
	_, err := svc.Upsert(context.Background(), "user-1", &model.Item{ID: "x"})
	assert.Error(t, err)
}

func TestGetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().GetItem(gomock.Any(), "user-1", "item-1").Return(&model.Item{ID: "item-1"}, nil)

	svc := NewVaultService(store)
	it, err := svc.Get(context.Background(), "user-1", "item-1")
	require.NoError(t, err)
	assert.Equal(t, "item-1", it.ID)
}

func TestGetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().GetItem(gomock.Any(), "user-1", "missing").Return(nil, repository.ErrItemNotFound)

	svc := NewVaultService(store)
	_, err := svc.Get(context.Background(), "user-1", "missing")
	assert.ErrorIs(t, err, ErrItemNotFound)
}

func TestDeleteMarksTombstone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().GetItem(gomock.Any(), "user-1", "item-1").Return(&model.Item{ID: "item-1"}, nil)
	store.EXPECT().UpsertItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, it *model.Item) (*model.Item, error) {
			assert.True(t, it.Deleted)
			return it, nil
		})

	svc := NewVaultService(store)
	res, err := svc.Delete(context.Background(), "user-1", "item-1")
	require.NoError(t, err)
	assert.True(t, res.Deleted)
}

func TestDeleteNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().GetItem(gomock.Any(), "user-1", "missing").Return(nil, repository.ErrItemNotFound)

	svc := NewVaultService(store)
	_, err := svc.Delete(context.Background(), "user-1", "missing")
	assert.ErrorIs(t, err, ErrItemNotFound)
}

func TestSyncSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	want := []*model.Item{{ID: "a", UpdatedAt: 10}, {ID: "b", UpdatedAt: 20}}
	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().ChangedSince(gomock.Any(), "user-1", int64(5)).Return(want, int64(20), nil)

	svc := NewVaultService(store)
	items, cursor, err := svc.Sync(context.Background(), "user-1", 5)
	require.NoError(t, err)
	assert.Equal(t, want, items)
	assert.Equal(t, int64(20), cursor)
}

func TestSyncError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockItemStore(ctrl)
	store.EXPECT().ChangedSince(gomock.Any(), "user-1", int64(0)).Return(nil, int64(0), errors.New("сбой"))

	svc := NewVaultService(store)
	_, _, err := svc.Sync(context.Background(), "user-1", 0)
	assert.Error(t, err)
}
