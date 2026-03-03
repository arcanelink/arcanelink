package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcane/arcanelink/internal/message/handler"
	"github.com/arcane/arcanelink/internal/message/longpoll"
	"github.com/arcane/arcanelink/internal/message/repository"
	"github.com/arcane/arcanelink/internal/message/service"
	"github.com/arcane/arcanelink/pkg/config"
	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/logger"
	pb "github.com/arcane/arcanelink/pkg/proto/message"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	if err := logger.Init(true); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Message Service...")

	// Load configuration
	cfg := config.LoadFromEnv()

	if cfg.Server.GRPCPort == 0 {
		cfg.Server.GRPCPort = 50052
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
	messageRepo := repository.NewMessageRepository(db)

	// Initialize long poll manager
	longPollManager := longpoll.NewManager()

	// Initialize service
	messageService := service.NewMessageService(messageRepo, longPollManager)

	// Initialize gRPC handler
	grpcHandler := handler.NewGRPCHandler(messageService)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, grpcHandler)

	logger.Info("Message service starting",
		zap.Int("port", cfg.Server.GRPCPort),
		zap.String("features", "long-polling, direct-messages"))

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

	logger.Info("Shutting down Message service...")
	grpcServer.GracefulStop()
	logger.Info("Message service stopped")
}
