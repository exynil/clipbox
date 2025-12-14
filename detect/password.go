package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 50
	minCharTypes      = 3
)

// IsPassword determines if content is likely a password based on heuristics.
func IsPassword(content []byte, ignorePatterns []string) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	// Check length
	length := len([]rune(text))
	if length < minPasswordLength || length > maxPasswordLength {
		return false
	}

	// Check if it's a date/time (before space check, as dates can contain spaces)
	if IsDateTime(content) {
		return false
	}

	// Check for spaces
	if strings.Contains(text, " ") {
		return false
	}

	// Check if it's an email address
	if IsEmail(content) {
		return false
	}

	// Check if it's a URL
	if IsURL(content) {
		return false
	}

	// Check if it's an IP address (IPv4 or IPv6)
	if IsIP(content) {
		return false
	}

	// Check if it's a UUID
	if IsUUID(content) {
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

	// Check for file path patterns
	if IsFilePath(content) {
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

	// Require at least minCharTypes different character types
	if charTypeCount < minCharTypes {
		return false
	}

	// Check user-defined ignore patterns
	for _, pattern := range ignorePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Skip invalid regex patterns
			continue
		}
		if re.MatchString(text) {
			return false
		}
	}

	return true
}
