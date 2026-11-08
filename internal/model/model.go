package model

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrNotFound   = errors.New("url not found")
	ErrURLExpired = errors.New("url has expired")
	ErrInvalidURL = errors.New("invalid url")
)

type URL struct {
	ID        uuid.UUID `db:"id"`
	Original  string    `db:"original"`
	Short     string    `db:"short"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
