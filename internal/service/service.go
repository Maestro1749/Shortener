package service

import (
	"context"
	"net/url"
	"shortener/internal/models"
	"shortener/internal/repository"
	"strconv"

	"github.com/lytics/base62"
	"go.uber.org/zap"
)

type ShortenerService struct {
	repo   repository.ShortenerRepository
	logger *zap.Logger
}

func NewShortenerService(repo repository.ShortenerRepository, logger *zap.Logger) *ShortenerService {
	return &ShortenerService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ShortenerService) ShortenLink(ctx context.Context, link string) (string, error) {
	if len(link) == 0 || len(link) >= 2048 {
		return "", models.ErrInvalidDataInput
	}

	u, err := url.ParseRequestURI(link)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "localhost" {
		return "", models.ErrInvalidDataInput
	}

	id, err := s.repo.GetOrCreateID(ctx, link)
	if err != nil {
		return "", err
	}
	id_data := strconv.Itoa(id)

	short := base62.StdEncoding.EncodeToString([]byte(id_data))
	return short, nil
}

func (s *ShortenerService) DecodeShortLink(ctx context.Context, shortLink string) (string, error) {
	byte_id, err := base62.StdEncoding.DecodeString(shortLink)
	if err != nil {
		s.logger.Error("Error to decode link ID", zap.Error(err))
		return "", models.ErrInternalServer
	}

	id, err := strconv.Atoi(string(byte_id))
	if err != nil {
		s.logger.Error("Error to convert id in integer", zap.Error(err))
		return "", models.ErrInvalidDataInput
	}

	link, err := s.repo.GetLongLinkByID(ctx, id)
	if err != nil {
		return "", err
	}

	return link, nil
}
