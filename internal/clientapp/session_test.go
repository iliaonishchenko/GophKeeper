package clientapp

import (
	"slices"
	"testing"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useTempConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	t.Setenv("HOME", t.TempDir())
}

func TestSaveLoadRoundTrip(t *testing.T) {
	useTempConfig(t)

	s := &Session{
		Login:         "alice",
		Token:         "tok",
		EncSalt:       []byte("salt"),
		ServerAddress: "localhost:3200",
		Cursor:        42,
		Items:         []*model.Item{{ID: "a", Name: "n", Version: 1}},
	}
	require.NoError(t, s.Save())

	got, err := LoadSession()
	require.NoError(t, err)
	assert.Equal(t, s.Login, got.Login)
	assert.Equal(t, s.Token, got.Token)
	assert.Equal(t, s.Cursor, got.Cursor)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "a", got.Items[0].ID)
}

func TestLoadSessionMissing(t *testing.T) {
	useTempConfig(t)
	got, err := LoadSession()
	require.NoError(t, err)
	assert.False(t, got.IsAuthenticated())
}

func TestMergeItemLWW(t *testing.T) {
	s := &Session{}

	s.mergeItem(&model.Item{ID: "a", Version: 1, Name: "старое"})
	assert.Len(t, s.Items, 1)

	s.mergeItem(&model.Item{ID: "a", Version: 2, Name: "новое"})
	require.Len(t, s.Items, 1)
	assert.Equal(t, "новое", s.Items[0].Name)

	s.mergeItem(&model.Item{ID: "a", Version: 1, Name: "устаревшее"})
	assert.Equal(t, "новое", s.Items[0].Name)

	s.mergeItem(&model.Item{ID: "b", Version: 1})
	assert.Len(t, s.Items, 2)
}

func TestActiveItemsAndFind(t *testing.T) {
	s := &Session{Items: []*model.Item{
		{ID: "a", Version: 1},
		{ID: "b", Version: 1, Deleted: true},
	}}

	active := slices.Collect(s.ActiveItems())
	require.Len(t, active, 1)
	assert.Equal(t, "a", active[0].ID)

	assert.NotNil(t, s.FindItem("b"))
	assert.Nil(t, s.FindItem("missing"))
}

func TestDeriveKeyFromSession(t *testing.T) {
	s := &Session{EncSalt: []byte("0123456789abcdef")}
	k1, err := s.DeriveKey("пароль")
	require.NoError(t, err)
	k2, err := s.DeriveKey("пароль")
	require.NoError(t, err)
	assert.Equal(t, k1, k2)
	assert.Len(t, k1, 32)
}
