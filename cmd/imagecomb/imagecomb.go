package imagecomb

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math"
	"strings"

	xdraw "golang.org/x/image/draw"
)

const (
	trimThreshold = 8
	minPixel      = 50
	quality       = 100
	targetW       = 1920
	targetH       = 1080
	padding       = 100
)

func CombineImages(base64Images []string) (string, error) {
	if len(base64Images) == 0 {
		return "", nil
	}

	var images []image.Image
	var origWidths []int
	var origHeights []int

	minHeight := int(^uint(0) >> 1) // Max int32 value
	for _, base64Str := range base64Images {
		img, err := decodeBase64Image(base64Str)
		if err != nil {
			log.Printf("Warning: Failed to decode image: %v", err)
			continue
		}

		trimmed := trimWhitespace(img, trimThreshold)
		images = append(images, trimmed)

		b := trimmed.Bounds()
		w, h := b.Dx(), b.Dy()
		origWidths = append(origWidths, w)
		origHeights = append(origHeights, h)

		if h < minHeight {
			minHeight = h
		}
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no valid images to combine")
	}

	var sumScaledW float64
	for i := range images {
		scale := float64(minHeight) / float64(origHeights[i])
		sumScaledW += float64(origWidths[i]) * scale
	}
	avgScaledW := sumScaledW / float64(len(images))

	spacing := int(math.Round(math.Min(minPixel, math.Max(0, 0.05*avgScaledW))))

	var scaledImages []*image.RGBA
	totalWidth := 0
	for _, img := range images {
		scaled := scaleImageToHeight(img, minHeight)
		scaledImages = append(scaledImages, scaled)
		totalWidth += scaled.Bounds().Dx()
	}

	if len(scaledImages) > 1 {
		totalWidth += spacing * (len(scaledImages) - 1)
	}

	combined := image.NewRGBA(image.Rect(0, 0, totalWidth, minHeight))

	white := image.NewUniform(image.White)
	draw.Draw(combined, combined.Bounds(), white, image.Point{}, draw.Src)

	currentX := 0
	for i, scaled := range scaledImages {
		draw.Draw(combined,
			image.Rect(currentX, 0, currentX+scaled.Bounds().Dx(), minHeight),
			scaled,
			image.Point{},
			draw.Over)
		currentX += scaled.Bounds().Dx()
		if i < len(scaledImages)-1 {
			currentX += spacing
		}
	}

	// Scale the combined image to 1920x1080
	combinedBounds := combined.Bounds()
	combinedW := combinedBounds.Dx()
	combinedH := combinedBounds.Dy()

	contentW := targetW - 2*padding
	contentH := targetH - 2*padding
	if contentW <= 0 || contentH <= 0 {
		return "", fmt.Errorf("invalid target size or padding")
	}

	scale := math.Min(math.Min(float64(contentW)/float64(combinedW), float64(contentH)/float64(combinedH)), 1)
	scaledW := int(math.Max(1, math.Round(float64(combinedW)*scale)))
	scaledH := int(math.Max(1, math.Round(float64(combinedH)*scale)))

	scaledFinal := scaleImage(combined, scaledW, scaledH)

	finalCanvas := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.Draw(finalCanvas, finalCanvas.Bounds(), white, image.Point{}, draw.Src)

	// Center inside the content area that has padding pixels as margins on all sides
	offsetX := padding + (contentW-scaledW)/2
	offsetY := padding + (contentH-scaledH)/2
	draw.Draw(finalCanvas,
		image.Rect(offsetX, offsetY, offsetX+scaledW, offsetY+scaledH),
		scaledFinal,
		image.Point{},
		draw.Over,
	)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, finalCanvas, &jpeg.Options{Quality: quality}); err != nil {
		return "", fmt.Errorf("failed to encode combined image: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Str, nil
}

func trimWhitespace(img image.Image, threshold uint8) image.Image {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y

	isNearWhite := func(c color.Color) bool {
		r, g, b, a := c.RGBA()
		if a == 0 {
			return true
		}
		rr := uint8(r >> 8)
		gg := uint8(g >> 8)
		bb := uint8(b >> 8)
		return rr >= 255-threshold && gg >= 255-threshold && bb >= 255-threshold
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if !isNearWhite(img.At(x, y)) {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x+1 > maxX {
					maxX = x + 1
				}
				if y+1 > maxY {
					maxY = y + 1
				}
			}
		}
	}

	if minX >= maxX || minY >= maxY {
		return img
	}

	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(minX, minY, maxX, maxY))
}

func scaleImageToHeight(img image.Image, targetHeight int) *image.RGBA {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Preserve aspect ratio
	newWidth := (originalWidth * targetHeight) / originalHeight

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, targetHeight))
	// High-quality scaler (Catmull-Rom). Alternatives: xdraw.ApproxBiLinear (fast), xdraw.BiLinear, xdraw.Lanczos3 (sharper).
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, xdraw.Over, nil)
	return dst
}

func decodeBase64Image(base64Str string) (image.Image, error) {
	if strings.Contains(base64Str, ",") {
		parts := strings.Split(base64Str, ",")
		if len(parts) > 1 {
			base64Str = parts[1]
		}
	}

	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func scaleImage(img image.Image, newWidth, newHeight int) *image.RGBA {
	// Generic high-quality scaler for arbitrary target size.
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}
