package inmem

import (
	"context"
	"sync"

	"github.com/pythonistD/url-shortener/internal/model"
)

type Repository struct {
	mu         sync.RWMutex
	byOriginal map[string]model.URL
	byShort    map[string]model.URL
}

func New() *Repository {
	return &Repository{
		byOriginal: make(map[string]model.URL),
		byShort:    make(map[string]model.URL),
	}
}

func (r *Repository) GetByOriginal(ctx context.Context, original string) (model.URL, error) {
	if err := ctx.Err(); err != nil {
		return model.URL{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.byOriginal[original]
	if !ok {
		return model.URL{}, model.ErrNotFound
	}

	return entity, nil
}

func (r *Repository) Save(ctx context.Context, url model.URL) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byOriginal[url.Original] = url
	r.byShort[url.Short] = url
	return nil
}

func (r *Repository) GetByShort(ctx context.Context, short string) (model.URL, error) {
	if err := ctx.Err(); err != nil {
		return model.URL{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.byShort[short]
	if !ok {
		return model.URL{}, model.ErrNotFound
	}

	return entity, nil
}
