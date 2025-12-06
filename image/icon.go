package image

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"clipbox/config"
)

const maxIconSize = 64 // Maximum icon size in pixels

// GetIconsDir returns the path to the icons directory (next to the database file)
func GetIconsDir() (string, error) {
	dbPath, err := config.GetDBPath()
	if err != nil {
		return "", err
	}
	dbDir := filepath.Dir(dbPath)
	iconsDir := filepath.Join(dbDir, "icons")
	return iconsDir, nil
}

// DeleteIconFile removes the icon file for a given entry ID
func DeleteIconFile(id int) error {
	iconsDir, err := GetIconsDir()
	if err != nil {
		return err
	}
	iconPath := filepath.Join(iconsDir, fmt.Sprintf("%d.png", id))
	if err := os.Remove(iconPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete icon file: %w", err)
	}
	return nil
}

// GetIconPath checks if an icon file exists for a given ID.
// Returns (absolutePath, true) if exists, ("", false) otherwise.
func GetIconPath(id int) (string, bool) {
	iconsDir, err := GetIconsDir()
	if err != nil {
		return "", false
	}
	iconPathCheck := filepath.Join(iconsDir, fmt.Sprintf("%d.png", id))
	if _, err := os.Stat(iconPathCheck); err != nil {
		return "", false
	}
	absPath, err := filepath.Abs(iconPathCheck)
	if err != nil {
		return "", false
	}
	return absPath, true
}

// ProcessImageIcon resizes an image to max 64x64 and saves it as a PNG icon
func ProcessImageIcon(id int, data []byte) (string, error) {
	iconsDir, err := GetIconsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get icons dir: %w", err)
	}

	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create icons dir: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var newWidth, newHeight int
	if width > height {
		if width > maxIconSize {
			newWidth = maxIconSize
			newHeight = height * maxIconSize / width
		} else {
			newWidth = width
			newHeight = height
		}
	} else {
		if height > maxIconSize {
			newHeight = maxIconSize
			newWidth = width * maxIconSize / height
		} else {
			newWidth = width
			newHeight = height
		}
	}

	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	ScaleImage(resized, img)

	iconPath := filepath.Join(iconsDir, fmt.Sprintf("%d.png", id))
	file, err := os.Create(iconPath)
	if err != nil {
		return "", fmt.Errorf("failed to create icon file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, resized); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	absPath, err := filepath.Abs(iconPath)
	if err != nil {
		return iconPath, nil
	}

	return absPath, nil
}
