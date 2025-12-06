package utils

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"strings"
)

// PangoReplacer escapes Pango markup characters in a single pass
var PangoReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
)

// Trunc truncates a string to max runes and appends ellipsis if truncated
func Trunc(in string, max int, ellip string) string {
	runes := []rune(in)
	if len(runes) > max {
		return string(runes[:max]) + ellip
	}
	return in
}

// FormatSize converts bytes to human-readable format (B, KiB, MiB, GiB)
func FormatSize(size int) string {
	units := []string{"B", "KiB", "MiB", "GiB"}
	var i int
	fsize := float64(size)
	for fsize >= 1024 && i < len(units)-1 {
		fsize /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", fsize, units[i])
}
