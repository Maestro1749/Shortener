package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"shortener/internal/logger"
	"shortener/internal/repository"
	"shortener/internal/service"
	"shortener/internal/transport"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	grpchandler "shortener/internal/transport/grpc"
	shortenerpc "shortener/internal/transport/grpc/proto/v1"
)

func main() {
	pid := os.Getpid()
	fmt.Println(pid)

	// Logger
	logger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Logger initialized successfully")

	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found, reading variables from environment")
	}

	appPort := os.Getenv("APP_PORT")
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	connectionString := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		panic(err)
	}
	logger.Info("Database connection established successfully")

	// Repository
	repo := repository.NewShortenerRepository(db, logger)
	logger.Info("Repositories initialized successfully")

	// Service
	service := service.NewShortenerService(repo, logger)
	logger.Info("Services initialized successfully")

	// Handler
	handler := transport.NewShortenerHandler(service, logger)
	logger.Info("handlers initialized successfully")

	router := mux.NewRouter()

	router.Path("/shortener").Methods("POST").HandlerFunc(handler.ShortenLink)
	router.Path("/{shortLink}").Methods("GET").HandlerFunc(handler.DecodeShortLink)

	// Server
	httpSrv := &http.Server{
		Addr:         ":" + appPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	// gRPC server
	grpcServer := grpc.NewServer()
	grpcHandler := grpchandler.NewServer(service, logger)

	shortenerpc.RegisterShortenerServiceServer(grpcServer, grpcHandler)

	go func() {
		logger.Info("Server started", zap.String("port", appPort))

		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("gRPC Server started", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Failed to start gRPC server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logger.Info("Shutting down server...")
	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
