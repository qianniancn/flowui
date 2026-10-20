// Package ganttstyle contains the visual recipe for Gantt providers.
// Gantt behavior remains in widget/gantt; this package owns palette, metrics,
// and the domain renderer strategies.
package ganttstyle

import (
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	taskBackgroundOpacity = 0.24
	baselineOpacity       = 0xa0 / 255.0
	selectedOpacity       = 0.42
	legendHiddenOpacity   = 0.38
	emphasisWidth         = 2
	dependencyArrowLength = 5
	dependencyArrowHeight = 3
	dependencyRouteMargin = 4
)

// BuildGantt returns the complete visual plan for the active theme.
func BuildGantt(active *theme.Theme, request theme.GanttRequest) theme.GanttVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.GanttChart
	semantic := active.Semantic()
	return theme.GanttVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.GanttOverrideStyle(request),
		Metrics: theme.GanttMetrics{
			AnimationDuration: 700 * time.Millisecond,
			UpdateDuration:    450 * time.Millisecond,
			Height:            tokens.Height, PlotPaddingTop: tokens.PlotPaddingTop, PlotPaddingRight: tokens.PlotPaddingRight,
			PlotPaddingBottom: tokens.PlotPaddingBottom, PlotPaddingLeft: tokens.PlotPaddingLeft,
			AxisNameGap: tokens.AxisNameGap, TickLabelGap: tokens.TickLabelGap, PlotLabelGap: 6, AxisTextSize: tokens.AxisTextSize,
			TextWeight: font.Normal, EditEdge: 6,
			GridWidth: tokens.GridWidth, AxisWidth: tokens.AxisWidth, RowHeight: tokens.RowHeight,
			BarHeight: tokens.BarHeight, BarRadius: tokens.BarRadius, TaskIndent: tokens.TaskIndent,
			TaskToggleSize: tokens.TaskToggleSize, TaskToggleGap: tokens.TaskToggleGap,
			BaselineHeight: tokens.BaselineHeight, TaskLabelPaddingX: tokens.TaskLabelPaddingX,
			DependencyWidth: tokens.DependencyWidth, DependencyDash: tokens.DependencyDash,
			MarkerWidth: tokens.MarkerWidth, MarkerLabelGap: tokens.MarkerLabelGap,
			LegendGap: tokens.LegendGap, LegendTextSize: tokens.LegendTextSize,
			LegendMarkerSize: tokens.LegendMarkerSize, LegendMarkerGap: tokens.LegendMarkerGap,
			LegendRadius:  4,
			LegendItemGap: tokens.LegendItemGap, LegendLineGap: tokens.LegendLineGap,
			TooltipGap: tokens.TooltipGap, TooltipRowGap: 4, TooltipMarkerSize: 7,
			TooltipTextSize: active.Components.Tooltip.TextSize,
			EmphasisWidth:   emphasisWidth, DependencyArrowLength: dependencyArrowLength,
			DependencyArrowHeight: dependencyArrowHeight, DependencyRouteMargin: dependencyRouteMargin,
		},
		Colors: theme.GanttColors{
			Axis: semantic.Border, Label: semantic.MutedForeground, Grid: withAlpha(semantic.Border, 0.75),
			Dependency: withAlpha(semantic.MutedForeground, 0.7), Marker: semantic.Tone(theme.TonePrimary).Fill,
			Background: semantic.SurfaceSecondary, Emphasis: semantic.Foreground,
			TaskLabel: semantic.OverlayForeground, Empty: semantic.MutedForeground,
			LegendHover: semantic.SurfaceHover, OverlayForeground: semantic.OverlayForeground,
			Series: tokens.SeriesColors,
		},
		Task:                  renderTask,
		Dependency:            renderDependency,
		Marker:                renderMarker,
		Toggle:                renderToggle,
		DisabledOpacity:       active.DisabledOpacityValue(),
		TaskBackgroundOpacity: taskBackgroundOpacity,
		BaselineOpacity:       baselineOpacity,
		SelectedOpacity:       selectedOpacity,
		LegendHiddenOpacity:   legendHiddenOpacity,
	}
}

func renderTask(gtx layout.Context, request theme.GanttTaskRequest) {
	if !request.Baseline.Empty() {
		baseline := withAlpha(request.Color, baselineOpacity)
		paint.FillShape(gtx.Ops, baseline, clip.UniformRRect(request.Baseline, min(request.Baseline.Dy()/2, 2)).Op(gtx.Ops))
	}
	if request.Rect.Empty() {
		return
	}
	opacity := float32(1)
	if request.HasSelection && !request.Selected {
		opacity = selectedOpacity
	}
	fade := paint.PushOpacity(gtx.Ops, opacity)
	if request.Milestone {
		renderMilestone(gtx, request)
		fade.Pop()
		return
	}
	background := withAlpha(request.Color, taskBackgroundOpacity)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(request.Rect, request.Radius).Op(gtx.Ops))
	if !request.ProgressRect.Empty() {
		progressRadius := min(request.Radius, min(request.ProgressRect.Dx(), request.ProgressRect.Dy())/2)
		paint.FillShape(gtx.Ops, request.Color, clip.UniformRRect(request.ProgressRect, progressRadius).Op(gtx.Ops))
	}
	fade.Pop()
}

func renderMilestone(gtx layout.Context, request theme.GanttTaskRequest) {
	center := f32.Pt(float32(request.Rect.Min.X+request.Rect.Dx()/2), float32(request.Rect.Min.Y+request.Rect.Dy()/2))
	radius := float32(min(request.Rect.Dx(), request.Rect.Dy())) / 2
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X, center.Y-radius))
	path.LineTo(f32.Pt(center.X+radius, center.Y))
	path.LineTo(f32.Pt(center.X, center.Y+radius))
	path.LineTo(f32.Pt(center.X-radius, center.Y))
	path.Close()
	spec := path.End()
	paint.FillShape(gtx.Ops, request.Color, clip.Outline{Path: spec}.Op())
	if request.Selected {
		paint.FillShape(gtx.Ops, request.Emphasis, clip.Stroke{Path: spec, Width: request.EmphasisWidth}.Op())
	}
}

func renderDependency(gtx layout.Context, request theme.GanttDependencyRequest) {
	strokeDependency(gtx, request.From, request.Middle, request.Width, request.Color, request.Dash)
	strokeDependency(gtx, request.Middle, f32.Pt(request.Middle.X, request.To.Y), request.Width, request.Color, request.Dash)
	strokeDependency(gtx, f32.Pt(request.Middle.X, request.To.Y), request.To, request.Width, request.Color, request.Dash)
	direction := float32(1)
	if !request.Forward {
		direction = -1
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(request.To)
	path.LineTo(f32.Pt(request.To.X-request.ArrowLength*direction, request.To.Y-request.ArrowHeight))
	path.LineTo(f32.Pt(request.To.X-request.ArrowLength*direction, request.To.Y+request.ArrowHeight))
	path.Close()
	paint.FillShape(gtx.Ops, request.Color, clip.Outline{Path: path.End()}.Op())
}

func strokeDependency(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA, dash float32) {
	dx, dy := to.X-from.X, to.Y-from.Y
	distance := float32(math.Hypot(float64(dx), float64(dy)))
	if distance <= 0 || dash <= 0 || width <= 0 || col.A == 0 {
		return
	}
	for offset := float32(0); offset < distance; offset += dash * 2 {
		end := min(offset+dash, distance)
		strokeLine(gtx,
			f32.Pt(from.X+dx*offset/distance, from.Y+dy*offset/distance),
			f32.Pt(from.X+dx*end/distance, from.Y+dy*end/distance), width, col)
	}
}

func strokeLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	if width <= 0 || col.A == 0 {
		return
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func renderMarker(gtx layout.Context, request theme.GanttMarkerRequest) {
	strokeLine(gtx, request.From, request.To, request.Width, request.Color)
}

func renderToggle(gtx layout.Context, request theme.GanttToggleRequest) {
	area := request.Area
	center := f32.Pt(float32(area.Min.X+area.Dx()/2), float32(area.Min.Y+area.Dy()/2))
	size := float32(min(max(area.Dx()/2, 3), max(area.Dy()/3, 3)))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	if request.Collapsed {
		path.MoveTo(f32.Pt(center.X-size/2, center.Y-size))
		path.LineTo(f32.Pt(center.X-size/2, center.Y+size))
		path.LineTo(f32.Pt(center.X+size, center.Y))
	} else {
		path.MoveTo(f32.Pt(center.X-size, center.Y-size/2))
		path.LineTo(f32.Pt(center.X+size, center.Y-size/2))
		path.LineTo(f32.Pt(center.X, center.Y+size))
	}
	path.Close()
	paint.FillShape(gtx.Ops, request.Color, clip.Outline{Path: path.End()}.Op())
}

func withAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
