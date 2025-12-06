package image

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"image"
	"io"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
)

// DetectImageFormat checks if data is an image and returns its format (jpg, png, gif)
func DetectImageFormat(data []byte) (string, bool) {
	limitedReader := io.LimitReader(bytes.NewReader(data), 1024*1024)
	_, format, err := image.DecodeConfig(limitedReader)
	if err != nil {
		return "", false
	}
	format = strings.ToLower(format)
	switch format {
	case "jpeg", "jpg":
		return "jpg", true
	case "png":
		return "png", true
	case "gif":
		return "gif", true
	default:
		return format, true
	}
}
