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

package config

import (
	"testing"
)

func TestBind(t *testing.T) {
	cfg := &Config{}
	err := cfg.Bind()
	if err != nil {
		t.Errorf("Bind() error = %v", err)
	}

	// Check defaults are set
	if cfg.Server == nil {
		t.Error("Bind() did not initialize Server")
	}
	if cfg.Database == nil {
		t.Error("Bind() did not initialize Database")
	}
	if cfg.Log == nil {
		t.Error("Bind() did not initialize Log")
	}
}

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Error("Load() returned nil config")
	}

	// Verify default values
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Load() Server.Host = %v, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Load() Server.Port = %v, want 8080", cfg.Server.Port)
	}
}

func TestGetDSN(t *testing.T) {
	db := &Database{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "pass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	dsn := db.GetDSN()
	expected := "host=localhost port=5432 user=test password=pass dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Errorf("GetDSN() = %v, want %v", dsn, expected)
	}
}

func TestGetServerAddr(t *testing.T) {
	server := &Server{
		Host: "localhost",
		Port: 8080,
	}

	addr := server.GetServerAddr()
	expected := "localhost:8080"
	if addr != expected {
		t.Errorf("GetServerAddr() = %v, want %v", addr, expected)
	}
}

func TestGetGRPCAddr(t *testing.T) {
	server := &Server{
		Host:     "localhost",
		GRPCPort: 9090,
	}

	addr := server.GetGRPCAddr()
	expected := "localhost:9090"
	if addr != expected {
		t.Errorf("GetGRPCAddr() = %v, want %v", addr, expected)
	}
}
