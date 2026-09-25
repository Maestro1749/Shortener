package grpc

import (
	"context"
	"errors"
	"shortener/internal/models"
	pb "shortener/internal/transport/grpc/proto/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShortenerService interface {
	ShortenLink(ctx context.Context, longURL string) (string, error)
	DecodeShortLink(ctx context.Context, shortCode string) (string, error)
}

type Server struct {
	pb.UnimplementedShortenerServiceServer
	service ShortenerService
	logger  *zap.Logger
}

func NewServer(service ShortenerService, logger *zap.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger,
	}
}

func (s *Server) CreateShortURL(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	if req.GetLongUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url cannot be empty")
	}

	shortCode, err := s.service.ShortenLink(ctx, req.GetLongUrl())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create short url: %v", err)
	}

	return &pb.CreateResponse{ShortUrl: shortCode}, nil
}

func (s *Server) GetOriginalURL(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.GetShortCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "short code cannot be empty")
	}

	originalURL, err := s.service.DecodeShortLink(ctx, req.GetShortCode())
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "link not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get url: %v", err)
	}

	return &pb.GetResponse{OriginalUrl: originalURL}, nil
}
