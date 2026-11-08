package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/pythonistD/url-shortener/internal/api"
	"github.com/pythonistD/url-shortener/internal/config"
	"github.com/pythonistD/url-shortener/internal/counter"
	"github.com/pythonistD/url-shortener/internal/external"
	"github.com/pythonistD/url-shortener/internal/logger"
	"github.com/pythonistD/url-shortener/internal/repo/inmem"
	"github.com/pythonistD/url-shortener/internal/repo/postgres"
	"github.com/pythonistD/url-shortener/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Env)
	log.Info("config loaded", "env", cfg.Env, "port", cfg.Server.Port, "storage", cfg.Storage.Type)

	var (
		storage service.Storage
		db      *sqlx.DB
	)

	cnt := counter.NewInMemory(0)

	switch cfg.Storage.Type {
	case "in-memory":
		storage = inmem.New()
	case "postgres":
		db, err = external.NewPostgres(cfg.Postgres)
		if err != nil {
			log.Error("failed to connect to postgres", "error", err)
			os.Exit(1)
		}

		storage = postgres.New(db)
	default:
		log.Error("unsupported storage type", "type", cfg.Storage.Type)
		os.Exit(1)
	}

	svc := service.New(storage, cnt)
	h := api.New(log, svc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(api.LoggingMiddleware(log))
	r.Use(middleware.Recoverer)

	h.RegisterRoutes(r)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("starting http server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped unexpectedly", "error", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	log.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown failed", "error", err)
	}

	wg.Wait()

	if db != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Error("failed to close postgres connection", "error", closeErr)
		}
	}

	log.Info("server stopped gracefully")
}
