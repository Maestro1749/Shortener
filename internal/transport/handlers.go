package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"shortener/internal/models"
	"shortener/internal/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ShortenerHandler struct {
	service *service.ShortenerService
	logger  *zap.Logger
}

func NewShortenerHandler(service *service.ShortenerService, logger *zap.Logger) *ShortenerHandler {
	return &ShortenerHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ShortenerHandler) ShortenLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var link models.ShortenLinkDTO

	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		h.logger.Error("Error to decode data", zap.Error(err))
		http.Error(w, "Error to decode data", http.StatusBadRequest)
		return
	}

	short, err := h.service.ShortenLink(ctx, link.Link)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidDataInput):
			http.Error(w, "Invalid input data", http.StatusBadRequest)
			return
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.ShortenLinkResponseDTO{ShortURL: short}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error to encode output data", http.StatusInternalServerError)
		return
	}
}

func (h *ShortenerHandler) DecodeShortLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	shortLink := vars["shortLink"]

	if shortLink == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	link, err := h.service.DecodeShortLink(ctx, shortLink)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			http.Error(w, "Not found current link", http.StatusNotFound)
			return
		case errors.Is(err, models.ErrInvalidDataInput):
			http.Error(w, "Invalid input short link", http.StatusBadRequest)
			return
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, link, http.StatusMovedPermanently)
}
