package theme

import (
	"github.com/qianniancn/flowui/style"
)

// SurfaceRequest describes the semantic surface variant and optional
// instance declaration needed by a surface visual provider.
type SurfaceRequest struct {
	Variant  SurfaceVariant
	Override *SurfaceOverride
}

// SurfaceOverride carries caller customization across the neutral provider
// boundary. The provider owns how this declaration is merged.
type SurfaceOverride struct {
	Declaration style.Style
}

// SurfaceOverrideStyle extracts the provider-owned instance layer.
func SurfaceOverrideStyle(request SurfaceRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SurfaceVisual is the complete visual plan consumed by surface behavior.
type SurfaceVisual struct {
	Defaults style.Style
	Variant  style.Style
	Override style.Style
}

// SurfaceProvider supplies semantic surface declarations without taking
// ownership of child layout or scope restoration.
type SurfaceProvider interface {
	Surface(active *Theme, request SurfaceRequest) SurfaceVisual
}

// ValidateSurfaceVisual checks the declaration contract before a widget stores
// a provider result in its cache. Surface behavior only supplies scope and
// child layout; it cannot invent border or inherited text declarations.
func ValidateSurfaceVisual(visual SurfaceVisual) {
	resolved := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Override)
	if resolved.Paint == nil {
		invalidVisual("Surface", "Resolved.Paint", "must be provided")
	}
	if resolved.Paint.Border == nil {
		invalidVisual("Surface", "Resolved.Paint.Border", "must be provided")
	}
	if resolved.Paint.Border.Width == nil {
		invalidVisual("Surface", "Resolved.Paint.Border.Width", "must be provided")
	}
	if !hasColor(resolved.Paint.Border.Color) {
		invalidVisual("Surface", "Resolved.Paint.Border.Color", "must resolve to a solid color")
	}
	if resolved.Paint.Opacity == nil {
		invalidVisual("Surface", "Resolved.Paint.Opacity", "must be provided")
	}
	if resolved.Text == nil {
		invalidVisual("Surface", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Surface", "Resolved.Text.Color", "must resolve to a solid color")
	}
}
