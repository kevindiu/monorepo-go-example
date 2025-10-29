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

package e2e

import (
	"context"
	"testing"
	"time"

	orderv1 "github.com/kevindiu/monorepo-go-example/apis/grpc/apis/proto/order/v1"
)

// TestOrderServiceE2E tests the order service end-to-end
func TestOrderServiceE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This test requires running services
	t.Skip("E2E test - requires running services")

	cluster := NewTestCluster(t)
	defer cluster.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for services to be ready
	if err := cluster.WaitForHealthy(ctx, 10*time.Second); err != nil {
		t.Fatalf("Services not healthy: %v", err)
	}

	// Connect to order service
	conn, err := cluster.ConnectToOrderService(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to order service: %v", err)
	}

	client := orderv1.NewOrderServiceClient(conn)

	// Test: Create Order
	t.Run("CreateOrder", func(t *testing.T) {
		req := &orderv1.CreateOrderRequest{
			UserId: "test-user-1",
			Items: []*orderv1.OrderItem{
				{
					ProductId: "prod-1",
					Quantity:  2,
					Price:     29.99,
				},
			},
		}

		resp, err := client.CreateOrder(ctx, req)
		if err != nil {
			t.Fatalf("CreateOrder failed: %v", err)
		}

		if resp.Order == nil {
			t.Fatal("CreateOrder returned nil order")
		}

		if resp.Order.UserId != req.UserId {
			t.Errorf("Order.UserId = %v, want %v", resp.Order.UserId, req.UserId)
		}

		orderID := resp.Order.Id

		// Test: Get Order
		t.Run("GetOrder", func(t *testing.T) {
			getReq := &orderv1.GetOrderRequest{
				Id: orderID,
			}

			getResp, err := client.GetOrder(ctx, getReq)
			if err != nil {
				t.Fatalf("GetOrder failed: %v", err)
			}

			if getResp.Order.Id != orderID {
				t.Errorf("Order.Id = %v, want %v", getResp.Order.Id, orderID)
			}
		})

		// Test: Update Order Status
		t.Run("UpdateOrderStatus", func(t *testing.T) {
			updateReq := &orderv1.UpdateOrderStatusRequest{
				Id:     orderID,
				Status: orderv1.OrderStatus_ORDER_STATUS_CONFIRMED,
			}

			updateResp, err := client.UpdateOrderStatus(ctx, updateReq)
			if err != nil {
				t.Fatalf("UpdateOrderStatus failed: %v", err)
			}

			if updateResp.Order.Status != orderv1.OrderStatus_ORDER_STATUS_CONFIRMED {
				t.Errorf("Order.Status = %v, want %v", updateResp.Order.Status, orderv1.OrderStatus_ORDER_STATUS_CONFIRMED)
			}
		})

		// Test: Cancel Order
		t.Run("CancelOrder", func(t *testing.T) {
			cancelReq := &orderv1.CancelOrderRequest{
				Id: orderID,
			}

			cancelResp, err := client.CancelOrder(ctx, cancelReq)
			if err != nil {
				t.Fatalf("CancelOrder failed: %v", err)
			}

			if !cancelResp.Success {
				t.Error("CancelOrder did not return success")
			}
		})
	})

	// Test: List Orders
	t.Run("ListOrders", func(t *testing.T) {
		req := &orderv1.ListOrdersRequest{
			PageSize: 10,
		}

		resp, err := client.ListOrders(ctx, req)
		if err != nil {
			t.Fatalf("ListOrders failed: %v", err)
		}

		if resp.Orders == nil {
			t.Error("ListOrders returned nil orders")
		}
	})
}
