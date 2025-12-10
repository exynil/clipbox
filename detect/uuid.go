package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import "regexp"

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID determines if content is a UUID.
func IsUUID(content []byte) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	return uuidRegex.MatchString(text)
}
