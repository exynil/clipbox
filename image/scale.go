package image

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"image"
	"image/color"
)

// ScaleImage scales an image using bilinear interpolation
func ScaleImage(dst *image.RGBA, src image.Image) {
	dstBounds := dst.Bounds()
	srcBounds := src.Bounds()

	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())
	dstW := float64(dstBounds.Dx())
	dstH := float64(dstBounds.Dy())

	for y := dstBounds.Min.Y; y < dstBounds.Max.Y; y++ {
		for x := dstBounds.Min.X; x < dstBounds.Max.X; x++ {
			// Map destination coordinates to source coordinates
			srcX := (float64(x-dstBounds.Min.X)+0.5)*srcW/dstW - 0.5
			srcY := (float64(y-dstBounds.Min.Y)+0.5)*srcH/dstH - 0.5

			// Get integer and fractional parts
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			// Clamp to source bounds
			if x0 < srcBounds.Min.X {
				x0 = srcBounds.Min.X
			}
			if x0 >= srcBounds.Max.X {
				x0 = srcBounds.Max.X - 1
			}
			if x1 >= srcBounds.Max.X {
				x1 = srcBounds.Max.X - 1
			}
			if y0 < srcBounds.Min.Y {
				y0 = srcBounds.Min.Y
			}
			if y0 >= srcBounds.Max.Y {
				y0 = srcBounds.Max.Y - 1
			}
			if y1 >= srcBounds.Max.Y {
				y1 = srcBounds.Max.Y - 1
			}

			// Get fractional parts
			fx := srcX - float64(x0)
			fy := srcY - float64(y0)

			// Get four corner pixels
			c00 := GetRGBA(src.At(x0, y0))
			c10 := GetRGBA(src.At(x1, y0))
			c01 := GetRGBA(src.At(x0, y1))
			c11 := GetRGBA(src.At(x1, y1))

			// Bilinear interpolation
			c0 := LerpColor(c00, c10, fx)
			c1 := LerpColor(c01, c11, fx)
			c := LerpColor(c0, c1, fy)

			dst.Set(x, y, c)
		}
	}
}

// GetRGBA converts a color.Color to color.RGBA
func GetRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

// LerpColor linearly interpolates between two colors
func LerpColor(c0, c1 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c0.R)*(1-t) + float64(c1.R)*t),
		G: uint8(float64(c0.G)*(1-t) + float64(c1.G)*t),
		B: uint8(float64(c0.B)*(1-t) + float64(c1.B)*t),
		A: uint8(float64(c0.A)*(1-t) + float64(c1.A)*t),
	}
}
