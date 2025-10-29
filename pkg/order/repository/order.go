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

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/kevindiu/monorepo-go-example/internal/db"
	"github.com/kevindiu/monorepo-go-example/internal/errors"
)

// Order represents an order entity
type Order struct {
	ID          string
	UserID      string
	Status      string
	TotalAmount float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrderItem represents an order item entity
type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int32
	Price     float64
	CreatedAt time.Time
}

// Repository defines the order repository interface
type Repository interface {
	Create(ctx context.Context, order *Order, items []*OrderItem) error
	GetByID(ctx context.Context, id string) (*Order, []*OrderItem, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error)
	List(ctx context.Context, limit, offset int) ([]*Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *db.DB
}

// New creates a new order repository
func New(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

// Create creates a new order with items
func (r *repository) Create(ctx context.Context, order *Order, items []*OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Insert order
	query := `
		INSERT INTO orders (id, user_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now()
	order.ID = uuid.New().String()
	order.CreatedAt = now
	order.UpdatedAt = now

	_, err = tx.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create order")
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, item := range items {
		item.ID = uuid.New().String()
		item.OrderID = order.ID
		item.CreatedAt = now

		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.Price,
			item.CreatedAt,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create order item")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// GetByID retrieves an order by ID with its items
func (r *repository) GetByID(ctx context.Context, id string) (*Order, []*OrderItem, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, errors.WithCode(errors.New("order not found"), errors.CodeNotFound)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get order")
	}

	// Get order items
	itemQuery := `
		SELECT id, order_id, product_id, quantity, price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get order items")
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.CreatedAt,
		); err != nil {
			return nil, nil, errors.Wrap(err, "failed to scan order item")
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "error iterating order items")
	}

	return &order, items, nil
}

// GetByUserID retrieves orders by user ID
func (r *repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list user orders")
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan order")
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating orders")
	}

	return orders, nil
}

// List retrieves all orders with pagination
func (r *repository) List(ctx context.Context, limit, offset int) ([]*Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list orders")
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan order")
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating orders")
	}

	return orders, nil
}

// UpdateStatus updates the order status
func (r *repository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return errors.Wrap(err, "failed to update order status")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get affected rows")
	}

	if rows == 0 {
		return errors.WithCode(errors.New("order not found"), errors.CodeNotFound)
	}

	return nil
}

// Delete deletes an order and its items
func (r *repository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Delete order items first (foreign key constraint)
	itemQuery := `DELETE FROM order_items WHERE order_id = $1`
	_, err = tx.ExecContext(ctx, itemQuery, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete order items")
	}

	// Delete order
	orderQuery := `DELETE FROM orders WHERE id = $1`
	result, err := tx.ExecContext(ctx, orderQuery, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete order")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get affected rows")
	}

	if rows == 0 {
		return errors.WithCode(errors.New("order not found"), errors.CodeNotFound)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}
