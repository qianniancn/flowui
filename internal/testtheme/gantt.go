package testtheme

import (
	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Gantt(active *theme.Theme, request theme.GanttRequest) theme.GanttVisual {
	tokens := active.Components.GanttChart
	return theme.GanttVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.GanttOverrideStyle(request),
		Metrics: theme.GanttMetrics{
			Height: tokens.Height, PlotPaddingTop: tokens.PlotPaddingTop, PlotPaddingRight: tokens.PlotPaddingRight,
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
			TooltipTextSize: active.Components.Tooltip.TextSize, EmphasisWidth: 2,
			DependencyArrowLength: 5, DependencyArrowHeight: 3, DependencyRouteMargin: 4,
		},
		Colors: theme.GanttColors{
			Axis: accent, Label: muted, Grid: muted, Dependency: muted, Marker: accent,
			Background: background, Emphasis: foreground, TaskLabel: foreground, Empty: muted,
			LegendHover: background, OverlayForeground: foreground, Series: tokens.SeriesColors,
		},
		Task:            func(layout.Context, theme.GanttTaskRequest) {},
		Dependency:      func(layout.Context, theme.GanttDependencyRequest) {},
		Marker:          func(layout.Context, theme.GanttMarkerRequest) {},
		Toggle:          func(layout.Context, theme.GanttToggleRequest) {},
		DisabledOpacity: active.DisabledOpacityValue(), TaskBackgroundOpacity: .24,
		BaselineOpacity: 0xa0 / 255.0, SelectedOpacity: .42, LegendHiddenOpacity: .38,
	}
}
