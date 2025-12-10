package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import "strings"

var urlSchemes = []string{
	"http://",
	"https://",
	"ftp://",
	"ftps://",
	"ws://",
	"wss://",
	"file://",
	"mailto:",
	"tel:",
}

// IsURL determines if content is a URL.
func IsURL(content []byte) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	textLower := strings.ToLower(text)
	for _, scheme := range urlSchemes {
		if strings.HasPrefix(textLower, scheme) {
			return true
		}
	}

	return false
}
