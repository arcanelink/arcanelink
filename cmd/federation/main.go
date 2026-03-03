package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcane/arcanelink/internal/federation/handler"
	"github.com/arcane/arcanelink/internal/federation/service"
	"github.com/arcane/arcanelink/pkg/config"
	"github.com/arcane/arcanelink/pkg/logger"
	pb "github.com/arcane/arcanelink/pkg/proto/federation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	if err := logger.Init(true); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Federation Service...")

	// Load configuration
	cfg := config.LoadFromEnv()

	if cfg.Server.GRPCPort == 0 {
		cfg.Server.GRPCPort = 50055
	}

	// Initialize service
	federationService := service.NewFederationService()

	// Initialize gRPC handler
	grpcHandler := handler.NewGRPCHandler(federationService)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFederationServiceServer(grpcServer, grpcHandler)

	logger.Info("Federation service starting",
		zap.Int("grpc_port", cfg.Server.GRPCPort),
		zap.String("features", "server-discovery, message-forwarding, retry-mechanism"))

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

	logger.Info("Shutting down Federation service...")
	grpcServer.GracefulStop()
	logger.Info("Federation service stopped")
}
