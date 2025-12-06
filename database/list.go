package database

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"strings"

	"clipbox/config"
)

// List outputs clipboard entries in rofi script mode format.
// Entries are sorted by ID (newest first), with pinned entries repeated at the end.
func List(limit int) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if limit <= 0 {
		limit = cfg.Limit
	}

	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	currentBuffer, err := GetCurrentBuffer(db)
	if err != nil {
		return fmt.Errorf("failed to get current buffer: %w", err)
	}

	// Rofi script mode configuration
	fmt.Print("\x00use-hot-keys\x1ftrue\n")
	fmt.Print("\x00keep-selection\x1ftrue\n")
	fmt.Print("\x00markup-rows\x1ftrue\n")

	var bufferName string
	if currentBuffer >= 1 && currentBuffer <= 5 {
		bufferName = cfg.BufferNames[currentBuffer-1]
	}
	if bufferName == "" {
		bufferName = fmt.Sprintf("Buffer %d", currentBuffer)
	}
	fmt.Printf("\x00prompt\x1f%s\n", bufferName)

	hasRows := false
	hasPinnedRows := false

	// Query all entries with limit
	query := `
    SELECT preview FROM clipboard
    WHERE buffer_id = ?
    ORDER BY id DESC
    LIMIT ?
    `

	rows, err := db.Query(query, currentBuffer, limit)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var preview string
		if err := rows.Scan(&preview); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		hasRows = true
		fmt.Printf("%s\n", preview)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	// Query pinned entries for the "pinned section" at the end
	queryPinned := `
    SELECT preview FROM clipboard
    WHERE buffer_id = ? AND is_pinned = 1
    ORDER BY id DESC
    LIMIT 1000
    `

	rowsPinned, err := db.Query(queryPinned, currentBuffer)
	if err != nil {
		return fmt.Errorf("failed to query pinned: %w", err)
	}
	defer rowsPinned.Close()

	for rowsPinned.Next() {
		var preview string
		if err := rowsPinned.Scan(&preview); err != nil {
			return fmt.Errorf("failed to scan pinned row: %w", err)
		}
		if !hasPinnedRows {
			hasPinnedRows = true
			if hasRows {
				separator := strings.Repeat("─", cfg.SeparatorLength)
				fmt.Printf("%s\x00info\x1f0\n", separator)
			}
		}
		fmt.Printf("%s\n", preview)
	}
	if err := rowsPinned.Err(); err != nil {
		return fmt.Errorf("error iterating pinned rows: %w", err)
	}

	if !hasRows && !hasPinnedRows {
		fmt.Printf(" (No entries in buffer %d)\x00info\x1f0\n", currentBuffer)
	}

	return nil
}
