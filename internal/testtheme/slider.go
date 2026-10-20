package testtheme

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Slider(active *theme.Theme, request theme.SliderRequest) theme.SliderVisual {
	tokens := active.Components.Slider
	track := style.Style{}.
		Background(style.Solid(background)).
		Radius(tokens.TrackRadius)
	fill := style.Style{}.Background(style.Solid(accent))
	thumbWidth := tokens.ThumbLength + tokens.ThumbExtra
	thumbHeight := tokens.TrackThickness
	if request.Orientation == theme.SliderVertical {
		track = track.Width(tokens.TrackThickness).FillHeight()
		fill = fill.RadiusBottomLeft(tokens.TrackRadius).RadiusBottomRight(tokens.TrackRadius)
		thumbWidth, thumbHeight = thumbHeight, thumbWidth
	} else {
		track = track.FillWidth().Height(tokens.TrackThickness)
		fill = fill.RadiusTopLeft(tokens.TrackRadius).RadiusBottomLeft(tokens.TrackRadius)
	}
	defaults := style.Recipe{
		style.PartTrack: {Base: track},
		style.PartFill:  {Base: fill},
		style.PartThumb: {Base: style.Style{}.
			Width(thumbWidth).
			Height(thumbHeight).
			Background(style.Solid(background)).
			BorderWidth(max(tokens.ThumbExtra/2, unit.Dp(1))).
			BorderColor(style.Solid(accent)).
			Radius(tokens.TrackRadius).
			BoxShadow(0, 2, 0, 2, style.Solid(color.NRGBA{A: 0x0a})).
			BoxShadow(0, 1, 0, 1, style.Solid(color.NRGBA{A: 0x0f})).
			Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(focus))},
		style.PartLabel: {Base: style.Style{}.
			TextColor(style.Solid(foreground)).
			FontSize(tokens.TextSize).
			FontWeight(int(font.Medium))},
	}.Compile()
	focusDuration := 100 * time.Millisecond
	draggingDuration := 150 * time.Millisecond
	if active.Motion.ControlDuration > 0 {
		focusDuration = active.Motion.ControlDuration
		draggingDuration = active.Motion.ControlDuration
	}
	return theme.SliderVisual{
		Defaults: defaults, Override: theme.SliderOverrideStyle(request),
		Metrics: theme.SliderMetrics{
			TrackThickness: tokens.TrackThickness, TrackRadius: tokens.TrackRadius,
			EdgeInset: tokens.EdgeInset, ThumbLength: tokens.ThumbLength,
			ThumbExtra: tokens.ThumbExtra, HeaderGap: tokens.HeaderGap,
			VerticalGap: tokens.VerticalGap, DraggingScale: tokens.DraggingScale,
			FocusRingWidth: tokens.FocusRingWidth, FocusRingOffset: tokens.FocusRingOffset,
		},
		Motion: active.Motion, FocusDuration: focusDuration, DraggingDuration: draggingDuration,
	}
}
