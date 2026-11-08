//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
	"uuid"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pythonistD/url-shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:18-alpine",
		tcpostgres.WithDatabase("test_db"),
		tcpostgres.WithUsername("test_user"),
		tcpostgres.WithPassword("test_pass"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start testcontainer: %v\n", err)
		os.Exit(1)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db: %v\n", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}
	testDB = db

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set goose dialect: %v\n", err)
		_ = db.Close()
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	if err := goose.Up(db.DB, "../../../migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to apply goose migrations: %v\n", err)
		_ = db.Close()
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	code := m.Run()

	_ = db.Close()
	_ = pgContainer.Terminate(context.Background())

	os.Exit(code)
}

func truncateTable(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE urls")
	require.NoError(t, err)
}

func TestRepository_SaveAndGet(t *testing.T) {
	truncateTable(t)
	repo := New(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)
	entity := model.URL{
		ID:        uuid.NewV7(),
		Original:  "https://github.com/",
		Short:     "aaaaaaaaab",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	err := repo.Save(ctx, entity)
	require.NoError(t, err)

	gotByOriginal, err := repo.GetByOriginal(ctx, entity.Original)
	require.NoError(t, err)
	assert.Equal(t, entity.ID, gotByOriginal.ID)
	assert.Equal(t, entity.Original, gotByOriginal.Original)
	assert.Equal(t, entity.Short, gotByOriginal.Short)
	assert.WithinDuration(t, entity.CreatedAt, gotByOriginal.CreatedAt, time.Second)
	assert.WithinDuration(t, entity.ExpiresAt, gotByOriginal.ExpiresAt, time.Second)

	gotByShort, err := repo.GetByShort(ctx, entity.Short)
	require.NoError(t, err)
	assert.Equal(t, entity.ID, gotByShort.ID)
	assert.Equal(t, entity.Original, gotByShort.Original)
	assert.Equal(t, entity.Short, gotByShort.Short)
}

func TestRepository_NotFound(t *testing.T) {
	truncateTable(t)
	repo := New(testDB)
	ctx := context.Background()

	_, err := repo.GetByOriginal(ctx, "https://fake.com")
	assert.ErrorIs(t, err, model.ErrNotFound)

	_, err = repo.GetByShort(ctx, "foooo")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestRepository_DuplicateShort(t *testing.T) {
	truncateTable(t)
	repo := New(testDB)
	ctx := context.Background()

	now := time.Now().UTC()
	entity1 := model.URL{
		ID:        uuid.NewV7(),
		Original:  "https://first.com",
		Short:     "short",
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}
	entity2 := model.URL{
		ID:        uuid.NewV7(),
		Original:  "https://second.com",
		Short:     "short",
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}

	err := repo.Save(ctx, entity1)
	require.NoError(t, err)

	err = repo.Save(ctx, entity2)
	assert.Error(t, err, "short must be unique")
}
