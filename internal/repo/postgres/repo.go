package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pythonistD/url-shortener/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByOriginal(ctx context.Context, original string) (model.URL, error) {
	if err := ctx.Err(); err != nil {
		return model.URL{}, err
	}

	var entity model.URL
	if err := r.db.GetContext(ctx, &entity, URLQueries.GetByOriginal, original); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, model.ErrNotFound
		}
		return model.URL{}, fmt.Errorf("get url by original: %w", err)
	}

	return entity, nil
}

//nolint:gocritic // implements service.Storage interface
func (r *Repository) Save(ctx context.Context, url model.URL) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, URLQueries.Save, url.ID, url.Original, url.Short, url.CreatedAt, url.ExpiresAt)
	if err != nil {
		return fmt.Errorf("save url: %w", err)
	}

	return nil
}

func (r *Repository) GetByShort(ctx context.Context, short string) (model.URL, error) {
	if err := ctx.Err(); err != nil {
		return model.URL{}, err
	}

	var entity model.URL
	if err := r.db.GetContext(ctx, &entity, URLQueries.GetByShort, short); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, model.ErrNotFound
		}
		return model.URL{}, fmt.Errorf("get url by short: %w", err)
	}

	return entity, nil
}
