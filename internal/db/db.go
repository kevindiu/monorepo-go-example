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

package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kevindiu/monorepo-go-example/internal/config"
	"github.com/kevindiu/monorepo-go-example/internal/errors"
	_ "github.com/lib/pq"
)

// DB wraps database connection
type DB struct {
	*sql.DB
}

// Connect establishes database connection
func Connect(cfg *config.Database) (*DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database connection")
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	return &DB{DB: db}, nil
}

// Close closes database connection
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	return tx, nil
}

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrate runs database migrations
func (db *DB) Migrate(migrations []Migration) error {
	// Create migrations table if not exists
	createTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(createTable); err != nil {
		return errors.Wrap(err, "failed to create migrations table")
	}

	// Get applied migrations
	applied := make(map[int]bool)
	rows, err := db.Query("SELECT version FROM migrations")
	if err != nil {
		return errors.Wrap(err, "failed to query applied migrations")
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return errors.Wrap(err, "failed to scan migration version")
		}
		applied[version] = true
	}

	// Run unapplied migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			return errors.Wrap(err, "failed to start migration transaction")
		}

		// Execute migration SQL
		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to execute migration %d: %s", migration.Version, migration.Name)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO migrations (version, name) VALUES ($1, $2)", migration.Version, migration.Name); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to record migration %d", migration.Version)
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrapf(err, "failed to commit migration %d", migration.Version)
		}

		fmt.Printf("Applied migration %d: %s\n", migration.Version, migration.Name)
	}

	return nil
}
