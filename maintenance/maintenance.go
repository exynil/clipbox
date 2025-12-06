package maintenance

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"os"

	"clipbox/config"
	"clipbox/database"
	"clipbox/image"
	"clipbox/preview"
	"clipbox/utils"
)

// VacuumDB runs VACUUM on the database to reclaim unused space
func VacuumDB() error {
	dbPath, err := config.GetDBPath()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("failed to get database file info: %w", err)
	}
	sizeBefore := fileInfo.Size()

	fmt.Fprintf(os.Stderr, "Database size before VACUUM: %s\n", utils.FormatSize(int(sizeBefore)))
	fmt.Fprintf(os.Stderr, "Running VACUUM...\n")

	db, err := database.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to execute VACUUM: %w", err)
	}

	fileInfo, err = os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("failed to get database file info after VACUUM: %w", err)
	}
	sizeAfter := fileInfo.Size()

	fmt.Fprintf(os.Stderr, "Database size after VACUUM: %s\n", utils.FormatSize(int(sizeAfter)))
	freed := sizeBefore - sizeAfter
	if freed > 0 {
		fmt.Fprintf(os.Stderr, "Freed: %s\n", utils.FormatSize(int(freed)))
	}
	fmt.Fprintf(os.Stderr, "VACUUM completed successfully\n")
	return nil
}

// RebuildAllPreviews regenerates preview strings for all entries using current config
func RebuildAllPreviews() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := `SELECT id, is_pinned, content FROM clipboard`
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query entries: %w", err)
	}

	type entry struct {
		id       int
		isPinned int
		content  []byte
	}

	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.isPinned, &e.content); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan row: %w", err)
		}
		entries = append(entries, e)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	updateStmt, err := db.Prepare("UPDATE clipboard SET preview = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	updatedCount := 0
	for _, e := range entries {
		var hasIcon bool
		var iconPath string
		if cfg.ShowImageIcons {
			// Check if icon already exists
			if path, ok := image.GetIconPath(e.id); ok {
				hasIcon = true
				iconPath = path
			} else {
				// Icon doesn't exist, check if content is an image and create icon
				if _, isImage := image.DetectImageFormat(e.content); isImage {
					path, err := image.ProcessImageIcon(e.id, e.content)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to process image icon for id %d: %v\n", e.id, err)
					} else {
						hasIcon = true
						iconPath = path
					}
				}
			}
		}

		previewText := preview.GeneratePreview(e.id, e.content, e.isPinned, hasIcon, iconPath, cfg)
		_, err := updateStmt.Exec(previewText, e.id)
		if err != nil {
			return fmt.Errorf("failed to update preview for id %d: %w", e.id, err)
		}

		updatedCount++
	}

	fmt.Fprintf(os.Stderr, "Updated %d previews\n", updatedCount)
	return nil
}
