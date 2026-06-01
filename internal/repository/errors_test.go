package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsRetriable(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil-ошибка",
			err:  nil,
			want: false,
		},
		{
			name: "обычная ошибка",
			err:  errors.New("какая-то ошибка"),
			want: false,
		},
		{
			name: "класс 08 — connection exception",
			err:  &pgconn.PgError{Code: "08006"},
			want: true,
		},
		{
			name: "класс 23 — integrity violation",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "пустой код",
			err:  &pgconn.PgError{Code: ""},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, classifier.IsRetriable(tt.err))
		})
	}
}
