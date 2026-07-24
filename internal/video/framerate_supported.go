//go:build darwin || windows

package video

import (
	"math"
	"strconv"
)

type captureMode struct {
	width      int
	height     int
	framerates []float64
}

func chooseCaptureFramerate(targetFPS, width, height int, modes []captureMode) float64 {
	candidates := make([]float64, 0)
	for _, mode := range modes {
		if mode.width == width && mode.height == height {
			candidates = append(candidates, mode.framerates...)
		}
	}

	if len(candidates) == 0 {
		for _, mode := range modes {
			candidates = append(candidates, mode.framerates...)
		}
	}

	return chooseNearestFramerate(targetFPS, candidates)
}

func chooseNearestFramerate(targetFPS int, candidates []float64) float64 {
	if targetFPS <= 0 {
		targetFPS = 30
	}

	target := float64(targetFPS)
	best := 0.0
	bestDiff := math.MaxFloat64

	for _, candidate := range candidates {
		if candidate <= 0 {
			continue
		}

		diff := math.Abs(candidate - target)
		if best == 0 || diff < bestDiff || (diff == bestDiff && candidate > best) {
			best = candidate
			bestDiff = diff
		}
	}

	return best
}

func formatFramerate(fps float64) string {
	return strconv.FormatFloat(fps, 'f', -1, 64)
}
