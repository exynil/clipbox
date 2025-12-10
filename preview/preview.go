package preview

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"strings"
	"unicode/utf8"

	_ "image/gif"
	_ "image/jpeg"

	"clipbox/config"
	"clipbox/detect"
	"clipbox/utils"
)

const (
	maskModeFull    = 2
	minMaskLength   = 6
	firstCharsCount = 2
	lastCharsCount  = 4
	maxImageSize    = 1024 * 1024
)

const (
	passwordMaskText = "[[ PASSWORD ]]"
	rofiIconSep      = "\x00icon\x1f"
	rofiInfoSep      = "\x00info\x1f"
)

// MaskPassword masks a password based on the masking mode.
// Mode 1: Shows first 2 and last 4 characters, masks middle with specified color and character.
// Mode 2: Fully masks password as [[ PASSWORD ]]
// If password is shorter than 6 characters in mode 1, returns it unmasked.
func MaskPassword(password string, mode int, maskColor string, maskChar string) string {
	if mode == maskModeFull {
		return passwordMaskText
	}

	// Mode 1: Partial masking
	runes := []rune(password)
	length := len(runes)

	// If password is shorter than minMaskLength characters, show it fully
	if length < minMaskLength {
		return password
	}

	// Extract parts
	firstPart := string(runes[:firstCharsCount])
	lastPart := string(runes[length-lastCharsCount:])
	maskedLength := length - (firstCharsCount + lastCharsCount)
	maskedPart := strings.Repeat(maskChar, maskedLength)

	// Escape parts for Pango markup
	firstPartEscaped := utils.PangoReplacer.Replace(firstPart)
	lastPartEscaped := utils.PangoReplacer.Replace(lastPart)
	maskColorEscaped := utils.PangoReplacer.Replace(maskColor)
	maskedPartEscaped := utils.PangoReplacer.Replace(maskedPart)

	// Combine with colored mask characters
	return fmt.Sprintf("%s<span color='%s'>%s</span>%s",
		firstPartEscaped, maskColorEscaped, maskedPartEscaped, lastPartEscaped)
}

// GeneratePreview creates a complete rofi display line with marker, preview text, and metadata.
func GeneratePreview(id int, content []byte, isPinned int, hasIcon bool, iconPath string, cfg *config.Config) string {
	previewText := generatePreviewText(content, cfg)
	marker := getMarker(isPinned, cfg)
	preview := marker + " " + previewText

	if hasIcon && iconPath != "" {
		hiddenID := utils.EncodeIDHidden(id)
		iconMeta := rofiIconSep + iconPath
		return fmt.Sprintf("%s%s%s%s%d", preview, hiddenID, iconMeta, rofiInfoSep, id)
	}

	return fmt.Sprintf("%s%s%d", preview, rofiInfoSep, id)
}

// generatePreviewText generates the preview text based on content type.
func generatePreviewText(content []byte, cfg *config.Config) string {
	// Try to decode as image first
	limitedReader := io.LimitReader(bytes.NewReader(content), maxImageSize)
	if imgConfig, format, err := image.DecodeConfig(limitedReader); err == nil {
		formatUpper := strings.ToUpper(format)
		return fmt.Sprintf("[[ %s: %dx%d • %s ]]",
			formatUpper, imgConfig.Width, imgConfig.Height, utils.FormatSize(len(content)))
	}

	// Check if binary content
	if !utf8.Valid(content) {
		return fmt.Sprintf("[[ BIN: %s ]]", utils.FormatSize(len(content)))
	}

	// Process text content
	return processTextContent(content, cfg)
}

// processTextContent processes text content, applying password masking if needed.
func processTextContent(content []byte, cfg *config.Config) string {
	text := strings.TrimSpace(string(content))

	// Check if content is a password and masking is enabled
	if cfg.MaskPasswords > 0 && detect.IsPassword(content) {
		text = utils.Trunc(text, cfg.PreviewWidth, "…")
		return MaskPassword(text, cfg.MaskPasswords, cfg.PasswordMaskColor, cfg.PasswordMaskChar)
	}

	// Normal text processing
	text = strings.Join(strings.Fields(text), " ")
	text = utils.Trunc(text, cfg.PreviewWidth, "…")
	return utils.PangoReplacer.Replace(text)
}

// getMarker returns the appropriate marker based on pinned status.
func getMarker(isPinned int, cfg *config.Config) string {
	if isPinned == 1 {
		return cfg.PinnedMarker
	}
	return cfg.UnpinnedMarker
}
