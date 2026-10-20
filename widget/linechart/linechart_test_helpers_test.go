package linechart

import "gioui.org/f32"

// Test-only helpers for validating point walking, viewport slicing, and
// min/max pixel sampling without keeping them on the production surface.

func walkLine(points []resolvedPoint, connectNulls bool, visit func(resolvedPoint, bool)) {
	started := false
	for _, point := range points {
		if !point.valid {
			if !connectNulls {
				started = false
			}
			continue
		}
		visit(point, !started)
		started = true
	}
}

func visiblePixelSegment(points []f32.Point, minimumX, maximumX float32) []f32.Point {
	start, end := visiblePixelRange(points, minimumX, maximumX)
	return points[start:end]
}

func minMaxPixelSample(points []f32.Point, pixelWidth int) []f32.Point {
	indices := minMaxPixelSampleIndices(points, pixelWidth)
	if indices == nil {
		return points
	}
	return selectPixelPoints(points, indices)
}
