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
	"fmt"

	"github.com/google/uuid"
	"github.com/kevindiu/monorepo-go-example/internal/errors"
	"github.com/kevindiu/monorepo-go-example/pkg/user/repository"
)

// UserService interface defines user business logic operations
type UserService interface {
	CreateUser(ctx context.Context, email, name string) (*repository.User, error)
	GetUser(ctx context.Context, id string) (*repository.User, error)
	ListUsers(ctx context.Context, pageSize int, pageToken string) ([]*repository.User, string, error)
	UpdateUser(ctx context.Context, id, email, name string) (*repository.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, email, name string) (*repository.User, error) {
	// Validate input
	if email == "" {
		return nil, errors.WithCode(errors.New("email is required"), errors.CodeInvalidInput)
	}
	if name == "" {
		return nil, errors.WithCode(errors.New("name is required"), errors.CodeInvalidInput)
	}

	// Check if user with email already exists
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, errors.WithCode(errors.New("user with this email already exists"), errors.CodeConflict)
	}

	// Create user
	user := &repository.User{
		ID:    uuid.New().String(),
		Email: email,
		Name:  name,
	}

	return s.repo.Create(ctx, user)
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*repository.User, error) {
	if id == "" {
		return nil, errors.WithCode(errors.New("user ID is required"), errors.CodeInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

// ListUsers retrieves users with pagination
func (s *userService) ListUsers(ctx context.Context, pageSize int, pageToken string) ([]*repository.User, string, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := 0
	if pageToken != "" {
		// In a real implementation, you'd decode the page token
		// For simplicity, we'll just use a basic offset
		fmt.Sscanf(pageToken, "%d", &offset)
	}

	users, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, "", err
	}

	// Generate next page token
	nextPageToken := ""
	if len(users) == pageSize {
		nextPageToken = fmt.Sprintf("%d", offset+pageSize)
	}

	return users, nextPageToken, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id, email, name string) (*repository.User, error) {
	if id == "" {
		return nil, errors.WithCode(errors.New("user ID is required"), errors.CodeInvalidInput)
	}
	if email == "" {
		return nil, errors.WithCode(errors.New("email is required"), errors.CodeInvalidInput)
	}
	if name == "" {
		return nil, errors.WithCode(errors.New("name is required"), errors.CodeInvalidInput)
	}

	// Check if user exists
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if email is taken by another user
	if user.Email != email {
		existing, err := s.repo.GetByEmail(ctx, email)
		if err == nil && existing != nil && existing.ID != id {
			return nil, errors.WithCode(errors.New("email is already taken"), errors.CodeConflict)
		}
	}

	// Update user
	user.Email = email
	user.Name = name

	return s.repo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.WithCode(errors.New("user ID is required"), errors.CodeInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}
