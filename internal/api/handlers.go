package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pythonistD/url-shortener/internal/model"
)

type Shortener interface {
	Shorten(ctx context.Context, original string, ttl *uint64) (string, error)
	GetURL(ctx context.Context, shortURL string) (string, error)
}

type Handler struct {
	logger  *slog.Logger
	service Shortener
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/url", h.RegisterURL)
	r.Get("/{short_url}", h.Redirect)
}

func (h *Handler) RegisterURL(w http.ResponseWriter, r *http.Request) {
	var req RegisterURLReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode req body", "error", err)
		http.Error(w, "failed to decode req body", http.StatusBadRequest)
		return
	}

	short, err := h.service.Shorten(r.Context(), req.URL, req.TTL)
	if err != nil {
		if errors.Is(err, model.ErrInvalidURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to shorten url", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(RegisterURLResp{ShortURL: short})
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to encode resp body", "error", err)
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "short_url")

	originalURL, err := h.service.GetURL(r.Context(), shortURL)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, model.ErrURLExpired) {
			http.Error(w, "url has expired", http.StatusGone)
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to get original url", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
