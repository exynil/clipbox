package utils

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// CopyToClipboard copies content to the Wayland clipboard using wl-copy
func CopyToClipboard(content []byte) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = bytes.NewReader(content)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

// ExtractID gets entry ID from ROFI_INFO env var or from hidden encoding in input
func ExtractID(input string) (int, error) {
	if rofiInfo := os.Getenv("ROFI_INFO"); rofiInfo != "" {
		id, err := strconv.Atoi(rofiInfo)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	if id, ok := DecodeIDHidden(input); ok {
		return id, nil
	}

	return 0, fmt.Errorf("input not prefixed with id and ROFI_INFO not set")
}

// EncodeIDHidden encodes an ID as invisible Unicode characters.
// Uses variation selectors (U+FE00-U+FE09) surrounded by zero-width spaces (U+200B).
func EncodeIDHidden(id int) string {
	var result strings.Builder
	result.WriteString("\u200b")
	idStr := strconv.Itoa(id)
	for _, r := range idStr {
		if r >= '0' && r <= '9' {
			vs := rune(0xFE00 + (r - '0'))
			result.WriteRune(vs)
		}
	}
	result.WriteString("\u200b")
	return result.String()
}

// DecodeIDHidden extracts an ID encoded by EncodeIDHidden from input string
func DecodeIDHidden(input string) (int, bool) {
	zwsp := "\u200b"
	startIdx := strings.Index(input, zwsp)
	if startIdx == -1 {
		return 0, false
	}
	afterStart := input[startIdx+len(zwsp):]
	endIdx := strings.Index(afterStart, zwsp)
	if endIdx == -1 {
		return 0, false
	}
	encoded := afterStart[:endIdx]

	var digits strings.Builder
	for _, r := range encoded {
		if r >= 0xFE00 && r <= 0xFE09 {
			digit := rune('0' + (r - 0xFE00))
			digits.WriteRune(digit)
		}
	}
	if digits.Len() == 0 {
		return 0, false
	}
	id, err := strconv.Atoi(digits.String())
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
