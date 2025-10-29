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

	"github.com/kevindiu/monorepo-go-example/internal/db"
	"github.com/kevindiu/monorepo-go-example/internal/errors"
)

// User represents a user entity
type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// UserRepository interface defines user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *db.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(database *db.DB) UserRepository {
	return &userRepository{db: database}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, name, created_at, updated_at
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	row := r.db.QueryRowContext(ctx, query, user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt)

	var created User
	err := row.Scan(&created.ID, &created.Email, &created.Name, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user")
	}

	return &created, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`

	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.WithCode(errors.New("user not found"), errors.CodeNotFound)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user by ID")
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.WithCode(errors.New("user not found"), errors.CodeNotFound)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user by email")
	}

	return &user, nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, email, name, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list users")
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user")
		}
		users = append(users, &user)
	}

	return users, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *User) (*User, error) {
	query := `
		UPDATE users 
		SET email = $2, name = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, email, name, created_at, updated_at
	`

	user.UpdatedAt = time.Now()

	row := r.db.QueryRowContext(ctx, query, user.ID, user.Email, user.Name, user.UpdatedAt)

	var updated User
	err := row.Scan(&updated.ID, &updated.Email, &updated.Name, &updated.CreatedAt, &updated.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.WithCode(errors.New("user not found"), errors.CodeNotFound)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update user")
	}

	return &updated, nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.WithCode(errors.New("user not found"), errors.CodeNotFound)
	}

	return nil
}
