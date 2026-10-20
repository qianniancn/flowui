package theme

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// TimelineRequest contains the instance-level visual declaration for a
// timeline. Item ordering, placement, and content remain widget-owned.
type TimelineRequest struct {
	Override *TimelineOverride
}

// TimelineOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type TimelineOverride struct {
	Declaration style.Style
}

// TimelineOverrideStyle extracts the provider-owned instance layer.
func TimelineOverrideStyle(request TimelineRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// TimelineRailRequest contains neutral geometry and semantic colors for one
// timeline rail. The widget computes placement and bounds; the provider owns
// the rail, marker, dash, and outline drawing strategy.
type TimelineRailRequest struct {
	Vertical      bool
	Center        int
	Start         int
	End           int
	MarkerX       int
	MarkerY       int
	Tail          bool
	Process       bool
	DotSize       int
	TailWidth     int
	BorderWidth   int
	DashLength    int
	DashGap       int
	DotBackground color.NRGBA
	TailColor     color.NRGBA
	MarkerColor   color.NRGBA
	CustomIcon    bool
	Filled        bool
}

// TimelineRailRenderer draws the provider-owned timeline chrome for one item.
type TimelineRailRenderer func(layout.Context, TimelineRailRequest)

// TimelineMetrics contains geometry consumed by timeline layout and drawing.
// It contains no theme or semantic objects.
type TimelineMetrics struct {
	TailWidth            unit.Dp
	DotBorderWidth       unit.Dp
	DotSize              unit.Dp
	ItemPaddingBottom    unit.Dp
	TitleContentGap      unit.Dp
	SectionGap           unit.Dp
	InlineInset          unit.Dp
	HorizontalTitleGap   unit.Dp
	HorizontalContentGap unit.Dp
	HorizontalItemGap    unit.Dp
	ProcessDashLength    unit.Dp
	ProcessDashGap       unit.Dp
}

// TimelineColors contains provider-owned colors used by timeline behavior.
// Explicit item colors remain widget inputs; the provider owns semantic
// fallbacks and the child background used while recording content.
type TimelineColors struct {
	Background    color.NRGBA
	Tail          color.NRGBA
	DotBackground color.NRGBA
	Title         color.NRGBA
	Content       color.NRGBA
	Muted         color.NRGBA
	Primary       color.NRGBA
	Error         color.NRGBA
	Success       color.NRGBA
	Disabled      color.NRGBA
}

// TimelineVisual is the complete visual plan consumed by timeline behavior.
// The root declaration owns disabled opacity and instance styling; metrics
// and colors own the remaining visual recipe.
type TimelineVisual struct {
	Defaults style.Style
	Override style.Style
	Metrics  TimelineMetrics
	Colors   TimelineColors
	Rail     TimelineRailRenderer
}

// TimelineProvider supplies timeline metrics and palette without taking
// ownership of item order, placement, or content behavior.
type TimelineProvider interface {
	Timeline(active *Theme, request TimelineRequest) TimelineVisual
}

// ValidateTimelineVisual checks the geometry, palette, and root declaration
// contract before a timeline visual enters its provider cache.
func ValidateTimelineVisual(visual TimelineVisual) {
	if visual.Rail == nil {
		invalidVisual("Timeline", "Rail", "must be provided")
	}

	m := visual.Metrics
	if !validPositiveDp(m.TailWidth) {
		invalidVisual("Timeline", "Metrics.TailWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.DotBorderWidth) {
		invalidVisual("Timeline", "Metrics.DotBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.DotSize) {
		invalidVisual("Timeline", "Metrics.DotSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.ItemPaddingBottom) {
		invalidVisual("Timeline", "Metrics.ItemPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TitleContentGap) {
		invalidVisual("Timeline", "Metrics.TitleContentGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.SectionGap) {
		invalidVisual("Timeline", "Metrics.SectionGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.InlineInset) {
		invalidVisual("Timeline", "Metrics.InlineInset", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.HorizontalTitleGap) {
		invalidVisual("Timeline", "Metrics.HorizontalTitleGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.HorizontalContentGap) {
		invalidVisual("Timeline", "Metrics.HorizontalContentGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.HorizontalItemGap) {
		invalidVisual("Timeline", "Metrics.HorizontalItemGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.ProcessDashLength) {
		invalidVisual("Timeline", "Metrics.ProcessDashLength", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ProcessDashGap) {
		invalidVisual("Timeline", "Metrics.ProcessDashGap", "must be finite and greater than 0")
	}

	if visual.Colors.Background.A == 0 {
		invalidVisual("Timeline", "Colors.Background", "must have nonzero alpha")
	}
	if visual.Colors.Tail.A == 0 {
		invalidVisual("Timeline", "Colors.Tail", "must have nonzero alpha")
	}
	if visual.Colors.DotBackground.A == 0 {
		invalidVisual("Timeline", "Colors.DotBackground", "must have nonzero alpha")
	}
	if visual.Colors.Title.A == 0 {
		invalidVisual("Timeline", "Colors.Title", "must have nonzero alpha")
	}
	if visual.Colors.Content.A == 0 {
		invalidVisual("Timeline", "Colors.Content", "must have nonzero alpha")
	}
	if visual.Colors.Muted.A == 0 {
		invalidVisual("Timeline", "Colors.Muted", "must have nonzero alpha")
	}
	if visual.Colors.Primary.A == 0 {
		invalidVisual("Timeline", "Colors.Primary", "must have nonzero alpha")
	}
	if visual.Colors.Error.A == 0 {
		invalidVisual("Timeline", "Colors.Error", "must have nonzero alpha")
	}
	if visual.Colors.Success.A == 0 {
		invalidVisual("Timeline", "Colors.Success", "must have nonzero alpha")
	}
	if visual.Colors.Disabled.A == 0 {
		invalidVisual("Timeline", "Colors.Disabled", "must have nonzero alpha")
	}
	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Timeline", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Timeline", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("Timeline", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
