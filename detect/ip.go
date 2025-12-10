package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import "net"

// IsIP determines if content is an IP address (IPv4 or IPv6).
func IsIP(content []byte) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	return net.ParseIP(text) != nil
}
