package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// StatusBarRequest describes the semantic chrome slots and optional
// instance declaration needed by a status-bar visual provider.
type StatusBarRequest struct {
	Override *StatusBarOverride
}

// StatusBarOverride carries caller customization across the neutral provider
// boundary. The provider owns how this declaration is merged.
type StatusBarOverride struct {
	Declaration style.Style
}

// StatusBarOverrideStyle extracts the provider-owned instance layer.
func StatusBarOverrideStyle(request StatusBarRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// StatusBarMetrics contains theme-owned chrome geometry used by status-bar
// behavior to size the root and arrange its semantic left and right slots.
// Padding is carried by the resolved root declaration.
type StatusBarMetrics struct {
	Height unit.Dp
	Gap    unit.Dp
}

// StatusBarVisual is the complete visual plan consumed by status-bar
// behavior. The separator is part of the root declaration, so no widget code
// needs to interpret a theme color or draw a fallback line.
type StatusBarVisual struct {
	Defaults style.Style
	Variant  style.Style
	Override style.Style
	Metrics  StatusBarMetrics
}

// StatusBarProvider supplies the status-bar declaration and layout metrics
// without taking ownership of slot ordering or width constraints.
type StatusBarProvider interface {
	StatusBar(active *Theme, request StatusBarRequest) StatusBarVisual
}

// ValidateStatusBarVisual checks the root declaration and geometry before a
// widget stores a provider result in its cache.
func ValidateStatusBarVisual(visual StatusBarVisual) {
	if visual.Metrics.Height <= 0 {
		invalidVisual("StatusBar", "Metrics.Height", "must be greater than 0")
	}
	if visual.Metrics.Gap < 0 {
		invalidVisual("StatusBar", "Metrics.Gap", "must be at least 0")
	}

	resolved := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Override)
	if resolved.Box == nil {
		invalidVisual("StatusBar", "Resolved.Box", "must be provided")
	}
	if resolved.Box.Height == nil {
		invalidVisual("StatusBar", "Resolved.Box.Height", "must be provided")
	}
	if *resolved.Box.Height <= 0 {
		invalidVisual("StatusBar", "Resolved.Box.Height", "must be greater than 0")
	}
	if resolved.Box.Padding == nil {
		invalidVisual("StatusBar", "Resolved.Box.Padding", "must be provided")
	}
	if resolved.Paint == nil {
		invalidVisual("StatusBar", "Resolved.Paint", "must be provided")
	}
	if resolved.Paint.Background == nil {
		invalidVisual("StatusBar", "Resolved.Paint.Background", "must be provided")
	}
	if resolved.Paint.BorderBottom == nil {
		invalidVisual("StatusBar", "Resolved.Paint.BorderBottom", "must be provided")
	}
	if resolved.Paint.BorderBottom.Width == nil {
		invalidVisual("StatusBar", "Resolved.Paint.BorderBottom.Width", "must be provided")
	}
	if !hasColor(resolved.Paint.BorderBottom.Color) {
		invalidVisual("StatusBar", "Resolved.Paint.BorderBottom.Color", "must resolve to a solid color")
	}
	if resolved.Text == nil {
		invalidVisual("StatusBar", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("StatusBar", "Resolved.Text.Color", "must resolve to a solid color")
	}
}
