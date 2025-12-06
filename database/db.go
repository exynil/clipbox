package database

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"clipbox/config"
)

// GetDBPath returns the path to the SQLite database file.
// Uses custom path from config if set, otherwise defaults to $XDG_CACHE_HOME/clipbox/clipbox.db
func GetDBPath() (string, error) {
	return config.GetDBPath()
}

// InitDB creates tables and indexes if they don't exist
func InitDB(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS clipboard (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        buffer_id INTEGER NOT NULL CHECK(buffer_id IN (1, 2, 3, 4, 5)),
        is_pinned INTEGER DEFAULT 0,
        preview TEXT NOT NULL DEFAULT '',
        content BLOB NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS current_buffer (
        id INTEGER PRIMARY KEY CHECK(id = 1),
        buffer_id INTEGER DEFAULT 1
    );
    CREATE INDEX IF NOT EXISTS idx_buffer_pinned ON clipboard(buffer_id, is_pinned DESC);
    CREATE INDEX IF NOT EXISTS idx_buffer_id ON clipboard(buffer_id);
    CREATE INDEX IF NOT EXISTS idx_pinned ON clipboard(is_pinned);
    CREATE INDEX IF NOT EXISTS idx_content ON clipboard(content);
    DELETE FROM current_buffer WHERE id != 1;
    INSERT OR IGNORE INTO current_buffer (id, buffer_id) VALUES (1, 1);
    `

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// GetCurrentBuffer returns the currently active buffer ID (1-5)
func GetCurrentBuffer(db *sql.DB) (int, error) {
	var bufferID int
	err := db.QueryRow("SELECT buffer_id FROM current_buffer LIMIT 1").Scan(&bufferID)
	if err != nil {
		return 1, fmt.Errorf("failed to get current buffer: %w", err)
	}
	return bufferID, nil
}

// OpenDB opens the SQLite database and initializes it if needed
func OpenDB() (*sql.DB, error) {
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := InitDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
