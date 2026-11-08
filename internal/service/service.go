package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
	"uuid"

	"github.com/pythonistD/url-shortener/internal/base63"
	"github.com/pythonistD/url-shortener/internal/model"
)

const (
	defaultTTL = 365 * 24 * time.Hour
	maxTTLSec  = 365 * 24 * 60 * 60
)

type Storage interface {
	GetByOriginal(ctx context.Context, original string) (model.URL, error)
	Save(ctx context.Context, url model.URL) error
	GetByShort(ctx context.Context, short string) (model.URL, error)
}

type Counter interface {
	Inc(ctx context.Context) (uint64, error)
}

type URLShorten struct {
	storage Storage
	counter Counter
}

func New(storage Storage, counter Counter) *URLShorten {
	return &URLShorten{
		storage: storage,
		counter: counter,
	}
}

func (s *URLShorten) Shorten(ctx context.Context, original string, ttl *uint64) (string, error) {
	parsed, err := url.ParseRequestURI(original)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", model.ErrInvalidURL
	}

	existing, err := s.storage.GetByOriginal(ctx, original)
	if err == nil {
		if time.Now().UTC().Before(existing.ExpiresAt) {
			return existing.Short, nil
		}
	} else if !errors.Is(err, model.ErrNotFound) {
		return "", fmt.Errorf("failed to get short by original: %w", err)
	}

	now := time.Now().UTC()
	var expiresAt time.Time
	if ttl != nil {
		if *ttl == 0 || *ttl > maxTTLSec {
			return "", fmt.Errorf("%w: must be 0 < ttl <= %d, got: %d", model.ErrInvalidURL, maxTTLSec, *ttl)
		}
		expiresAt = now.Add(time.Duration(*ttl) * time.Second)
	} else {
		expiresAt = now.Add(defaultTTL)
	}

	num, err := s.counter.Inc(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}
	short, err := base63.Encode(num)
	if err != nil {
		return "", fmt.Errorf("failed to encode short url: %w", err)
	}

	entity := model.URL{
		ID:        uuid.NewV7(),
		Original:  original,
		Short:     short,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	if err := s.storage.Save(ctx, entity); err != nil {
		return "", fmt.Errorf("failed to save url: %w", err)
	}

	return short, nil
}

func (s *URLShorten) GetURL(ctx context.Context, shortURL string) (string, error) {
	if _, err := base63.Decode(shortURL); err != nil {
		return "", model.ErrNotFound
	}

	entity, err := s.storage.GetByShort(ctx, shortURL)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("failed to get original: %w", err)
	}

	if time.Now().UTC().After(entity.ExpiresAt) {
		return "", model.ErrURLExpired
	}

	return entity.Original, nil
}
