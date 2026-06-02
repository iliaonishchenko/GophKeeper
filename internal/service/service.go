package service

import (
	"context"
	"errors"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
)

//go:generate mockgen -source=service.go -destination=mocks/store_mock.go -package=mocks

type UserStore interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type ItemStore interface {
	UpsertItem(ctx context.Context, it *model.Item) (*model.Item, error)
	GetItem(ctx context.Context, userID, id string) (*model.Item, error)
	ChangedSince(ctx context.Context, userID string, since int64) ([]*model.Item, int64, error)
}

var (
	ErrUserExists         = errors.New("пользователь уже существует")
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrEmptyCredentials   = errors.New("логин и пароль обязательны")
	ErrItemNotFound       = errors.New("элемент не найден")
)
