package clientapp

import (
	"context"
	"errors"
	"testing"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRemote реализует Upserter и Syncer для тестов.
type fakeRemote struct {
	upserted *model.Item
	items    []*model.Item
	cursor   int64
	err      error
}

func (f *fakeRemote) Upsert(_ context.Context, it *model.Item) (*model.Item, error) {
	if f.err != nil {
		return nil, f.err
	}
	it.ID = "server-id"
	it.Version = 1
	it.UpdatedAt = 100
	f.upserted = it
	return it, nil
}

func (f *fakeRemote) Sync(_ context.Context, _ int64) ([]*model.Item, int64, error) {
	return f.items, f.cursor, f.err
}

func TestAddEncryptsAndCaches(t *testing.T) {
	key := testKey(t)
	s := &Session{}
	remote := &fakeRemote{}

	it := &model.Item{Type: model.TypeText, Name: "заметка"}
	saved, err := Add(context.Background(), remote, s, key, it, []byte("секрет"))
	require.NoError(t, err)

	assert.Equal(t, "server-id", saved.ID)
	assert.NotEmpty(t, remote.upserted.Ciphertext)
	assert.NotEqual(t, []byte("секрет"), remote.upserted.Ciphertext)
	require.Len(t, s.Items, 1)
	assert.Equal(t, int64(100), s.Cursor)

	plain, err := Decrypt(key, s.Items[0])
	require.NoError(t, err)
	assert.Equal(t, "секрет", string(plain))
}

func TestAddEncryptError(t *testing.T) {
	key := testKey(t)
	s := &Session{}
	_, err := Add(context.Background(), &fakeRemote{}, s, key, &model.Item{Type: model.TypeUnspecified}, []byte("x"))
	assert.Error(t, err)
}

func TestAddRemoteError(t *testing.T) {
	key := testKey(t)
	s := &Session{}
	_, err := Add(context.Background(), &fakeRemote{err: errors.New("сбой")}, s, key,
		&model.Item{Type: model.TypeText}, []byte("x"))
	assert.Error(t, err)
}

func TestSyncMergesAndAdvancesCursor(t *testing.T) {
	s := &Session{Cursor: 0, Items: []*model.Item{{ID: "a", Version: 1, Name: "старое"}}}
	remote := &fakeRemote{
		items:  []*model.Item{{ID: "a", Version: 2, Name: "новое"}, {ID: "b", Version: 1}},
		cursor: 50,
	}

	n, err := Sync(context.Background(), remote, s)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, int64(50), s.Cursor)
	assert.Equal(t, "новое", s.FindItem("a").Name)
	assert.NotNil(t, s.FindItem("b"))
}

func TestSyncError(t *testing.T) {
	s := &Session{}
	_, err := Sync(context.Background(), &fakeRemote{err: errors.New("сбой")}, s)
	assert.Error(t, err)
}
