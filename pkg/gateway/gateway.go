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

package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	orderv1 "github.com/kevindiu/monorepo-go-example/apis/grpc/apis/proto/order/v1"
	userv1 "github.com/kevindiu/monorepo-go-example/apis/grpc/apis/proto/user/v1"
	"github.com/kevindiu/monorepo-go-example/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Gateway represents the API gateway
type Gateway struct {
	userServiceEndpoint  string
	orderServiceEndpoint string
	logger               *log.Logger
	mux                  *runtime.ServeMux
}

// Config holds gateway configuration
type Config struct {
	UserServiceEndpoint  string
	OrderServiceEndpoint string
	Logger               *log.Logger
}

// New creates a new gateway
func New(cfg Config) (*Gateway, error) {
	if cfg.Logger == nil {
		cfg.Logger = log.NewDefault()
	}

	// Create gRPC-Gateway mux
	mux := runtime.NewServeMux()

	gw := &Gateway{
		userServiceEndpoint:  cfg.UserServiceEndpoint,
		orderServiceEndpoint: cfg.OrderServiceEndpoint,
		logger:               cfg.Logger,
		mux:                  mux,
	}

	return gw, nil
}

// Start initializes connections to backend services and registers handlers
func (g *Gateway) Start(ctx context.Context) error {
	// Connect to user service
	g.logger.Info("Connecting to user service", log.String("endpoint", g.userServiceEndpoint))
	userConn, err := grpc.DialContext(
		ctx,
		g.userServiceEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %w", err)
	}

	// Register user service handler
	if err := userv1.RegisterUserServiceHandler(ctx, g.mux, userConn); err != nil {
		return fmt.Errorf("failed to register user service handler: %w", err)
	}

	// Connect to order service
	g.logger.Info("Connecting to order service", log.String("endpoint", g.orderServiceEndpoint))
	orderConn, err := grpc.DialContext(
		ctx,
		g.orderServiceEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to order service: %w", err)
	}

	// Register order service handler
	if err := orderv1.RegisterOrderServiceHandler(ctx, g.mux, orderConn); err != nil {
		return fmt.Errorf("failed to register order service handler: %w", err)
	}

	g.logger.Info("Gateway initialized successfully")
	return nil
}

// Handler returns the HTTP handler
func (g *Gateway) Handler() http.Handler {
	// Wrap the mux with middleware
	handler := g.loggingMiddleware(g.mux)
	handler = g.corsMiddleware(handler)
	handler = g.healthCheckMiddleware(handler)
	return handler
}

// healthCheckMiddleware adds health check endpoints
func (g *Gateway) healthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs incoming requests
func (g *Gateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.logger.Info("Request",
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.String("remote_addr", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles CORS
func (g *Gateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
