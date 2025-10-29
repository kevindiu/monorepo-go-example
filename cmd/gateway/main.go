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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kevindiu/monorepo-go-example/internal/config"
	"github.com/kevindiu/monorepo-go-example/internal/log"
	"github.com/kevindiu/monorepo-go-example/pkg/gateway"
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

	logger.Info("Starting gateway service",
		log.String("version", "1.0.0"),
		log.Int("port", cfg.Server.Port),
	)

	// Get service endpoints from environment
	userServiceEndpoint := os.Getenv("USER_SERVICE_ENDPOINT")
	if userServiceEndpoint == "" {
		userServiceEndpoint = "localhost:9091"
	}

	orderServiceEndpoint := os.Getenv("ORDER_SERVICE_ENDPOINT")
	if orderServiceEndpoint == "" {
		orderServiceEndpoint = "localhost:9092"
	}

	logger.Info("Backend service endpoints",
		log.String("user_service", userServiceEndpoint),
		log.String("order_service", orderServiceEndpoint),
	)

	// Create gateway
	gw, err := gateway.New(gateway.Config{
		UserServiceEndpoint:  userServiceEndpoint,
		OrderServiceEndpoint: orderServiceEndpoint,
		Logger:               logger,
	})
	if err != nil {
		logger.Fatal("Failed to create gateway", log.Error(err))
	}

	// Start gateway (initialize connections)
	ctx := context.Background()
	if err := gw.Start(ctx); err != nil {
		logger.Fatal("Failed to start gateway", log.Error(err))
	}

	// Create HTTP server
	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      gw.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", log.Error(err))
	}

	logger.Info("Server stopped")
}
