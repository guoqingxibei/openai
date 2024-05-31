package util

import "math"

// CountImageTokens return max is 1445
func CountImageTokens(width, height int) int {
	if width > 2048 || height > 2048 {
		aspectRatio := float64(width) / float64(height)
		if aspectRatio > 1 {
			width, height = 2048, int(2048/aspectRatio)
		} else {
			width, height = int(2048*aspectRatio), 2048
		}
	}

	if width >= height && height > 768 {
		width, height = int((768/float64(height))*float64(width)), 768
	} else if height > width && width > 768 {
		width, height = 768, int((768/float64(width))*float64(height))
	}

	tilesWidth := int(math.Ceil(float64(width) / 512.0))
	tilesHeight := int(math.Ceil(float64(height) / 512.0))
	totalTokens := 85 + 170*(tilesWidth*tilesHeight)

	return totalTokens
}
