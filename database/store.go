package database

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	"clipbox/config"
	"clipbox/image"
	"clipbox/preview"
)

const maxFileSize = 12 * 1e6 // 12MB

// Store reads content from stdin and saves it to the clipboard database.
// Handles deduplication, icon generation for images, and max items limit.
func Store() error {
	limitedReader := io.LimitReader(os.Stdin, maxFileSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(content) > maxFileSize {
		return nil
	}

	trimmedContent := bytes.TrimSpace(content)
	if len(trimmedContent) == 0 {
		return nil
	}

	// Load config to check MinStoreLength
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.MinStoreLength > 0 && len(trimmedContent) < cfg.MinStoreLength {
		return nil
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

	// Find and delete duplicates (keep their IDs to delete icon files)
	getDuplicatesQuery := `
    SELECT id FROM clipboard
    WHERE buffer_id = ?
    AND content = ?
    AND is_pinned = 0
    ORDER BY id DESC
    LIMIT ?
    `

	dupRows, err := db.Query(getDuplicatesQuery, currentBuffer, content, cfg.MaxDedupeSearch)
	if err != nil {
		return fmt.Errorf("failed to get duplicates: %w", err)
	}

	var duplicateIDs []int
	for dupRows.Next() {
		var id int
		if err := dupRows.Scan(&id); err != nil {
			dupRows.Close()
			return fmt.Errorf("failed to scan duplicate id: %w", err)
		}
		duplicateIDs = append(duplicateIDs, id)
	}
	dupRows.Close()

	if len(duplicateIDs) > 0 {
		placeholders := strings.Repeat("?,", len(duplicateIDs))
		placeholders = placeholders[:len(placeholders)-1]

		deleteQuery := fmt.Sprintf(
			"DELETE FROM clipboard WHERE buffer_id = ? AND content = ? AND is_pinned = 0 AND id IN (%s)",
			placeholders,
		)

		args := make([]interface{}, len(duplicateIDs)+2)
		args[0] = currentBuffer
		args[1] = content
		for i, id := range duplicateIDs {
			args[i+2] = id
		}

		_, err = db.Exec(deleteQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to delete duplicates: %w", err)
		}

		for _, id := range duplicateIDs {
			_ = image.DeleteIconFile(id)
		}
	}

	insertQuery := `
    INSERT INTO clipboard (buffer_id, content, preview) VALUES (?, ?, ?)
    `

	result, err := db.Exec(insertQuery, currentBuffer, content, "")
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}

	var hasIcon bool
	var iconPath string
	if cfg.ShowImageIcons {
		if _, isImage := image.DetectImageFormat(content); isImage {
			path, err := image.ProcessImageIcon(int(insertedID), content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to process image icon: %v\n", err)
			} else {
				hasIcon = true
				iconPath = path
			}
		}
	}

	previewText := preview.GeneratePreview(int(insertedID), content, 0, hasIcon, iconPath, cfg)
	_, err = db.Exec("UPDATE clipboard SET preview = ? WHERE id = ?", previewText, insertedID)
	if err != nil {
		return fmt.Errorf("failed to update preview: %w", err)
	}

	if cfg.MaxItems > 0 {
		if err := EnforceMaxItems(db, currentBuffer, cfg.MaxItems); err != nil {
			return fmt.Errorf("failed to enforce max_items limit: %w", err)
		}
	}

	return nil
}

// EnforceMaxItems removes oldest unpinned entries exceeding maxItems limit.
// Pinned entries are always kept and don't count towards the limit.
func EnforceMaxItems(db *sql.DB, bufferID int, maxItems int) error {
	if maxItems <= 0 {
		return nil
	}

	getIdsToDeleteQuery := `
		SELECT id FROM clipboard
		WHERE buffer_id = ?
		AND is_pinned = 0
		AND id NOT IN (
			SELECT id FROM clipboard
			WHERE buffer_id = ? AND is_pinned = 0
			ORDER BY id DESC
			LIMIT ?
		)
	`

	rows, err := db.Query(getIdsToDeleteQuery, bufferID, bufferID, maxItems)
	if err != nil {
		return fmt.Errorf("failed to query entries to delete: %w", err)
	}
	defer rows.Close()

	var idsToDelete []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		idsToDelete = append(idsToDelete, id)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	if len(idsToDelete) > 0 {
		placeholders := strings.Repeat("?,", len(idsToDelete))
		placeholders = placeholders[:len(placeholders)-1]

		deleteQuery := fmt.Sprintf(
			"DELETE FROM clipboard WHERE buffer_id = ? AND id IN (%s)",
			placeholders,
		)

		args := make([]interface{}, len(idsToDelete)+1)
		args[0] = bufferID
		for i, id := range idsToDelete {
			args[i+1] = id
		}

		_, err = db.Exec(deleteQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to delete excess entries: %w", err)
		}

		for _, id := range idsToDelete {
			_ = image.DeleteIconFile(id)
		}
	}

	return nil
}
