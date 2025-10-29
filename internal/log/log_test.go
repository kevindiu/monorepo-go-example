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

package log

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "debug level with console format",
			config: &Config{
				Level:  "debug",
				Format: "console",
			},
			wantErr: false,
		},
		{
			name: "info level with json format",
			config: &Config{
				Level:  "info",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "warn level",
			config: &Config{
				Level:  "warn",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "error level",
			config: &Config{
				Level:  "error",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "default level (invalid)",
			config: &Config{
				Level:  "invalid",
				Format: "json",
			},
			wantErr: false, // falls back to production config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger")
			}
			if logger != nil {
				logger.Sync()
			}
		})
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	if logger == nil {
		t.Error("NewDefault() returned nil")
	}
	defer logger.Sync()

	// Test that we can log without panicking
	logger.Info("test message")
}

func TestLoggerMethods(t *testing.T) {
	logger := NewDefault()
	defer logger.Sync()

	// Test Named
	namedLogger := logger.Named("test")
	if namedLogger == nil {
		t.Error("Named() returned nil")
	}

	// Test With
	childLogger := logger.With(String("key", "value"))
	if childLogger == nil {
		t.Error("With() returned nil")
	}
}

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "String field",
			fn: func() {
				_ = String("key", "value")
			},
		},
		{
			name: "Int field",
			fn: func() {
				_ = Int("key", 42)
			},
		},
		{
			name: "Int32 field",
			fn: func() {
				_ = Int32("key", int32(42))
			},
		},
		{
			name: "Int64 field",
			fn: func() {
				_ = Int64("key", int64(42))
			},
		},
		{
			name: "Error field",
			fn: func() {
				_ = Error(nil)
			},
		},
		{
			name: "Any field",
			fn: func() {
				_ = Any("key", "value")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.fn()
		})
	}
}
