package service

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"time"

	"github.com/iliaonishchenko/gophkeeper/internal/crypto"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/iliaonishchenko/gophkeeper/internal/repository"
	"github.com/iliaonishchenko/gophkeeper/internal/token"
)

type AuthResult struct {
	Token   string
	EncSalt []byte
}

type AuthService struct {
	users  UserStore
	secret string
	ttl    time.Duration
}

func NewAuthService(users UserStore, secret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, secret: secret, ttl: ttl}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (*AuthResult, error) {
	if login == "" || password == "" {
		return nil, ErrEmptyCredentials
	}

	authSalt, err := crypto.NewSalt()
	if err != nil {
		return nil, err
	}
	encSalt, err := crypto.NewSalt()
	if err != nil {
		return nil, err
	}
	authHash, err := crypto.DeriveKey(password, authSalt)
	if err != nil {
		return nil, err
	}

	id, err := model.NewID()
	if err != nil {
		return nil, err
	}

	u := &model.User{
		ID:       id,
		Login:    login,
		AuthHash: authHash,
		AuthSalt: authSalt,
		EncSalt:  encSalt,
	}
	if err := s.users.CreateUser(ctx, u); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	return &AuthResult{Token: token.Issue(s.secret, u.ID, s.ttl), EncSalt: encSalt}, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*AuthResult, error) {
	if login == "" || password == "" {
		return nil, ErrEmptyCredentials
	}

	u, err := s.users.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("поиск пользователя: %w", err)
	}

	hash, err := crypto.DeriveKey(password, u.AuthSalt)
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(hash, u.AuthHash) {
		return nil, ErrInvalidCredentials
	}

	return &AuthResult{Token: token.Issue(s.secret, u.ID, s.ttl), EncSalt: u.EncSalt}, nil
}

func isUniqueViolation(err error) bool {
	type sqlStateError interface{ SQLState() string }
	var stateErr sqlStateError
	if errors.As(err, &stateErr) {
		return stateErr.SQLState() == "23505"
	}
	return false
}
