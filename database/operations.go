package database

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"

	"clipbox/config"
	"clipbox/image"
	"clipbox/preview"
)

// GetContentByID retrieves the content of a clipboard entry by its ID
func GetContentByID(id int) ([]byte, error) {
	db, err := OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var content []byte
	err = db.QueryRow("SELECT content FROM clipboard WHERE id = ?", id).Scan(&content)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	return content, nil
}

// SwitchBuffer changes the active buffer to the specified ID (1-5)
func SwitchBuffer(bufferID int) error {
	if bufferID < 1 || bufferID > 5 {
		return fmt.Errorf("buffer_id must be between 1 and 5, got %d", bufferID)
	}

	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT OR REPLACE INTO current_buffer (id, buffer_id) VALUES (1, ?)", bufferID)
	if err != nil {
		return fmt.Errorf("failed to switch buffer: %w", err)
	}

	return nil
}

// SwitchToNextBuffer switches to the next buffer cyclically (1->2->3->4->5->1)
func SwitchToNextBuffer() error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	currentBuffer, err := GetCurrentBuffer(db)
	if err != nil {
		return fmt.Errorf("failed to get current buffer: %w", err)
	}

	nextBuffer := currentBuffer + 1
	if nextBuffer > 5 {
		nextBuffer = 1
	}

	_, err = db.Exec("INSERT OR REPLACE INTO current_buffer (id, buffer_id) VALUES (1, ?)", nextBuffer)
	if err != nil {
		return fmt.Errorf("failed to switch buffer: %w", err)
	}

	return nil
}

// SwitchToPreviousBuffer switches to the previous buffer cyclically (1->5->4->3->2->1)
func SwitchToPreviousBuffer() error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	currentBuffer, err := GetCurrentBuffer(db)
	if err != nil {
		return fmt.Errorf("failed to get current buffer: %w", err)
	}

	prevBuffer := currentBuffer - 1
	if prevBuffer < 1 {
		prevBuffer = 5
	}

	_, err = db.Exec("INSERT OR REPLACE INTO current_buffer (id, buffer_id) VALUES (1, ?)", prevBuffer)
	if err != nil {
		return fmt.Errorf("failed to switch buffer: %w", err)
	}

	return nil
}

// TogglePin toggles the pinned status of an entry and regenerates its preview
func TogglePin(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %d", id)
	}

	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var currentPinned int
	var content []byte
	err = db.QueryRow("SELECT is_pinned, content FROM clipboard WHERE id = ?", id).Scan(&currentPinned, &content)
	if err != nil {
		return fmt.Errorf("failed to get current pinned status: %w", err)
	}

	var newPinned int
	if currentPinned == 1 {
		newPinned = 0
	} else {
		newPinned = 1
	}

	var hasIcon bool
	var iconPath string
	if cfg.ShowImageIcons {
		if path, ok := image.GetIconPath(id); ok {
			hasIcon = true
			iconPath = path
		}
	}

	previewText := preview.GeneratePreview(id, content, newPinned, hasIcon, iconPath, cfg)
	_, err = db.Exec("UPDATE clipboard SET is_pinned = ?, preview = ? WHERE id = ?", newPinned, previewText, id)
	if err != nil {
		return fmt.Errorf("failed to toggle pin: %w", err)
	}

	return nil
}

// DeleteEntry deletes a clipboard entry by its ID and removes associated icon file
func DeleteEntry(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %d", id)
	}

	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM clipboard WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("entry with id %d not found", id)
	}

	_ = image.DeleteIconFile(id)

	return nil
}
