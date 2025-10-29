//
// Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	userv1 "github.com/kevindiu/monorepo-go-example/apis/grpc/apis/proto/user/v1"
	"github.com/kevindiu/monorepo-go-example/internal/config"
	"github.com/kevindiu/monorepo-go-example/internal/db"
	"github.com/kevindiu/monorepo-go-example/internal/log"
	"github.com/kevindiu/monorepo-go-example/internal/middleware"
	"github.com/kevindiu/monorepo-go-example/pkg/user/repository"
	"github.com/kevindiu/monorepo-go-example/pkg/user/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logCfg := &log.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}
	logger, err := log.New(logCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting user service",
		log.String("version", "1.0.0"),
		log.Int("grpc_port", cfg.Server.GRPCPort),
		log.Int("http_port", cfg.Server.Port),
	)

	// Connect to database
	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", log.Error(err))
	}
	defer database.Close()

	// Run migrations - skipping for now as migrations should be handled separately
	// In production, use a migration tool like golang-migrate
	logger.Info("Skipping automatic migrations - use migration tool separately")

	// Initialize repository and service
	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo)
	_ = userService // TODO: create gRPC handler wrapper

	// TODO: User service needs a gRPC handler wrapper since the business logic
	// service interface doesn't match the protobuf-generated gRPC interface.
	// For now, the service is created but not registered.
	// Create a handler in pkg/user/handler that wraps userService and implements userv1.UserServiceServer

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLoggingInterceptor(logger),
			middleware.UnaryRecoveryInterceptor(logger),
		),
	)

	// Register service - DISABLED until gRPC handler is implemented
	// userv1.RegisterUserServiceServer(grpcServer, userService)
	logger.Warn("User service gRPC handler not yet implemented - service will start but won't handle requests")
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", log.Error(err))
	}

	go func() {
		logger.Info("Starting gRPC server", log.String("address", grpcAddr))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("Failed to serve gRPC", log.Error(err))
		}
	}()

	// Create HTTP server with gRPC-Gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	// Register gateway
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := userv1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		logger.Fatal("Failed to register gateway", log.Error(err))
	}

	// Add health check endpoints
	handler := addHealthCheckEndpoints(mux, logger)

	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting HTTP server", log.String("address", httpAddr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to serve HTTP", log.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", log.Error(err))
	}

	grpcServer.GracefulStop()

	logger.Info("Server stopped")
}

func addHealthCheckEndpoints(mux *runtime.ServeMux, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
			logger.Debug("Health check", log.String("path", r.URL.Path))
			return
		}
		mux.ServeHTTP(w, r)
	})
}
