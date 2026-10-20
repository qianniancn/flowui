package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ToolbarOrientation identifies the neutral axis used by toolbar behavior.
// Numeric values keep the provider protocol independent from widget packages.
type ToolbarOrientation uint8

const (
	ToolbarHorizontal ToolbarOrientation = iota
	ToolbarVertical
)

// ToolbarRequest describes the behavior-owned toolbar axes and optional
// instance declaration needed by a visual provider.
type ToolbarRequest struct {
	Orientation ToolbarOrientation
	Attached    bool
	Override    *ToolbarOverride
}

// ToolbarOverride carries caller customization across the neutral provider
// boundary. The provider owns how the declaration is merged.
type ToolbarOverride struct {
	Declaration style.Style
}

// ToolbarOverrideStyle extracts the provider-owned instance layer.
func ToolbarOverrideStyle(request ToolbarRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ToolbarSeparatorOverride carries an instance declaration for one separator.
// It remains separate from ToolbarOverride so separator transitions keep their
// own identity and do not invalidate the toolbar root declaration.
type ToolbarSeparatorOverride struct {
	Declaration style.Style
}

// ToolbarSeparatorOverrideStyle extracts the separator instance layer.
func ToolbarSeparatorOverrideStyle(override *ToolbarSeparatorOverride) style.Style {
	if override == nil {
		return style.Style{}
	}
	return override.Declaration
}

// ToolbarMetrics contains theme-owned geometry used by toolbar behavior.
type ToolbarMetrics struct {
	Gap unit.Dp
}

// ToolbarVisual is the complete visual plan consumed by toolbar behavior.
// Root declarations are optional for unattached toolbars; separator
// declarations are always concrete so the widget never invents a visual
// fallback.
type ToolbarVisual struct {
	Defaults          style.Style
	Variant           style.Style
	Override          style.Style
	Separator         style.Style
	VerticalSeparator style.Style
	Metrics           ToolbarMetrics
}

// ToolbarProvider supplies toolbar declarations and layout metrics without
// taking ownership of child ordering, focus navigation, or semantic state.
type ToolbarProvider interface {
	Toolbar(active *Theme, request ToolbarRequest) ToolbarVisual
}

// ValidateToolbarVisual checks the provider-owned geometry and separator
// declarations before toolbar behavior stores a visual plan in its cache.
// An empty root declaration means that an unattached toolbar inherits its
// parent's surface. Any non-empty root declaration must be complete.
func ValidateToolbarVisual(visual ToolbarVisual) {
	if visual.Metrics.Gap < 0 {
		invalidVisual("Toolbar", "Metrics.Gap", "must be at least 0")
	}

	validateToolbarSeparator("Separator", visual.Separator)
	validateToolbarSeparator("VerticalSeparator", visual.VerticalSeparator)
	resolved := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Override)
	if resolved.Box == nil && resolved.Paint == nil && resolved.Text == nil && resolved.Trans == nil {
		return
	}
	if resolved.Box == nil {
		invalidVisual("Toolbar", "Resolved.Box", "must be provided")
	}
	if resolved.Box.Padding == nil {
		invalidVisual("Toolbar", "Resolved.Box.Padding", "must be provided")
	}
	if resolved.Paint == nil {
		invalidVisual("Toolbar", "Resolved.Paint", "must be provided")
	}
	if resolved.Paint.Background == nil {
		invalidVisual("Toolbar", "Resolved.Paint.Background", "must be provided")
	}
	if resolved.Paint.Radius == nil {
		invalidVisual("Toolbar", "Resolved.Paint.Radius", "must be provided")
	}
	if resolved.Text == nil {
		invalidVisual("Toolbar", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Toolbar", "Resolved.Text.Color", "must resolve to a solid color")
	}
}

func validateToolbarSeparator(field string, declaration style.Style) {
	resolved := style.Cascade(style.State{}, declaration)
	if resolved.Box == nil {
		invalidVisual("Toolbar", field+".Box", "must be provided")
	}
	if resolved.Box.Width == nil {
		invalidVisual("Toolbar", field+".Box.Width", "must be provided")
	}
	if *resolved.Box.Width <= 0 {
		invalidVisual("Toolbar", field+".Box.Width", "must be greater than 0")
	}
	if resolved.Box.Height == nil {
		invalidVisual("Toolbar", field+".Box.Height", "must be provided")
	}
	if *resolved.Box.Height <= 0 {
		invalidVisual("Toolbar", field+".Box.Height", "must be greater than 0")
	}
	if resolved.Paint == nil {
		invalidVisual("Toolbar", field+".Paint", "must be provided")
	}
	if resolved.Paint.Background == nil {
		invalidVisual("Toolbar", field+".Paint.Background", "must be provided")
	}
	if !hasPaint(resolved.Paint.Background) {
		invalidVisual("Toolbar", field+".Paint.Background", "must provide a solid color or nonempty gradient")
	}
}
