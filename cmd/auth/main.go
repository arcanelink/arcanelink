package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcane/arcanelink/internal/auth/handler"
	"github.com/arcane/arcanelink/internal/auth/repository"
	"github.com/arcane/arcanelink/internal/auth/service"
	"github.com/arcane/arcanelink/pkg/config"
	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/logger"
	pb "github.com/arcane/arcanelink/pkg/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	if err := logger.Init(true); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Auth Service...")

	// Load configuration from environment
	cfg := config.LoadFromEnv()

	// Set defaults if not provided
	if cfg.Server.GRPCPort == 0 {
		cfg.Server.GRPCPort = 50051
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "default-secret-change-in-production"
	}
	if cfg.JWT.ExpiresIn == 0 {
		cfg.JWT.ExpiresIn = 86400 // 24 hours
	}

	// Connect to database
	db, err := database.NewPostgresDB(&database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Connected to database successfully")

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize service
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	// Initialize gRPC handler
	grpcHandler := handler.NewGRPCHandler(authService)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, grpcHandler)

	logger.Info("Auth service starting", zap.Int("port", cfg.Server.GRPCPort))

	// Graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Auth service...")
	grpcServer.GracefulStop()
	logger.Info("Auth service stopped")
}
