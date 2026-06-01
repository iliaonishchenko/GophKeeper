package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophkeeper/internal/crypto"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/iliaonishchenko/gophkeeper/internal/repository"
	"github.com/iliaonishchenko/gophkeeper/internal/service/mocks"
	"github.com/iliaonishchenko/gophkeeper/internal/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

func TestRegisterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockUserStore(ctrl)
	store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)

	svc := NewAuthService(store, testSecret, time.Hour)
	res, err := svc.Register(context.Background(), "alice", "pass123")
	require.NoError(t, err)
	assert.NotEmpty(t, res.Token)
	assert.Len(t, res.EncSalt, 16)

	uid, err := token.Verify(testSecret, res.Token)
	require.NoError(t, err)
	assert.NotEmpty(t, uid)
}

func TestRegisterEmptyCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewAuthService(mocks.NewMockUserStore(ctrl), testSecret, time.Hour)
	_, err := svc.Register(context.Background(), "", "pass")
	assert.ErrorIs(t, err, ErrEmptyCredentials)
}

type uniqueErr struct{}

func (uniqueErr) Error() string    { return "duplicate" }
func (uniqueErr) SQLState() string { return "23505" }

func TestRegisterDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockUserStore(ctrl)
	store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(uniqueErr{})

	svc := NewAuthService(store, testSecret, time.Hour)
	_, err := svc.Register(context.Background(), "alice", "pass")
	assert.ErrorIs(t, err, ErrUserExists)
}

func TestLoginSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	salt, _ := crypto.NewSalt()
	hash, _ := crypto.DeriveKey("pass123", salt)
	encSalt, _ := crypto.NewSalt()

	store := mocks.NewMockUserStore(ctrl)
	store.EXPECT().GetUserByLogin(gomock.Any(), "alice").Return(&model.User{
		ID:       "user-1",
		Login:    "alice",
		AuthHash: hash,
		AuthSalt: salt,
		EncSalt:  encSalt,
	}, nil)

	svc := NewAuthService(store, testSecret, time.Hour)
	res, err := svc.Login(context.Background(), "alice", "pass123")
	require.NoError(t, err)
	assert.Equal(t, encSalt, res.EncSalt)

	uid, err := token.Verify(testSecret, res.Token)
	require.NoError(t, err)
	assert.Equal(t, "user-1", uid)
}

func TestLoginWrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	salt, _ := crypto.NewSalt()
	hash, _ := crypto.DeriveKey("correct", salt)

	store := mocks.NewMockUserStore(ctrl)
	store.EXPECT().GetUserByLogin(gomock.Any(), "alice").Return(&model.User{
		AuthHash: hash,
		AuthSalt: salt,
	}, nil)

	svc := NewAuthService(store, testSecret, time.Hour)
	_, err := svc.Login(context.Background(), "alice", "wrong")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLoginUnknownUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockUserStore(ctrl)
	store.EXPECT().GetUserByLogin(gomock.Any(), "ghost").Return(nil, repository.ErrUserNotFound)

	svc := NewAuthService(store, testSecret, time.Hour)
	_, err := svc.Login(context.Background(), "ghost", "pass")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLoginStoreError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockUserStore(ctrl)
	store.EXPECT().GetUserByLogin(gomock.Any(), "alice").Return(nil, errors.New("сбой БД"))

	svc := NewAuthService(store, testSecret, time.Hour)
	_, err := svc.Login(context.Background(), "alice", "pass")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidCredentials)
}
