package theme

import (
	"image/color"
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SkeletonRequest describes the shape a skeleton placeholder should take.
// Pulsing and layout stay with the widget.
type SkeletonRequest struct {
	Circle   bool
	Override *SkeletonOverride
}

// SkeletonOverride carries caller customization across the provider boundary.
type SkeletonOverride struct {
	Declaration style.Style
}

// SkeletonOverrideStyle extracts the provider-owned instance layer.
func SkeletonOverrideStyle(request SkeletonRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SkeletonMetrics is the provider-owned geometry and pulse of a placeholder.
type SkeletonMetrics struct {
	Height unit.Dp
	Radius unit.Dp
	Period time.Duration
}

// SkeletonColors is the provider-owned pulse range.
type SkeletonColors struct {
	Base      color.NRGBA
	Highlight color.NRGBA
}

// SkeletonVisual is the complete visual plan consumed by skeleton behavior.
type SkeletonVisual struct {
	Metrics  SkeletonMetrics
	Colors   SkeletonColors
	Motion   MotionTheme
	Override style.Style
}

// SkeletonProvider supplies placeholder shape and pulse without taking
// ownership of sizing, layout, or animation timing.
type SkeletonProvider interface {
	Skeleton(active *Theme, request SkeletonRequest) SkeletonVisual
}

// ValidateSkeletonVisual checks the provider-owned placeholder before a widget
// caches it.
func ValidateSkeletonVisual(visual SkeletonVisual) {
	metrics := visual.Metrics
	if metrics.Height < 0 {
		invalidVisual("Skeleton", "Metrics.Height", "must be at least 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Skeleton", "Metrics.Radius", "must be at least 0")
	}
	if metrics.Period < 0 {
		invalidVisual("Skeleton", "Metrics.Period", "must be at least 0")
	}
	if visual.Colors.Base.A == 0 {
		invalidVisual("Skeleton", "Colors.Base", "must be opaque")
	}
	if visual.Colors.Highlight.A == 0 {
		invalidVisual("Skeleton", "Colors.Highlight", "must be opaque")
	}
}
