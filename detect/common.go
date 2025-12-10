package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"strings"
	"unicode/utf8"
)

// validateAndTrim validates UTF-8 encoding and returns trimmed text.
// Returns empty string and false if content is invalid UTF-8 or empty after trimming.
func validateAndTrim(content []byte) (string, bool) {
	if !utf8.Valid(content) {
		return "", false
	}

	text := strings.TrimSpace(string(content))
	if text == "" {
		return "", false
	}

	return text, true
}
