package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
)

var ErrUserNotFound = errors.New("пользователь не найден")

type UsersRepository struct {
	baseRepo
}

func NewUsersRepository(db *sql.DB, classifier ErrorClassifier) *UsersRepository {
	return &UsersRepository{baseRepo{db: db, classifier: classifier}}
}

func (r *UsersRepository) CreateUser(ctx context.Context, u *model.User) error {
	return executeWithRetry(r.classifier, func() error {
		const query = `
			INSERT INTO users (id, login, auth_hash, auth_salt, enc_salt)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err := r.db.ExecContext(ctx, query, u.ID, u.Login, u.AuthHash, u.AuthSalt, u.EncSalt)
		return err
	})
}

func (r *UsersRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var u model.User
	err := executeWithRetry(r.classifier, func() error {
		const query = `
			SELECT id, login, auth_hash, auth_salt, enc_salt
			FROM users WHERE login = $1
		`
		row := r.db.QueryRowContext(ctx, query, login)
		return row.Scan(&u.ID, &u.Login, &u.AuthHash, &u.AuthSalt, &u.EncSalt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
