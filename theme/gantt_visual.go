package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SeriesColor resolves a default task color for behavior-owned data.
func (c GanttColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// SeriesSignature returns the provider-owned palette identity used by caches.
func (c GanttColors) SeriesSignature() uint64 {
	hash := uint64(1469598103934665603)
	for _, col := range c.Series {
		hash = (hash ^ uint64(col.R)) * 1099511628211
		hash = (hash ^ uint64(col.G)) * 1099511628211
		hash = (hash ^ uint64(col.B)) * 1099511628211
		hash = (hash ^ uint64(col.A)) * 1099511628211
	}
	return hash
}

// GanttRequest contains the instance-level visual declaration for a Gantt
// chart. Task data, hierarchy, time mapping, and pointer behavior remain
// widget-owned.
type GanttRequest struct {
	Override *GanttOverride
}

// GanttOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type GanttOverride struct {
	Declaration style.Style
}

// GanttOverrideStyle extracts the provider-owned instance layer.
func GanttOverrideStyle(request GanttRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// GanttMetrics contains geometry and typography consumed by Gantt layout and
// drawing. It contains no theme or semantic objects.
type GanttMetrics struct {
	AnimationDuration     time.Duration
	UpdateDuration        time.Duration
	Height                unit.Dp
	PlotPaddingTop        unit.Dp
	PlotPaddingRight      unit.Dp
	PlotPaddingBottom     unit.Dp
	PlotPaddingLeft       unit.Dp
	AxisNameGap           unit.Dp
	TickLabelGap          unit.Dp
	PlotLabelGap          unit.Dp
	AxisTextSize          unit.Sp
	TextWeight            font.Weight
	EditEdge              unit.Dp
	GridWidth             unit.Dp
	AxisWidth             unit.Dp
	RowHeight             unit.Dp
	BarHeight             unit.Dp
	BarRadius             unit.Dp
	TaskIndent            unit.Dp
	TaskToggleSize        unit.Dp
	TaskToggleGap         unit.Dp
	BaselineHeight        unit.Dp
	TaskLabelPaddingX     unit.Dp
	DependencyWidth       unit.Dp
	DependencyDash        unit.Dp
	MarkerWidth           unit.Dp
	MarkerLabelGap        unit.Dp
	LegendGap             unit.Dp
	LegendTextSize        unit.Sp
	LegendMarkerSize      unit.Dp
	LegendMarkerGap       unit.Dp
	LegendRadius          unit.Dp
	LegendItemGap         unit.Dp
	LegendLineGap         unit.Dp
	TooltipGap            unit.Dp
	TooltipRowGap         unit.Dp
	TooltipMarkerSize     unit.Dp
	TooltipTextSize       unit.Sp
	EmphasisWidth         unit.Dp
	DependencyArrowLength unit.Dp
	DependencyArrowHeight unit.Dp
	DependencyRouteMargin unit.Dp
}

// GanttColors contains every palette decision used by Gantt behavior.
// Explicit task and marker colors remain widget inputs; the provider owns all
// semantic fallbacks and supporting chrome colors.
type GanttColors struct {
	Axis              color.NRGBA
	Label             color.NRGBA
	Grid              color.NRGBA
	Dependency        color.NRGBA
	Marker            color.NRGBA
	Background        color.NRGBA
	Emphasis          color.NRGBA
	TaskLabel         color.NRGBA
	Empty             color.NRGBA
	LegendHover       color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// GanttTaskRequest contains neutral geometry and state for one task visual.
type GanttTaskRequest struct {
	Rect          image.Rectangle
	ProgressRect  image.Rectangle
	Baseline      image.Rectangle
	Color         color.NRGBA
	Emphasis      color.NRGBA
	HasSelection  bool
	Selected      bool
	Milestone     bool
	Radius        int
	EmphasisWidth float32
}

// GanttTaskRenderer draws a task bar, baseline, progress, or milestone.
type GanttTaskRenderer func(layout.Context, GanttTaskRequest)

// GanttDependencyRequest contains the routed points for one dependency edge.
// The widget owns task ordering and routing facts; the provider owns stroke
// and arrow geometry.
type GanttDependencyRequest struct {
	From        f32.Point
	To          f32.Point
	Middle      f32.Point
	Forward     bool
	Width       float32
	Dash        float32
	ArrowLength float32
	ArrowHeight float32
	Color       color.NRGBA
}

// GanttDependencyRenderer draws one dependency edge.
type GanttDependencyRenderer func(layout.Context, GanttDependencyRequest)

// GanttMarkerRequest contains the neutral line geometry for one time marker.
type GanttMarkerRequest struct {
	From  f32.Point
	To    f32.Point
	Width float32
	Color color.NRGBA
}

// GanttMarkerRenderer draws one time marker line.
type GanttMarkerRenderer func(layout.Context, GanttMarkerRequest)

// GanttToggleRequest contains the neutral hit area and collapsed state for a
// hierarchical task toggle.
type GanttToggleRequest struct {
	Area      image.Rectangle
	Collapsed bool
	Color     color.NRGBA
}

// GanttToggleRenderer draws one hierarchy toggle glyph.
type GanttToggleRenderer func(layout.Context, GanttToggleRequest)

// GanttVisual is the complete visual plan consumed by Gantt behavior. The
// root declaration owns disabled opacity and instance styling; metrics,
// colors, and renderer closures own the remaining visual recipe.
type GanttVisual struct {
	Defaults              style.Style
	Override              style.Style
	Metrics               GanttMetrics
	Colors                GanttColors
	Task                  GanttTaskRenderer
	Dependency            GanttDependencyRenderer
	Marker                GanttMarkerRenderer
	Toggle                GanttToggleRenderer
	DisabledOpacity       float32
	TaskBackgroundOpacity float32
	BaselineOpacity       float32
	SelectedOpacity       float32
	LegendHiddenOpacity   float32
}

// GanttProvider supplies Gantt metrics, palette, and renderer strategies
// without taking ownership of task data, hierarchy, or pointer behavior.
type GanttProvider interface {
	Gantt(active *Theme, request GanttRequest) GanttVisual
}

// ValidateGanttVisual checks the geometry, palette, renderer, and root
// declaration contract before a Gantt visual enters its provider cache.
func ValidateGanttVisual(visual GanttVisual) {
	if visual.Task == nil {
		invalidVisual("Gantt", "Task", "must be provided")
	}
	if visual.Dependency == nil {
		invalidVisual("Gantt", "Dependency", "must be provided")
	}
	if visual.Marker == nil {
		invalidVisual("Gantt", "Marker", "must be provided")
	}
	if visual.Toggle == nil {
		invalidVisual("Gantt", "Toggle", "must be provided")
	}

	m := visual.Metrics
	if !validPositiveDp(m.Height) {
		invalidVisual("Gantt", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.PlotPaddingTop) {
		invalidVisual("Gantt", "Metrics.PlotPaddingTop", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingRight) {
		invalidVisual("Gantt", "Metrics.PlotPaddingRight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingBottom) {
		invalidVisual("Gantt", "Metrics.PlotPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingLeft) {
		invalidVisual("Gantt", "Metrics.PlotPaddingLeft", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.AxisNameGap) {
		invalidVisual("Gantt", "Metrics.AxisNameGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TickLabelGap) {
		invalidVisual("Gantt", "Metrics.TickLabelGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotLabelGap) {
		invalidVisual("Gantt", "Metrics.PlotLabelGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.AxisTextSize) {
		invalidVisual("Gantt", "Metrics.AxisTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EditEdge) {
		invalidVisual("Gantt", "Metrics.EditEdge", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.GridWidth) {
		invalidVisual("Gantt", "Metrics.GridWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AxisWidth) {
		invalidVisual("Gantt", "Metrics.AxisWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.RowHeight) {
		invalidVisual("Gantt", "Metrics.RowHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.BarHeight) {
		invalidVisual("Gantt", "Metrics.BarHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.BarRadius) {
		invalidVisual("Gantt", "Metrics.BarRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TaskIndent) {
		invalidVisual("Gantt", "Metrics.TaskIndent", "must be finite and non-negative")
	}
	if !validPositiveDp(m.TaskToggleSize) {
		invalidVisual("Gantt", "Metrics.TaskToggleSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.TaskToggleGap) {
		invalidVisual("Gantt", "Metrics.TaskToggleGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.BaselineHeight) {
		invalidVisual("Gantt", "Metrics.BaselineHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.TaskLabelPaddingX) {
		invalidVisual("Gantt", "Metrics.TaskLabelPaddingX", "must be finite and non-negative")
	}
	if !validPositiveDp(m.DependencyWidth) {
		invalidVisual("Gantt", "Metrics.DependencyWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.DependencyDash) {
		invalidVisual("Gantt", "Metrics.DependencyDash", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MarkerWidth) {
		invalidVisual("Gantt", "Metrics.MarkerWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.MarkerLabelGap) {
		invalidVisual("Gantt", "Metrics.MarkerLabelGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendGap) {
		invalidVisual("Gantt", "Metrics.LegendGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.LegendTextSize) {
		invalidVisual("Gantt", "Metrics.LegendTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LegendMarkerSize) {
		invalidVisual("Gantt", "Metrics.LegendMarkerSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.LegendMarkerGap) {
		invalidVisual("Gantt", "Metrics.LegendMarkerGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendRadius) {
		invalidVisual("Gantt", "Metrics.LegendRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendItemGap) {
		invalidVisual("Gantt", "Metrics.LegendItemGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendLineGap) {
		invalidVisual("Gantt", "Metrics.LegendLineGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipGap) {
		invalidVisual("Gantt", "Metrics.TooltipGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipRowGap) {
		invalidVisual("Gantt", "Metrics.TooltipRowGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.TooltipMarkerSize) {
		invalidVisual("Gantt", "Metrics.TooltipMarkerSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.TooltipTextSize) {
		invalidVisual("Gantt", "Metrics.TooltipTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EmphasisWidth) {
		invalidVisual("Gantt", "Metrics.EmphasisWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.DependencyArrowLength) {
		invalidVisual("Gantt", "Metrics.DependencyArrowLength", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.DependencyArrowHeight) {
		invalidVisual("Gantt", "Metrics.DependencyArrowHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.DependencyRouteMargin) {
		invalidVisual("Gantt", "Metrics.DependencyRouteMargin", "must be finite and non-negative")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("Gantt", "DisabledOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.TaskBackgroundOpacity) {
		invalidVisual("Gantt", "TaskBackgroundOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.BaselineOpacity) {
		invalidVisual("Gantt", "BaselineOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.SelectedOpacity) {
		invalidVisual("Gantt", "SelectedOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.LegendHiddenOpacity) {
		invalidVisual("Gantt", "LegendHiddenOpacity", "must be finite and between 0 and 1")
	}

	if visual.Colors.Axis.A == 0 {
		invalidVisual("Gantt", "Colors.Axis", "must have nonzero alpha")
	}
	if visual.Colors.Label.A == 0 {
		invalidVisual("Gantt", "Colors.Label", "must have nonzero alpha")
	}
	if visual.Colors.Grid.A == 0 {
		invalidVisual("Gantt", "Colors.Grid", "must have nonzero alpha")
	}
	if visual.Colors.Dependency.A == 0 {
		invalidVisual("Gantt", "Colors.Dependency", "must have nonzero alpha")
	}
	if visual.Colors.Marker.A == 0 {
		invalidVisual("Gantt", "Colors.Marker", "must have nonzero alpha")
	}
	if visual.Colors.Background.A == 0 {
		invalidVisual("Gantt", "Colors.Background", "must have nonzero alpha")
	}
	if visual.Colors.Emphasis.A == 0 {
		invalidVisual("Gantt", "Colors.Emphasis", "must have nonzero alpha")
	}
	if visual.Colors.TaskLabel.A == 0 {
		invalidVisual("Gantt", "Colors.TaskLabel", "must have nonzero alpha")
	}
	if visual.Colors.Empty.A == 0 {
		invalidVisual("Gantt", "Colors.Empty", "must have nonzero alpha")
	}
	if visual.Colors.LegendHover.A == 0 {
		invalidVisual("Gantt", "Colors.LegendHover", "must have nonzero alpha")
	}
	if visual.Colors.OverlayForeground.A == 0 {
		invalidVisual("Gantt", "Colors.OverlayForeground", "must have nonzero alpha")
	}
	validSeries := false
	for _, value := range visual.Colors.Series {
		if value.A != 0 {
			validSeries = true
			break
		}
	}
	if !validSeries {
		invalidVisual("Gantt", "Colors.Series", "must contain at least one color with nonzero alpha")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Gantt", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Gantt", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("Gantt", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
