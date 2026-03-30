package repository

import (
	"context"
	"database/sql"
	"errors"
	"shortener/internal/models"
	"time"

	"go.uber.org/zap"
)

type ShortenerRepository interface {
	GetOrCreateID(ctx context.Context, link string) (int, error)
	GetLongLinkByID(ctx context.Context, id int) (string, error)
}

type shortenerRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewShortenerRepository(db *sql.DB, logger *zap.Logger) ShortenerRepository {
	return &shortenerRepo{
		db:     db,
		logger: logger,
	}
}

func (r *shortenerRepo) GetOrCreateID(ctx context.Context, link string) (int, error) {
	query := `
		INSERT INTO Links (link) VALUES ($1)
		ON CONFLICT (link) DO UPDATE SET link = EXCLUDED.link
		RETURNING id;
	`

	ctx_timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var id int
	if err := r.db.QueryRowContext(ctx_timeout, query, link).Scan(&id); err != nil {
		r.logger.Error("Error to complete query", zap.Error(err))
		return 0, models.ErrInternalServer
	}

	return id, nil
}

func (r *shortenerRepo) GetLongLinkByID(ctx context.Context, id int) (string, error) {
	var link string
	query := `SELECT link FROM Links WHERE id = $1`

	ctx_timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.db.QueryRowContext(ctx_timeout, query, id).Scan(&link); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info("Not found link by id", zap.Int("ID:", id))
			return "", models.ErrNotFound
		}
		r.logger.Error("Error to complete query", zap.Error(err))
		return "", models.ErrInternalServer
	}

	return link, nil
}
