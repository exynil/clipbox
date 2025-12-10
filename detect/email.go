package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import "regexp"

// emailRegex matches basic email format: something@domain.tld
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsEmail determines if content is an email address.
func IsEmail(content []byte) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	return emailRegex.MatchString(text)
}
