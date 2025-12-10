package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"path/filepath"
	"strings"
	"unicode"
)

const minPathComponents = 2

// IsFilePath determines if content is a file path.
func IsFilePath(content []byte) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	// Normalize the path
	cleaned := filepath.Clean(text)
	if cleaned == "." || cleaned == ".." {
		return false
	}

	// Check for absolute paths (Unix and Windows)
	if filepath.IsAbs(text) {
		return true
	}

	// Check for relative paths starting with ./
	if strings.HasPrefix(text, "./") || strings.HasPrefix(text, ".\\") {
		return true
	}

	// Check for parent directory references
	if strings.HasPrefix(text, "../") || strings.HasPrefix(text, "..\\") {
		return true
	}

	// Check for Windows drive letters (C:, D:, etc.)
	if isWindowsDrive(text) {
		return true
	}

	// Check for UNC paths (\\server\share)
	if strings.HasPrefix(text, "\\\\") {
		return true
	}

	// Check if path contains path separators and has multiple components
	return hasPathSeparators(text)
}

// isWindowsDrive checks if text starts with a Windows drive letter (C:, D:, etc.).
func isWindowsDrive(text string) bool {
	if len(text) < 2 || text[1] != ':' {
		return false
	}

	if !unicode.IsLetter(rune(text[0])) {
		return false
	}

	// Drive letter only (C:) or followed by separator (C:\ or C:/)
	return len(text) == 2 || text[2] == '/' || text[2] == '\\'
}

// hasPathSeparators checks if text contains path separators and has valid path structure.
func hasPathSeparators(text string) bool {
	hasUnixSep := strings.Contains(text, "/")
	hasWinSep := strings.Contains(text, "\\")

	if !hasUnixSep && !hasWinSep {
		return false
	}

	// Count path components
	components := strings.FieldsFunc(text, func(r rune) bool {
		return r == '/' || r == '\\'
	})

	// If there are multiple components, it's likely a path
	if len(components) >= minPathComponents {
		return true
	}

	// Single component with separator might be a path if it starts with separator
	return (hasUnixSep && strings.HasPrefix(text, "/")) ||
		(hasWinSep && strings.HasPrefix(text, "\\"))
}
