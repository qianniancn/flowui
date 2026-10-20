package testtheme

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Checkbox(active *theme.Theme, request theme.CheckboxRequest) theme.CheckboxVisual {
	tokens := active.Components.Checkbox
	indicator := style.Style{}.
		Width(tokens.Size).
		Height(tokens.Size).
		Background(style.Solid(background)).
		BorderWidth(tokens.BorderWidth).
		BorderColor(style.Solid(foreground)).
		Radius(tokens.Radius).
		Outline(tokens.FocusRingWidth, 0, style.Solid(focus)).
		TextColor(style.Solid(foreground)).
		OnAny(style.Checked|style.Indeterminate, style.Style{}.
			Background(style.Solid(accent)).
			TextColor(style.Solid(accentFg)))
	defaults := style.Style{}.
		Part(style.PartIndicator, indicator).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(foreground)).
			FontSize(active.Typography.ControlSize)).
		Part(style.PartDescription, style.Style{}.
			TextColor(style.Solid(muted)))
	return theme.CheckboxVisual{
		Defaults: defaults,
		Override: theme.CheckboxOverrideStyle(request),
		Metrics: theme.CheckboxMetrics{
			FocusSpace: tokens.FocusSpace, IndicatorSize: tokens.IndicatorSize,
			IndeterminateStroke: tokens.IndeterminateStroke, CheckStroke: tokens.CheckStroke,
			DescriptionGap: tokens.DescriptionGap, DescriptionIndent: tokens.DescriptionIndent,
			LabelGap: tokens.LabelGap, RequiredMarkOffset: active.Components.Label.RequiredMarkOffset,
		},
		Colors:            theme.CheckboxColors{RequiredMark: focus, DisabledRequiredMark: muted},
		IndicatorRenderer: neutralCheckboxIndicatorRenderer,
		Motion:            active.Motion, SelectionDuration: 100 * time.Millisecond, FocusDuration: 100 * time.Millisecond,
	}
}

// neutralCheckboxIndicatorRenderer keeps behavior fixtures independent of the
// shipped checkbox recipe while satisfying the renderer contract.
func neutralCheckboxIndicatorRenderer(gtx layout.Context, _ theme.CheckboxIndicatorRenderRequest) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Min}
}
