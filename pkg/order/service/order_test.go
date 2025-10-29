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

package service

import (
	"context"
	"testing"

	orderv1 "github.com/kevindiu/monorepo-go-example/apis/grpc/apis/proto/order/v1"
	"github.com/kevindiu/monorepo-go-example/internal/log"
	"github.com/kevindiu/monorepo-go-example/pkg/order/repository"
)

// mockRepository implements repository.Repository for testing
type mockRepository struct {
	orders     map[string]*repository.Order
	orderItems map[string][]*repository.OrderItem
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		orders:     make(map[string]*repository.Order),
		orderItems: make(map[string][]*repository.OrderItem),
	}
}

func (m *mockRepository) Create(ctx context.Context, order *repository.Order, items []*repository.OrderItem) error {
	m.orders[order.ID] = order
	m.orderItems[order.ID] = items
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*repository.Order, []*repository.OrderItem, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, nil, nil
	}
	items := m.orderItems[id]
	return order, items, nil
}

func (m *mockRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*repository.Order, error) {
	orders := []*repository.Order{}
	for _, order := range m.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]*repository.Order, error) {
	orders := []*repository.Order{}
	for _, order := range m.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id, status string) error {
	order, ok := m.orders[id]
	if !ok {
		return nil
	}
	order.Status = status
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	delete(m.orders, id)
	delete(m.orderItems, id)
	return nil
}

func TestNew(t *testing.T) {
	repo := newMockRepository()
	logger := log.NewDefault()

	svc := New(repo, logger)
	if svc == nil {
		t.Error("New() returned nil")
	}
}

func TestCreateOrder(t *testing.T) {
	repo := newMockRepository()
	logger := log.NewDefault()
	svc := New(repo, logger)

	tests := []struct {
		name    string
		req     *orderv1.CreateOrderRequest
		wantErr bool
	}{
		{
			name: "valid order",
			req: &orderv1.CreateOrderRequest{
				UserId: "user-1",
				Items: []*orderv1.OrderItem{
					{
						ProductId: "prod-1",
						Quantity:  2,
						Price:     10.50,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			req: &orderv1.CreateOrderRequest{
				UserId: "",
				Items: []*orderv1.OrderItem{
					{
						ProductId: "prod-1",
						Quantity:  1,
						Price:     10.00,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no items",
			req: &orderv1.CreateOrderRequest{
				UserId: "user-1",
				Items:  []*orderv1.OrderItem{},
			},
			wantErr: true,
		},
		{
			name: "invalid quantity",
			req: &orderv1.CreateOrderRequest{
				UserId: "user-1",
				Items: []*orderv1.OrderItem{
					{
						ProductId: "prod-1",
						Quantity:  0,
						Price:     10.00,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			req: &orderv1.CreateOrderRequest{
				UserId: "user-1",
				Items: []*orderv1.OrderItem{
					{
						ProductId: "prod-1",
						Quantity:  1,
						Price:     -10.00,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateOrder(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp == nil {
				t.Error("CreateOrder() returned nil response")
			}
			if !tt.wantErr && resp.Order == nil {
				t.Error("CreateOrder() returned nil order")
			}
		})
	}
}

func TestStatusConversion(t *testing.T) {
	tests := []struct {
		name       string
		statusStr  string
		statusEnum orderv1.OrderStatus
	}{
		{"pending", "pending", orderv1.OrderStatus_ORDER_STATUS_PENDING},
		{"confirmed", "confirmed", orderv1.OrderStatus_ORDER_STATUS_CONFIRMED},
		{"shipped", "shipped", orderv1.OrderStatus_ORDER_STATUS_SHIPPED},
		{"delivered", "delivered", orderv1.OrderStatus_ORDER_STATUS_DELIVERED},
		{"cancelled", "cancelled", orderv1.OrderStatus_ORDER_STATUS_CANCELLED},
		{"unknown", "unknown", orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test string to proto
			got := statusToProto(tt.statusStr)
			if got != tt.statusEnum {
				t.Errorf("statusToProto(%v) = %v, want %v", tt.statusStr, got, tt.statusEnum)
			}

			// Test proto to string (except unknown)
			if tt.statusStr != "unknown" {
				gotStr := statusFromProto(tt.statusEnum)
				if gotStr != tt.statusStr {
					t.Errorf("statusFromProto(%v) = %v, want %v", tt.statusEnum, gotStr, tt.statusStr)
				}
			}
		})
	}
}
