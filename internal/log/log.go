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
	"go.uber.org/zap"
)

// Logger wraps zap logger
type Logger struct {
	*zap.Logger
}

// Config represents logger configuration
type Config struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// New creates a new logger instance
func New(cfg *Config) (*Logger, error) {
	var zapConfig zap.Config

	switch cfg.Level {
	case "debug":
		zapConfig = zap.NewDevelopmentConfig()
	case "info", "warn", "error":
		zapConfig = zap.NewProductionConfig()
	default:
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	switch cfg.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	// Set format
	if cfg.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: logger}, nil
}

// NewDefault creates a default logger
func NewDefault() *Logger {
	logger, _ := zap.NewProduction()
	return &Logger{Logger: logger}
}

// With creates a child logger with additional fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// Named creates a named logger
func (l *Logger) Named(name string) *Logger {
	return &Logger{Logger: l.Logger.Named(name)}
}

// String creates a string field for structured logging
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int creates an int field for structured logging
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int32 creates an int32 field for structured logging
func Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)
}

// Int64 creates an int64 field for structured logging
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Error creates an error field for structured logging
func Error(err error) zap.Field {
	return zap.Error(err)
}

// Any creates a field for any type
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// Duration creates a duration field
func Duration(key string, val interface{}) zap.Field {
	if d, ok := val.(interface{ Duration() interface{} }); ok {
		return zap.Any(key, d.Duration())
	}
	return zap.Any(key, val)
}
