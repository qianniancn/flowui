package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ImageRequest describes the instance-level declaration a theme should apply to
// an image. Loading, fitting, sizing, opacity, and accessibility stay with the
// widget; only the provider-owned shape crosses this boundary.
type ImageRequest struct {
	Override *ImageOverride
}

// ImageOverride carries caller customization across the provider boundary.
type ImageOverride struct {
	Declaration style.Style
}

// ImageOverrideStyle extracts the provider-owned instance layer.
func ImageOverrideStyle(request ImageRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ImageMetrics is the provider-owned shape of an image.
type ImageMetrics struct {
	// Radius clips the image's corners. Zero keeps square corners.
	Radius unit.Dp
}

// ImageVisual is the complete visual plan consumed by image behavior.
type ImageVisual struct {
	Metrics  ImageMetrics
	Override style.Style
}

// ImageProvider supplies image shape without taking ownership of loading,
// fit, sizing, opacity, or accessibility.
type ImageProvider interface {
	Image(active *Theme, request ImageRequest) ImageVisual
}

// ValidateImageVisual checks the provider-owned shape before an image caches it.
func ValidateImageVisual(visual ImageVisual) {
	if visual.Metrics.Radius < 0 {
		invalidVisual("Image", "Metrics.Radius", "must be at least 0")
	}
}
