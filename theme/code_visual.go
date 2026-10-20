package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// CodeRequest describes the language and instance declaration of a code block.
// Selection, copying, highlighting, and layout stay with the widget; the
// provider owns the block surface and its geometry.
type CodeRequest struct {
	Language string
	Override *CodeOverride
}

// CodeOverride carries caller customization across the provider boundary.
type CodeOverride struct {
	Declaration style.Style
}

// CodeOverrideStyle extracts the provider-owned instance layer.
func CodeOverrideStyle(request CodeRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// CodeMetrics is the provider-owned geometry of a code block.
type CodeMetrics struct {
	Padding       unit.Dp
	Radius        unit.Dp
	HeaderGap     unit.Dp
	LineNumberGap unit.Dp
	TextSize      unit.Sp
}

// CodeVisual is the complete visual plan consumed by code behavior.
type CodeVisual struct {
	// Root is the block surface: padding, radius, and background.
	Root     style.Style
	Metrics  CodeMetrics
	Override style.Style
}

// CodeProvider supplies the code-block surface without taking ownership of
// selection, clipboard, highlighting, or layout.
type CodeProvider interface {
	Code(active *Theme, request CodeRequest) CodeVisual
}

// ValidateCodeVisual checks the provider-owned surface before a code block
// caches it.
func ValidateCodeVisual(visual CodeVisual) {
	metrics := visual.Metrics
	if metrics.Padding < 0 {
		invalidVisual("Code", "Metrics.Padding", "must be at least 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Code", "Metrics.Radius", "must be at least 0")
	}
	if metrics.HeaderGap < 0 {
		invalidVisual("Code", "Metrics.HeaderGap", "must be at least 0")
	}
	if metrics.LineNumberGap < 0 {
		invalidVisual("Code", "Metrics.LineNumberGap", "must be at least 0")
	}
	if metrics.TextSize <= 0 {
		invalidVisual("Code", "Metrics.TextSize", "must be greater than 0")
	}
}
