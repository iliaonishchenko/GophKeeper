package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
)

var ErrItemNotFound = errors.New("элемент не найден")

type ItemsRepository struct {
	baseRepo
}

func NewItemsRepository(db *sql.DB, classifier ErrorClassifier) *ItemsRepository {
	return &ItemsRepository{baseRepo{db: db, classifier: classifier}}
}

func (r *ItemsRepository) UpsertItem(ctx context.Context, it *model.Item) (*model.Item, error) {
	err := executeWithRetry(r.classifier, func() error {
		const query = `
			INSERT INTO items (id, user_id, type, name, ciphertext, metadata, version, updated_at, deleted)
			VALUES ($1, $2, $3, $4, $5, $6, 1, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				type       = EXCLUDED.type,
				name       = EXCLUDED.name,
				ciphertext = EXCLUDED.ciphertext,
				metadata   = EXCLUDED.metadata,
				version    = items.version + 1,
				updated_at = EXCLUDED.updated_at,
				deleted    = EXCLUDED.deleted
			WHERE items.user_id = EXCLUDED.user_id
			RETURNING version, updated_at
		`
		row := r.db.QueryRowContext(ctx, query,
			it.ID, it.UserID, it.Type, it.Name, it.Ciphertext, it.Metadata, it.UpdatedAt, it.Deleted)
		return row.Scan(&it.Version, &it.UpdatedAt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return it, nil
}

func (r *ItemsRepository) GetItem(ctx context.Context, userID, id string) (*model.Item, error) {
	var it model.Item
	err := executeWithRetry(r.classifier, func() error {
		const query = `
			SELECT id, user_id, type, name, ciphertext, metadata, version, updated_at, deleted
			FROM items WHERE id = $1 AND user_id = $2
		`
		row := r.db.QueryRowContext(ctx, query, id, userID)
		return row.Scan(&it.ID, &it.UserID, &it.Type, &it.Name, &it.Ciphertext,
			&it.Metadata, &it.Version, &it.UpdatedAt, &it.Deleted)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *ItemsRepository) ChangedSince(ctx context.Context, userID string, since int64) ([]*model.Item, int64, error) {
	var items []*model.Item
	cursor := since
	err := executeWithRetry(r.classifier, func() error {
		const query = `
			SELECT id, user_id, type, name, ciphertext, metadata, version, updated_at, deleted
			FROM items WHERE user_id = $1 AND updated_at > $2
			ORDER BY updated_at
		`
		rows, err := r.db.QueryContext(ctx, query, userID, since)
		if err != nil {
			return err
		}
		defer rows.Close()

		items = items[:0]
		cursor = since
		for rows.Next() {
			var it model.Item
			if err := rows.Scan(&it.ID, &it.UserID, &it.Type, &it.Name, &it.Ciphertext,
				&it.Metadata, &it.Version, &it.UpdatedAt, &it.Deleted); err != nil {
				return err
			}
			items = append(items, &it)
			if it.UpdatedAt > cursor {
				cursor = it.UpdatedAt
			}
		}
		return rows.Err()
	})
	if err != nil {
		return nil, since, err
	}
	return items, cursor, nil
}
