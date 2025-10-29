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

// Package e2e provides end-to-end testing utilities
package e2e

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// TestCluster represents a test environment with all services
type TestCluster struct {
	UserServiceAddr  string
	OrderServiceAddr string
	GatewayAddr      string
	DatabaseAddr     string

	userConn  *grpc.ClientConn
	orderConn *grpc.ClientConn

	cleanup []func()
}

// NewTestCluster creates a new test cluster
func NewTestCluster(t *testing.T) *TestCluster {
	t.Helper()

	// In a real implementation, this would:
	// 1. Start PostgreSQL using testcontainers
	// 2. Start each microservice
	// 3. Wait for services to be healthy
	// 4. Return cluster info

	return &TestCluster{
		UserServiceAddr:  "localhost:9091",
		OrderServiceAddr: "localhost:9092",
		GatewayAddr:      "localhost:8080",
		DatabaseAddr:     "localhost:5432",
		cleanup:          []func(){},
	}
}

// ConnectToUserService establishes connection to user service
func (tc *TestCluster) ConnectToUserService(ctx context.Context) (*grpc.ClientConn, error) {
	if tc.userConn != nil {
		return tc.userConn, nil
	}

	conn, err := grpc.DialContext(
		ctx,
		tc.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	tc.userConn = conn
	tc.cleanup = append(tc.cleanup, func() { conn.Close() })
	return conn, nil
}

// ConnectToOrderService establishes connection to order service
func (tc *TestCluster) ConnectToOrderService(ctx context.Context) (*grpc.ClientConn, error) {
	if tc.orderConn != nil {
		return tc.orderConn, nil
	}

	conn, err := grpc.DialContext(
		ctx,
		tc.OrderServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	tc.orderConn = conn
	tc.cleanup = append(tc.cleanup, func() { conn.Close() })
	return conn, nil
}

// WaitForHealthy waits for all services to be healthy
func (tc *TestCluster) WaitForHealthy(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	services := []string{tc.UserServiceAddr, tc.OrderServiceAddr}

	for _, addr := range services {
		if err := tc.waitForService(ctx, addr); err != nil {
			return fmt.Errorf("service %s not healthy: %w", addr, err)
		}
	}

	return nil
}

func (tc *TestCluster) waitForService(ctx context.Context, addr string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if tc.isServiceHealthy(addr) {
				return nil
			}
		}
	}
}

func (tc *TestCluster) isServiceHealthy(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// CheckHealth checks service health using gRPC health check
func (tc *TestCluster) CheckHealth(ctx context.Context, addr string) error {
	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return err
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving")
	}

	return nil
}

// Cleanup cleans up all test resources
func (tc *TestCluster) Cleanup() {
	for i := len(tc.cleanup) - 1; i >= 0; i-- {
		tc.cleanup[i]()
	}
}
