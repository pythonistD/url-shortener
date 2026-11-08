package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pythonistD/url-shortener/internal/model"
)

type counterMock struct {
	val uint64
}

func (c *counterMock) Inc() uint64 {
	c.val++
	return c.val
}

type storageMock struct {
	byOriginal map[string]model.URL
	byShort    map[string]model.URL
}

func newStorageMock() *storageMock {
	return &storageMock{
		byOriginal: make(map[string]model.URL),
		byShort:    make(map[string]model.URL),
	}
}

func (s *storageMock) GetByOriginal(ctx context.Context, original string) (model.URL, error) {
	u, ok := s.byOriginal[original]
	if !ok {
		return model.URL{}, model.ErrNotFound
	}
	return u, nil
}

//nolint:gocritic // implements Storage interface
func (s *storageMock) Save(ctx context.Context, url model.URL) error {
	s.byOriginal[url.Original] = url
	s.byShort[url.Short] = url
	return nil
}

func (s *storageMock) GetByShort(ctx context.Context, short string) (model.URL, error) {
	u, ok := s.byShort[short]
	if !ok {
		return model.URL{}, model.ErrNotFound
	}
	return u, nil
}

func TestShorten_Positive(t *testing.T) {
	t.Run("valid url with default ttl", func(t *testing.T) {
		storage := newStorageMock()
		counter := &counterMock{}
		service := New(storage, counter)

		got, err := service.Shorten(context.Background(), "https://google.com", nil)
		if err != nil {
			t.Fatalf("Shorten unexpected error: %v", err)
		}
		if len(got) != 10 {
			t.Errorf("Shorten returned short len = %d, want 10", len(got))
		}
		if counter.val != 1 {
			t.Errorf("counter.val = %d, want 1", counter.val)
		}

		stored, ok := storage.byShort[got]
		if !ok {
			t.Fatalf("expected stored url for short %q", got)
		}
		if stored.Original != "https://google.com" {
			t.Errorf("stored.Original = %q, want https://google.com", stored.Original)
		}
		expectedExpire := stored.CreatedAt.Add(defaultTTL)
		if !stored.ExpiresAt.Equal(expectedExpire) {
			t.Errorf("stored.ExpiresAt = %v, want %v", stored.ExpiresAt, expectedExpire)
		}
	})

	t.Run("valid url with custom ttl", func(t *testing.T) {
		storage := newStorageMock()
		counter := &counterMock{}
		service := New(storage, counter)

		ttl := uint64(3600)
		got, err := service.Shorten(context.Background(), "https://github.com", &ttl)
		if err != nil {
			t.Fatalf("Shorten unexpected error: %v", err)
		}

		stored := storage.byShort[got]
		expectedExpire := stored.CreatedAt.Add(3600 * time.Second)
		if !stored.ExpiresAt.Equal(expectedExpire) {
			t.Errorf("stored.ExpiresAt = %v, want %v", stored.ExpiresAt, expectedExpire)
		}
	})

	t.Run("valid url with max ttl", func(t *testing.T) {
		storage := newStorageMock()
		counter := &counterMock{}
		service := New(storage, counter)

		ttl := uint64(maxTTLSec)
		got, err := service.Shorten(context.Background(), "https://golang.org", &ttl)
		if err != nil {
			t.Fatalf("Shorten unexpected error: %v", err)
		}
		if len(got) != 10 {
			t.Errorf("Shorten returned short len = %d, want 10", len(got))
		}
	})
}

func TestShorten_Idempotency(t *testing.T) {
	t.Run("same url returns same short", func(t *testing.T) {
		storage := newStorageMock()
		counter := &counterMock{}
		service := New(storage, counter)

		short1, err := service.Shorten(context.Background(), "https://example.com", nil)
		if err != nil {
			t.Fatalf("first Shorten unexpected error: %v", err)
		}

		short2, err := service.Shorten(context.Background(), "https://example.com", nil)
		if err != nil {
			t.Fatalf("second Shorten unexpected error: %v", err)
		}

		if short1 != short2 {
			t.Errorf("idempotency violated: short1 = %q, short2 = %q", short1, short2)
		}
		if counter.val != 1 {
			t.Errorf("counter incremented on duplicate request: counter.val = %d, want 1", counter.val)
		}
	})

	t.Run("expired url generates new short", func(t *testing.T) {
		storage := newStorageMock()
		counter := &counterMock{}
		service := New(storage, counter)

		short1, err := service.Shorten(context.Background(), "https://example.com", nil)
		if err != nil {
			t.Fatalf("first Shorten unexpected error: %v", err)
		}

		stored := storage.byOriginal["https://example.com"]
		stored.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)
		storage.byOriginal["https://example.com"] = stored

		short2, err := service.Shorten(context.Background(), "https://example.com", nil)
		if err != nil {
			t.Fatalf("Shorten expired unexpected error: %v", err)
		}
		if short2 == short1 {
			t.Errorf("expected new short for expired url, got same: %q", short2)
		}
		if counter.val != 2 {
			t.Errorf("counter.val = %d, want 2 after expired url re-creation", counter.val)
		}
	})
}

func TestShorten_Negative(t *testing.T) {
	tests := []struct {
		name     string
		original string
		ttl      *uint64
		wantErr  error
	}{
		{
			name:     "empty url",
			original: "",
			ttl:      nil,
			wantErr:  model.ErrInvalidURL,
		},
		{
			name:     "missing scheme",
			original: "google.com",
			ttl:      nil,
			wantErr:  model.ErrInvalidURL,
		},
		{
			name:     "invalid url format",
			original: "://aaaav",
			ttl:      nil,
			wantErr:  model.ErrInvalidURL,
		},
		{
			name:     "zero ttl",
			original: "https://google.com",
			ttl:      new(uint64(0)),
			wantErr:  model.ErrInvalidURL,
		},
		{
			name:     "ttl greater than max",
			original: "https://google.com",
			ttl:      new(uint64(maxTTLSec + 1)),
			wantErr:  model.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newStorageMock()
			counter := &counterMock{}
			service := New(storage, counter)

			got, err := service.Shorten(context.Background(), tt.original, tt.ttl)
			if err == nil {
				t.Fatalf("Shorten(%q) expected error %v, got %q", tt.original, tt.wantErr, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Shorten(%q) error = %v, want %v", tt.original, err, tt.wantErr)
			}
			if counter.val != 0 {
				t.Errorf("counter incremented on failed validation: %d", counter.val)
			}
		})
	}
}

func TestGetURL_Positive(t *testing.T) {
	storage := newStorageMock()
	counter := &counterMock{}
	service := New(storage, counter)

	short, err := service.Shorten(context.Background(), "https://google.com", nil)
	if err != nil {
		t.Fatalf("Shorten unexpected error: %v", err)
	}

	got, err := service.GetURL(context.Background(), short)
	if err != nil {
		t.Fatalf("GetURL(%q) unexpected error: %v", short, err)
	}
	if got != "https://google.com" {
		t.Errorf("GetURL(%q) = %q, want https://google.com", short, got)
	}
}

func TestGetURL_Negative(t *testing.T) {
	tests := []struct {
		name     string
		shortURL string
		wantErr  error
	}{
		{
			name:     "too short length",
			shortURL: "abc",
			wantErr:  model.ErrNotFound,
		},
		{
			name:     "too long length",
			shortURL: "aaaaaaaaaaa",
			wantErr:  model.ErrNotFound,
		},
		{
			name:     "invalid character in short",
			shortURL: "aaaaaaaaa!",
			wantErr:  model.ErrNotFound,
		},
		{
			name:     "non-existent valid short",
			shortURL: "aaaaaaaaaz",
			wantErr:  model.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newStorageMock()
			counter := &counterMock{}
			service := New(storage, counter)

			got, err := service.GetURL(context.Background(), tt.shortURL)
			if err == nil {
				t.Fatalf("GetURL(%q) expected error %v, got %q", tt.shortURL, tt.wantErr, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetURL(%q) error = %v, want %v", tt.shortURL, err, tt.wantErr)
			}
		})
	}

	t.Run("expired url", func(t *testing.T) {
		storage := newStorageMock()
		storage.byShort["aaaaaaaaab"] = model.URL{
			Original:  "https://expired.com",
			Short:     "aaaaaaaaab",
			ExpiresAt: time.Now().UTC().Add(-5 * time.Minute),
		}
		counter := &counterMock{}
		service := New(storage, counter)

		got, err := service.GetURL(context.Background(), "aaaaaaaaab")
		if err == nil {
			t.Fatalf("GetURL expected error %v, got %q", model.ErrURLExpired, got)
		}
		if !errors.Is(err, model.ErrURLExpired) {
			t.Errorf("GetURL error = %v, want %v", err, model.ErrURLExpired)
		}
	})
}
