package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arcane/arcanelink/internal/api-gateway/handler"
	"github.com/arcane/arcanelink/internal/api-gateway/middleware"
	"github.com/arcane/arcanelink/internal/api-gateway/router"
	"github.com/arcane/arcanelink/pkg/config"
	"github.com/arcane/arcanelink/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := logger.Init(true); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting API Gateway...")

	// Load configuration
	cfg := config.LoadFromEnv()

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	// Initialize API handler
	apiHandler, err := handler.NewAPIHandler(
		cfg.Services.Auth,
		cfg.Services.Message,
		cfg.Services.Room,
		cfg.Server.Domain,
	)
	if err != nil {
		logger.Fatal("Failed to initialize API handler", zap.Error(err))
	}

	logger.Info("API Gateway configuration",
		zap.String("server_domain", cfg.Server.Domain),
		zap.String("auth_service", cfg.Services.Auth))

	// Initialize auth middleware
	authMiddleware, err := middleware.NewAuthMiddleware(cfg.Services.Auth)
	if err != nil {
		logger.Fatal("Failed to initialize auth middleware", zap.Error(err))
	}

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Setup router
	r := router.SetupRouter(apiHandler, authMiddleware, rateLimiter)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  35 * time.Second, // Slightly longer than long poll timeout
		WriteTimeout: 35 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("API Gateway starting",
		zap.Int("port", cfg.Server.Port),
		zap.String("auth_service", cfg.Services.Auth),
		zap.String("message_service", cfg.Services.Message))

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("API Gateway is ready to accept requests")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("API Gateway stopped")
}
