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

	"github.com/kevindiu/monorepo-go-example/internal/errors"
	"github.com/kevindiu/monorepo-go-example/pkg/user/repository"
)

// mockUserRepository is a mock implementation of repository.UserRepository
type mockUserRepository struct {
	users map[string]*repository.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*repository.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *repository.User) (*repository.User, error) {
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*repository.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.WithCode(errors.New("user not found"), errors.CodeNotFound)
	}
	return user, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*repository.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *repository.User) (*repository.User, error) {
	if _, ok := m.users[user.ID]; !ok {
		return nil, errors.WithCode(errors.New("user not found"), errors.CodeNotFound)
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*repository.User, error) {
	users := make([]*repository.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func TestNewUserService(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)
	if svc == nil {
		t.Fatal("NewUserService returned nil service")
	}
}

func TestCreateUser(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	tests := []struct {
		name     string
		email    string
		userName string
		wantErr  bool
	}{
		{
			name:     "valid user",
			email:    "test@example.com",
			userName: "Test User",
			wantErr:  false,
		},
		{
			name:     "empty email",
			email:    "",
			userName: "Test User",
			wantErr:  true,
		},
		{
			name:     "empty name",
			email:    "test@example.com",
			userName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := svc.CreateUser(context.Background(), tt.email, tt.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && user == nil {
				t.Error("CreateUser() returned nil user for valid request")
			}
			if !tt.wantErr && user.Email != tt.email {
				t.Errorf("CreateUser() user.Email = %v, want %v", user.Email, tt.email)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	// Create a user first
	createdUser, err := svc.CreateUser(context.Background(), "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "existing user",
			userID:  createdUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			userID:  "non-existent-id",
			wantErr: true,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := svc.GetUser(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && user == nil {
				t.Error("GetUser() returned nil user for existing user")
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	// Create a user first
	createdUser, err := svc.CreateUser(context.Background(), "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	updatedName := "Updated Name"
	updatedEmail := "updated@example.com"
	updatedUser, err := svc.UpdateUser(context.Background(), createdUser.ID, updatedEmail, updatedName)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if updatedUser.Name != updatedName {
		t.Errorf("UpdateUser() user.Name = %v, want %v", updatedUser.Name, updatedName)
	}
	if updatedUser.Email != updatedEmail {
		t.Errorf("UpdateUser() user.Email = %v, want %v", updatedUser.Email, updatedEmail)
	}
}

func TestDeleteUser(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	// Create a user first
	createdUser, err := svc.CreateUser(context.Background(), "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	err = svc.DeleteUser(context.Background(), createdUser.ID)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	// Verify user is deleted
	_, err = svc.GetUser(context.Background(), createdUser.ID)
	if err == nil {
		t.Error("GetUser() succeeded for deleted user, want error")
	}
}

func TestListUsers(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	// Create some users
	_, err := svc.CreateUser(context.Background(), "user1@example.com", "User 1")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	_, err = svc.CreateUser(context.Background(), "user2@example.com", "User 2")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	users, _, err := svc.ListUsers(context.Background(), 10, "")
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("ListUsers() returned %d users, want 2", len(users))
	}
}
