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
	"strconv"

	orderv1 "github.com/kevindiu/monorepo-go-example/apis/grpc/apis/proto/order/v1"
	"github.com/kevindiu/monorepo-go-example/internal/errors"
	"github.com/kevindiu/monorepo-go-example/internal/log"
	"github.com/kevindiu/monorepo-go-example/pkg/order/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// statusToProto converts string status to proto enum
func statusToProto(status string) orderv1.OrderStatus {
	switch status {
	case "pending":
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case "confirmed":
		return orderv1.OrderStatus_ORDER_STATUS_CONFIRMED
	case "shipped":
		return orderv1.OrderStatus_ORDER_STATUS_SHIPPED
	case "delivered":
		return orderv1.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

// statusFromProto converts proto enum to string status
func statusFromProto(status orderv1.OrderStatus) string {
	switch status {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		return "pending"
	case orderv1.OrderStatus_ORDER_STATUS_CONFIRMED:
		return "confirmed"
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		return "shipped"
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "pending"
	}
}

// Service defines the order service interface
type Service interface {
	orderv1.OrderServiceServer
}

type service struct {
	orderv1.UnimplementedOrderServiceServer
	repo   repository.Repository
	logger *log.Logger
}

// New creates a new order service
func New(repo repository.Repository, logger *log.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateOrder creates a new order
func (s *service) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	s.logger.Info("Creating order", log.String("user_id", req.GetUserId()))

	// Validate request
	if req.GetUserId() == "" {
		return nil, errors.WithCode(errors.New("user_id is required"), errors.CodeInvalidInput)
	}

	if len(req.GetItems()) == 0 {
		return nil, errors.WithCode(errors.New("at least one item is required"), errors.CodeInvalidInput)
	}

	// Calculate total amount
	var totalAmount float64
	items := make([]*repository.OrderItem, len(req.GetItems()))
	for i, item := range req.GetItems() {
		if item.GetProductId() == "" {
			return nil, errors.WithCode(errors.New("product_id is required"), errors.CodeInvalidInput)
		}
		if item.GetQuantity() <= 0 {
			return nil, errors.WithCode(errors.New("quantity must be positive"), errors.CodeInvalidInput)
		}
		if item.GetPrice() <= 0 {
			return nil, errors.WithCode(errors.New("price must be positive"), errors.CodeInvalidInput)
		}

		items[i] = &repository.OrderItem{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
			Price:     item.GetPrice(),
		}
		totalAmount += float64(item.GetQuantity()) * item.GetPrice()
	}

	// Create order
	order := &repository.Order{
		UserID:      req.GetUserId(),
		Status:      "pending",
		TotalAmount: totalAmount,
	}

	if err := s.repo.Create(ctx, order, items); err != nil {
		s.logger.Error("Failed to create order", log.Error(err))
		return nil, err
	}

	s.logger.Info("Order created successfully", log.String("order_id", order.ID))

	return &orderv1.CreateOrderResponse{
		Order: &orderv1.Order{
			Id:          order.ID,
			UserId:      order.UserID,
			Status:      statusToProto(order.Status),
			TotalAmount: order.TotalAmount,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// GetOrder retrieves an order by ID
func (s *service) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	s.logger.Info("Getting order", log.String("order_id", req.GetId()))

	if req.GetId() == "" {
		return nil, errors.WithCode(errors.New("id is required"), errors.CodeInvalidInput)
	}

	order, items, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		s.logger.Error("Failed to get order", log.Error(err))
		return nil, err
	}

	// Convert items
	orderItems := make([]*orderv1.OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = &orderv1.OrderItem{
			Id:        item.ID,
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &orderv1.GetOrderResponse{
		Order: &orderv1.Order{
			Id:          order.ID,
			UserId:      order.UserID,
			Status:      statusToProto(order.Status),
			TotalAmount: order.TotalAmount,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
			Items:       orderItems,
		},
	}, nil
}

// ListOrders lists orders with pagination
func (s *service) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	s.logger.Info("Listing orders", log.String("user_id", req.GetUserId()), log.Int32("page_size", req.GetPageSize()))

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse page token
	offset := 0
	if req.GetPageToken() != "" {
		parsedOffset, err := strconv.Atoi(req.GetPageToken())
		if err == nil && parsedOffset > 0 {
			offset = parsedOffset
		}
	}

	var orders []*repository.Order
	var err error

	if req.GetUserId() != "" {
		orders, err = s.repo.GetByUserID(ctx, req.GetUserId(), pageSize, offset)
	} else {
		orders, err = s.repo.List(ctx, pageSize, offset)
	}

	if err != nil {
		s.logger.Error("Failed to list orders", log.Error(err))
		return nil, err
	}

	// Convert to protobuf
	pbOrders := make([]*orderv1.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = &orderv1.Order{
			Id:          order.ID,
			UserId:      order.UserID,
			Status:      statusToProto(order.Status),
			TotalAmount: order.TotalAmount,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		}
	}

	nextPageToken := ""
	if len(orders) == pageSize {
		nextPageToken = strconv.Itoa(offset + pageSize)
	}

	return &orderv1.ListOrdersResponse{
		Orders:        pbOrders,
		NextPageToken: nextPageToken,
	}, nil
}

// UpdateOrderStatus updates the order status
func (s *service) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest) (*orderv1.UpdateOrderStatusResponse, error) {
	s.logger.Info("Updating order status", log.String("order_id", req.GetId()), log.String("status", req.GetStatus().String()))

	if req.GetId() == "" {
		return nil, errors.WithCode(errors.New("id is required"), errors.CodeInvalidInput)
	}

	if req.GetStatus() == orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		return nil, errors.WithCode(errors.New("status is required"), errors.CodeInvalidInput)
	}

	// Convert status to string
	status := statusFromProto(req.GetStatus())

	if err := s.repo.UpdateStatus(ctx, req.GetId(), status); err != nil {
		s.logger.Error("Failed to update order status", log.Error(err))
		return nil, err
	}

	// Get updated order
	order, items, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		s.logger.Error("Failed to get updated order", log.Error(err))
		return nil, err
	}

	// Convert items
	orderItems := make([]*orderv1.OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = &orderv1.OrderItem{
			Id:        item.ID,
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	s.logger.Info("Order status updated successfully", log.String("order_id", order.ID))

	return &orderv1.UpdateOrderStatusResponse{
		Order: &orderv1.Order{
			Id:          order.ID,
			UserId:      order.UserID,
			Status:      statusToProto(order.Status),
			TotalAmount: order.TotalAmount,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
			Items:       orderItems,
		},
	}, nil
}

// CancelOrder cancels an order
func (s *service) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	s.logger.Info("Cancelling order", log.String("order_id", req.GetId()))

	if req.GetId() == "" {
		return nil, errors.WithCode(errors.New("id is required"), errors.CodeInvalidInput)
	}

	// Get order to check status
	order, _, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		s.logger.Error("Failed to get order", log.Error(err))
		return nil, err
	}

	// Check if order can be cancelled
	if order.Status == "cancelled" {
		return nil, errors.WithCode(errors.New("order is already cancelled"), errors.CodeInvalidInput)
	}

	if order.Status == "delivered" {
		return nil, errors.WithCode(errors.New("cannot cancel delivered order"), errors.CodeInvalidInput)
	}

	// Update status to cancelled
	if err := s.repo.UpdateStatus(ctx, req.GetId(), "cancelled"); err != nil {
		s.logger.Error("Failed to cancel order", log.Error(err))
		return nil, err
	}

	s.logger.Info("Order cancelled successfully", log.String("order_id", req.GetId()))

	return &orderv1.CancelOrderResponse{
		Success: true,
	}, nil
}
