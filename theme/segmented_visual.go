package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SegmentedRequest describes the instance declaration of a segmented control.
// Selection, keyboard, and change callbacks stay with the widget; the provider
// owns the track surface and its geometry.
type SegmentedRequest struct {
	Override *SegmentedOverride
}

// SegmentedOverride carries caller customization across the provider boundary.
type SegmentedOverride struct {
	Declaration style.Style
}

// SegmentedOverrideStyle extracts the provider-owned instance layer.
func SegmentedOverrideStyle(request SegmentedRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SegmentedMetrics is the provider-owned track geometry.
type SegmentedMetrics struct {
	Padding unit.Dp
	Radius  unit.Dp
	Gap     unit.Dp
}

// SegmentedVisual is the complete visual plan consumed by segmented behavior.
type SegmentedVisual struct {
	// Root is the track surface: padding, radius, and background.
	Root     style.Style
	Metrics  SegmentedMetrics
	Override style.Style
}

// SegmentedProvider supplies the segmented track without taking ownership of
// selection, focus, or keyboard behavior.
type SegmentedProvider interface {
	Segmented(active *Theme, request SegmentedRequest) SegmentedVisual
}

// ValidateSegmentedVisual checks the provider-owned track before a control
// caches it.
func ValidateSegmentedVisual(visual SegmentedVisual) {
	metrics := visual.Metrics
	if metrics.Padding < 0 {
		invalidVisual("Segmented", "Metrics.Padding", "must be at least 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Segmented", "Metrics.Radius", "must be at least 0")
	}
	if metrics.Gap < 0 {
		invalidVisual("Segmented", "Metrics.Gap", "must be at least 0")
	}
}
