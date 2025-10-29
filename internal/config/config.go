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
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// GlobalConfig represents the global configuration interface
type GlobalConfig interface {
	Bind() error
}

// Config holds the entire application configuration
type Config struct {
	Server   *Server   `yaml:"server" mapstructure:"server"`
	Database *Database `yaml:"database" mapstructure:"database"`
	Log      *Log      `yaml:"log" mapstructure:"log"`
}

// Server configuration
type Server struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	GRPCPort int    `yaml:"grpc_port" mapstructure:"grpc_port"`
	Mode     string `yaml:"mode" mapstructure:"mode"`
}

// Database configuration
type Database struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Name     string `yaml:"name" mapstructure:"name"`
	SSLMode  string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
}

// Log configuration
type Log struct {
	Level  string `yaml:"level" mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
}

// Bind binds environment variables to config struct
func (c *Config) Bind() error {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set default values
	setDefaults(v)

	// Bind environment variables
	bindEnvs(v, "", reflect.TypeOf(*c))

	// Unmarshal into struct
	return v.Unmarshal(c)
}

// GetDSN returns database connection string
func (d *Database) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetServerAddr returns server address
func (s *Server) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetGRPCAddr returns gRPC server address
func (s *Server) GetGRPCAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.GRPCPort)
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := cfg.Bind(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return cfg, nil
}

// LoadFromFile loads configuration from YAML file
func LoadFromFile(filename string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filename)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// ToYAML converts config to YAML string
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	return string(data), nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("server.mode", "development")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "monorepo")
	v.SetDefault("database.ssl_mode", "disable")

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}

func bindEnvs(v *viper.Viper, prefix string, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		if field.Type.Kind() == reflect.Struct {
			bindEnvs(v, key, field.Type)
		} else {
			v.BindEnv(key)
		}
	}
}
