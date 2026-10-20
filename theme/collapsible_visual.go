package theme

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// CollapsibleRequest carries the optional instance declaration supplied by a
// collapsible behavior widget. Expansion, focus, and keyboard state remain
// owned by the widget; the provider owns the resulting visual declaration.
type CollapsibleRequest struct {
	Override *CollapsibleOverride
}

// CollapsibleOverride carries caller customization across the neutral
// provider boundary.
type CollapsibleOverride struct {
	Declaration style.Style
	// Trigger is scoped to the internal disclosure button. Keeping it beside
	// the root declaration lets provider-aware wrappers carry both visual
	// layers without putting either one on the behavior widget.
	Trigger style.Style
}

// CollapsibleOverrideStyle extracts the provider-owned instance layer.
func CollapsibleOverrideStyle(request CollapsibleRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// CollapsibleTriggerOverrideStyle extracts the provider-owned trigger layer.
func CollapsibleTriggerOverrideStyle(request CollapsibleRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Trigger
}

// CollapsibleMetrics contains neutral geometry and animation timing used by
// collapsible behavior layout.
type CollapsibleMetrics struct {
	BodyPadding       unit.Dp
	ContentGap        unit.Dp
	ContentDuration   time.Duration
	IndicatorDuration time.Duration
}

// CollapsibleIndicatorRequest contains behavior-owned state needed by a
// theme-owned disclosure indicator renderer.
type CollapsibleIndicatorRequest struct {
	Progress float32
	Disabled bool
}

// CollapsibleIndicatorRenderer paints the disclosure indicator and returns
// its dimensions under the supplied constraints.
type CollapsibleIndicatorRenderer func(layout.Context, CollapsibleIndicatorRequest) layout.Dimensions

// CollapsibleVisual is the complete visual plan consumed by collapsible
// behavior. The widget does not interpret theme tokens or choose indicator
// colors and geometry.
type CollapsibleVisual struct {
	Label     style.Style
	Trigger   style.Style
	Metrics   CollapsibleMetrics
	Motion    MotionTheme
	Indicator CollapsibleIndicatorRenderer
	Override  style.Style
}

// CollapsibleProvider supplies collapsible declarations, layout metrics,
// motion, and indicator drawing without taking ownership of expansion state.
type CollapsibleProvider interface {
	Collapsible(active *Theme, request CollapsibleRequest) CollapsibleVisual
}

// ValidateCollapsibleVisual checks the declaration, geometry, motion, and
// indicator contracts before a widget stores a provider result in its cache.
// Trigger is an instance-level button override, so an empty trigger layer is
// valid and is intentionally not treated as a missing theme recipe.
func ValidateCollapsibleVisual(visual CollapsibleVisual) {
	metrics := visual.Metrics
	if metrics.BodyPadding < 0 {
		invalidVisual("Collapsible", "Metrics.BodyPadding", "must be at least 0")
	}
	if metrics.ContentGap < 0 {
		invalidVisual("Collapsible", "Metrics.ContentGap", "must be at least 0")
	}
	if metrics.ContentDuration <= 0 {
		invalidVisual("Collapsible", "Metrics.ContentDuration", "must be greater than 0")
	}
	if metrics.IndicatorDuration <= 0 {
		invalidVisual("Collapsible", "Metrics.IndicatorDuration", "must be greater than 0")
	}

	validateMotion("Collapsible", "Motion", visual.Motion)
	if visual.Indicator == nil {
		invalidVisual("Collapsible", "Indicator", "must be provided")
	}

	requireLabel := func(state style.State) {
		resolved := style.Cascade(state, visual.Label, visual.Override)
		if resolved.Text == nil {
			invalidVisual("Collapsible", "Resolved.Text", "must be provided")
		}
		if resolved.Text.FontSize == nil {
			invalidVisual("Collapsible", "Resolved.Text.FontSize", "must be provided")
		}
		if *resolved.Text.FontSize <= 0 {
			invalidVisual("Collapsible", "Resolved.Text.FontSize", "must be greater than 0")
		}
		if !hasColor(resolved.Text.Color) {
			invalidVisual("Collapsible", "Resolved.Text.Color", "must resolve to a solid color")
		}
	}
	for _, state := range []style.State{
		{},
		{Expanded: true},
		{Disabled: true},
		{Disabled: true, Expanded: true},
	} {
		requireLabel(state)
	}
}
