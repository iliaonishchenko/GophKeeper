package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type ErrorClassifier interface {
	IsRetriable(err error) bool
}

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08"
	}

	return false
}
