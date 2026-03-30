package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	// .env
	if err := godotenv.Load(); err != nil {
		logger.Error("Error to load .env file")
		panic(err)
	}

	userDB := os.Getenv("DB_USER")
	passwordDB := os.Getenv("DB_PASSWORD")
	hostDB := os.Getenv("DB_HOST")
	portDB := os.Getenv("DB_PORT")
	nameDB := os.Getenv("DB_NAME")

	// Database
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", userDB, passwordDB, hostDB, portDB, nameDB)

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
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Server started on :8080")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
