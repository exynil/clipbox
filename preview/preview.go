package preview

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	_ "image/gif"
	_ "image/jpeg"

	"clipbox/config"
	"clipbox/utils"
)

// emailRegex matches basic email format: something@domain.tld
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsPassword determines if content is likely a password based on heuristics
func IsPassword(content []byte) bool {
	if !utf8.Valid(content) {
		return false
	}

	text := string(content)
	text = strings.TrimSpace(text)

	// Check length
	length := len([]rune(text))
	if length < 8 || length > 50 {
		return false
	}

	// Check for spaces
	if strings.Contains(text, " ") {
		return false
	}

	// Check if it's an email address
	if emailRegex.MatchString(text) {
		return false
	}

	// Check for line breaks
	if strings.Contains(text, "\n") || strings.Contains(text, "\r") {
		return false
	}

	// Check that all characters are printable
	for _, r := range text {
		if !unicode.IsPrint(r) {
			return false
		}
	}

	// Check for URL patterns
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") ||
		strings.HasPrefix(text, "ws://") || strings.HasPrefix(text, "wss://") ||
		strings.HasPrefix(text, "ftp://") {
		return false
	}

	// Check for file path patterns
	if strings.Contains(text, "/") && (strings.Count(text, "/") > 2 ||
		strings.HasPrefix(text, "/") || strings.HasPrefix(text, "./")) {
		return false
	}

	if strings.Contains(text, "\\") && strings.Count(text, "\\") > 1 {
		return false
	}

	// Check for at least 3 different character types
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range text {
		if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasSpecial = true
		}
	}

	charTypeCount := 0
	if hasLower {
		charTypeCount++
	}
	if hasUpper {
		charTypeCount++
	}
	if hasDigit {
		charTypeCount++
	}
	if hasSpecial {
		charTypeCount++
	}

	// Require at least 3 different character types
	if charTypeCount < 3 {
		return false
	}

	return true
}

// MaskPassword masks a password based on the masking mode.
// Mode 1: Shows first 2 and last 3 characters, masks middle with specified color and character.
// Mode 2: Fully masks password as [[ PASSWORD ]]
// If password is shorter than 5 characters in mode 1, returns it unmasked.
func MaskPassword(password string, mode int, maskColor string, maskChar string) string {
	if mode == 2 {
		return "[[ PASSWORD ]]"
	}

	// Mode 1: Partial masking
	runes := []rune(password)
	length := len(runes)

	// If password is shorter than 5 characters, show it fully
	if length < 5 {
		return password
	}

	// Extract parts
	firstPart := string(runes[:2])
	lastPart := string(runes[length-3:])
	maskedLength := length - 5
	maskedPart := strings.Repeat(maskChar, maskedLength)

	// Escape parts for Pango markup
	firstPartEscaped := utils.PangoReplacer.Replace(firstPart)
	lastPartEscaped := utils.PangoReplacer.Replace(lastPart)
	maskColorEscaped := utils.PangoReplacer.Replace(maskColor)
	maskedPartEscaped := utils.PangoReplacer.Replace(maskedPart)

	// Combine with colored mask characters
	masked := fmt.Sprintf("%s<span color='%s'>%s</span>%s", firstPartEscaped, maskColorEscaped, maskedPartEscaped, lastPartEscaped)

	return masked
}

// GeneratePreview creates a complete rofi display line with marker, preview text, and metadata
func GeneratePreview(id int, content []byte, isPinned int, hasIcon bool, iconPath string, cfg *config.Config) string {
	var previewText string

	limitedReader := io.LimitReader(bytes.NewReader(content), 1024*1024)
	if imgConfig, format, err := image.DecodeConfig(limitedReader); err == nil {
		formatUpper := strings.ToUpper(format)
		previewText = fmt.Sprintf("[[ %s: %dx%d • %s ]]",
			formatUpper, imgConfig.Width, imgConfig.Height, utils.FormatSize(len(content)))
	} else if !utf8.Valid(content) {
		previewText = fmt.Sprintf("[[ BIN: %s ]]", utils.FormatSize(len(content)))
	} else {
		prev := string(content)
		prev = strings.TrimSpace(prev)

		// Check if content is a password and masking is enabled
		if cfg.MaskPasswords > 0 && IsPassword(content) {
			prev = MaskPassword(prev, cfg.MaskPasswords, cfg.PasswordMaskColor, cfg.PasswordMaskChar)
			prev = utils.Trunc(prev, cfg.PreviewWidth, "…")
		} else {
			prev = strings.Join(strings.Fields(prev), " ")
			prev = utils.Trunc(prev, cfg.PreviewWidth, "…")
			prev = utils.PangoReplacer.Replace(prev)
		}
		previewText = prev
	}

	var marker string
	if isPinned == 1 {
		marker = cfg.PinnedMarker
	} else {
		marker = cfg.UnpinnedMarker
	}

	preview := marker + " " + previewText

	if hasIcon && iconPath != "" {
		hiddenID := utils.EncodeIDHidden(id)
		iconMeta := fmt.Sprintf("\x00icon\x1f%s", iconPath)
		return fmt.Sprintf("%s%s%s\x00info\x1f%d", preview, hiddenID, iconMeta, id)
	}

	return fmt.Sprintf("%s\x00info\x1f%d", preview, id)
}
